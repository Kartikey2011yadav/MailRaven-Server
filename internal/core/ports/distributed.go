package ports

import (
	"context"
	"time"
)

// Cache provides distributed key-value caching with TTL.
type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Increment(ctx context.Context, key string, delta int64) (int64, error)
}

// PubSub provides distributed publish/subscribe messaging for real-time events.
type PubSub interface {
	Publish(ctx context.Context, channel string, payload []byte) error
	Subscribe(ctx context.Context, channel string) (<-chan []byte, error)
	Unsubscribe(ctx context.Context, channel string) error
	Close() error
}

// DistributedRateLimiter provides rate limiting that works across multiple instances.
type DistributedRateLimiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
}

// LockHandle represents a held distributed lock.
type LockHandle interface {
	Release(ctx context.Context) error
	Extend(ctx context.Context, ttl time.Duration) error
}

// DistributedLock provides mutual exclusion across multiple pods/instances.
type DistributedLock interface {
	Acquire(ctx context.Context, name string, ttl time.Duration) (LockHandle, error)
	TryAcquire(ctx context.Context, name string, ttl time.Duration) (LockHandle, bool, error)
}

// MessageBroker provides distributed work queue semantics for async task processing.
type MessageBroker interface {
	Publish(ctx context.Context, subject string, data []byte) error
	// QueueSubscribe subscribes to a subject with a queue group (competing consumers).
	// Only one subscriber in the group receives each message.
	QueueSubscribe(ctx context.Context, subject string, queue string, handler func(data []byte) error) error
	Close() error
}

// NotificationEvent represents a mailbox change event.
type NotificationEvent struct {
	UserID    string
	Mailbox   string
	EventType string // "new_message", "message_deleted", "flags_changed"
	MessageID string
}

// NotificationBus provides cross-instance notification delivery for IMAP IDLE.
type NotificationBus interface {
	Notify(ctx context.Context, event NotificationEvent) error
	Listen(ctx context.Context, userID string) (<-chan NotificationEvent, func(), error)
}
