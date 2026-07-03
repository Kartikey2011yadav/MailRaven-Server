package infra

import (
	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/broker/local"
	memorycache "github.com/Kartikey2011yadav/mailraven-server/internal/adapters/cache/memory"
	memorylock "github.com/Kartikey2011yadav/mailraven-server/internal/adapters/lock/memory"
	memorypubsub "github.com/Kartikey2011yadav/mailraven-server/internal/adapters/pubsub/memory"
	memoryratelimit "github.com/Kartikey2011yadav/mailraven-server/internal/adapters/ratelimit/memory"
	"github.com/Kartikey2011yadav/mailraven-server/internal/config"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/ports"
	"github.com/Kartikey2011yadav/mailraven-server/internal/observability"
)

// Infrastructure holds all distributed adapters wired based on deployment mode.
type Infrastructure struct {
	Cache           ports.Cache
	PubSub          ports.PubSub
	RateLimiter     ports.DistributedRateLimiter
	Lock            ports.DistributedLock
	Broker          ports.MessageBroker
	Notifications   ports.NotificationBus
	closers         []func()
}

// Close gracefully shuts down all infrastructure connections.
func (i *Infrastructure) Close() {
	for _, fn := range i.closers {
		fn()
	}
}

// Build creates the infrastructure adapters based on the config mode.
// In standalone mode, all adapters are in-memory.
// When Redis/NATS are enabled, distributed adapters are used.
func Build(cfg *config.Config, logger *observability.Logger) (*Infrastructure, error) {
	infra := &Infrastructure{}

	// Cache + PubSub + RateLimiter + Lock
	if cfg.Redis.Enabled {
		// TODO: Wire Redis adapters when implemented (Phase 3)
		logger.Info("redis configured but adapters not yet implemented, falling back to in-memory")
		infra.wireInMemory()
	} else {
		infra.wireInMemory()
	}

	// Message Broker
	if cfg.NATS.Enabled {
		// TODO: Wire NATS adapter when implemented (Phase 3)
		logger.Info("nats configured but adapter not yet implemented, falling back to local broker")
		broker := local.NewBroker()
		infra.Broker = broker
		infra.closers = append(infra.closers, func() { broker.Close() })
	} else {
		broker := local.NewBroker()
		infra.Broker = broker
		infra.closers = append(infra.closers, func() { broker.Close() })
	}

	logger.Info("infrastructure initialized", "mode", string(cfg.Mode), "redis", cfg.Redis.Enabled, "nats", cfg.NATS.Enabled)
	return infra, nil
}

func (i *Infrastructure) wireInMemory() {
	cache := memorycache.NewCache()
	pubsub := memorypubsub.NewPubSub()
	notifications := memorypubsub.NewNotificationBus(pubsub)
	rateLimiter := memoryratelimit.NewRateLimiter()
	lock := memorylock.NewLock()

	i.Cache = cache
	i.PubSub = pubsub
	i.Notifications = notifications
	i.RateLimiter = rateLimiter
	i.Lock = lock

	i.closers = append(i.closers, func() {
		cache.Close()
		pubsub.Close()
	})
}
