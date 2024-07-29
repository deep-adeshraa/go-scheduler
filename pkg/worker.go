package pkg

import (
	context "context"
	"fmt"

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
	j.grpc_client = NewScheduleJobServiceClient(j.grpc_conn)
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
			}
		}
	}
}
