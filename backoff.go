package backoff

import (
	"context"
	"time"
)

// Func represents functions that can be retried.
type Func func(attempt int) (retry bool, err error)

// Strategy is a backoff methodology for retrying an operation.
type Strategy interface {
	// Pause returns the duration of the next pause, according to the
	// attempt and Func ran duration.
	Pause(attempt int, ran time.Duration) time.Duration
}

// Retry keeps trying the fn until it returns false, or no error is
// returned. It will pause between retries, according to backoff
// strategy.
func Retry(ctx context.Context, strategy Strategy, fn Func) (err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	var attempt int
	var cont bool
	timer := time.NewTimer(0)
	defer timer.Stop()

quit:
	for {
		select {
		case start := <-timer.C:
			cont, err = fn(attempt)
			if !cont || err == nil {
				break quit
			}
			attempt++
			timer.Reset(strategy.Pause(attempt, time.Since(start)))
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return err
}
