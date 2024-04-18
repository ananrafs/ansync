package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/ananrafs/ansync"
	"time"
)

func main() {
	//DoWorkerPoolExample()
	//DoCancelAndRetry()
	//DoCancelTaskAndRetry()
}

func DoWorkerPoolExample() {
	wp := ansync.DoWithWorkerPool(
		[]ansync.Task{
			// define task to perform
			func() (interface{}, error) {
				for i := 0; i < 10; i++ {
					fmt.Printf("%d from task 1 \n", i)
					time.Sleep(time.Second)
				}
				return 12, nil
			},
			func() (interface{}, error) {
				for i := 0; i < 10; i++ {
					fmt.Printf("%d from task 2 \n", i)
					time.Sleep(time.Second)
				}
				return nil, errors.New("matatata matuta")
			},
			func() (interface{}, error) {
				for i := 0; i < 10; i++ {
					fmt.Printf("%d from task 3 \n", i)
					time.Sleep(time.Second)
				}
				return "natsu", nil
			},
		},
		ansync.WithMaxWorker(2),
	)
	// need to defer closer
	defer wp.Await()

	wp.Submit(func() (interface{}, error) {
		for i := 0; i < 10; i++ {
			fmt.Printf("%d from task 4 \n", i)
			time.Sleep(time.Second)
		}
		return "oh my god", nil
	})

	wp.Submit(func() (interface{}, error) {
		for i := 0; i < 10; i++ {
			fmt.Printf("%d from task 5 \n", i)
			time.Sleep(time.Second)
		}
		return "ululu", nil
	})

	// handling response and error
	wp.Handle(func(res <-chan interface{}) {
		for r := range res {
			fmt.Println(r, "rrrrrrrrr")
		}
	}, func(errors <-chan error) {
		for err := range errors {
			fmt.Println("[error]", err)
		}
	})
}

func DoCancelAndRetry() {
	parentCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	err := ansync.DoActionWithCancellation(parentCtx, func() error {
		return ansync.DoActionWithRetry(
			func() error {
				childCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
				return ansync.DoActionWithCancellation(childCtx, func() error {
					for i := 0; i < 10; i++ {
						fmt.Printf("%d second \n", i)
						time.Sleep(time.Second)
					}
					return nil
				})
			},
			ansync.WithRetry(3),
			ansync.WithDelay(func(attempt int) time.Duration {
				return time.Duration(attempt+1) * time.Second
			}),
		)
	})
	if err != nil {
		fmt.Println("[error]", err)
	}
	//time.Sleep(20 * time.Second)
}

func DoCancelTaskAndRetry() {
	parentCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	resp, err := ansync.DoTaskWithCancellation(parentCtx, func() (interface{}, error) {
		return ansync.DoTaskWithRetry(
			func() (interface{}, error) {
				childCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
				return ansync.DoTaskWithCancellation(childCtx, func() (interface{}, error) {
					for i := 0; i < 10; i++ {
						fmt.Printf("%d second \n", i)
						time.Sleep(time.Second)
					}
					return 31, nil
				})
			},
			ansync.WithRetry(3),
			ansync.WithDelay(func(attempt int) time.Duration {
				return time.Duration(attempt+1) * time.Second
			}),
		)
	})
	if err != nil {
		fmt.Println("[error]", err)
	} else {
		fmt.Println("resp", resp)
	}
	//time.Sleep(20 * time.Second)
}
