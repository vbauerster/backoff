package backoff

import (
	"context"
	"time"
)

// Func represents functions that can be retried.
type Func func(count int, now time.Time) (retry bool, err error)

// Strategy is a backoff methodology for retrying an operation.
type Strategy interface {
	// Pause returns the duration of the next pause, according to attempt number.
	// Attempt = 1 indicates that strategy should reset to its initial state.
	Pause(attempt int) time.Duration
}

// Retry keeps trying the fn until it returns false, or no error is
// returned. It will pause between retries, according to backoff
// strategy.
func Retry(ctx context.Context, strategy Strategy, reset time.Duration, fn Func) error {
	if ctx == nil {
		ctx = context.Background()
	}
	var err error
	var cont bool
	var attempt int
	var count int
	timer := time.NewTimer(0)
	defer timer.Stop()

quit:
	for {
		select {
		case now := <-timer.C:
			cont, err = fn(count, now)
			if !cont || err == nil {
				break quit
			}
			if time.Since(now) >= reset {
				attempt = 0
			}
			count++
			attempt++
			timer.Reset(strategy.Pause(attempt))
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return err
}
