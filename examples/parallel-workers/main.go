package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/jfenske89/gobasrv"
)

// an example "message queue"
var queue sync.Map

func main() {
	service := gobasrv.NewService()

	// the run function can accept multiple instructions that will run in parallel,
	// the following examples run a "message publisher" and a "message subscriber",
	// just to give a basic working example which can run and produce output
	if err := service.Run(
		// mimic a message publisher that sends data into a queue
		func(ctx context.Context) error {
			fmt.Println("[publisher] starting...")

			// this example will "publish" 10 messages into a "queue" and exit
			for i := 0; i < 10; i++ {
				go func() {
					queue.Store(fmt.Sprint(i), fmt.Sprintf("message id %d with a string value", i))

					fmt.Printf("[publisher] pushed message id %d\n", i)
				}()
			}

			fmt.Println("[publisher] stopping...")

			return nil
		},

		// mimic a message subscriber that reads data out of a queue
		func(ctx context.Context) error {
			fmt.Println("[subscriber] starting...")

			// in a real world scenario, you may want to register a shutdown handler
			// that will stop the subscriber from accepting new messages, and wait for
			// current messages to finish processing before returning
			var wg sync.WaitGroup

			// this example will "process" exactly 10 messages from a "queue" and exit
			for i := 0; i < 10; i++ {
				wg.Add(1)

				go func() {
					defer wg.Done()

					for {
						if message, ok := queue.LoadAndDelete(fmt.Sprint(i)); ok {
							fmt.Printf("[subscriber] processed message id %d: %s\n", i, message)
							return
						}
					}
				}()
			}

			wg.Wait()

			fmt.Println("[subscriber] stopping...")

			return nil
		},
	); err != nil {
		panic(fmt.Errorf("failed to run example: %s", err.Error()))
	}
}
