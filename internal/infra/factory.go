package infra

import (
	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/broker/local"
	natsBroker "github.com/Kartikey2011yadav/mailraven-server/internal/adapters/broker/nats"
	memorycache "github.com/Kartikey2011yadav/mailraven-server/internal/adapters/cache/memory"
	rediscache "github.com/Kartikey2011yadav/mailraven-server/internal/adapters/cache/redis"
	memorylock "github.com/Kartikey2011yadav/mailraven-server/internal/adapters/lock/memory"
	memorypubsub "github.com/Kartikey2011yadav/mailraven-server/internal/adapters/pubsub/memory"
	memoryratelimit "github.com/Kartikey2011yadav/mailraven-server/internal/adapters/ratelimit/memory"
	minioBlobStore "github.com/Kartikey2011yadav/mailraven-server/internal/adapters/storage/minio"
	"github.com/Kartikey2011yadav/mailraven-server/internal/config"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/ports"
	"github.com/Kartikey2011yadav/mailraven-server/internal/observability"
)

// Infrastructure holds all distributed adapters wired based on deployment mode.
type Infrastructure struct {
	Cache         ports.Cache
	PubSub        ports.PubSub
	RateLimiter   ports.DistributedRateLimiter
	Lock          ports.DistributedLock
	Broker        ports.MessageBroker
	Notifications ports.NotificationBus
	BlobStore     ports.BlobStore // nil if using disk (configured separately)
	closers       []func()
}

// Close gracefully shuts down all infrastructure connections.
func (i *Infrastructure) Close() {
	for _, fn := range i.closers {
		fn()
	}
}

// Build creates the infrastructure adapters based on the config mode.
func Build(cfg *config.Config, logger *observability.Logger) (*Infrastructure, error) {
	infra := &Infrastructure{}

	// Cache + PubSub + RateLimiter + Lock + Notifications
	if cfg.Redis.Enabled {
		redisClient, err := rediscache.NewClient(cfg.Redis)
		if err != nil {
			logger.Warn("redis connection failed, falling back to in-memory", "error", err)
			infra.wireInMemory()
		} else {
			logger.Info("redis connected", "addr", cfg.Redis.Addr)
			infra.Cache = rediscache.NewCache(redisClient)
			infra.RateLimiter = rediscache.NewRateLimiter(redisClient)
			infra.Notifications = rediscache.NewNotificationBus(redisClient)
			infra.Lock = memorylock.NewLock() // TODO: Redis lock adapter
			pubsub := memorypubsub.NewPubSub()
			infra.PubSub = pubsub
			infra.closers = append(infra.closers, func() { redisClient.Close() })
		}
	} else {
		infra.wireInMemory()
	}

	// Message Broker
	if cfg.NATS.Enabled {
		broker, err := natsBroker.NewBroker(cfg.NATS, logger)
		if err != nil {
			logger.Warn("nats connection failed, falling back to local broker", "error", err)
			localBroker := local.NewBroker()
			infra.Broker = localBroker
			infra.closers = append(infra.closers, func() { localBroker.Close() })
		} else {
			logger.Info("nats connected", "url", cfg.NATS.URL)
			infra.Broker = broker
			infra.closers = append(infra.closers, func() { broker.Close() })
		}
	} else {
		localBroker := local.NewBroker()
		infra.Broker = localBroker
		infra.closers = append(infra.closers, func() { localBroker.Close() })
	}

	// Object Store (MinIO)
	if cfg.ObjectStore.Driver == "minio" {
		store, err := minioBlobStore.NewBlobStore(cfg.ObjectStore)
		if err != nil {
			logger.Warn("minio connection failed, blob store must be configured separately", "error", err)
		} else {
			logger.Info("minio connected", "endpoint", cfg.ObjectStore.Endpoint, "bucket", cfg.ObjectStore.Bucket)
			infra.BlobStore = store
		}
	}

	logger.Info("infrastructure initialized", "mode", string(cfg.Mode), "redis", cfg.Redis.Enabled, "nats", cfg.NATS.Enabled, "object_store", cfg.ObjectStore.Driver)
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
