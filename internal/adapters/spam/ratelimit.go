package spam

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter limits requests per IP
type RateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
	rate     rate.Limit
	burst    int
}

// NewRateLimiter creates a new rate limiter
// window is the time window (e.g. 1h)
// count is the max requests allowed in that window
func NewRateLimiter(window time.Duration, count int) *RateLimiter {
	r := rate.Limit(float64(count) / window.Seconds())
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     r,
		burst:    count, // Allow burst up to the full count initially? Or stricter?
		// Usually burst should be small if we want smooth rate, but for "100 emails per hour" usually means "can dump 100 now then wait".
		// I'll set burst to count.
	}
}

// Allow checks if the IP is allowed to proceed
func (l *RateLimiter) Allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	limiter, exists := l.limiters[ip]
	if !exists {
		limiter = rate.NewLimiter(l.rate, l.burst)
		l.limiters[ip] = limiter
	}

	return limiter.Allow()
}
