package main

import (
	"errors"
	"fmt"

	"sis/internal/task"
)

func main() {
	svc := task.NewService(task.NewMemoryRepository(), task.TextAnalyzer{})

	inputs := []string{
		"Hello Go world",
		"Go makes concurrency simple\nand tooling pleasant",
		"   ",
	}

	for _, in := range inputs {
		created, err := svc.Create(in)
		if err != nil {
			fmt.Printf("create rejected: %v\n", err)
			continue
		}
		fmt.Printf("created task %d (%s)\n", created.ID, created.State)
	}

	for _, t := range svc.List() {
		done, err := svc.Process(t.ID)
		if err != nil {
			fmt.Printf("task %d failed: %v\n", t.ID, err)
			continue
		}
		fmt.Printf("task %d %s: %+v\n", done.ID, done.State, *done.Result)
	}

	if _, err := svc.Process(1); errors.Is(err, task.ErrInvalidTransition) {
		fmt.Printf("re-processing rejected: %v\n", err)
	}

	if _, err := svc.Get(99); errors.Is(err, task.ErrTaskNotFound) {
		fmt.Printf("lookup rejected: %v\n", err)
	}
}
