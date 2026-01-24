package middleware

import (
	"encoding/json"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/http/dto"
)

// RateLimiter tracks request counts per IP address
type RateLimiter struct {
	mu       sync.RWMutex
	requests map[string]*ipLimit // IP address -> request info
	limit    int                 // Max requests per window
	window   time.Duration       // Time window
}

type ipLimit struct {
	count      int       // Request count in current window
	windowEnd  time.Time // End of current window
	lastAccess time.Time // For cleanup
}

// NewRateLimiter creates a rate limiter with specified limit per window
func NewRateLimiter(requestsPerMinute int) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string]*ipLimit),
		limit:    requestsPerMinute,
		window:   time.Minute,
	}

	// Start cleanup goroutine to remove stale entries
	go rl.cleanup()

	return rl
}

// cleanup removes entries older than 2 windows
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, limit := range rl.requests {
			if now.Sub(limit.lastAccess) > 2*rl.window {
				delete(rl.requests, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Allow checks if request from IP is allowed under rate limit
func (rl *RateLimiter) Allow(ip string) (allowed bool, retryAfter int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Get or create limit for this IP
	limit, exists := rl.requests[ip]
	if !exists {
		limit = &ipLimit{
			count:      0,
			windowEnd:  now.Add(rl.window),
			lastAccess: now,
		}
		rl.requests[ip] = limit
	}

	// Check if window expired, reset if so
	if now.After(limit.windowEnd) {
		limit.count = 0
		limit.windowEnd = now.Add(rl.window)
	}

	limit.lastAccess = now

	// Check if limit exceeded
	if limit.count >= rl.limit {
		retryAfter = int(time.Until(limit.windowEnd).Seconds())
		if retryAfter < 0 {
			retryAfter = 0
		}
		return false, retryAfter
	}

	// Increment and allow
	limit.count++
	return true, 0
}

// RateLimit creates middleware that enforces rate limiting per IP
func RateLimit(requestsPerMinute int) func(http.Handler) http.Handler {
	limiter := NewRateLimiter(requestsPerMinute)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract IP from request
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				ip = r.RemoteAddr // Fallback if no port
			}

			// Check rate limit
			allowed, retryAfter := limiter.Allow(ip)
			if !allowed {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", string(rune(retryAfter)))
				w.WriteHeader(http.StatusTooManyRequests)

				resp := dto.RateLimitResponse{
					Error:      "Rate limit exceeded",
					Message:    "Maximum 100 requests per minute exceeded",
					RetryAfter: retryAfter,
				}
				json.NewEncoder(w).Encode(resp)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
