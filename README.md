# gobasrv

***Go base service***: *because all the best names are taken.*

This project is meant to serve as a bare minimum base for Go services.

It offers support for running business logic within a graceful shutdown wrapper, and that's it.

An example real world use case would be a pub/sub worker running in Kubernetes, which processes messages.
- Run the service with pub/sub message processing logic
- Register a shutdown handler to gracefully handle the shutdown (examples below)
- Subscribe to and process messages
- Kubernetes HPA sends a SIGTERM to scale down
- A shutdown handler informs the subscriber to stop accepting new messages
- A shutdown handler waits for active messages to finish, or Nacks incomplete messages, so that another pod can process them instead.

## Original project

I renamed the original project, and archived it. It can be seen here: https://github.com/jfenske89/go-service

## Running

Simply execute your logic using the `Run` or `RunContext` function.

```go
// Run executes the provided logic functions in parallel and executes shutdown handlers after
Run(...func(context.Context) error) error

// RunContext executes the provided logic functions in parallel with a context and executes shutdown handlers after
RunContext(context.Context, ...func(context.Context) error) error
```

## Graceful shutdown

Define graceful shutdown logic with `RegisterShutdownHandler`.

For example: flush logs, close connections, wait for active work to finish, etc...

```go
// RegisterShutdownHandler registers a graceful shutdown handler
RegisterShutdownHandler(...func(context.Context) error)
```

These functions are executed in parallel before the application exits.

Shutdown has a 30-second deadline by default. This can be customized.

### Initiate shutdown

There are two options for initiating a shutdown while the logic is processing:

1. `RequestShutdown`: will cancel the inner context and wait for the goroutines to finish.
2. `Shutdown`: will cancel the inner context and immediately execute the shutdown handlers.

```go
// RequestShutdown cancels the run context giving the main logic time to exit
RequestShutdown()

// Shutdown cancels the run context and executes graceful shutdown handlers immediately
Shutdown() error
```

## Examples

- Basic example: [./examples/basic/main.go](./examples/basic/main.go)
- Parallel workers: [./examples/parallel-workers/main.go](./examples/parallel-workers/main.go)
