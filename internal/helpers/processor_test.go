package helpers_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jfenske89/gobasrv/internal/helpers"
)

func TestProcessLogicGroupOk(t *testing.T) {
	results := &atomic.Uint32{}

	if err := helpers.ProcessLogic(
		context.Background(),
		func(context.Context) error {
			results.Add(1)
			return nil
		},
		func(context.Context) error {
			results.Add(1)
			return nil
		},
		func(context.Context) error {
			results.Add(1)
			return nil
		},
	); err != nil {
		t.Errorf("err expected nil, got %v", err)
	} else if val := results.Load(); val != 3 {
		t.Errorf("results expected 3, got %v", val)
	}
}

func TestProcessLogicGroupDeadline(t *testing.T) {
	results := &atomic.Uint32{}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()

	if err := helpers.ProcessLogic(
		ctx,
		func(context.Context) error {
			time.Sleep(10 * time.Millisecond)
			results.Add(2)
			return nil
		},
		func(context.Context) error {
			results.Add(1)
			return nil
		},
	); !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("err expected context.DeadlineExceeded, got %v", err)
	} else if val := results.Load(); val != 1 {
		t.Errorf("results expected 1, got %v", val)
	}
}
