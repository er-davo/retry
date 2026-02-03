package retry

import (
	"math"
	"math/rand/v2"
	"time"
)

// Backoff defines a strategy for calculating delay durations
// between retry attempts.
type Backoff interface {
	// Next returns the duration to wait before the next retry attempt.
	// The attempt parameter is zero-based (first retry = attempt 0).
	Next(attempt int) time.Duration
}

// FixedBackoff implements a constant delay between attempts.
//
// Interval defines the base delay duration.
// Jitter adds a random variation in the range [-Jitter, +Jitter]
// as a fraction of Interval (e.g. 0.2 = ±20%).
type FixedBackoff struct {
	Interval time.Duration
	Jitter   float64
}

// Next returns a constant delay with optional jitter applied.
func (f FixedBackoff) Next(attempt int) time.Duration {
	return addJitter(f.Interval, f.Jitter)
}

// LinearBackoff increases the delay linearly with each attempt.
//
// Base is the initial delay.
// Step is added for each subsequent attempt.
// Max caps the maximum delay (0 means no limit).
// Jitter adds a random variation as a fraction of the computed delay.
type LinearBackoff struct {
	Base   time.Duration
	Step   time.Duration
	Max    time.Duration
	Jitter float64
}

// Next returns a linearly increasing delay with optional max cap and jitter.
func (l LinearBackoff) Next(attempt int) time.Duration {
	d := l.Base + time.Duration(attempt)*l.Step
	if l.Max > 0 && d > l.Max {
		return l.Max
	}
	return addJitter(d, l.Jitter)
}

// ExponentialBackoff increases the delay exponentially with each attempt.
//
// Base is the initial delay.
// Factor is the exponential multiplier (e.g. 2.0).
// Max caps the maximum delay (0 means no limit).
// Jitter adds a random variation as a fraction of the computed delay.
type ExponentialBackoff struct {
	Base   time.Duration
	Factor float64
	Max    time.Duration
	Jitter float64
}

// Next returns an exponentially increasing delay with optional max cap and jitter.
func (e ExponentialBackoff) Next(attempt int) time.Duration {
	d := float64(e.Base) * math.Pow(e.Factor, float64(attempt))
	if e.Max > 0 && d > float64(e.Max) {
		return e.Max
	}
	return addJitter(time.Duration(d), e.Jitter)
}

// addJitter applies random jitter to a duration.
// Jitter must be in the range (0, 1). Values outside this range
// disable jitter and return the original duration.
func addJitter(d time.Duration, jitter float64) time.Duration {
	if jitter <= 0 || jitter >= 1 {
		return d
	}
	delta := (rand.Float64()*2 - 1) * jitter
	return time.Duration(float64(d) * (1 + delta))
}
