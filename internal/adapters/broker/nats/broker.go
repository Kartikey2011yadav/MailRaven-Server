package nats

import (
	"context"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/Kartikey2011yadav/mailraven-server/internal/config"
	"github.com/Kartikey2011yadav/mailraven-server/internal/observability"
)

// Broker implements ports.MessageBroker using NATS JetStream.
type Broker struct {
	conn   *nats.Conn
	js     nats.JetStreamContext
	logger *observability.Logger
	subs   []*nats.Subscription
}

// NewBroker connects to NATS and initializes JetStream.
func NewBroker(cfg config.NATSConfig, logger *observability.Logger) (*Broker, error) {
	nc, err := nats.Connect(cfg.URL,
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(10),
		nats.ReconnectWait(2*time.Second),
		nats.Timeout(5*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("nats connection failed: %w", err)
	}

	js, err := nc.JetStream()
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("jetstream init failed: %w", err)
	}

	// Ensure streams exist for our subjects
	streams := []struct {
		name     string
		subjects []string
	}{
		{"MAILRAVEN_DELIVERY", []string{"mailraven.delivery.>"}},
		{"MAILRAVEN_WORKERS", []string{"mailraven.spam.>", "mailraven.search.>"}},
	}

	for _, s := range streams {
		_, err := js.AddStream(&nats.StreamConfig{
			Name:      s.name,
			Subjects:  s.subjects,
			Retention: nats.WorkQueuePolicy,
			MaxAge:    24 * time.Hour,
		})
		if err != nil {
			logger.Warn("stream create/update", "stream", s.name, "error", err)
		}
	}

	return &Broker{conn: nc, js: js, logger: logger}, nil
}

func (b *Broker) Publish(_ context.Context, subject string, data []byte) error {
	_, err := b.js.Publish(subject, data)
	return err
}

func (b *Broker) QueueSubscribe(_ context.Context, subject string, queue string, handler func(data []byte) error) error {
	sub, err := b.js.QueueSubscribe(subject, queue, func(msg *nats.Msg) {
		if err := handler(msg.Data); err != nil {
			b.logger.Error("message handler failed", "subject", subject, "error", err)
			msg.Nak()
			return
		}
		msg.Ack()
	}, nats.Durable(queue), nats.ManualAck(), nats.AckWait(2*time.Minute))
	if err != nil {
		return fmt.Errorf("queue subscribe failed: %w", err)
	}
	b.subs = append(b.subs, sub)
	return nil
}

func (b *Broker) Close() error {
	for _, sub := range b.subs {
		sub.Drain()
	}
	b.conn.Close()
	return nil
}

// Ping checks if NATS is connected.
func (b *Broker) Ping() error {
	if !b.conn.IsConnected() {
		return fmt.Errorf("nats not connected")
	}
	return nil
}
