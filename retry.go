package retry

import (
	"context"
	"fmt"
	"time"
)

// RetryOption configures a Retrier.
type RetryOption func(*retrier)

// AttemptFunc represents a single retryable operation.
// The argument is the zero-based attempt number.
// Returning nil indicates success; a non-nil error triggers retry logic.
type AttemptFunc func(int) error

// IsRetryableFunc determines whether an error is retryable.
// Returning false stops retries immediately.
type IsRetryableFunc func(error) bool

// Retrier executes an operation with retry semantics.
type Retrier interface {
	// Do executes the provided AttemptFunc until it succeeds,
	// the context is canceled, or retry limits are exceeded.
	Do(context.Context, AttemptFunc) error
}

type retrier struct {
	backoff     Backoff
	maxAttempts int
	isRetryable IsRetryableFunc
}

// New creates a new Retrier with optional configuration.
// By default, it uses:
//   - a linear backoff
//   - a maximum of 3 attempts
//   - a retryable check that retries on any non-nil error
func New(opts ...RetryOption) Retrier {
	r := &retrier{
		backoff:     defaultBackoff(),
		maxAttempts: defaultAttempts(),
		isRetryable: defaultIsRetryableFunc(),
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// Do runs the provided AttemptFunc according to the retry configuration.
//
// The function:
//   - stops immediately if the context is canceled
//   - retries while attempts remain (or indefinitely if maxAttempts == 0)
//   - applies the configured backoff between attempts
//   - stops early if an error is deemed non-retryable
func (r retrier) Do(ctx context.Context, f AttemptFunc) error {
	var err error

	for attempt := 0; r.maxAttempts == 0 || attempt < r.maxAttempts; attempt++ {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}

		if err = f(attempt); err == nil {
			return nil
		}

		if r.isRetryable != nil && !r.isRetryable(err) {
			return newUnretryableError(err)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(r.backoff.Next(attempt)):
		}
	}

	return fmt.Errorf("all attempts failed: %w", err)
}

// defaultAttempts returns the default maximum number of retry attempts.
func defaultAttempts() int {
	return 3
}

// defaultBackoff returns the default backoff strategy.
func defaultBackoff() Backoff {
	return LinearBackoff{
		Base:   time.Second,
		Step:   time.Second,
		Max:    10 * time.Second,
		Jitter: 0.1,
	}
}

// defaultIsRetryableFunc retries on any non-nil error.
func defaultIsRetryableFunc() IsRetryableFunc {
	return func(err error) bool {
		return err != nil
	}
}

// WithMaxAttempts sets the maximum number of retry attempts.
// A value of 0 means unlimited retries.
func WithMaxAttempts(maxAttempts int) RetryOption {
	return func(r *retrier) {
		r.maxAttempts = maxAttempts
	}
}

// WithBackoff sets a custom backoff strategy.
func WithBackoff(backoff Backoff) RetryOption {
	return func(r *retrier) {
		r.backoff = backoff
	}
}

// WithIsRetryableFunc sets a custom function to determine
// whether an error should be retried.
func WithIsRetryableFunc(isRetryable IsRetryableFunc) RetryOption {
	return func(r *retrier) {
		r.isRetryable = isRetryable
	}
}
