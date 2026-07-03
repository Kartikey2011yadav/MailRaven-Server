package memory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Kartikey2011yadav/mailraven-server/internal/core/ports"
)

// Lock implements ports.DistributedLock for single-instance use.
type Lock struct {
	mu    sync.Mutex
	locks map[string]time.Time
}

func NewLock() *Lock {
	return &Lock{locks: make(map[string]time.Time)}
}

func (l *Lock) Acquire(ctx context.Context, name string, ttl time.Duration) (ports.LockHandle, error) {
	for {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		h, ok, err := l.TryAcquire(ctx, name, ttl)
		if err != nil {
			return nil, err
		}
		if ok {
			return h, nil
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func (l *Lock) TryAcquire(_ context.Context, name string, ttl time.Duration) (ports.LockHandle, bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if exp, held := l.locks[name]; held && time.Now().Before(exp) {
		return nil, false, nil
	}

	l.locks[name] = time.Now().Add(ttl)
	return &handle{lock: l, name: name}, true, nil
}

type handle struct {
	lock *Lock
	name string
}

func (h *handle) Release(_ context.Context) error {
	h.lock.mu.Lock()
	defer h.lock.mu.Unlock()
	if _, ok := h.lock.locks[h.name]; !ok {
		return fmt.Errorf("lock %s not held", h.name)
	}
	delete(h.lock.locks, h.name)
	return nil
}

func (h *handle) Extend(_ context.Context, ttl time.Duration) error {
	h.lock.mu.Lock()
	defer h.lock.mu.Unlock()
	h.lock.locks[h.name] = time.Now().Add(ttl)
	return nil
}
