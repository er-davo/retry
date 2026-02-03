package retry

import (
	"context"
)

// Do executes the provided function with retry semantics.
//
// Parameters:
//   - ctx controls cancellation and timeouts.
//   - maxAttempts defines the maximum number of attempts.
//     A value of 0 means retry indefinitely until the context is canceled.
//   - f is the function to execute; it receives the zero-based attempt number.
//
// This is a convenience wrapper around New(...) with default configuration
// and a custom maxAttempts value.
func Do(ctx context.Context, maxAttempts int, f AttemptFunc) error {
	return New(
		WithMaxAttempts(maxAttempts),
	).Do(ctx, f)
}
