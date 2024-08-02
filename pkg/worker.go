package pkg

import (
	context "context"
	"database/sql"
	"fmt"
	"time"

	grpc "google.golang.org/grpc"
)

type JobWorker struct {
	workerID    string
	grpc_conn   *grpc.ClientConn
	grpc_client ScheduleJobServiceClient
	context     context.Context
	cancel      context.CancelFunc
	port        string // co-ordinator port
	host        string // co-ordinator host
	db          *sql.DB
}

func NewJobWorker(workerID string) *JobWorker {
	db, err := ConnectToDB()
	if err != nil {
		panic(err)
	}

	context, cancel := context.WithCancel(context.Background())
	return &JobWorker{
		workerID: workerID,
		context:  context,
		cancel:   cancel,
		port:     "50051",
		host:     "localhost",
		db:       db,
	}
}

func (j *JobWorker) connectToCoordinator() {
	url := j.host + ":" + j.port
	conn, err := grpc.Dial(url, grpc.WithInsecure())

	if err != nil {
		panic(err)
	}

	j.grpc_conn = conn
	j.grpc_client = NewScheduleJobServiceClient(j.grpc_conn)
}

func (j *JobWorker) ProcessJob(job *Job) {
	fmt.Println("Processing job: ", job)
	scheduledAt := job.ScheduledAt
	now := time.Now().UTC()

	// convert string scheduledAt to time.Time
	scheduledAtTime, err := time.Parse(time.RFC3339, scheduledAt)

	if err != nil {
		panic(err)
	}

	// if the job is scheduled for the future, sleep until it's time to process the job
	if scheduledAtTime.After(now) {
		fmt.Println("Job scheduled for the future. Sleeping until scheduled time")
		time.Sleep(scheduledAtTime.Sub(now))
	} else {
		fmt.Println("Now = ", now, " ScheduledAt = ", scheduledAtTime, scheduledAtTime.After(now))
	}

	tx, err := j.db.Begin()
	if err != nil {
		panic(err)
	}

	_, err = tx.Exec("SELECT * FROM personal_test_job_schedules WHERE id = $1 FOR UPDATE", job.Id)
	if err != nil {
		tx.Rollback()
		panic(err)
	}
	fmt.Println("Job started processing")
	_, err = tx.Exec("UPDATE personal_test_job_schedules SET started_at=NOW() WHERE id = $1", job.Id)

	if err != nil {
		tx.Rollback()
		panic(err)
	}
	fmt.Println("Job completed processing")
	_, err = tx.Exec("UPDATE personal_test_job_schedules SET completed_at=NOW() WHERE id = $1", job.Id)

	if err != nil {
		tx.Rollback()
		panic(err)
	}
	tx.Commit()

	fmt.Println("Job processed successfully")
}

func (j *JobWorker) StartReceivingJobs() {
	j.connectToCoordinator()
	defer j.grpc_conn.Close()

	stream, err := j.grpc_client.GetUpcomingJobs(j.context, &Worker{Name: j.workerID})

	if err != nil {
		panic(err)
	}

	for {
		select {
		case <-j.context.Done():
			return
		default:
			jobs, err := stream.Recv()
			if err != nil {
				panic(err)
			}
			// Do something with the job
			for _, job := range jobs.Jobs {
				fmt.Println("Received job: ", job)
				j.ProcessJob(job)
			}
		}
	}
}
