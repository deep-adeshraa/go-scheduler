package coordinator

import (
	context "context"
	"database/sql"
	api_grpc "deep-adeshraa/task-scheduler/pkg/api_grpc"
	utils "deep-adeshraa/task-scheduler/pkg/utils"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	grpc "google.golang.org/grpc"

	"net"
	"strings"
	"time"
)

type workerStream api_grpc.ScheduleJobService_GetUpcomingJobsServer

type RegisteredWorkerSteam struct {
	worker      *api_grpc.Worker
	stream      workerStream
	upcomingJob chan *api_grpc.Job
}

type JobCoordinator struct {
	api_grpc.UnimplementedScheduleJobServiceServer
	db              *sql.DB
	listener        net.Listener
	grpcServer      *grpc.Server
	port            string
	host            string
	context         context.Context
	workerStreams   []RegisteredWorkerSteam
	cancel          context.CancelFunc
	wg              sync.WaitGroup
	upcomingJobs    chan []*api_grpc.Job
	roundRobinIndex int
}

func NewJobCoordinator() *JobCoordinator {
	context, cancel := context.WithCancel(context.Background())

	return &JobCoordinator{
		context:         context,
		port:            "50051",
		host:            "localhost",
		cancel:          cancel,
		workerStreams:   []RegisteredWorkerSteam{},
		wg:              sync.WaitGroup{},
		roundRobinIndex: 0,
		upcomingJobs:    make(chan []*api_grpc.Job),
	}
}

func (j *JobCoordinator) Start() error {
	db, err := utils.ConnectToDB()
	if err != nil {
		fmt.Println("Failed to connect to db: ", err)
		return err
	}
	j.db = db
	j.StartGRPCServer()
	j.wg.Add(2)

	// start the go routine to get the next jobs
	go j.GetNextJobs()
	go j.SendJobsToWorkers()

	fmt.Println("Job coordinator started")
	return j.awaitShutdown()
}

func (j *JobCoordinator) StartGRPCServer() {
	url := j.host + ":" + j.port
	listener, err := net.Listen("tcp", url)

	if err != nil {
		fmt.Println("Failed to listen: ", err)
		return
	}

	grpcServer := grpc.NewServer()
	j.listener = listener
	j.grpcServer = grpcServer
	api_grpc.RegisterScheduleJobServiceServer(grpcServer, j)

	go func() {
		if err := j.grpcServer.Serve(j.listener); err != nil {
			fmt.Println("gRPC server failed: ", err)
		}
	}()
}

func (j *JobCoordinator) awaitShutdown() error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// this like blocks the main thread until a signal is received
	sig := <-stop
	fmt.Println("Received signal to stop: ", sig)

	return j.Stop()
}

func (j *JobCoordinator) GetUpcomingJobs(worker *api_grpc.Worker, stream api_grpc.ScheduleJobService_GetUpcomingJobsServer) error {
	fmt.Println("Worker connected: ", worker.Name)

	registeredWorkerStream := RegisteredWorkerSteam{
		worker,
		stream,
		make(chan *api_grpc.Job),
	}

	// register the worker stream
	j.workerStreams = append(j.workerStreams, registeredWorkerStream)

	fmt.Println("total workers: ", len(j.workerStreams))

	// Keep the stream open
	for {
		select {
		case upcomingJob := <-registeredWorkerStream.upcomingJob:
			if err := stream.Send(upcomingJob); err != nil {
				return err
			}
		// Wait for the client to close the stream
		case <-stream.Context().Done():
			// Remove the worker from the list
			for i, workerStream := range j.workerStreams {
				if workerStream.worker.Name == worker.Name {
					j.workerStreams = append(j.workerStreams[:i], j.workerStreams[i+1:]...)
					break
				}
			}
			fmt.Println("Worker disconnected: ", worker.Name)
			return nil
			// Wait for the broker to shutdown
		case <-j.context.Done():
			return nil
		}
	}
}

func (j *JobCoordinator) UpdateJobStatus(ctx context.Context, in *api_grpc.UpdateJobStatusRequest) (*api_grpc.Job, error) {
	update_column := "started_at"

	switch in.Status {
	case api_grpc.JobStatus_COMPLETED:
		update_column = "completed_at"
	case api_grpc.JobStatus_FAILED:
		update_column = "failed_at"
	case api_grpc.JobStatus_STARTED:
		update_column = "started_at"
	}

	_, err := j.db.Exec("UPDATE personal_test_job_schedules SET "+update_column+"=$1 WHERE id=$2", time.Now(), in.Job.Id)

	if err != nil {
		fmt.Println("Failed to update job status: ", err)
		return nil, err
	}

	return in.Job, nil
}

func (j *JobCoordinator) Stop() error {
	// send a signal to all the go routines to stop
	j.cancel()

	// j.wg.Wait() // wait for all the go routines to finish

	j.grpcServer.GracefulStop()
	j.listener.Close()
	j.db.Close()

	for _, workerStream := range j.workerStreams {
		// close the streams
		workerStream.stream.Context().Done()
	}

	return nil
}

func (j *JobCoordinator) SendJobsToWorkers() {
	defer j.wg.Done()

	for {
		select {
		case jobs := <-j.upcomingJobs:
			if len(j.workerStreams) == 0 {
				fmt.Println("No workers available")
				continue
			}
			for _, job := range jobs {
				steamIndex := j.roundRobinIndex % len(j.workerStreams)
				workerStream := j.workerStreams[steamIndex]
				fmt.Printf("Sending job %s to worker %s\n", job.Id, workerStream.worker.Name)
				workerStream.upcomingJob <- job
				j.roundRobinIndex++
			}
		case <-j.context.Done():
			return
		}
	}
}

func (j *JobCoordinator) GetNextJobs() {
	defer j.wg.Done()
	ticker := time.NewTicker(10 * time.Second)

	for {
		select {
		case <-j.context.Done():
			fmt.Println("Stopping job scheduler")
			return
		case t := <-ticker.C:
			tx, err := j.db.Begin()
			if err != nil {
				panic(err)
			}

			// Get the next jobs from the database
			// return the jobs
			rows, err := tx.Query("SELECT id, function, scheduled_at FROM personal_test_job_schedules " +
				"where scheduled_at > NOW() and scheduled_at < NOW() + INTERVAL '30 seconds' " +
				"AND picked_at IS NULL " +
				"ORDER BY scheduled_at ASC FOR UPDATE SKIP LOCKED")

			if err != nil {
				fmt.Println("Failed to get jobs: ", err)
				tx.Rollback()
				return
			}

			jobs := []*api_grpc.Job{}

			for rows.Next() {
				job := api_grpc.Job{}
				err := rows.Scan(&job.Id, &job.Function, &job.ScheduledAt)

				if err != nil {
					continue
				}

				jobs = append(jobs, &job)
			}
			if len(jobs) == 0 {
				err = tx.Commit()
				if err != nil {
					fmt.Println("Failed to commit transaction: ", err)
					return
				}
				time.Sleep(10 * time.Second)
				continue
			}

			jobIds := make([]string, len(jobs))
			for i, job := range jobs {
				jobIds[i] = job.Id
			}

			tx.Exec("UPDATE personal_test_job_schedules SET picked_at=NOW() WHERE id IN (" + strings.Join(jobIds, ",") + ")")

			err = tx.Commit()
			if err != nil {
				fmt.Println("Failed to commit transaction: ", err)
				return
			}
			fmt.Printf("Got %d jobs at %s\n", len(jobs), t)
			j.upcomingJobs <- jobs
		}
	}

}
