package services

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/spam"
	"github.com/Kartikey2011yadav/mailraven-server/internal/config"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
	"github.com/Kartikey2011yadav/mailraven-server/internal/observability"
)

// SpamProtectionService implements ports.SpamFilter
type SpamProtectionService struct {
	dnsbl     *spam.DNSBLChecker
	rateLimit *spam.RateLimiter
	rspamd    *spam.Client
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

	var rspamdClient *spam.Client
	if cfg.RspamdURL != "" {
		rspamdClient = spam.NewClient(cfg.RspamdURL)
	}

	return &SpamProtectionService{
		dnsbl:     spam.NewDNSBLChecker(cfg.DNSBLs),
		rateLimit: spam.NewRateLimiter(window, cfg.RateLimit.Count),
		rspamd:    rspamdClient,
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

// CheckContent checks the message content for spam
func (s *SpamProtectionService) CheckContent(ctx context.Context, content io.Reader, headers map[string]string) (*domain.SpamCheckResult, error) {
	if !s.config.Enabled || s.rspamd == nil {
		return &domain.SpamCheckResult{Action: domain.SpamActionPass}, nil
	}

	res, err := s.rspamd.Check(content, headers)
	if err != nil {
		return nil, err
	}

	action := domain.SpamActionPass
	switch res.Action {
	case "reject":
		action = domain.SpamActionReject
	case "add header", "rewrite subject":
		action = domain.SpamActionAddHeader
	case "soft reject":
		action = domain.SpamActionSoftReject
	}

	// Override based on scores defined in config
	if s.config.RejectScore > 0 && res.Score >= s.config.RejectScore {
		action = domain.SpamActionReject
	} else if s.config.HeaderScore > 0 && res.Score >= s.config.HeaderScore {
		if action == domain.SpamActionPass {
			action = domain.SpamActionAddHeader
		}
	}

	spamHeaders := make(map[string]string)
	spamHeaders["X-Spam-Score"] = fmt.Sprintf("%.2f", res.Score)
	if action != domain.SpamActionPass {
		spamHeaders["X-Spam-Status"] = "Yes"
	} else {
		spamHeaders["X-Spam-Status"] = "No"
	}

	return &domain.SpamCheckResult{
		Action:  action,
		Score:   res.Score,
		Headers: spamHeaders,
	}, nil
}
