package main

import pkg "deep-adeshraa/task-scheduler/pkg"
import "os"
import "fmt"

func main() {
	// take command line arguments for starting scheduler or co-ordinator or worker

	args := os.Args[1:]

	if args[0] == "scheduler" {
		s := pkg.NewScheduler()
		s.Start()
	} else if args[0] == "coordinator" {
		c := pkg.NewJobCoordinator()
		c.StartGRPCServer()
	} else if args[0] == "worker" {
		w := pkg.NewJobWorker(args[1])
		w.StartReceivingJobs()
	} else {
		fmt.Println("Invalid argument")
	}
}
