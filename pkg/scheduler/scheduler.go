package scheduler

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"strings"

	api_grpc "deep-adeshraa/task-scheduler/pkg/api_grpc"
	utils "deep-adeshraa/task-scheduler/pkg/utils"
	_ "github.com/lib/pq"
)



type Scheduler struct {
	db *sql.DB
}

func NewScheduler() *Scheduler {
	db, err := utils.ConnectToDB()
	if err != nil {
		panic(err)
	}
	return &Scheduler{db: db}
}

func (s *Scheduler) Close() {
	s.db.Close()
}

func (s *Scheduler) CreateJob(job *api_grpc.Job) error {
	_, err := s.db.Exec("INSERT INTO personal_test_job_schedules (id, function, scheduled_at) VALUES ($1, $2, $3)", job.Id, job.Function, job.ScheduledAt)
	return err
}

func (s *Scheduler) UpdateJob(job *api_grpc.Job) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	// Lock the row for update
	_, err = tx.Exec("SELECT * FROM personal_test_job_schedules WHERE id = $1 FOR UPDATE", job.Id)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Update the job
	_, err = tx.Exec("UPDATE personal_test_job_schedules SET function = $1, scheduled_at = $2 WHERE id = $3", job.Function, job.ScheduledAt, job.Id)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (s *Scheduler) DeleteJob(name string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	// Lock the row for update
	_, err = tx.Exec("SELECT * FROM personal_test_job_schedules WHERE name = $1 FOR UPDATE", name)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Delete the job
	_, err = tx.Exec("DELETE FROM personal_test_job_schedules WHERE name = $1", name)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (s *Scheduler) GetJob(name string) (*api_grpc.Job, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	name = strings.TrimSpace(name)

	row := tx.QueryRow("SELECT id, function, scheduled_at FROM personal_test_job_schedules WHERE name = '" + name + "' FOR SHARE")
	job := &api_grpc.Job{}
	err = row.Scan(&job.Id, &job.Function, &job.ScheduledAt)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()

	return job, nil
}

func createJobCommand(commandParts []string, s *Scheduler) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Error creating job: ", r)
		}
	}()

	// CREATE JOB job1 hello_world 2021-01-01T00:00:00Z
	job := &api_grpc.Job{Id: commandParts[2], Function: commandParts[3], ScheduledAt: commandParts[4]}

	err := s.CreateJob(job)
	if err != nil {
		fmt.Println("Error creating job: ", err)
	} else {
		fmt.Println("Job created successfully")
	}
}

func updateJobCommand(commandParts []string, s *Scheduler) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Error updating job: ", r)
		}
	}()

	// UPDATE JOB job1 hello_world_updated 2021-01-01T00:00:00Z
	job := &api_grpc.Job{Id: commandParts[2], Function: commandParts[3], ScheduledAt: commandParts[4]}

	err := s.UpdateJob(job)
	if err != nil {
		fmt.Println("Error updating job: ", err)
	} else {
		fmt.Println("Job updated successfully")
	}
}

func deleteJobCommand(commandParts []string, s *Scheduler) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Error deleting job: ", r)
		}
	}()

	err := s.DeleteJob(commandParts[2])
	if err != nil {
		fmt.Println("Error deleting job: ", err)
	} else {
		fmt.Println("Job deleted successfully")
	}
}

func getJobCommand(commandParts []string, s *Scheduler) *api_grpc.Job {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Error getting job: ", r)
		}
	}()

	job, err := s.GetJob(commandParts[2])
	if err != nil {
		fmt.Println("Error getting job: ", err)
	} else {
		fmt.Println("Job: ", job)
	}
	return job
}

func (s *Scheduler) Start() {
	// start the scheduler and start taking commands from CMD
	// to create, update, delete and get jobs
	fmt.Println("Scheduler started")
	fmt.Println("Commands Examples: " +
		"\n: CREATE JOB job1 hello_world 2021-01-01T00:00:00Z" +
		"\n: UPDATE JOB job1 hello_world_updated 2021-01-01T00:00:00Z" +
		"\n: DELETE JOB job1" +
		"\n: GET JOB job1")

	for {
		in := bufio.NewReader(os.Stdin)

		command, err := in.ReadString('\n')

		if err != nil {
			fmt.Println("Error reading command: ", err)
			continue
		}

		if command == "EXIT" {
			break
		}

		if command == "" {
			continue
		}

		commandParts := strings.Split(command, " ")
		if len(commandParts) < 3 {
			fmt.Println("Invalid command")
			continue
		}

		switch commandParts[0] {
		case "CREATE":
			createJobCommand(commandParts, s)
		case "UPDATE":
			updateJobCommand(commandParts, s)
		case "DELETE":
			deleteJobCommand(commandParts, s)
		case "GET":
			job := getJobCommand(commandParts, s)
			fmt.Println("Job: ", job)
		default:
			fmt.Println("Invalid command")
		}
	}
}
