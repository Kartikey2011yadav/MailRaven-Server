package sieve

import (
	"context"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/ports"
	"github.com/google/uuid"
)

// VacationManager handles vacation auto-replies.
type VacationManager struct {
	vacationRepo ports.VacationRepository
	queueRepo    ports.QueueRepository
	blobStore    ports.BlobStore
	// domainRepo removed as unused
}

// NewVacationManager creates a new VacationManager.
func NewVacationManager(
	vacationRepo ports.VacationRepository,
	queueRepo ports.QueueRepository,
	blobStore ports.BlobStore,
) *VacationManager {
	return &VacationManager{
		vacationRepo: vacationRepo,
		queueRepo:    queueRepo,
		blobStore:    blobStore,
	}
}

type VacationConfig struct {
	Subject string
	Reason  string // Body
	Days    int    // Frequency (default 7)
	Mime    string // "text/plain" usually
	From    string // Alternate from address
	Handle  string // Unused key for tracking? Usually just sender-recipient pair.
}

// ProcessVacation handles the "vacation" action.
// args: :days, :subject, :from, :addresses, :mime, :handle, string (reason)
// See RFC 5230.
func (m *VacationManager) ProcessVacation(ctx context.Context, recipient string, msg *mail.Message, args map[string]interface{}, reason string) error {
	// 1. Determine Sender (Envelope From)
	// Sieve vacation replies are sent to the Envelope From (Return-Path).
	// In local processing, we might not have Envelope structure handy, but `msg.Header.Get("Return-Path")`
	// or `Sender` passed via context/arguments is safer.
	// For now, let's assume `msg` headers or we need to pass Envelope Sender explicitly to Execute.
	// We'll trust Return-Path header if present, else From.
	// However, auto-replies should NOT be sent to null return-path (<>) or mailing lists.

	sender := msg.Header.Get("Return-Path")
	if sender == "" {
		sender = msg.Header.Get("From") // Fallback, risky (mailing lists)
		// Parse address
		addr, err := mail.ParseAddress(sender)
		if err == nil {
			sender = addr.Address
		}
	}
	sender = strings.Trim(sender, "<>")
	if sender == "" || sender == "mailer-daemon" || sender == "postmaster" {
		return nil // Do not reply to bounces
	}

	// 2. Check for loop prevention headers
	if msg.Header.Get("Precedence") == "list" || msg.Header.Get("Precedence") == "bulk" || msg.Header.Get("Precedence") == "junk" {
		return nil
	}
	if msg.Header.Get("Auto-Submitted") != "" && msg.Header.Get("Auto-Submitted") != "no" {
		return nil
	}
	// Check for List-Id, List-Unsubscribe, etc.?
	if msg.Header.Get("List-Id") != "" {
		return nil
	}

	// 3. Rate Limit Check
	days := 7 // Default
	if d, ok := args["days"].(int); ok && d > 0 {
		days = d
	}
	// Min days check (e.g. 1 day) to avoid spam
	if days < 1 {
		days = 1
	}

	lastReply, err := m.vacationRepo.LastReply(ctx, recipient, sender)
	if err == nil && !lastReply.IsZero() {
		if time.Since(lastReply) < time.Duration(days)*24*time.Hour {
			// Already replied recently
			return nil
		}
	}

	// 4. Construct Reply
	var subject string
	if s, ok := args["subject"].(string); ok {
		subject = "Auto: " + s
	} else {
		// "Auto: " + Original Subject
		subject = "Auto: " + msg.Header.Get("Subject")
	}

	replyBody := reason
	// If :mime specified, use it? Simple text/plain implementation for MVP.

	// 5. Enqueue Reply
	// Construct outbound message
	replyID := uuid.New().String()

	// Store body in BlobStore? Outbound messages expect a BlobKey.
	blobKey := "outbound/" + replyID
	timestamp := time.Now()

	// Simple email construction
	rawEmail := fmt.Sprintf("From: <%s>\r\nTo: <%s>\r\nSubject: %s\r\nAuto-Submitted: auto-replied\r\n\r\n%s",
		recipient, sender, subject, replyBody)

	if _, err := m.blobStore.Write(ctx, blobKey, []byte(rawEmail)); err != nil {
		return fmt.Errorf("failed to write vacation body blob: %w", err)
	}

	outMsg := &domain.OutboundMessage{
		ID:          replyID,
		Sender:      recipient,
		Recipient:   sender,
		BlobKey:     blobKey,
		Status:      domain.QueueStatusPending,
		CreatedAt:   timestamp,
		RetryCount:  0,
		NextRetryAt: timestamp,
	}

	if err := m.queueRepo.Enqueue(ctx, outMsg); err != nil {
		return fmt.Errorf("failed to enqueue vacation reply: %w", err)
	}

	// 6. Update Tracking
	if err := m.vacationRepo.RecordReply(ctx, recipient, sender); err != nil {
		// Log but don't fail, worst case we send duplicate after timeout
		// But we just sent the email, so it's fine.
		return nil
	}

	return nil
}
