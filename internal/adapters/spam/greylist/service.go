package greylist

import (
	"context"
	"fmt"
	"time"

	"github.com/Kartikey2011yadav/mailraven-server/internal/config"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/ports"
)

// Service implements ports.Greylister
type Service struct {
	repo       ports.GreylistRepository
	config     config.GreylistConfig
	retryDelay time.Duration
	expiration time.Duration
}

// NewService creates a new Greylist service
func NewService(repo ports.GreylistRepository, cfg config.GreylistConfig) (*Service, error) {
	retry, err := time.ParseDuration(cfg.RetryDelay)
	if err != nil {
		retry = 5 * time.Minute
	}
	expire, err := time.ParseDuration(cfg.Expiration)
	if err != nil {
		expire = 24 * time.Hour
	}
	return &Service{
		repo:       repo,
		config:     cfg,
		retryDelay: retry,
		expiration: expire,
	}, nil
}

// Check determines if a tuple should be allowed or temporarily rejected.
func (s *Service) Check(ctx context.Context, tuple domain.GreylistTuple) error {
	if !s.config.Enabled {
		return nil
	}

	entry, err := s.repo.Get(ctx, tuple)
	if err != nil {
		// If DB error, fail closed (temp error) to be safe and avoid passing spam
		return fmt.Errorf("greylist check failed: %w", err)
	}

	now := time.Now().Unix()

	// 1. New Triplet -> Block
	if entry == nil {
		newEntry := &domain.GreylistEntry{
			Tuple:        tuple,
			FirstSeenAt:  now,
			LastSeenAt:   now,
			BlockedCount: 1,
		}
		if err := s.repo.Upsert(ctx, newEntry); err != nil {
			return fmt.Errorf("failed to record new greylist entry: %w", err)
		}
		return fmt.Errorf("greylisted: new connection, please retry in %v", s.retryDelay)
	}

	// 2. Existing Triplet
	elapsed := time.Duration(now-entry.FirstSeenAt) * time.Second

	// Too Soon -> Block
	if elapsed < s.retryDelay {
		entry.LastSeenAt = now
		entry.BlockedCount++
		// Best effort update of last_seen
		_ = s.repo.Upsert(ctx, entry)

		remaining := s.retryDelay - elapsed
		return fmt.Errorf("greylisted: retry too soon, please wait %v", remaining.Round(time.Second))
	}

	// 3. Valid Retry -> Allow
	entry.LastSeenAt = now
	// We do not reset blocked count, it serves as metric
	if err := s.repo.Upsert(ctx, entry); err != nil {
		// Non-critical if update fails
		// log.Warn? We don't have logger here, return nil is fine.
	}
	return nil
}

// Prune removes expired entries.
func (s *Service) Prune(ctx context.Context) (int64, error) {
	cutoff := time.Now().Add(-s.expiration).Unix()
	return s.repo.DeleteOlderThan(ctx, cutoff)
}

// Check implementation
var _ ports.Greylister = (*Service)(nil)
