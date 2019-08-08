package exponential

import (
	"math/rand"
	"time"

	"github.com/vbauerster/backoff"
)

type Option func(*Exponential)

type Exponential struct {
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

// New initialises Strategy, which is NOT safe for concurrent use,
// as it stores some state to keep track of when to reset.
// For concurrent use, call New for each goroutine.
func New(options ...Option) backoff.Strategy {
	eb := Exponential{
		maxDelay:  180 * time.Second,
		baseDelay: 1.0 * time.Second,
		factor:    1.6,
		jitter:    0.2,
		rand:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	for _, option := range options {
		option(&eb)
	}
	return eb
}

func (eb Exponential) Pause(attempt int) time.Duration {
	if attempt == 0 {
		return eb.baseDelay
	}
	backoff, max := float64(eb.baseDelay), float64(eb.maxDelay)
	for backoff < max && attempt > 0 {
		backoff *= eb.factor
		attempt--
	}
	if backoff > max {
		backoff = max
	}
	// Randomize backoff delays so that if a cluster of requests start at
	// the same time, they won't operate in lockstep.
	backoff *= 1 + eb.jitter*(eb.rand.Float64()*2-1)
	if backoff < 0 {
		return 0
	}
	return time.Duration(backoff)
}

// WithMaxDelay is the upper bound of backoff delay.
// Default is 180 seconds.
func WithMaxDelay(d time.Duration) Option {
	return func(eb *Exponential) {
		eb.maxDelay = d
	}
}

// failure. Default is 1 second.
func WithBaseDelay(d time.Duration) Option {
	return func(eb *Exponential) {
		eb.baseDelay = d
	}
}

// WithFactor is applied to the backoff after each retry.
// Default value is 1.6
func WithFactor(factor float64) Option {
	return func(eb *Exponential) {
		eb.factor = factor
	}
}

// WithJitter provides a range to randomize backoff delays.
// Default value is 0.2
func WithJitter(jitter float64) Option {
	return func(eb *Exponential) {
		eb.jitter = jitter
	}
}

// WithCustomRand override default *rand.Rand which is
// seeded with rand.NewSource(time.Now().UnixNano()).
func WithCustomRand(rand *rand.Rand) Option {
	return func(eb *Exponential) {
		if rand == nil {
			return
		}
		eb.rand = rand
	}
}
