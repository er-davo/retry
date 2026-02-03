# retry

A small Go package for executing operations with retries.

The package provides:

* Pluggable backoff strategies (fixed, linear, exponential)
* Context-aware cancellation and timeouts
* Retry limits (including infinite retries)
* Explicit support for non-retryable errors

---

## Installation

```bash
go get github.com/your-org/retry
```

---

## Quick start

```go
err := retry.Do(ctx, 3, func(attempt int) error {
    return callExternalService()
})
```

* `attempt` is zero-based
* returning `nil` stops retries immediately
* returning an error triggers retry logic

---

## Retrier

For more control, use the configurable `Retrier`:

```go
r := retry.New(
    retry.WithMaxAttempts(5),
    retry.WithBackoff(retry.ExponentialBackoff{
        Base:   time.Second,
        Factor: 2,
        Max:    30 * time.Second,
        Jitter: 0.2,
    }),
)

err := r.Do(ctx, func(attempt int) error {
    return doWork()
})
```

---

## Backoff strategies

### Fixed backoff

```go
retry.FixedBackoff{
    Interval: time.Second,
    Jitter:   0.1,
}
```

### Linear backoff

```go
retry.LinearBackoff{
    Base:   time.Second,
    Step:   time.Second,
    Max:    10 * time.Second,
    Jitter: 0.1,
}
```

### Exponential backoff

```go
retry.ExponentialBackoff{
    Base:   time.Second,
    Factor: 2,
    Max:    30 * time.Second,
    Jitter: 0.2,
}
```

All backoff strategies support optional jitter to reduce coordinated retries
(thundering herd problem).

---

## Retry limits

```go
retry.WithMaxAttempts(3)
```

* `maxAttempts > 0` — retry up to the specified number of attempts
* `maxAttempts == 0` — retry indefinitely until the context is canceled

---

## Non-retryable errors

Retryability is controlled by a user-provided function, not by returning a special error from the attempt itself.

When creating a `Retrier`, you can provide an `IsRetryableFunc`:

```go
r := retry.New(
    retry.WithIsRetryableFunc(func(err error) bool {
        // return false for errors that should NOT be retried
        return !errors.Is(err, ErrPermanent)
    }),
)
```

If an attempt returns an error and `IsRetryableFunc` returns `false`, the retry loop stops immediately.

In this case, the retrier wraps the original error into an `UnretryableError`:

```go
return newUnretryableError(err)
```

This design allows callers to distinguish *why* the retry stopped:

* the operation succeeded
* retries were exhausted
* the context was canceled
* a non-retryable error was encountered

The original error is preserved and can be inspected using `errors.Unwrap` or `errors.As`:

```go
var ure *retry.UnretryableError
if errors.As(err, &ure) {
    // retry stopped because the error was marked as non-retryable
}
```

---

## Context handling

The retry loop respects `context.Context`:

* retries stop immediately when the context is canceled
* backoff waiting is interrupted on cancellation

This makes the package safe to use in:

* HTTP handlers
* gRPC requests
* background workers

---

## Design notes

* `Retrier` instances are **not thread-safe** and should not be reused
  concurrently
* Backoff strategies are fully decoupled from retry logic

---

## License

MIT
