package middleware

import (
	"testing"
	"time"
)

func TestRateLimiterPrunesExpiredWindows(t *testing.T) {
	now := time.Date(2026, 5, 30, 12, 0, 0, 0, time.UTC)
	limiter := NewRateLimiter()
	limiter.clock = func() time.Time { return now }

	allowed, _, _ := limiter.Allow("partner|route", 1)
	if !allowed {
		t.Fatal("expected first request to be allowed")
	}
	if limiter.WindowCount() != 1 {
		t.Fatalf("expected one active window, got %d", limiter.WindowCount())
	}

	now = now.Add(time.Minute)
	if limiter.WindowCount() != 0 {
		t.Fatalf("expected expired window to be pruned, got %d", limiter.WindowCount())
	}
}
