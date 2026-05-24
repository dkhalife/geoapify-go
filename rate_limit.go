package geoapify

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Rate limit reason constants for [RateLimitError.Reason].
const (
	// RateLimitReasonHTTP429 indicates the server responded with HTTP 429.
	RateLimitReasonHTTP429 = "http_429"
	// RateLimitReasonDailyExceeded indicates the client-side daily quota
	// configured via [WithDailyLimit] was exhausted before the request was
	// dispatched.
	RateLimitReasonDailyExceeded = "daily_quota_exceeded"
)

// defaultRetryAfterFallback is used when an HTTP 429 response has no
// Retry-After header or the header value cannot be parsed.
const defaultRetryAfterFallback = 60 * time.Second

// RateLimitError signals that a request was rejected due to rate limiting,
// either by the GeoApify server (HTTP 429) or by a client-side quota
// configured via [WithDailyLimit].
//
// Callers can switch on [RateLimitError.Reason] to distinguish the two
// situations. For HTTP 429, the embedded [APIError] is reachable via
// [errors.As] so existing handling of [APIError] continues to work.
type RateLimitError struct {
	// RetryAfter is the suggested wait before retrying. For http_429 this is
	// derived from the response's Retry-After header (falling back to 60s
	// when missing or unparseable). For daily_quota_exceeded this is the
	// duration until the next UTC midnight.
	RetryAfter time.Duration
	// Reason is one of [RateLimitReasonHTTP429] or
	// [RateLimitReasonDailyExceeded].
	Reason string
	// APIError is set for http_429 responses and exposes the underlying
	// server error. It is nil for daily_quota_exceeded.
	APIError *APIError
}

// Error implements the error interface.
func (e *RateLimitError) Error() string {
	if e.APIError != nil {
		return fmt.Sprintf("geoapify: rate limited (%s, retry after %s): %s", e.Reason, e.RetryAfter, e.APIError.Error())
	}
	return fmt.Sprintf("geoapify: rate limited (%s, retry after %s)", e.Reason, e.RetryAfter)
}

// Unwrap returns the wrapped [APIError] so [errors.As] / [errors.Is] can
// reach it. Returns nil for daily-quota errors which have no server-side
// counterpart.
func (e *RateLimitError) Unwrap() error {
	if e.APIError == nil {
		return nil
	}
	return e.APIError
}

// IsRateLimitError reports whether err is or wraps a [*RateLimitError] and
// returns it when so.
func IsRateLimitError(err error) (*RateLimitError, bool) {
	var rle *RateLimitError
	if errors.As(err, &rle) {
		return rle, true
	}
	return nil, false
}

// parseRetryAfter parses the value of an HTTP Retry-After header per
// RFC 7231 section 7.1.3 (delta-seconds or HTTP-date). It returns fallback
// when value is empty or unparseable. Negative or past values clamp to 0.
func parseRetryAfter(value string, fallback time.Duration, now time.Time) time.Duration {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	if seconds, err := strconv.Atoi(value); err == nil {
		if seconds <= 0 {
			return 0
		}
		return time.Duration(seconds) * time.Second
	}
	if t, err := http.ParseTime(value); err == nil {
		d := t.Sub(now)
		if d < 0 {
			return 0
		}
		return d
	}
	return fallback
}

// WithRateLimit caps client throughput at the given average requests per
// second using a token bucket. Burst capacity equals ceil(requestsPerSecond)
// (minimum 1), so short bursts up to that size are admitted immediately
// before throttling kicks in.
//
// All HTTP requests block on the limiter before being sent, including
// retries. A value <= 0 disables the limiter.
func WithRateLimit(requestsPerSecond float64) Option {
	return func(c *Client) {
		if requestsPerSecond <= 0 {
			c.limiter = nil
			return
		}
		c.limiter = newTokenBucket(requestsPerSecond)
	}
}

// WithDailyLimit caps the number of requests the client will dispatch per
// UTC day. When the cap is reached, subsequent calls return a
// [*RateLimitError] with Reason [RateLimitReasonDailyExceeded] without
// issuing an HTTP request and without incrementing the counter. The counter
// resets at the next UTC midnight.
//
// A value <= 0 disables daily tracking (the default).
func WithDailyLimit(n int) Option {
	return func(c *Client) {
		if n <= 0 {
			c.daily = nil
			return
		}
		c.daily = newDailyCounter(n)
	}
}

// tokenBucket is a small goroutine-free token bucket. It allows tokens to
// go negative so that callers waiting for capacity effectively reserve a
// future slot; subsequent callers then queue behind them.
type tokenBucket struct {
	mu     sync.Mutex
	rate   float64 // tokens per second
	burst  float64
	tokens float64
	last   time.Time
	nowFn  func() time.Time
	// sleepFn is overridable in tests so we don't have to sleep in real time.
	sleepFn func(ctx context.Context, d time.Duration) error
}

func newTokenBucket(rps float64) *tokenBucket {
	burst := math.Max(1, math.Ceil(rps))
	return &tokenBucket{
		rate:    rps,
		burst:   burst,
		tokens:  burst,
		last:    time.Now(),
		nowFn:   time.Now,
		sleepFn: ctxSleep,
	}
}

// Wait blocks until a token is available or ctx is cancelled.
func (b *tokenBucket) Wait(ctx context.Context) error {
	b.mu.Lock()
	now := b.nowFn()
	if !b.last.IsZero() {
		elapsed := now.Sub(b.last).Seconds()
		if elapsed > 0 {
			b.tokens = math.Min(b.burst, b.tokens+elapsed*b.rate)
		}
	}
	b.last = now

	var wait time.Duration
	if b.tokens < 1 {
		need := 1 - b.tokens
		wait = time.Duration(need / b.rate * float64(time.Second))
	}
	b.tokens--
	b.mu.Unlock()

	if wait > 0 {
		return b.sleepFn(ctx, wait)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return nil
}

func ctxSleep(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// dailyCounter tracks request count per UTC day and resets at the next UTC
// midnight relative to the most recently observed timestamp.
type dailyCounter struct {
	mu       sync.Mutex
	limit    int
	count    int
	dayStart time.Time // UTC midnight of the currently tracked day
	nowFn    func() time.Time
}

func newDailyCounter(limit int) *dailyCounter {
	return &dailyCounter{limit: limit, nowFn: time.Now}
}

// checkAndIncrement atomically validates the daily cap and (on success)
// increments the counter. On exhaustion it returns a [*RateLimitError]
// describing time until the next UTC midnight and does NOT increment.
func (d *dailyCounter) checkAndIncrement() *RateLimitError {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := d.nowFn().UTC()
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	if dayStart.After(d.dayStart) {
		d.dayStart = dayStart
		d.count = 0
	}
	if d.count >= d.limit {
		next := dayStart.Add(24 * time.Hour)
		retry := next.Sub(now)
		if retry < 0 {
			retry = 0
		}
		return &RateLimitError{
			RetryAfter: retry,
			Reason:     RateLimitReasonDailyExceeded,
		}
	}
	d.count++
	return nil
}
