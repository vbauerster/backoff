package backoff

import (
	"context"
	"time"

	"github.com/vbauerster/backoff/exponential"
)

// DefaultStrategy is default Strategy interface implementation.
var DefaultStrategy Strategy = exponential.New()

// Func represents functions that can be retried.
// Initial attempt is zero.
type Func func(attempt uint) (retry bool, err error)

// Strategy is a backoff methodology for retrying an operation.
type Strategy interface {
	// Pause returns duration of the next pause.
	Pause() time.Duration
	// Reset so that the next Pause returns BaseDelay duration.
	Reset()
}

// RetryWithContext keeps trying the fn until it returns false or no error.
// It will pause between retries according to backoff strategy.
// If context is canceled during a pause context error is returned without
// re-calling provided fn.
func RetryWithContext(ctx context.Context, strategy Strategy, fn Func) error {
	if ctx == nil {
		ctx = context.Background()
	}
	var attempt uint
	for {
		cont, err := fn(attempt)
		if !cont || err == nil {
			return err
		}
		timer := time.NewTimer(strategy.Pause())
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
