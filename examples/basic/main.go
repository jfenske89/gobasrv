package main

import (
	"context"
	"fmt"
	"time"

	"github.com/jfenske89/gobasrv"
)

// CustomService embed the base service and override any methods as needed
type CustomService struct {
	gobasrv.Service
}

func NewCustomService() gobasrv.Service {
	return &CustomService{
		// Create a new service with a 30 second shutdown deadline. Kubernetes
		// will send a SIGTERM to the process when it's time to stop. Generally
		// 30 seconds is given before it sends a SIGKILL, but this deadline is
		// configurable to support other environments.
		gobasrv.NewServiceWithShutdownDeadline(10 * time.Second),
	}
}

func main() {
	service := NewCustomService()

	if err := service.Run(func(ctx context.Context) error {
		// Connect to databases, message queues, etc.
		// ...

		// Configure graceful shutdown. For example: wait for messages to be processed, close connections, etc.
		service.RegisterShutdownHandler(func(ctx context.Context) error {
			fmt.Println("disconnecting...")

			// an error will be returned to the caller if a shutdown handler takes longer than the deadline
			// time.Sleep(15 * time.Second)
			return nil
		})

		// Write your logic here, for example some kind of server or message processor
		fmt.Println("your service logic...")
		time.Sleep(1 * time.Second)

		// Return at any time to initiate shutdown logic (errors are returned to the caller)
		return nil
	}); err != nil {
		// Handle any error from the service or graceful shutdown logic
		panic("error running service: " + err.Error())
	}
}
