package gobasrv

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/jfenske89/gobasrv/internal/helpers"
)

const defaultShutdownDeadline = 30 * time.Second

// Service is a generic service interface to run logic functions in parallel and handle graceful shutdown
type Service interface {
	// Run executes the provided logic functions in parallel and executes shutdown handlers after
	Run(...func(context.Context) error) error

	// RunContext executes the provided logic functions in parallel with a context and executes shutdown handlers after
	RunContext(context.Context, ...func(context.Context) error) error

	// RegisterShutdownHandler registers a graceful shutdown handler
	RegisterShutdownHandler(...func(context.Context) error)

	// RequestShutdown cancels the run context giving the main logic time to exit
	RequestShutdown()

	// Shutdown cancels the run context and executes graceful shutdown handlers immediately
	Shutdown() error
}

type serviceImpl struct {
	ctx    context.Context
	err    error
	cancel context.CancelFunc

	shutdownHandlers     []func(context.Context) error
	shutdownDeadline     time.Duration
	shutdownOnce         sync.Once
	shutdownHandlerMutex sync.Mutex
}

// NewService builds a new service with the default 30-second shutdown deadline
func NewService() Service {
	return NewServiceWithShutdownDeadline(defaultShutdownDeadline)
}

// NewServiceWithShutdownDeadline builds a new service with a specific shutdown deadline (default if invalid)
func NewServiceWithShutdownDeadline(shutdownDeadline time.Duration) Service {
	if shutdownDeadline <= 0 {
		shutdownDeadline = defaultShutdownDeadline
	}

	return &serviceImpl{shutdownDeadline: shutdownDeadline}
}

func (s *serviceImpl) Run(logic ...func(context.Context) error) error {
	return s.RunContext(context.Background(), logic...)
}

func (s *serviceImpl) RunContext(parentCtx context.Context, logic ...func(context.Context) error) error {
	// create a new context that will be canceled on interrupt signals
	s.ctx, s.cancel = signal.NotifyContext(
		parentCtx,
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	defer s.cancel()

	// execute the provided logic functions in parallel and store the first error
	s.err = helpers.ProcessLogic(s.ctx, logic...)

	// execute the shutdown handlers in parallel and join any errors for return
	if err := s.Shutdown(); err != nil {
		if s.err != nil {
			s.err = errors.Join(s.err, err)
		} else {
			s.err = err
		}
	}

	return s.err
}

func (s *serviceImpl) RegisterShutdownHandler(handlers ...func(context.Context) error) {
	s.shutdownHandlerMutex.Lock()
	defer s.shutdownHandlerMutex.Unlock()

	s.shutdownHandlers = append(s.shutdownHandlers, handlers...)
}

func (s *serviceImpl) RequestShutdown() {
	// signal the run logic to exit
	if s.cancel != nil {
		s.cancel()
	}
}

func (s *serviceImpl) Shutdown() error {
	var err error

	// only execute the shutdown handlers once
	s.shutdownOnce.Do(func() {
		// signal the run logic to exit
		if s.cancel != nil {
			s.cancel()
		}

		// process the shutdown handlers with the configured deadline
		s.shutdownHandlerMutex.Lock()
		defer s.shutdownHandlerMutex.Unlock()

		// create a context that will be canceled when the shutdown deadline is reached
		ctx, cancel := context.WithTimeout(context.Background(), s.shutdownDeadline)
		defer cancel()

		// execute the shutdown handlers in parallel and store the first error
		err = helpers.ProcessLogic(ctx, s.shutdownHandlers...)
	})

	return err
}
