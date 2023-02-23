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

// Retry keeps trying the fn until it returns false, or no error is
// returned. It will pause between retries, according to backoff
// strategy.
func Retry(ctx context.Context, strategy Strategy, fn Func) (err error) {
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
