package backoff

import (
	"context"
	"time"
)

// Func represents functions that can be retried,
// reset param shouldn't be called in new goroutine.
type Func func(attempt uint, reset func()) (retry bool, err error)

// Strategy is a backoff methodology for retrying an operation.
type Strategy interface {
	// Pause returns duration of the next pause, according to the attempt number.
	Pause(attempt uint) time.Duration
}

// RetryWithContext keeps trying the fn until it returns false or no error.
// It will pause between retries according to backoff strategy.
// If context is canceled during a pause context error is returned without
// re-calling provided fn.
func RetryWithContext(ctx context.Context, strategy Strategy, fn Func) error {
	if ctx == nil {
		ctx = context.Background()
	}
	var attempt, reset uint
	cb := func() { reset = attempt }

	for {
		cont, err := fn(attempt, cb)
		if !cont || err == nil {
			return err
		}
		timer := time.NewTimer(strategy.Pause(attempt - reset))
		select {
		case <-timer.C:
			attempt++
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		}
	}
}

// Retry keeps trying the fn until it returns false or no error.
// It will pause between retries according to backoff strategy.
func Retry(strategy Strategy, fn Func) error {
	return RetryWithContext(context.Background(), strategy, fn)
}
