package redis

import (
	"context"
	"encoding/json"

	"github.com/Kartikey2011yadav/mailraven-server/internal/core/ports"
)

// NotificationBus implements ports.NotificationBus using Redis Pub/Sub.
type NotificationBus struct {
	client *Client
	prefix string
}

func NewNotificationBus(client *Client) *NotificationBus {
	return &NotificationBus{client: client, prefix: "notifications:"}
}

func (n *NotificationBus) Notify(ctx context.Context, event ports.NotificationEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return n.client.rdb.Publish(ctx, n.prefix+event.UserID, data).Err()
}

func (n *NotificationBus) Listen(ctx context.Context, userID string) (<-chan ports.NotificationEvent, func(), error) {
	channel := n.prefix + userID
	sub := n.client.rdb.Subscribe(ctx, channel)

	out := make(chan ports.NotificationEvent, 64)
	done := make(chan struct{})

	go func() {
		defer close(out)
		ch := sub.Channel()
		for {
			select {
			case <-done:
				return
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				var event ports.NotificationEvent
				if json.Unmarshal([]byte(msg.Payload), &event) == nil {
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
		sub.Close()
	}
	return out, cancel, nil
}
