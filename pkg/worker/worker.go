package worker

import (
	context "context"
	"fmt"
	"time"

	api_grpc "deep-adeshraa/task-scheduler/pkg/api_grpc"
	grpc "google.golang.org/grpc"
)

type JobWorker struct {
	workerID    string
	grpc_conn   *grpc.ClientConn
	grpc_client api_grpc.ScheduleJobServiceClient
	context     context.Context
	cancel      context.CancelFunc
	port        string // co-ordinator port
	host        string // co-ordinator host
}

func NewJobWorker(workerID string) *JobWorker {
	context, cancel := context.WithCancel(context.Background())
	return &JobWorker{
		workerID: workerID,
		context:  context,
		cancel:   cancel,
		port:     "50051",
		host:     "localhost",
	}
}

func (j *JobWorker) connectToCoordinator() {
	url := j.host + ":" + j.port
	conn, err := grpc.Dial(url, grpc.WithInsecure())

	if err != nil {
		panic(err)
	}

	j.grpc_conn = conn
	j.grpc_client = api_grpc.NewScheduleJobServiceClient(j.grpc_conn)
	fmt.Println("Connected to coordinator")
}

func (j *JobWorker) ProcessJob(job *api_grpc.Job) {
	fmt.Println("Processing job: ", job)
	scheduledAt := job.ScheduledAt
	now := time.Now()
	// convert string scheduledAt to time.Time
	scheduledAtTime, _ := time.Parse(time.RFC3339, scheduledAt)

	// if the job is scheduled for the future, sleep until it's time to process the job
	if scheduledAtTime.After(now) {
		fmt.Println("Job scheduled for the future. Sleeping until scheduled time")
		time.Sleep(scheduledAtTime.Sub(now))
	}

	updateJobRequest := &api_grpc.UpdateJobStatusRequest{
		Job:    job,
		Status: api_grpc.JobStatus_STARTED,
	}

	funcToRun, exists := functionNameObjMap[job.Function]
	fmt.Println("Job started processing")
	j.grpc_client.UpdateJobStatus(j.context, updateJobRequest)

	if !exists {
		updateJobRequest.Status = api_grpc.JobStatus_FAILED
		j.grpc_client.UpdateJobStatus(j.context, updateJobRequest)
		fmt.Println("Function not found")
		return
	}

	if err := funcToRun(); err != nil {
		fmt.Println("Job failed with error: ", err)
		updateJobRequest.Status = api_grpc.JobStatus_FAILED
		j.grpc_client.UpdateJobStatus(j.context, updateJobRequest)
		return
	}

	// Job processed successfully
	updateJobRequest.Status = api_grpc.JobStatus_COMPLETED
	j.grpc_client.UpdateJobStatus(j.context, updateJobRequest)
	fmt.Println("Job processed successfully")
}

func (j *JobWorker) StartReceivingJobs() {
	j.connectToCoordinator()
	defer j.grpc_conn.Close()

	stream, err := j.grpc_client.GetUpcomingJobs(j.context, &api_grpc.Worker{Name: j.workerID})

	if err != nil {
		panic(err)
	}

	for {
		select {
		case <-j.context.Done():
			return
		case <-stream.Context().Done():
			fmt.Println("Coordinator disconnected")
			return
		default:
			job, err := stream.Recv()
			if err != nil {
				fmt.Println("Failed to receive jobs: ", err)
				continue
			}
			// Do something with the job
			fmt.Println("Received job: ", job)
			j.ProcessJob(job)
		}
	}
}
