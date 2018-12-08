package backoff

import (
	"math/rand"
	"time"
)

type BackoffStrategy interface {
	Backoff(retries int) time.Duration
}

type BackoffOption func(*BackoffConfig)

type BackoffConfig struct {
	state

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

	rand *rand.Rand
}

type state struct {
	count  int
	reset  int
	offdur time.Duration
	iT     time.Time
}

// New initialises BackoffStrategy, which is NOT safe for concurrent use,
// as it stores some state to keep track of when to reset.
// For concurrent use, call New for each goroutine.
func New(options ...BackoffOption) BackoffStrategy {
	b := &BackoffConfig{
		maxDelay:   180 * time.Second,
		baseDelay:  1.0 * time.Second,
		resetDelay: 1.0 * time.Hour,
		factor:     1.6,
		jitter:     0.2,
		// not shared, not concurrent safe *rand.Rand instance
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	for _, option := range options {
		option(b)
	}
	return b
}

func (bc *BackoffConfig) Backoff(retries int) (offdur time.Duration) {
	defer func() {
		bc.count++
		bc.offdur = offdur
		bc.iT = time.Now()
	}()

	if !bc.iT.IsZero() && time.Since(bc.iT) >= bc.resetDelay+bc.offdur {
		bc.reset = retries - (retries - bc.count)
	}

	retries -= bc.reset
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
	backoff *= 1 + bc.jitter*(bc.rand.Float64()*2-1)
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

// WithCustomRand override default *rand.Rand which is
// seeded with rand.NewSource(time.Now().UnixNano()).
func WithCustomRand(rand *rand.Rand) BackoffOption {
	return func(b *BackoffConfig) {
		if rand == nil {
			return
		}
		b.rand = rand
	}
}
