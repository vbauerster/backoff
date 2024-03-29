package exponential

import (
	"math/rand"
	"time"

	"github.com/vbauerster/backoff"
)

type Option func(*strategy)

type strategy struct {
	// maxDelay is the upper bound of backoff delay.
	maxDelay time.Duration

	// baseDelay is the amount of time to wait before retrying after the first
	// failure.
	baseDelay time.Duration

	// factor is applied to the backoff after each retry.
	factor float64

	// jitter provides a range to randomize backoff delays.
	jitter float64

	rand *rand.Rand
}

// New initializes exponential backoff.Strategy.
// It implements exponential backoff algorithm as defined in
// https://github.com/grpc/grpc/blob/master/doc/connection-backoff.md.
func New(options ...Option) backoff.Strategy {
	s := strategy{
		maxDelay:  180 * time.Second,
		baseDelay: 1.0 * time.Second,
		factor:    1.6,
		jitter:    0.2,
		rand:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	for _, option := range options {
		option(&s)
	}
	return s
}

func (s strategy) Pause(attempt uint) time.Duration {
	if attempt == 0 {
		return s.baseDelay
	}
	backoff, max := float64(s.baseDelay), float64(s.maxDelay)
	for backoff < max && attempt > 0 {
		backoff *= s.factor
		attempt--
	}
	if backoff > max {
		backoff = max
	}
	// Randomize backoff delays so that if a cluster of requests start at
	// the same time, they won't operate in lockstep.
	backoff *= 1 + s.jitter*(s.rand.Float64()*2-1)
	if backoff < 0 {
		return 0
	}
	return time.Duration(backoff)
}

// WithMaxDelay is the upper bound of backoff delay.
// Default is 180 seconds.
func WithMaxDelay(d time.Duration) Option {
	return func(s *strategy) {
		s.maxDelay = d
	}
}

// WithBaseDelay is the amount of time to wait before retrying after the first
// failure. Default is 1 second.
func WithBaseDelay(d time.Duration) Option {
	return func(s *strategy) {
		s.baseDelay = d
	}
}

// WithFactor is applied to the backoff after each retry.
// Default value is 1.6
func WithFactor(factor float64) Option {
	return func(s *strategy) {
		s.factor = factor
	}
}

// WithJitter provides a range to randomize backoff delays.
// Default value is 0.2
func WithJitter(jitter float64) Option {
	return func(s *strategy) {
		s.jitter = jitter
	}
}
