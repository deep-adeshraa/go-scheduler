package main

import (
	coordinator "deep-adeshraa/task-scheduler/pkg/coordinator"
	scheduler "deep-adeshraa/task-scheduler/pkg/scheduler"
	worker "deep-adeshraa/task-scheduler/pkg/worker"
	"fmt"
	"os"
)

func main() {
	// take command line arguments for starting scheduler or co-ordinator or worker

	args := os.Args[1:]

	if args[0] == "scheduler" {
		s := scheduler.NewScheduler()
		s.Start()
	} else if args[0] == "coordinator" {
		c := coordinator.NewJobCoordinator()
		c.Start()
	} else if args[0] == "worker" {
		w := worker.NewJobWorker(args[1])
		w.Start(1)
	} else {
		fmt.Println("Invalid argument")
	}
}
