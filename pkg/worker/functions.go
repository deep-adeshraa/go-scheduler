package worker

import (
	"errors"
	"fmt"
	"math/rand"
	"time"
)

// 	_, err = j.db.Exec("UPDATE personal_test_job_schedules SET started_at=NOW() WHERE id = $1", job.Id)

func function1() error {
	time.Sleep(1 * time.Second)
	fmt.Println("Function 1")
	return nil
}

func function2() error {
	time.Sleep(2 * time.Second)
	fmt.Println("Function 2")
	return nil
}

func function3() error {
	time.Sleep(3 * time.Second)
	fmt.Println("Function 3")
	return nil
}

func function4() error {
	time.Sleep(4 * time.Second)
	fmt.Println("Function 4")
	return nil
}

func function5() error {
	// get random number between 1 and 10
	rand.Seed(time.Now().UnixNano())  // Seeding the random number generator
	randomNumber := rand.Intn(10) + 1 // Generating a random number between 1 and 10

	if randomNumber < 3 {
		return errors.New("random number less than 3")
	}

	time.Sleep(5 * time.Second)
	fmt.Println("Function 5")
	return nil
}

var functionNameObjMap = map[string]func() error{
	"function1": function1,
	"function2": function2,
	"function3": function3,
	"function4": function4,
	"function5": function5,
}
