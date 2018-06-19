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
	DefaultStrategy = BackoffConfig{
		maxDelay:  180 * time.Second,
		baseDelay: 1.0 * time.Second,
		factor:    1.6,
		jitter:    0.2,
	}
}

type BackoffOption func(*BackoffConfig)

type BackoffConfig struct {
	// maxDelay is the upper bound of backoff delay.
	maxDelay time.Duration

	// baseDelay is the amount of time to wait before retrying after the first
	// failure.
	baseDelay time.Duration

	// factor is applied to the backoff after each retry.
	factor float64

	// jitter provides a range to randomize backoff delays.
	jitter float64
}

func New(options ...BackoffOption) BackoffStrategy {
	b := DefaultStrategy.(BackoffConfig)
	for _, option := range options {
		option(&b)
	}
	return b
}

func (bc BackoffConfig) Backoff(retries int) time.Duration {
	if retries == 0 {
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

func WithMaxDelay(md time.Duration) BackoffOption {
	return func(b *BackoffConfig) {
		b.maxDelay = md
	}
}

func WithBaseDelay(bd time.Duration) BackoffOption {
	return func(b *BackoffConfig) {
		b.baseDelay = bd
	}
}

func WithFactor(factor float64) BackoffOption {
	return func(b *BackoffConfig) {
		b.factor = factor
	}
}

func WithJitter(jitter float64) BackoffOption {
	return func(b *BackoffConfig) {
		b.jitter = jitter
	}
}
