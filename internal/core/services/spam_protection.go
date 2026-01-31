package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/spam"
	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/spam/bayesian"
	"github.com/Kartikey2011yadav/mailraven-server/internal/config"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/ports"
	"github.com/Kartikey2011yadav/mailraven-server/internal/observability"
)

// SpamProtectionService implements ports.SpamFilter
type SpamProtectionService struct {
	dnsbl      *spam.DNSBLChecker
	rateLimit  *spam.RateLimiter
	rspamd     *spam.Client
	config     config.SpamConfig
	logger     *observability.Logger
	greylister ports.Greylister      // Core greylist logic
	bayes      ports.BayesRepository // Placeholder for training usage
	classifier ports.BayesClassifier // Naive Bayes classifier
	trainer    ports.BayesTrainer    // Naive Bayes trainer
}

// NewSpamProtectionService creates a new spam protection service
func NewSpamProtectionService(cfg config.SpamConfig, logger *observability.Logger, greylister ports.Greylister, bayes ports.BayesRepository) (*SpamProtectionService, error) {
	window, err := time.ParseDuration(cfg.RateLimit.Window)
	if err != nil {
		// Fallback or error?
		window = time.Hour
	}

	var rspamdClient *spam.Client
	if cfg.RspamdURL != "" {
		rspamdClient = spam.NewClient(cfg.RspamdURL)
	}

	// Initialize Classifier and Trainer if repo provided
	var classifier ports.BayesClassifier
	var trainer ports.BayesTrainer
	if bayes != nil {
		classifier = bayesian.NewClassifier(bayes)
		trainer = bayesian.NewTrainer(bayes)
	}

	return &SpamProtectionService{
		dnsbl:      spam.NewDNSBLChecker(cfg.DNSBLs),
		rateLimit:  spam.NewRateLimiter(window, cfg.RateLimit.Count),
		rspamd:     rspamdClient,
		config:     cfg,
		logger:     logger,
		greylister: greylister,
		bayes:      bayes,
		classifier: classifier,
		trainer:    trainer,
	}, nil
}

// TrainSpam learns from spam content
func (s *SpamProtectionService) TrainSpam(ctx context.Context, content io.Reader) error {
	if s.trainer == nil {
		return nil
	}
	return s.trainer.TrainSpam(ctx, content)
}

// TrainHam learns from ham (non-spam) content
func (s *SpamProtectionService) TrainHam(ctx context.Context, content io.Reader) error {
	if s.trainer == nil {
		return nil
	}
	return s.trainer.TrainHam(ctx, content)
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

// CheckRecipient checks if the sender/recipient pair allows delivery (Greylisting)
func (s *SpamProtectionService) CheckRecipient(ctx context.Context, ip, sender, recipient string) error {
	// Construct the tuple for checking
	tuple := domain.GreylistTuple{
		IPNet:     normalizeIP(ip),
		Sender:    sender,
		Recipient: recipient,
	}

	// Helper log
	s.logger.DebugContext(ctx, "checking greylist status",
		"ip_net", tuple.IPNet,
		"sender", sender,
		"recipient", recipient,
	)

	// Delegate to the specialized greylist service
	// If greylisting is disabled via config, the service.Check returns nil immediately.
	if err := s.greylister.Check(ctx, tuple); err != nil {
		s.logger.InfoContext(ctx, "greylist check prevented delivery",
			"reason", err,
			"tuple", tuple,
		)
		return err
	}

	return nil
}

// normalizeIP masks the IP to /24 (IPv4) or /64 (IPv6)
func normalizeIP(ipStr string) string {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return ipStr
	}
	if v4 := ip.To4(); v4 != nil {
		mask := net.CIDRMask(24, 32)
		return v4.Mask(mask).String()
	}
	mask := net.CIDRMask(64, 128)
	return ip.Mask(mask).String()
}

// CheckContent checks the message content for spam
func (s *SpamProtectionService) CheckContent(ctx context.Context, content io.Reader, headers map[string]string) (*domain.SpamCheckResult, error) {
	if !s.config.Enabled {
		return &domain.SpamCheckResult{Action: domain.SpamActionPass}, nil
	}

	// Buffer the content so we can read it multiple times (Rspamd + Bayes)
	bodyBytes, err := io.ReadAll(content)
	if err != nil {
		return nil, fmt.Errorf("failed to read content for spam check: %w", err)
	}

	var totalScore float64
	var rspamdAction domain.SpamAction
	var details string

	// 1. Check Rspamd
	if s.rspamd != nil {
		res, err := s.rspamd.Check(bytes.NewReader(bodyBytes), headers)
		if err != nil {
			s.logger.WarnContext(ctx, "rspamd check failed", "error", err)
		} else {
			totalScore += res.Score
			// Map Rspamd action to domain action
			switch res.Action {
			case "reject":
				rspamdAction = domain.SpamActionReject
			case "add header", "rewrite subject":
				rspamdAction = domain.SpamActionAddHeader
			case "soft reject":
				rspamdAction = domain.SpamActionSoftReject
			default:
				rspamdAction = domain.SpamActionPass
			}
		}
	}

	// 2. Check Naive Bayes
	var bayesScore float64
	if s.classifier != nil {
		prob, err := s.classifier.Classify(ctx, bytes.NewReader(bodyBytes))
		if err != nil {
			s.logger.WarnContext(ctx, "bayes check failed", "error", err)
		} else {
			// Convert Probability to Score modifier
			// < 0.1 -> -2.0 (Hammy)
			// > 0.9 -> +5.0 (Spammy)
			// > 0.7 -> +2.0 (Likely Spam)
			if prob > 0.9 {
				bayesScore = 5.0
			} else if prob > 0.7 {
				bayesScore = 2.0
			} else if prob < 0.1 {
				bayesScore = -2.0
			}
			totalScore += bayesScore
			details += fmt.Sprintf("Bayes:%.2f;", prob)
		}
	}

	// Determine Final Action
	// Priority: Reject > SoftReject > AddHeader > Pass
	// We use the aggregated Score vs Config thresholds.

	finalAction := domain.SpamActionPass

	// Use Rspamd action as baseline if available and severe
	if rspamdAction == domain.SpamActionReject {
		finalAction = domain.SpamActionReject
	} else if rspamdAction == domain.SpamActionSoftReject {
		finalAction = domain.SpamActionSoftReject
	}

	// Apply Score Thresholds
	if s.config.RejectScore > 0 && totalScore >= s.config.RejectScore {
		finalAction = domain.SpamActionReject
	} else if s.config.HeaderScore > 0 && totalScore >= s.config.HeaderScore {
		if finalAction == domain.SpamActionPass {
			finalAction = domain.SpamActionAddHeader
		}
	}

	spamHeaders := make(map[string]string)
	spamHeaders["X-Spam-Score"] = fmt.Sprintf("%.2f", totalScore)
	if finalAction != domain.SpamActionPass {
		spamHeaders["X-Spam-Status"] = "Yes"
	} else {
		spamHeaders["X-Spam-Status"] = "No"
	}
	if details != "" {
		spamHeaders["X-Spam-Details"] = details
	}

	return &domain.SpamCheckResult{
		Action:  finalAction,
		Score:   totalScore,
		Headers: spamHeaders,
	}, nil
}
