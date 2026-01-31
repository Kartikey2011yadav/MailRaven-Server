package greylist

import (
	"context"
	"time"

	"github.com/Kartikey2011yadav/mailraven-server/internal/observability"
)

// StartPruning runs a background scheduler to prune expired entries
func (s *Service) StartPruning(ctx context.Context, interval time.Duration, logger *observability.Logger) {
	if !s.config.Enabled {
		return
	}

	logger.Info("Greylist pruner started", "interval", interval.String())
	ticker := time.NewTicker(interval)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				logger.Info("Greylist pruner stopped")
				return
			case <-ticker.C:
				count, err := s.Prune(ctx)
				if err != nil {
					logger.Error("Failed to prune greylist", "error", err)
				} else if count > 0 {
					logger.Info("Pruned greylist entries", "count", count)
				}
			}
		}
	}()
}
