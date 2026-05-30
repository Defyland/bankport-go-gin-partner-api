package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/allanflavio/bankport-go-gin-partner-api/internal/observability"
)

type RateLimiter struct {
	mu      sync.Mutex
	clock   func() time.Time
	windows map[string]rateWindow
}

type rateWindow struct {
	start time.Time
	count int
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		clock:   func() time.Time { return time.Now().UTC() },
		windows: make(map[string]rateWindow),
	}
}

func (l *RateLimiter) Allow(key string, limit int) (bool, int, time.Duration) {
	if limit <= 0 {
		return true, 0, 0
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.clock()
	l.pruneExpiredLocked(now)
	window := l.windows[key]
	if window.start.IsZero() || now.Sub(window.start) >= time.Minute {
		window = rateWindow{start: now, count: 0}
	}
	window.count++
	l.windows[key] = window

	remaining := limit - window.count
	if remaining < 0 {
		remaining = 0
	}
	resetAfter := time.Minute - now.Sub(window.start)
	if window.count > limit {
		return false, remaining, resetAfter
	}
	return true, remaining, resetAfter
}

func (l *RateLimiter) WindowCount() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.pruneExpiredLocked(l.clock())
	return len(l.windows)
}

func (l *RateLimiter) pruneExpiredLocked(now time.Time) {
	for key, window := range l.windows {
		if !window.start.IsZero() && now.Sub(window.start) >= time.Minute {
			delete(l.windows, key)
		}
	}
}

func RateLimit(limiter *RateLimiter, metrics *observability.Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		partner, ok := Partner(c)
		if !ok {
			c.Next()
			return
		}

		routeName := route(c)
		allowed, remaining, resetAfter := limiter.Allow(partner.ID+"|"+routeName, partner.RateLimitPerMinute)
		c.Header("X-RateLimit-Limit", strconv.Itoa(partner.RateLimitPerMinute))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset-Seconds", strconv.Itoa(int(resetAfter.Seconds())))
		if !allowed {
			c.Header("Retry-After", strconv.Itoa(int(resetAfter.Seconds())))
			metrics.RateLimitExceeded.WithLabelValues(partner.ID, routeName).Inc()
			Abort(c, http.StatusTooManyRequests, "rate_limit_exceeded", "The partner rate limit was exceeded.", map[string]any{
				"limit_per_minute": partner.RateLimitPerMinute,
				"retry_after":      int(resetAfter.Seconds()),
			})
			return
		}
		c.Next()
	}
}
