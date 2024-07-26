package helpers

import (
	"context"
	"errors"

	"golang.org/x/sync/errgroup"
)

func ProcessLogic(
	incomingCtx context.Context,
	logic ...func(context.Context) error,
) error {
	// create a context that will be canceled when the first error occurs
	ctx, cancel := context.WithCancelCause(incomingCtx)

	// run logic asynchronously
	go func() {
		// cancel the context after all logic has been executed
		defer cancel(nil)

		// run each piece of logic in a goroutine
		eg := &errgroup.Group{}

		for i := range logic {
			fn := logic[i]

			eg.Go(func() error {
				if err := fn(ctx); err != nil {
					// immediately cancel the context and stop processing logic when an error occurs
					cancel(err)
				}

				return nil
			})
		}

		// block until all logic has been executed
		if err := eg.Wait(); err != nil {
			cancel(err)
		}
	}()

	// wait until asynchronous logic is finished (or canceled)
	<-ctx.Done()

	// check for and return errors other than context.Canceled
	if err := context.Cause(ctx); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}

	return nil
}
