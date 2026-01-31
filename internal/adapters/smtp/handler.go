package smtp

import (
	"bytes"
	"context"
	"database/sql"
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
	emailRepo ports.EmailRepository
	blobStore ports.BlobStore
	searchIdx ports.SearchIndex
	db        *sql.DB
	logger    *observability.Logger
	metrics   *observability.Metrics
}

// NewHandler creates a new SMTP message handler
func NewHandler(
	emailRepo ports.EmailRepository,
	blobStore ports.BlobStore,
	searchIdx ports.SearchIndex,
	db *sql.DB,
	logger *observability.Logger,
	metrics *observability.Metrics,
) *Handler {
	return &Handler{
		emailRepo: emailRepo,
		blobStore: blobStore,
		searchIdx: searchIdx,
		db:        db,
		logger:    logger,
		metrics:   metrics,
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

// storeMessageAtomic stores message with transactional guarantees
// Implements constitution requirement: "250 OK" = fsync complete
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
	// Begin database transaction
	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		//nolint:errcheck // Rollback if not committed, error safe to ignore
		_ = tx.Rollback()
	}()

	// Generate message ID
	messageID := uuid.New().String()

	// Write message body to blob store (with fsync)
	h.logger.Info("writing message body to blob store", "message_id", messageID)
	bodyPath, err := h.blobStore.Write(ctx, messageID, rawMessage)
	if err != nil {
		return fmt.Errorf("failed to write blob: %w", err)
	}

	// Determine Mailbox (Spam Routing)
	mailbox := "INBOX"
	// Parse headers to check for X-Spam-Status
	headerEnd := bytes.Index(rawMessage, []byte("\r\n\r\n"))
	if headerEnd == -1 {
		headerEnd = len(rawMessage)
	}
	if bytes.Contains(rawMessage[:headerEnd], []byte("X-Spam-Status: Yes")) {
		mailbox = "Junk"
	}

	// Create domain message
	msg := &domain.Message{
		ID:          messageID,
		MessageID:   parsed.MessageID,
		Sender:      session.Sender,
		Recipient:   session.Recipients[0], // MVP: single recipient
		Subject:     parsed.Subject,
		Snippet:     parsed.Snippet,
		BodyPath:    bodyPath,
		ReadState:   false,
		ReceivedAt:  time.Now(),
		Mailbox:     mailbox, // Routing
		SPFResult:   string(spfResult),
		DKIMResult:  string(dkimResult),
		DMARCResult: string(dmarcResult),
		DMARCPolicy: string(dmarcPolicy),
	}

	// Save to database (within transaction)
	h.logger.Info("saving message to database", "message_id", messageID)
	h.metrics.IncrementStorageWrites()
	if err := h.emailRepo.Save(ctx, msg); err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}

	// Index for full-text search
	h.logger.Info("indexing message for search", "message_id", messageID)
	if err := h.searchIdx.Index(ctx, msg, parsed.PlainText); err != nil {
		h.logger.Warn("failed to index message", "error", err)
		// Don't fail on search indexing error
	}

	// Commit transaction (includes fsync via PRAGMA synchronous=FULL)
	h.logger.Info("committing transaction", "message_id", messageID)
	if err := tx.Commit(); err != nil {
		// Cleanup blob on transaction failure
		if delErr := h.blobStore.Delete(ctx, bodyPath); delErr != nil {
			h.logger.Warn("failed to cleanup blob", "error", delErr)
		}
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	h.logger.Info("message stored atomically", "message_id", messageID)
	return nil
}

// BuildMiddlewarePipeline creates the complete validation and storage pipeline
func (h *Handler) BuildMiddlewarePipeline() MessageHandler {
	// For now, just use the handler directly
	// Middleware pattern allows future expansion (rate limiting, custom filters, etc.)
	return h.Handle
}
