package memory

import (
	"context"
	"sync"
	"time"
)

type entry struct {
	value     []byte
	expiresAt time.Time
}

// Cache implements ports.Cache with in-memory storage and TTL expiration.
type Cache struct {
	mu      sync.RWMutex
	entries map[string]entry
	stop    chan struct{}
}

func NewCache() *Cache {
	c := &Cache{
		entries: make(map[string]entry),
		stop:    make(chan struct{}),
	}
	go c.evictLoop()
	return c
}

func (c *Cache) Get(_ context.Context, key string) ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	e, ok := c.entries[key]
	if !ok || (!e.expiresAt.IsZero() && time.Now().After(e.expiresAt)) {
		return nil, nil
	}
	return e.value, nil
}

func (c *Cache) Set(_ context.Context, key string, value []byte, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}
	c.entries[key] = entry{value: value, expiresAt: expiresAt}
	return nil
}

func (c *Cache) Delete(_ context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, key)
	return nil
}

func (c *Cache) Increment(_ context.Context, key string, delta int64) (int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var current int64
	if e, ok := c.entries[key]; ok && (e.expiresAt.IsZero() || time.Now().Before(e.expiresAt)) {
		if len(e.value) == 8 {
			current = int64(e.value[0]) | int64(e.value[1])<<8 | int64(e.value[2])<<16 | int64(e.value[3])<<24 |
				int64(e.value[4])<<32 | int64(e.value[5])<<40 | int64(e.value[6])<<48 | int64(e.value[7])<<56
		}
	}
	current += delta
	b := make([]byte, 8)
	b[0] = byte(current)
	b[1] = byte(current >> 8)
	b[2] = byte(current >> 16)
	b[3] = byte(current >> 24)
	b[4] = byte(current >> 32)
	b[5] = byte(current >> 40)
	b[6] = byte(current >> 48)
	b[7] = byte(current >> 56)

	e := c.entries[key]
	e.value = b
	c.entries[key] = e
	return current, nil
}

func (c *Cache) Close() {
	close(c.stop)
}

func (c *Cache) evictLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-c.stop:
			return
		case <-ticker.C:
			c.mu.Lock()
			now := time.Now()
			for k, e := range c.entries {
				if !e.expiresAt.IsZero() && now.After(e.expiresAt) {
					delete(c.entries, k)
				}
			}
			c.mu.Unlock()
		}
	}
}
