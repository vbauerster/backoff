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

	reset *reset
	rand  *rand.Rand
}

type reset struct {
	attempt int
	d       time.Duration
}

func (s *reset) check(attempt int, ran time.Duration) int {
	if ran >= s.d {
		s.attempt = attempt
	}
	return attempt - s.attempt
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
		reset:     &reset{0, time.Hour},
	}
	for _, option := range options {
		option(&s)
	}
	return s
}

func (s strategy) Pause(attempt int, ran time.Duration) time.Duration {
	attempt = s.reset.check(attempt, ran)
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

// WithReset resets backoff strategy if previous backoff.Func ran
// duration has exceeded provided duration. Default is 1 hour.
func WithReset(d time.Duration) Option {
	return func(s *strategy) {
		s.reset.d = d
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

// WithCustomRand override default *rand.Rand which is
// seeded with rand.NewSource(time.Now().UnixNano()).
func WithCustomRand(rand *rand.Rand) Option {
	return func(s *strategy) {
		if rand == nil {
			return
		}
		s.rand = rand
	}
}
