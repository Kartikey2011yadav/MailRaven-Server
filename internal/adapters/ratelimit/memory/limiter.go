package memory

import (
	"context"
	"sync"
	"time"
)

type window struct {
	count     int
	expiresAt time.Time
}

// RateLimiter implements ports.DistributedRateLimiter with in-memory sliding windows.
type RateLimiter struct {
	mu      sync.Mutex
	windows map[string]*window
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		windows: make(map[string]*window),
	}
}

func (r *RateLimiter) Allow(_ context.Context, key string, limit int, dur time.Duration) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	w, ok := r.windows[key]
	if !ok || now.After(w.expiresAt) {
		r.windows[key] = &window{count: 1, expiresAt: now.Add(dur)}
		return true, nil
	}

	if w.count >= limit {
		return false, nil
	}

	w.count++
	return true, nil
}
