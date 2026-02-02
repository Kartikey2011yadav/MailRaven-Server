package services

import (
	"context"
	"errors"
	"strings"

	"github.com/Kartikey2011yadav/mailraven-server/internal/core/ports"
)

var (
	ErrAccessDenied = errors.New("access denied")
)

type EmailService struct {
	emailRepo ports.EmailRepository
}

func NewEmailService(emailRepo ports.EmailRepository) *EmailService {
	return &EmailService{emailRepo: emailRepo}
}

// UpdateACL updates the Access Control List for a mailbox
func (s *EmailService) UpdateACL(ctx context.Context, ownerID, mailboxName, identifier, rights string) error {
	// 1. Verify rights format
	if err := s.validateRights(rights); err != nil {
		return err
	}

	// 2. Verify mailbox exists
	_, err := s.emailRepo.GetMailbox(ctx, ownerID, mailboxName)
	if err != nil {
		return err
	}

	// 3. Delegate to repository
	return s.emailRepo.SetACL(ctx, ownerID, mailboxName, identifier, rights)
}

func (s *EmailService) validateRights(rights string) error {
	// RFC 4314 standard rights + 'a' (admin)
	// l: lookup, r: read, s: seen, w: write, i: insert, p: post, k: create, x: delete, t: delete msgs, e: expunge, a: admin
	allowed := "lrswipkxtea"
	for _, char := range rights {
		if !strings.ContainsRune(allowed, char) {
			return errors.New("invalid right: " + string(char))
		}
	}
	return nil
}

// CheckAccess verifies if a user has the required rights on a mailbox
func (s *EmailService) CheckAccess(ctx context.Context, ownerID, mailboxName, user string, requiredRights string) error {
	// 1. Owner always has full access (simplification for MVP)
	if ownerID == user {
		return nil
	}

	// 2. Fetch Mailbox ACLs
	mb, err := s.emailRepo.GetMailbox(ctx, ownerID, mailboxName)
	if err != nil {
		return err
	}

	// 3. Check rights
	// Identifier can be "anyone" or specific user
	// Rights format: "lrs"
	userRights := ""

	// Check specific user
	if r, ok := mb.ACL[user]; ok {
		userRights += r
	}
	// Check "anyone"
	if r, ok := mb.ACL["anyone"]; ok {
		userRights += r
	}
	// Check "authenticated" (if user != "")
	if user != "" {
		if r, ok := mb.ACL["authenticated"]; ok {
			userRights += r
		}
	}

	// Check required rights
	for _, char := range requiredRights {
		if !strings.ContainsRune(userRights, char) {
			return ErrAccessDenied
		}
	}

	return nil
}

// EnsureMailbox ensures a mailbox exists (e.g. for delivery)
func (s *EmailService) EnsureMailbox(ctx context.Context, userID, name string) error {
	// Check existence
	_, err := s.emailRepo.GetMailbox(ctx, userID, name)
	if err == nil {
		return nil // Exists
	}

	// Create if missing
	return s.emailRepo.CreateMailbox(ctx, userID, name)
}
