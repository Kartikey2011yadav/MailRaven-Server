package memory

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/Kartikey2011yadav/mailraven-server/internal/core/ports"
)

// PubSub implements ports.PubSub and ports.NotificationBus with local channels.
type PubSub struct {
	mu          sync.RWMutex
	subscribers map[string][]chan []byte
}

func NewPubSub() *PubSub {
	return &PubSub{
		subscribers: make(map[string][]chan []byte),
	}
}

func (p *PubSub) Publish(_ context.Context, channel string, payload []byte) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, ch := range p.subscribers[channel] {
		select {
		case ch <- payload:
		default:
		}
	}
	return nil
}

func (p *PubSub) Subscribe(_ context.Context, channel string) (<-chan []byte, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	ch := make(chan []byte, 64)
	p.subscribers[channel] = append(p.subscribers[channel], ch)
	return ch, nil
}

func (p *PubSub) Unsubscribe(_ context.Context, channel string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	subs := p.subscribers[channel]
	for _, ch := range subs {
		close(ch)
	}
	delete(p.subscribers, channel)
	return nil
}

func (p *PubSub) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, subs := range p.subscribers {
		for _, ch := range subs {
			close(ch)
		}
	}
	p.subscribers = make(map[string][]chan []byte)
	return nil
}

// NotificationBus wraps PubSub to implement ports.NotificationBus
type NotificationBus struct {
	pubsub *PubSub
}

func NewNotificationBus(ps *PubSub) *NotificationBus {
	return &NotificationBus{pubsub: ps}
}

func (n *NotificationBus) Notify(ctx context.Context, event ports.NotificationEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return n.pubsub.Publish(ctx, "notifications:"+event.UserID, data)
}

func (n *NotificationBus) Listen(ctx context.Context, userID string) (<-chan ports.NotificationEvent, func(), error) {
	channel := "notifications:" + userID
	raw, err := n.pubsub.Subscribe(ctx, channel)
	if err != nil {
		return nil, nil, err
	}

	out := make(chan ports.NotificationEvent, 64)
	done := make(chan struct{})

	go func() {
		defer close(out)
		for {
			select {
			case <-done:
				return
			case <-ctx.Done():
				return
			case data, ok := <-raw:
				if !ok {
					return
				}
				var event ports.NotificationEvent
				if json.Unmarshal(data, &event) == nil {
					select {
					case out <- event:
					default:
					}
				}
			}
		}
	}()

	cancel := func() {
		close(done)
		//nolint:errcheck
		_ = n.pubsub.Unsubscribe(ctx, channel)
	}
	return out, cancel, nil
}
