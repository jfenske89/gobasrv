package gobasrv

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewService(t *testing.T) {
	if svc := NewService(); svc == nil {
		t.Fatal("expected NewService to return non-nil")
	}
}

func TestNewServiceWithShutdownDeadline(t *testing.T) {
	if svc := NewServiceWithShutdownDeadline(5 * time.Second); svc == nil {
		t.Fatal("expected NewServiceWithShutdownDeadline to return non-nil")
	}
}

func TestService_Run(t *testing.T) {
	svc := NewService()

	if err := svc.Run(func(ctx context.Context) error {
		return nil
	}); err != nil {
		t.Fatalf("expected Run to return nil, got %v", err)
	}
}

func TestService_RunContext(t *testing.T) {
	svc := NewService()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := svc.RunContext(ctx, func(ctx context.Context) error {
		return nil
	}); err != nil {
		t.Fatalf("expected RunContext to return nil, got %v", err)
	}
}

func TestService_RegisterShutdownHandler(t *testing.T) {
	var results atomic.Int32

	svc := NewService()
	svc.RegisterShutdownHandler(func(ctx context.Context) error {
		results.Add(1)
		return nil
	})

	if err := svc.Run(func(ctx context.Context) error {
		return nil
	}); err != nil {
		t.Fatalf("expected Run to return nil, got %v", err)
	} else if results.Load() != 1 {
		t.Fatalf("expected shutdown handler to set results to 1, got %v", results.Load())
	}
}

func TestService_RegisterShutdownHandlerWithError(t *testing.T) {
	svc := NewService()
	svc.RegisterShutdownHandler(func(ctx context.Context) error {
		return errors.New("test")
	})

	if err := svc.Run(func(ctx context.Context) error {
		return nil
	}); err == nil {
		t.Fatalf("expected Run to return an error, got nil")
	} else if err.Error() != "test" {
		t.Fatalf("expected Run to return an error with message 'test', got %v", err.Error())
	}
}

func TestService_RegisterShutdownHandlerWithDoubleError(t *testing.T) {
	svc := NewService()
	svc.RegisterShutdownHandler(func(ctx context.Context) error {
		return errors.New("test2")
	})

	if err := svc.Run(func(ctx context.Context) error {
		return errors.New("test1")
	}); err == nil {
		t.Fatalf("expected Run to return an error, got nil")
	} else if err.Error() != "test1\ntest2" {
		t.Fatalf("expected Run to return an error with message 'test1\ntest2', got %v", err.Error())
	}
}

func TestService_RequestShutdown(t *testing.T) {
	var results atomic.Int32

	svc := NewService()

	go func() {
		time.Sleep(5 * time.Millisecond)
		svc.RequestShutdown()
	}()

	if err := svc.Run(func(ctx context.Context) error {
		results.Add(1)
		time.Sleep(time.Minute)
		results.Add(1)
		return nil
	}); err != nil {
		t.Fatalf("expected Run to return nil, got %v", err)
	}

	if results.Load() != 1 {
		t.Fatalf("expected results to be 1, got %v", results.Load())
	}
}

func TestService_Shutdown(t *testing.T) {
	var results atomic.Int32

	svc := NewService()

	var shutdownErr error

	go func() {
		time.Sleep(5 * time.Millisecond)
		shutdownErr = svc.Shutdown()
	}()

	if err := svc.Run(func(ctx context.Context) error {
		results.Add(1)
		time.Sleep(time.Minute)
		results.Add(1)
		return nil
	}); err != nil {
		t.Fatalf("expected Run to return nil, got %v", err)
	}

	if results.Load() != 1 {
		t.Fatalf("expected results to be 1, got %v", results.Load())
	} else if shutdownErr != nil {
		t.Fatalf("expected Shutdown to return nil, got %v", shutdownErr)
	}
}
