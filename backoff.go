package backoff

import (
	"context"
	"time"
)

// Func represents functions that can be retried.
type Func func(attempt int, now time.Time) (retry bool, err error)

// Strategy is a backoff methodology for retrying an operation.
type Strategy interface {
	NextAttempt(attempt int) time.Duration
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
	timer := time.NewTimer(0)

quit:
	for {
		select {
		case now := <-timer.C:
			cont, err = fn(attempt, now)
			if !cont || err == nil {
				break quit
			}
			if time.Since(now) >= reset {
				attempt = 0
			}
			attempt++
			timer.Reset(strategy.NextAttempt(attempt))
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		}
	}
	return err
}
