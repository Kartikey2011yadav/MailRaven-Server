package imap

import (
	"context"
	"fmt"
	"strings"
)

// RFC 4314 IMAP4 Access Control List (ACL) Extension

// handleSetACL SETACL <mailbox> <identifier> <rights>
func (s *Session) handleSetACL(cmd *Command) {
	if s.state != StateAuthenticated && s.state != StateSelected {
		s.send(fmt.Sprintf("%s NO [AUTH] Must be authenticated", cmd.Tag))
		return
	}
	if len(cmd.Args) != 3 {
		s.send(fmt.Sprintf("%s BAD Missing arguments", cmd.Tag))
		return
	}

	mailboxName := cmd.Args[0]
	identifier := cmd.Args[1]
	rights := cmd.Args[2]

	ctx := context.Background()

	// 1. Check mailbox exists and ownership
	// For MVP, we only allow setting ACL on owned mailboxes found by GetMailbox
	_, err := s.emailRepo.GetMailbox(ctx, s.user.Email, mailboxName)
	if err != nil {
		s.send(fmt.Sprintf("%s NO Mailbox not found", cmd.Tag))
		return
	}

	// 2. Set ACL
	if err := s.emailRepo.SetACL(ctx, s.user.Email, mailboxName, identifier, rights); err != nil {
		s.send(fmt.Sprintf("%s NO SetACL failed", cmd.Tag))
		return
	}

	s.send(fmt.Sprintf("%s OK SetACL completed", cmd.Tag))
}

// handleDeleteACL DELETEACL <mailbox> <identifier>
func (s *Session) handleDeleteACL(cmd *Command) {
	if s.state != StateAuthenticated && s.state != StateSelected {
		s.send(fmt.Sprintf("%s NO [AUTH] Must be authenticated", cmd.Tag))
		return
	}
	if len(cmd.Args) != 2 {
		s.send(fmt.Sprintf("%s BAD Missing arguments", cmd.Tag))
		return
	}

	mailboxName := cmd.Args[0]
	identifier := cmd.Args[1]

	ctx := context.Background()

	// 1. Check mailbox exists
	_, err := s.emailRepo.GetMailbox(ctx, s.user.Email, mailboxName)
	if err != nil {
		s.send(fmt.Sprintf("%s NO Mailbox not found", cmd.Tag))
		return
	}

	// 2. Remove ACL by setting empty rights (logic in Repo)
	// Or implementation of explicit delete if Repo requires different method.
	// My Repo SetACL deletes if rights == "".
	if err := s.emailRepo.SetACL(ctx, s.user.Email, mailboxName, identifier, ""); err != nil {
		s.send(fmt.Sprintf("%s NO DeleteACL failed", cmd.Tag))
		return
	}

	s.send(fmt.Sprintf("%s OK DeleteACL completed", cmd.Tag))
}

// handleGetACL GETACL <mailbox>
func (s *Session) handleGetACL(cmd *Command) {
	if s.state != StateAuthenticated && s.state != StateSelected {
		s.send(fmt.Sprintf("%s NO [AUTH] Must be authenticated", cmd.Tag))
		return
	}
	if len(cmd.Args) != 1 {
		s.send(fmt.Sprintf("%s BAD Missing arguments", cmd.Tag))
		return
	}

	mailboxName := cmd.Args[0]

	ctx := context.Background()
	mb, err := s.emailRepo.GetMailbox(ctx, s.user.Email, mailboxName)
	if err != nil {
		s.send(fmt.Sprintf("%s NO Mailbox not found", cmd.Tag))
		return
	}

	// Response format: * ACL <mailbox> <id> <rights> <id> <rights> ...
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("* ACL %s", mailboxName))

	for id, rights := range mb.ACL {
		sb.WriteString(fmt.Sprintf(" %s %s", id, rights))
	}

	s.send(sb.String())
	s.send(fmt.Sprintf("%s OK GetACL completed", cmd.Tag))
}

// handleListRights LISTRIGHTS <mailbox> <identifier>
// Required by RFC 4314
func (s *Session) handleListRights(cmd *Command) {
	if len(cmd.Args) < 2 {
		s.send(fmt.Sprintf("%s BAD Missing arguments", cmd.Tag))
		return
	}
	// Stub implementation: Return all rights as allowed
	// Standard rights: l (lookup), r (read), s (seen), w (write), i (insert), p (post), k (create), x (delete), t (delete msg), e (expunge), a (admin)
	s.send(fmt.Sprintf("* LISTRIGHTS %s %s l r s w i p k x t e a", cmd.Args[0], cmd.Args[1]))
	s.send(fmt.Sprintf("%s OK ListRights completed", cmd.Tag))
}

// handleMyRights MYRIGHTS <mailbox>
// Required by RFC 4314
func (s *Session) handleMyRights(cmd *Command) {
	if len(cmd.Args) != 1 {
		s.send(fmt.Sprintf("%s BAD Missing arguments", cmd.Tag))
		return
	}
	// Since we currently only expose Owned mailboxes, the user has full rights.
	// "acdilrsw" is a common set + kxte
	s.send(fmt.Sprintf("* MYRIGHTS %s lrswipkxte", cmd.Args[0]))
	s.send(fmt.Sprintf("%s OK MyRights completed", cmd.Tag))
}
