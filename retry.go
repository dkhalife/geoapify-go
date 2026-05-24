package geoapify

import (
	"math"
	"math/rand/v2"
	"time"
)

type retryConfig struct {
	maxRetries   int
	initialDelay time.Duration
	maxDelay     time.Duration
}

// WithRetry enables retry with exponential backoff and jitter.
//
// Retries are attempted on transient failures:
//   - HTTP 429 Too Many Requests with a Retry-After value that fits inside
//     maxDelay (or no Retry-After at all, in which case exponential backoff
//     is used).
//   - HTTP 5xx responses.
//
// When a 429 carries a Retry-After larger than maxDelay, or when retries
// are exhausted, the SDK surfaces a [*RateLimitError] so the caller can
// implement its own cooldown. Daily-quota errors raised by
// [WithDailyLimit] are never retried.
//
// maxRetries is the maximum number of retry attempts (0 means no retries).
// initialDelay is the delay before the first retry. maxDelay caps both the
// exponential backoff and the acceptable server-suggested Retry-After.
func WithRetry(maxRetries int, initialDelay, maxDelay time.Duration) Option {
	return func(c *Client) {
		c.retry = &retryConfig{
			maxRetries:   maxRetries,
			initialDelay: initialDelay,
			maxDelay:     maxDelay,
		}
	}
}

// calculateDelay returns the exponential-backoff delay for the given
// attempt index, capped by maxDelay and with 50-100% jitter applied.
func (r *retryConfig) calculateDelay(attempt int) time.Duration {
	backoff := float64(r.initialDelay) * math.Pow(2, float64(attempt))
	if backoff > float64(r.maxDelay) {
		backoff = float64(r.maxDelay)
	}
	jitter := backoff * (0.5 + rand.Float64()*0.5)
	return time.Duration(jitter)
}
