package backoff

import (
	"time"

	"github.com/vbauerster/backoff/saferand"
)

type BackoffStrategy interface {
	Backoff(retries int) time.Duration
}

var DefaultStrategy BackoffStrategy

func init() {
	// DefaultStrategy uses values specified for backoff in
	// https://github.com/grpc/grpc/blob/master/doc/connection-backoff.md.
	DefaultStrategy = &BackoffConfig{
		maxDelay:   180 * time.Second,
		baseDelay:  1.0 * time.Second,
		resetDelay: 1.0 * time.Hour,
		factor:     1.6,
		jitter:     0.2,
	}
}

type BackoffOption func(*BackoffConfig)

type BackoffConfig struct {
	// maxDelay is the upper bound of backoff delay.
	maxDelay time.Duration

	// baseDelay is the amount of time to wait before retrying after the first
	// failure.
	baseDelay time.Duration

	// resetDelay iteration run duration, which if passed starts backoff from scratch.
	resetDelay time.Duration

	// factor is applied to the backoff after each retry.
	factor float64

	// jitter provides a range to randomize backoff delays.
	jitter float64

	retryOffset int
	lastBackoff time.Duration
	lastNow     time.Time
}

func New(options ...BackoffOption) BackoffStrategy {
	b := DefaultStrategy.(*BackoffConfig)
	for _, option := range options {
		option(b)
	}
	return b
}

func (bc *BackoffConfig) Backoff(retries int) (offdur time.Duration) {
	defer func() {
		bc.lastBackoff = offdur
		bc.lastNow = time.Now()
	}()

	retries -= bc.retryOffset
	if retries == 0 || time.Since(bc.lastNow) >= bc.resetDelay+bc.lastBackoff {
		bc.retryOffset += retries
		return bc.baseDelay
	}
	backoff, max := float64(bc.baseDelay), float64(bc.maxDelay)
	for backoff < max && retries > 0 {
		backoff *= bc.factor
		retries--
	}
	if backoff > max {
		backoff = max
	}
	// Randomize backoff delays so that if a cluster of requests start at
	// the same time, they won't operate in lockstep.
	backoff *= 1 + bc.jitter*(saferand.Float64()*2-1)
	if backoff < 0 {
		return 0
	}
	return time.Duration(backoff)
}

// WithMaxDelay is the upper bound of backoff delay.
// Default is 180 seconds.
func WithMaxDelay(d time.Duration) BackoffOption {
	return func(b *BackoffConfig) {
		b.maxDelay = d
	}
}

// WithBaseDelay is the amount of time to wait before retrying after the first
// failure. Default is 1 second.
func WithBaseDelay(d time.Duration) BackoffOption {
	return func(b *BackoffConfig) {
		b.baseDelay = d
	}
}

// WithResetDelay is iteration run duration between retry to check,
// which if passed starts backoff from scratch, i.e. from base delay.
// Default is 1 hour.
func WithResetDelay(d time.Duration) BackoffOption {
	return func(b *BackoffConfig) {
		b.resetDelay = d
	}
}

// WithFactor is applied to the backoff after each retry.
// Default value is 1.6
func WithFactor(factor float64) BackoffOption {
	return func(b *BackoffConfig) {
		b.factor = factor
	}
}

// WithJitter provides a range to randomize backoff delays.
// Default value is 0.2
func WithJitter(jitter float64) BackoffOption {
	return func(b *BackoffConfig) {
		b.jitter = jitter
	}
}
