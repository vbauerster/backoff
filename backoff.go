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
	var cont bool
	var attempt, reset uint32
	cb := func() { reset = attempt }
	timer := time.NewTimer(0)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			cont, err = fn(uint(attempt+1), cb)
			if !cont || err == nil {
				return
			}
			pause := strategy.Pause(uint(attempt - reset))
			timer.Reset(pause)
			attempt++
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
