package local

import (
	"context"
	"sync"
)

// Broker implements ports.MessageBroker with direct synchronous execution.
// Used in standalone mode where no external message queue is available.
type Broker struct {
	mu       sync.RWMutex
	handlers map[string]func(data []byte) error
}

func NewBroker() *Broker {
	return &Broker{
		handlers: make(map[string]func(data []byte) error),
	}
}

func (b *Broker) Publish(_ context.Context, subject string, data []byte) error {
	b.mu.RLock()
	handler, ok := b.handlers[subject]
	b.mu.RUnlock()

	if ok {
		return handler(data)
	}
	return nil
}

func (b *Broker) QueueSubscribe(_ context.Context, subject string, _ string, handler func(data []byte) error) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[subject] = handler
	return nil
}

func (b *Broker) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers = make(map[string]func(data []byte) error)
	return nil
}
