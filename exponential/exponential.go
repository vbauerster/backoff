package exponential

import (
	"math/rand"
	"sync/atomic"
	"time"
)

type Option func(*BackoffStrategy)

type BackoffStrategy struct {
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

	count atomic.Uint32
}

// New initializes exponential backoff.Strategy.
// It implements exponential backoff algorithm as defined in
// https://github.com/grpc/grpc/blob/master/doc/connection-backoff.md.
func New(options ...Option) *BackoffStrategy {
	strategy := &BackoffStrategy{
		maxDelay:  180 * time.Second,
		baseDelay: 1.0 * time.Second,
		factor:    1.6,
		jitter:    0.2,
		rand:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	for _, option := range options {
		option(strategy)
	}
	return strategy
}

func (s *BackoffStrategy) Pause() time.Duration {
	count := s.count.Load()
	defer s.count.Add(1)
	if count == 0 {
		return s.baseDelay
	}
	backoff, max := float64(s.baseDelay), float64(s.maxDelay)
	for backoff < max && count > 0 {
		backoff *= s.factor
		count--
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

func (s *BackoffStrategy) Reset() {
	s.count.Store(0)
}

// WithMaxDelay is the upper bound of backoff delay.
// Default is 180 seconds.
func WithMaxDelay(d time.Duration) Option {
	return func(s *BackoffStrategy) {
		s.maxDelay = d
	}
}

// WithBaseDelay is the amount of time to wait before retrying after the first
// failure. Default is 1 second.
func WithBaseDelay(d time.Duration) Option {
	return func(s *BackoffStrategy) {
		s.baseDelay = d
	}
}

// WithFactor is applied to the backoff after each retry.
// Default value is 1.6
func WithFactor(factor float64) Option {
	return func(s *BackoffStrategy) {
		s.factor = factor
	}
}

// WithJitter provides a range to randomize backoff delays.
// Default value is 0.2
func WithJitter(jitter float64) Option {
	return func(s *BackoffStrategy) {
		s.jitter = jitter
	}
}
