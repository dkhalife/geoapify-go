package geoapify

import (
	"context"
	"math"
	"math/rand/v2"
	"strconv"
	"time"
)

type retryConfig struct {
	maxRetries   int
	initialDelay time.Duration
	maxDelay     time.Duration
}

type retryHint struct {
	retryAfter string
}

// WithRetry enables retry with exponential backoff and jitter.
// Retries are attempted on 429 (rate limit) and 5xx (server error) responses.
// maxRetries is the maximum number of retry attempts (0 means no retries).
// initialDelay is the delay before the first retry.
// maxDelay is the maximum delay between retries.
func WithRetry(maxRetries int, initialDelay, maxDelay time.Duration) Option {
	return func(c *Client) {
		c.retry = &retryConfig{
			maxRetries:   maxRetries,
			initialDelay: initialDelay,
			maxDelay:     maxDelay,
		}
	}
}

func (r *retryConfig) do(ctx context.Context, fn func() (*retryHint, error)) error {
	var lastErr error
	for attempt := range r.maxRetries + 1 {
		hint, err := fn()
		if err == nil {
			return nil
		}
		lastErr = err

		if hint == nil || attempt == r.maxRetries {
			break
		}

		delay := r.calculateDelay(attempt, hint)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}
	return lastErr
}

func (r *retryConfig) calculateDelay(attempt int, hint *retryHint) time.Duration {
	// Respect Retry-After header if present.
	if hint != nil && hint.retryAfter != "" {
		if seconds, err := strconv.Atoi(hint.retryAfter); err == nil {
			return time.Duration(seconds) * time.Second
		}
	}

	// Exponential backoff with jitter.
	backoff := float64(r.initialDelay) * math.Pow(2, float64(attempt))
	if backoff > float64(r.maxDelay) {
		backoff = float64(r.maxDelay)
	}

	// Add jitter: 50-100% of computed backoff.
	jitter := backoff * (0.5 + rand.Float64()*0.5)
	return time.Duration(jitter)
}
