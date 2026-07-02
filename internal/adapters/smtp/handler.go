package smtp

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/smtp/mime"
	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/smtp/validators"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/ports"
	"github.com/Kartikey2011yadav/mailraven-server/internal/observability"
)

// Handler processes SMTP messages with validation and storage
type Handler struct {
	emailRepo     ports.EmailRepository
	userRepo      ports.UserRepository
	blobStore     ports.BlobStore
	searchIdx     ports.SearchIndex
	sieveExecutor ports.SieveExecutor
	logger        *observability.Logger
	metrics       *observability.Metrics
}

// NewHandler creates a new SMTP message handler
func NewHandler(
	emailRepo ports.EmailRepository,
	userRepo ports.UserRepository,
	blobStore ports.BlobStore,
	searchIdx ports.SearchIndex,
	sieveExecutor ports.SieveExecutor,
	logger *observability.Logger,
	metrics *observability.Metrics,
) *Handler {
	return &Handler{
		emailRepo:     emailRepo,
		userRepo:      userRepo,
		blobStore:     blobStore,
		searchIdx:     searchIdx,
		sieveExecutor: sieveExecutor,
		logger:        logger,
		metrics:       metrics,
	}
}

// Handle processes an incoming SMTP message
func (h *Handler) Handle(session *domain.SMTPSession, rawMessage []byte) error {
	ctx := context.Background()

	sessionLogger := h.logger.WithSMTPSession(session.SessionID, session.RemoteIP)

	// Step 1: Validate SPF
	sessionLogger.Info("validating SPF", "sender", session.Sender)
	spfResult, err := validators.ValidateSPF(ctx, session.RemoteIP, session.Sender, "")
	if err != nil {
		sessionLogger.Warn("SPF validation error", "error", err)
	}
	sessionLogger.Info("SPF result", "result", spfResult)

	// Step 2: Verify DKIM
	sessionLogger.Info("verifying DKIM")
	dkimResult, err := validators.VerifyDKIM(ctx, rawMessage)
	if err != nil {
		sessionLogger.Warn("DKIM verification error", "error", err)
	}
	sessionLogger.Info("DKIM result", "result", dkimResult)

	// Step 3: Evaluate DMARC
	sessionLogger.Info("evaluating DMARC")
	dmarcResult, dmarcPolicy, err := validators.EvaluateDMARC(ctx, session.Sender, spfResult, dkimResult)
	if err != nil {
		sessionLogger.Warn("DMARC evaluation error", "error", err)
	}
	sessionLogger.Info("DMARC result", "result", dmarcResult, "policy", dmarcPolicy)

	// Step 4: Check DMARC policy enforcement
	if validators.ShouldRejectMessage(dmarcResult, dmarcPolicy) {
		sessionLogger.Warn("message rejected by DMARC policy")
		h.metrics.IncrementMessagesRejected()
		return fmt.Errorf("rejected by DMARC policy")
	}

	// Step 5: Parse MIME message
	sessionLogger.Info("parsing MIME message")
	parsed, err := mime.ParseMessage(rawMessage)
	if err != nil {
		sessionLogger.Error("failed to parse MIME", "error", err)
		return fmt.Errorf("failed to parse message: %w", err)
	}

	// Step 6: Store message atomically (transaction + fsync)
	sessionLogger.Info("storing message atomically")
	if err := h.storeMessageAtomic(ctx, session, parsed, rawMessage, spfResult, dkimResult, dmarcResult, dmarcPolicy); err != nil {
		sessionLogger.Error("failed to store message", "error", err)
		h.metrics.IncrementStorageErrors()
		return fmt.Errorf("storage failed: %w", err)
	}

	sessionLogger.Info("message stored successfully")
	return nil
}

// storeMessageAtomic stores message with blob write + DB save and compensating cleanup on failure.
// Blob storage is non-transactional, so we write blob first and delete it if DB save fails.
func (h *Handler) storeMessageAtomic(
	ctx context.Context,
	session *domain.SMTPSession,
	parsed *mime.ParsedMessage,
	rawMessage []byte,
	spfResult validators.SPFResult,
	dkimResult validators.DKIMResult,
	dmarcResult validators.DMARCResult,
	dmarcPolicy validators.DMARCPolicy,
) error {
	// Generate message ID
	messageID := uuid.New().String()

	// Write message body to blob store (with fsync)
	h.logger.Info("writing message body to blob store", "message_id", messageID)
	bodyPath, err := h.blobStore.Write(ctx, messageID, rawMessage)
	if err != nil {
		return fmt.Errorf("failed to write blob: %w", err)
	}

	// Determine Mailbox (Routing via Sieve)
	targets, err := h.sieveExecutor.Execute(ctx, session.Recipients[0], rawMessage)
	if err != nil {
		h.logger.Error("sieve execution failed (fallback to INBOX)", "error", err)
		targets = []string{"INBOX"}
	}

	// Check Quota for Recipient
	user, err := h.userRepo.FindByEmail(ctx, session.Recipients[0])
	if err == nil && user.StorageQuota > 0 {
		requiredSize := int64(len(rawMessage)) * int64(len(targets))
		if user.StorageUsed+requiredSize > user.StorageQuota {
			h.logger.Warn("delivery rejected: quota exceeded", "user", user.Email)
			if delErr := h.blobStore.Delete(ctx, bodyPath); delErr != nil {
				h.logger.Warn("failed to cleanup blob after quota reject", "error", delErr)
			}
			return fmt.Errorf("quota exceeded")
		}
	}

	// Handle Discard
	if len(targets) == 0 {
		h.logger.Info("message discarded by sieve", "message_id", messageID)
		if err := h.blobStore.Delete(ctx, bodyPath); err != nil {
			h.logger.Warn("failed to cleanup discarded blob", "error", err)
		}
		return nil
	}

	// Save to database for each target mailbox
	for _, folder := range targets {
		msg := &domain.Message{
			MessageID:   parsed.MessageID,
			Sender:      session.Sender,
			Recipient:   session.Recipients[0],
			Subject:     parsed.Subject,
			Snippet:     parsed.Snippet,
			BodyPath:    bodyPath,
			ReadState:   false,
			ReceivedAt:  time.Now(),
			Mailbox:     folder,
			SPFResult:   string(spfResult),
			DKIMResult:  string(dkimResult),
			DMARCResult: string(dmarcResult),
			DMARCPolicy: string(dmarcPolicy),
		}

		if len(targets) > 1 {
			msg.ID = uuid.New().String()
		} else {
			msg.ID = messageID
		}

		h.logger.Info("saving message to database", "message_id", msg.ID, "mailbox", folder)
		h.metrics.IncrementStorageWrites()
		if err := h.emailRepo.Save(ctx, msg); err != nil {
			if delErr := h.blobStore.Delete(ctx, bodyPath); delErr != nil {
				h.logger.Warn("failed to cleanup blob after DB error", "error", delErr)
			}
			return fmt.Errorf("failed to save message to %s: %w", folder, err)
		}

		if err := h.searchIdx.Index(ctx, msg, parsed.PlainText); err != nil {
			h.logger.Warn("failed to index message", "error", err)
		}
	}

	h.logger.Info("message stored successfully", "message_id", messageID)

	// Increment storage usage (best effort)
	if user != nil {
		//nolint:errcheck
		_ = h.userRepo.IncrementStorageUsed(ctx, user.Email, int64(len(rawMessage))*int64(len(targets)))
	}

	return nil
}

// BuildMiddlewarePipeline creates the complete validation and storage pipeline
func (h *Handler) BuildMiddlewarePipeline() MessageHandler {
	// For now, just use the handler directly
	// Middleware pattern allows future expansion (rate limiting, custom filters, etc.)
	return h.Handle
}
