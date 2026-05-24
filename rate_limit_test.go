package geoapify

import (
	"context"
	"errors"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestParseRetryAfter_DeltaSeconds(t *testing.T) {
	got := parseRetryAfter("30", 60*time.Second, time.Now())
	assertEqual(t, got, 30*time.Second)
}

func TestParseRetryAfter_HTTPDate(t *testing.T) {
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	future := now.Add(45 * time.Second).Format(http.TimeFormat)
	got := parseRetryAfter(future, 60*time.Second, now)
	// HTTP-date has second-level precision; tolerate a tiny rounding window.
	if got < 44*time.Second || got > 46*time.Second {
		t.Fatalf("want ~45s, got %v", got)
	}
}

func TestParseRetryAfter_PastDate(t *testing.T) {
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	past := now.Add(-time.Hour).Format(http.TimeFormat)
	got := parseRetryAfter(past, 60*time.Second, now)
	assertEqual(t, got, time.Duration(0))
}

func TestParseRetryAfter_MissingFallsBack(t *testing.T) {
	got := parseRetryAfter("", 60*time.Second, time.Now())
	assertEqual(t, got, 60*time.Second)
}

func TestParseRetryAfter_GarbageFallsBack(t *testing.T) {
	got := parseRetryAfter("not-a-date", 60*time.Second, time.Now())
	assertEqual(t, got, 60*time.Second)
}

func TestParseRetryAfter_NegativeSecondsClampsToZero(t *testing.T) {
	got := parseRetryAfter("-5", 60*time.Second, time.Now())
	assertEqual(t, got, time.Duration(0))
}

func TestRateLimitError_ErrorMessages(t *testing.T) {
	rle := &RateLimitError{Reason: RateLimitReasonDailyExceeded, RetryAfter: time.Hour}
	if rle.Error() == "" {
		t.Fatal("expected non-empty message")
	}
	rle2 := &RateLimitError{
		Reason:     RateLimitReasonHTTP429,
		RetryAfter: 30 * time.Second,
		APIError:   &APIError{StatusCode: 429, Message: "slow down"},
	}
	if rle2.Error() == "" {
		t.Fatal("expected non-empty message")
	}
}

func TestIsRateLimitError(t *testing.T) {
	rle := &RateLimitError{Reason: RateLimitReasonHTTP429}
	got, ok := IsRateLimitError(rle)
	if !ok || got != rle {
		t.Fatal("expected to unwrap RateLimitError")
	}
	if _, ok := IsRateLimitError(nil); ok {
		t.Fatal("nil should not be a rate-limit error")
	}
	if _, ok := IsRateLimitError(errors.New("other")); ok {
		t.Fatal("plain error should not be a rate-limit error")
	}
}

func TestRateLimitError_UnwrapsToAPIError(t *testing.T) {
	apiErr := &APIError{StatusCode: 429, Message: "rate limited"}
	rle := &RateLimitError{Reason: RateLimitReasonHTTP429, RetryAfter: time.Second, APIError: apiErr}
	var unwrapped *APIError
	if !errors.As(rle, &unwrapped) {
		t.Fatal("errors.As should reach embedded APIError")
	}
	assertEqual(t, unwrapped.StatusCode, 429)

	// IsAPIError sugar should also recognize it.
	got, ok := IsAPIError(rle)
	if !ok || got.StatusCode != 429 {
		t.Fatal("IsAPIError should find the wrapped APIError")
	}
}

func TestRateLimitError_DailyHasNilUnwrap(t *testing.T) {
	rle := &RateLimitError{Reason: RateLimitReasonDailyExceeded, RetryAfter: time.Hour}
	if rle.Unwrap() != nil {
		t.Fatal("daily quota error should not expose a wrapped error")
	}
}

// ----- HTTP 429 surfacing -----

func TestDo_429_NoRetry_SurfacesRateLimitErrorWithRetryAfter(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "30")
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"message":"slow down"}`))
	})

	err := client.doGet(context.Background(), "/x", nil, nil)
	rle, ok := IsRateLimitError(err)
	if !ok {
		t.Fatalf("expected *RateLimitError, got %T: %v", err, err)
	}
	assertEqual(t, rle.Reason, RateLimitReasonHTTP429)
	assertEqual(t, rle.RetryAfter, 30*time.Second)
	if rle.APIError == nil || rle.APIError.StatusCode != 429 {
		t.Fatal("expected wrapped APIError with 429 status")
	}
}

func TestDo_429_HTTPDateRetryAfter(t *testing.T) {
	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)
	future := now.Add(2 * time.Minute).Format(http.TimeFormat)
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", future)
		w.WriteHeader(http.StatusTooManyRequests)
	})
	client.nowFn = func() time.Time { return now }

	err := client.doGet(context.Background(), "/x", nil, nil)
	rle, ok := IsRateLimitError(err)
	if !ok {
		t.Fatalf("expected *RateLimitError, got %v", err)
	}
	if rle.RetryAfter < 119*time.Second || rle.RetryAfter > 121*time.Second {
		t.Fatalf("want ~120s, got %v", rle.RetryAfter)
	}
}

func TestDo_429_NoHeaderUsesFallback(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	})

	err := client.doGet(context.Background(), "/x", nil, nil)
	rle, ok := IsRateLimitError(err)
	if !ok {
		t.Fatalf("expected *RateLimitError, got %v", err)
	}
	assertEqual(t, rle.RetryAfter, defaultRetryAfterFallback)
}

func TestDo_429_WithRetry_ResolvesWithinMaxRetries(t *testing.T) {
	var calls atomic.Int32
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Write([]byte(`{"ok":true}`))
	})
	client.retry = &retryConfig{maxRetries: 3, initialDelay: time.Millisecond, maxDelay: time.Hour}

	var result struct{ OK bool }
	err := client.doGet(context.Background(), "/x", nil, &result)
	assertNoError(t, err)
	assertEqual(t, result.OK, true)
	assertEqual(t, calls.Load(), int32(2))
}

func TestDo_429_WithRetry_ExhaustedReturnsRateLimitError(t *testing.T) {
	var calls atomic.Int32
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("Retry-After", "0")
		w.WriteHeader(http.StatusTooManyRequests)
	})
	client.retry = &retryConfig{maxRetries: 2, initialDelay: time.Millisecond, maxDelay: time.Hour}

	err := client.doGet(context.Background(), "/x", nil, nil)
	rle, ok := IsRateLimitError(err)
	if !ok {
		t.Fatalf("expected *RateLimitError, got %v", err)
	}
	assertEqual(t, rle.Reason, RateLimitReasonHTTP429)
	assertEqual(t, calls.Load(), int32(3))
}

func TestDo_429_RetryAfterExceedsMaxDelay_Surfaces(t *testing.T) {
	var calls atomic.Int32
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("Retry-After", "3600")
		w.WriteHeader(http.StatusTooManyRequests)
	})
	client.retry = &retryConfig{maxRetries: 5, initialDelay: time.Millisecond, maxDelay: 10 * time.Second}

	err := client.doGet(context.Background(), "/x", nil, nil)
	rle, ok := IsRateLimitError(err)
	if !ok {
		t.Fatalf("expected *RateLimitError, got %v", err)
	}
	assertEqual(t, rle.RetryAfter, time.Hour)
	// Should NOT retry: server asked for longer than maxDelay.
	assertEqual(t, calls.Load(), int32(1))
}

// ----- WithDailyLimit -----

func TestWithDailyLimit_Option(t *testing.T) {
	c := NewClient("k", WithDailyLimit(100))
	if c.daily == nil {
		t.Fatal("expected daily counter to be configured")
	}
	assertEqual(t, c.daily.limit, 100)

	c2 := NewClient("k", WithDailyLimit(0))
	if c2.daily != nil {
		t.Fatal("WithDailyLimit(0) should disable tracking")
	}
}

func TestDailyCounter_RefusesAfterLimitWithoutDispatch(t *testing.T) {
	var calls atomic.Int32
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Write([]byte(`{}`))
	})
	client.daily = newDailyCounter(2)

	assertNoError(t, client.doGet(context.Background(), "/x", nil, nil))
	assertNoError(t, client.doGet(context.Background(), "/x", nil, nil))

	err := client.doGet(context.Background(), "/x", nil, nil)
	rle, ok := IsRateLimitError(err)
	if !ok {
		t.Fatalf("expected *RateLimitError, got %v", err)
	}
	assertEqual(t, rle.Reason, RateLimitReasonDailyExceeded)
	assertEqual(t, rle.APIError == nil, true)
	if rle.RetryAfter <= 0 || rle.RetryAfter > 24*time.Hour {
		t.Fatalf("unreasonable RetryAfter: %v", rle.RetryAfter)
	}
	assertEqual(t, calls.Load(), int32(2)) // third call never dispatched
}

func TestDailyCounter_ResetsAtUTCMidnight(t *testing.T) {
	now := time.Date(2025, 6, 1, 23, 59, 0, 0, time.UTC)
	d := newDailyCounter(1)
	d.nowFn = func() time.Time { return now }

	if rle := d.checkAndIncrement(); rle != nil {
		t.Fatalf("first call should pass: %v", rle)
	}
	if rle := d.checkAndIncrement(); rle == nil {
		t.Fatal("expected daily quota exhaustion on second call")
	} else {
		// 1 minute until UTC midnight.
		if rle.RetryAfter < 30*time.Second || rle.RetryAfter > 90*time.Second {
			t.Fatalf("want ~60s until midnight, got %v", rle.RetryAfter)
		}
	}

	// Advance past midnight.
	now = time.Date(2025, 6, 2, 0, 0, 1, 0, time.UTC)
	if rle := d.checkAndIncrement(); rle != nil {
		t.Fatalf("counter should have reset after midnight: %v", rle)
	}
}

func TestDailyCounter_QuotaExceededNotRetried(t *testing.T) {
	var calls atomic.Int32
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Write([]byte(`{}`))
	})
	client.daily = newDailyCounter(0) // immediately exhausted
	// (newDailyCounter(0) keeps limit=0; WithDailyLimit(0) would disable.)
	client.retry = &retryConfig{maxRetries: 5, initialDelay: time.Millisecond, maxDelay: time.Second}

	err := client.doGet(context.Background(), "/x", nil, nil)
	rle, ok := IsRateLimitError(err)
	if !ok {
		t.Fatalf("expected *RateLimitError, got %v", err)
	}
	assertEqual(t, rle.Reason, RateLimitReasonDailyExceeded)
	// Daily errors are never retried, even with WithRetry set, and the
	// HTTP layer must not be touched.
	assertEqual(t, calls.Load(), int32(0))
}

// ----- WithRateLimit -----

func TestWithRateLimit_OptionDisabledOnZero(t *testing.T) {
	c := NewClient("k", WithRateLimit(0))
	if c.limiter != nil {
		t.Fatal("WithRateLimit(0) should disable the limiter")
	}
	c2 := NewClient("k", WithRateLimit(5))
	if c2.limiter == nil {
		t.Fatal("expected limiter to be configured")
	}
}

func TestTokenBucket_SerializesBeyondBurst(t *testing.T) {
	b := newTokenBucket(10) // 10 rps, burst=10
	// Drain the burst.
	for i := 0; i < 10; i++ {
		if err := b.Wait(context.Background()); err != nil {
			t.Fatal(err)
		}
	}

	var sleeps []time.Duration
	var mu sync.Mutex
	b.sleepFn = func(ctx context.Context, d time.Duration) error {
		mu.Lock()
		sleeps = append(sleeps, d)
		mu.Unlock()
		return nil
	}
	for i := 0; i < 3; i++ {
		if err := b.Wait(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	if len(sleeps) != 3 {
		t.Fatalf("expected 3 throttled waits, got %d", len(sleeps))
	}
	for _, d := range sleeps {
		if d <= 0 {
			t.Fatalf("expected positive sleep, got %v", d)
		}
	}
}

func TestTokenBucket_ContextCancelled(t *testing.T) {
	b := newTokenBucket(1)
	// Drain burst then drive into negative.
	_ = b.Wait(context.Background())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := b.Wait(ctx); err == nil {
		t.Fatal("expected ctx error")
	}
}

func TestWithRateLimit_AppliedToEachAttempt(t *testing.T) {
	// Drive the limiter into negative-token territory before the request
	// runs so each subsequent Wait() actually blocks; then verify the
	// limiter is consulted on every retry attempt (not just the first).
	var calls atomic.Int32
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		if n < 3 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Write([]byte(`{"ok":true}`))
	})
	client.retry = &retryConfig{maxRetries: 5, initialDelay: time.Millisecond, maxDelay: time.Hour}
	client.limiter = newTokenBucket(1)
	// Drain the burst so the next Wait() must "sleep".
	_ = client.limiter.Wait(context.Background())

	var waits atomic.Int32
	client.limiter.sleepFn = func(ctx context.Context, d time.Duration) error {
		waits.Add(1)
		return nil
	}

	var result struct{ OK bool }
	assertNoError(t, client.doGet(context.Background(), "/x", nil, &result))
	// Three HTTP attempts → three Wait() calls → three sleeps.
	if got := waits.Load(); got != 3 {
		t.Fatalf("expected limiter sleepFn called 3 times (once per attempt), got %d", got)
	}
	assertEqual(t, calls.Load(), int32(3))
}

// ----- POST retry body / daily-per-logical-request -----

func TestDo_POST_BodyResentOnRetry(t *testing.T) {
	// Server asserts the request body is non-empty on every attempt.
	var calls atomic.Int32
	var emptyBody atomic.Bool
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		buf, _ := io.ReadAll(r.Body)
		if len(buf) == 0 {
			emptyBody.Store(true)
		}
		if n < 3 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Write([]byte(`{"ok":true}`))
	})
	client.retry = &retryConfig{maxRetries: 5, initialDelay: time.Millisecond, maxDelay: time.Hour}

	body := map[string]string{"mode": "drive"}
	var result struct{ OK bool }
	assertNoError(t, client.doPost(context.Background(), "/x", nil, body, &result))
	assertEqual(t, calls.Load(), int32(3))
	if emptyBody.Load() {
		t.Fatal("retry sent an empty body — GetBody not invoked")
	}
}

func TestDailyCounter_NotChargedPerRetryAttempt(t *testing.T) {
	// A single logical request that retries internally must consume exactly
	// one daily-quota token, not one per HTTP attempt.
	var calls atomic.Int32
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		if n < 3 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Write([]byte(`{"ok":true}`))
	})
	client.retry = &retryConfig{maxRetries: 5, initialDelay: time.Millisecond, maxDelay: time.Hour}
	client.daily = newDailyCounter(1)

	var result struct{ OK bool }
	assertNoError(t, client.doGet(context.Background(), "/x", nil, &result))
	assertEqual(t, calls.Load(), int32(3))
	// Three HTTP attempts but only one token consumed.
	assertEqual(t, client.daily.count, 1)
}
