package pkg

import (
	context "context"
	"database/sql"
	"fmt"
	"net"
	"time"

	"golang.org/x/sync/semaphore"

	grpc "google.golang.org/grpc"
)

type workerStream ScheduleJobService_GetUpcomingJobsServer

type RegisteredWorkerSteams struct {
	worker    *Worker
	stream    workerStream
	semaphore *semaphore.Weighted
}

type JobCoordinator struct {
	UnimplementedScheduleJobServiceServer
	db            *sql.DB
	listener      net.Listener
	grpcServer    *grpc.Server
	port          string
	host          string
	context       context.Context
	workerStreams []RegisteredWorkerSteams
	cancel        context.CancelFunc
}

func NewJobCoordinator() *JobCoordinator {
	db, err := ConnectToDB()
	if err != nil {
		panic(err)
	}
	context, cancel := context.WithCancel(context.Background())

	return &JobCoordinator{
		db:            db,
		context:       context,
		port:          "50051",
		host:          "localhost",
		cancel:        cancel,
		workerStreams: []RegisteredWorkerSteams{},
	}
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
	RegisterScheduleJobServiceServer(grpcServer, j)

	if err := j.grpcServer.Serve(j.listener); err != nil {
		fmt.Println("Failed to serve: ", err)
	}
}

func (j *JobCoordinator) GetUpcomingJobs(worker *Worker, stream ScheduleJobService_GetUpcomingJobsServer) error {
	j.workerStreams = append(j.workerStreams, RegisteredWorkerSteams{
		worker,
		stream,
		semaphore.NewWeighted(1),
	})

	// Keep the stream open
	for {
		select {
		// Wait for the client to close the stream
		case <-stream.Context().Done():
			return nil
			// Wait for the broker to shutdown
		case <-j.context.Done():
			return nil
		}
	}
}

func (j *JobCoordinator) Stop() {
	j.cancel()
	j.grpcServer.Stop()
	j.listener.Close()
	j.db.Close()
}

func (j *JobCoordinator) GetNextJobs() []*Job {
	// Get the next jobs from the database
	// return the jobs
	rows, err := j.db.Query("SELECT * FROM jobs ORDER BY RunAt ASC LIMIT 10")

	if err != nil {
		fmt.Println("Failed to get jobs: ", err)
		return nil
	}

	defer rows.Close()

	jobs := []*Job{}

	for rows.Next() {
		job := Job{}
		err := rows.Scan(&job.Name, &job.Function, &job.RunAt)

		if err != nil {
			continue
		}

		jobs = append(jobs, &job)
	}

	return jobs
}

func (j *JobCoordinator) SendJobsToWorkers() {
	for {
		if j.context.Done() != nil {
			break
		}

		for _, workerStream := range j.workerStreams {
			jobs := j.GetNextJobs()
			upcomingJobs := &UpcomingJobs{Jobs: jobs}

			if workerStream.semaphore.TryAcquire(1) {
				workerStream.stream.Send(upcomingJobs)
				workerStream.semaphore.Release(1)
			} else {
				continue
			}

		}
		time.Sleep(2 * time.Second)
	}
}
