package retry

import (
	"errors"
	"fmt"
)

// IsUnretryable reports whether the error is marked as unretryable.
func IsUnretryable(err error) bool {
	var e *UnretryableError
	return errors.As(err, &e)
}

// UnretryableError marks an error as non-retryable.
//
// When this error is returned (or wrapped), the retry mechanism
// should stop immediately and propagate the error to the caller.
// The original cause is preserved and can be accessed via errors.Unwrap
// or errors.As.
type UnretryableError struct {
	err error
}

func newUnretryableError(err error) error {
	if err == nil {
		return nil
	}
	return &UnretryableError{err: err}
}

func (e *UnretryableError) Error() string { return fmt.Sprintf("unretryable error: %v", e.err) }
func (e *UnretryableError) Unwrap() error { return e.err }
