package geoapify

import (
	"context"
	"net/http"
	"sync/atomic"
	"testing"
	"time"
)

func TestWithRetry_Option(t *testing.T) {
	client := NewClient("key", WithRetry(3, 100*time.Millisecond, 5*time.Second))
	if client.retry == nil {
		t.Fatal("expected retry config")
	}
	assertEqual(t, client.retry.maxRetries, 3)
	assertEqual(t, client.retry.initialDelay, 100*time.Millisecond)
	assertEqual(t, client.retry.maxDelay, 5*time.Second)
}

func TestRetry_SuccessOnFirstAttempt(t *testing.T) {
	var calls atomic.Int32
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Write([]byte(`{"ok":true}`))
	})
	client.retry = &retryConfig{maxRetries: 3, initialDelay: time.Millisecond, maxDelay: 10 * time.Millisecond}

	var result struct{ OK bool }
	err := client.doGet(context.Background(), "/test", nil, &result)
	assertNoError(t, err)
	assertEqual(t, calls.Load(), int32(1))
	assertEqual(t, result.OK, true)
}

func TestRetry_SuccessAfterRetries(t *testing.T) {
	var calls atomic.Int32
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		if n < 3 {
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"message":"rate limited"}`))
			return
		}
		w.Write([]byte(`{"ok":true}`))
	})
	client.retry = &retryConfig{maxRetries: 5, initialDelay: time.Millisecond, maxDelay: 10 * time.Millisecond}

	var result struct{ OK bool }
	err := client.doGet(context.Background(), "/test", nil, &result)
	assertNoError(t, err)
	assertEqual(t, calls.Load(), int32(3))
}

func TestRetry_ExhaustedRetries(t *testing.T) {
	var calls atomic.Int32
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"server error"}`))
	})
	client.retry = &retryConfig{maxRetries: 2, initialDelay: time.Millisecond, maxDelay: 10 * time.Millisecond}

	err := client.doGet(context.Background(), "/test", nil, nil)
	assertError(t, err)
	assertEqual(t, calls.Load(), int32(3)) // initial + 2 retries
}

func TestRetry_NonRetryableError(t *testing.T) {
	var calls atomic.Int32
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"bad request"}`))
	})
	client.retry = &retryConfig{maxRetries: 3, initialDelay: time.Millisecond, maxDelay: 10 * time.Millisecond}

	err := client.doGet(context.Background(), "/test", nil, nil)
	assertError(t, err)
	assertEqual(t, calls.Load(), int32(1)) // no retries for 400
}

func TestRetry_ContextCancelled(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"message":"rate limited"}`))
	})
	client.retry = &retryConfig{maxRetries: 10, initialDelay: time.Second, maxDelay: 10 * time.Second}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := client.doGet(ctx, "/test", nil, nil)
	assertError(t, err)
}

func TestRetry_RespectsRetryAfterHeader(t *testing.T) {
	r := &retryConfig{initialDelay: time.Millisecond, maxDelay: time.Hour}
	delay := r.calculateDelay(0, &retryHint{retryAfter: "2"})
	assertEqual(t, delay, 2*time.Second)
}

func TestRetry_CalculateDelay_ExponentialBackoff(t *testing.T) {
	r := &retryConfig{initialDelay: 100 * time.Millisecond, maxDelay: 10 * time.Second}

	delay0 := r.calculateDelay(0, &retryHint{})
	delay1 := r.calculateDelay(1, &retryHint{})
	delay2 := r.calculateDelay(2, &retryHint{})

	// Delays should generally increase (with jitter they may vary).
	if delay0 > 200*time.Millisecond {
		t.Errorf("delay0 too large: %v", delay0)
	}
	if delay1 > 400*time.Millisecond {
		t.Errorf("delay1 too large: %v", delay1)
	}
	if delay2 > 800*time.Millisecond {
		t.Errorf("delay2 too large: %v", delay2)
	}
}

func TestRetry_CalculateDelay_CappedAtMax(t *testing.T) {
	r := &retryConfig{initialDelay: time.Second, maxDelay: 2 * time.Second}
	delay := r.calculateDelay(10, &retryHint{})
	if delay > 2*time.Second {
		t.Errorf("delay exceeded max: %v", delay)
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		code int
		want bool
	}{
		{200, false},
		{400, false},
		{401, false},
		{429, true},
		{500, true},
		{502, true},
		{503, true},
	}
	for _, tt := range tests {
		got := isRetryable(tt.code)
		if got != tt.want {
			t.Errorf("isRetryable(%d) = %v, want %v", tt.code, got, tt.want)
		}
	}
}
