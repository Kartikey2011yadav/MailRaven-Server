package services

import (
	"context"
	"fmt"
	"time"

	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/spam"
	"github.com/Kartikey2011yadav/mailraven-server/internal/config"
	"github.com/Kartikey2011yadav/mailraven-server/internal/observability"
)

// SpamProtectionService implements ports.SpamFilter
type SpamProtectionService struct {
	dnsbl     *spam.DNSBLChecker
	rateLimit *spam.RateLimiter
	config    config.SpamConfig
	logger    *observability.Logger
}

// NewSpamProtectionService creates a new spam protection service
func NewSpamProtectionService(cfg config.SpamConfig, logger *observability.Logger) (*SpamProtectionService, error) {
	window, err := time.ParseDuration(cfg.RateLimit.Window)
	if err != nil {
		// Fallback or error?
		window = time.Hour
	}

	return &SpamProtectionService{
		dnsbl:     spam.NewDNSBLChecker(cfg.DNSBLs),
		rateLimit: spam.NewRateLimiter(window, cfg.RateLimit.Count),
		config:    cfg,
		logger:    logger,
	}, nil
}

// CheckConnection checks if the connection is allowed
func (s *SpamProtectionService) CheckConnection(ctx context.Context, ip string) error {
	// 1. Check Rate Limit
	if !s.rateLimit.Allow(ip) {
		s.logger.Warn("connection rate limited", "ip", ip)
		return fmt.Errorf("rate limit exceeded")
	}

	// 2. Check DNSBL
	if len(s.config.DNSBLs) > 0 {
		if err := s.dnsbl.Check(ip); err != nil {
			s.logger.Warn("connection rejected by DNSBL", "ip", ip, "reason", err)
			return err
		}
	}

	return nil
}
