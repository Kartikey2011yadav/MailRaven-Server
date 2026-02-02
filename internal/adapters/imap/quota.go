package imap

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
)

func (s *Session) handleGetQuotaRoot(cmd *Command) {
	if s.state != StateAuthenticated && s.state != StateSelected {
		s.send(fmt.Sprintf("%s NO [AUTH] Must be authenticated", cmd.Tag))
		return
	}
	if len(cmd.Args) < 1 {
		s.send(fmt.Sprintf("%s BAD Missing arguments", cmd.Tag))
		return
	}
	mailbox := cmd.Args[0]

	ctx := context.Background()
	user, err := s.userRepo.FindByEmail(ctx, s.user.Email)
	if err != nil {
		s.logger.Error("Storage error during GETQUOTAROOT", "error", err)
		s.send(fmt.Sprintf("%s NO Storage error", cmd.Tag))
		return
	}

	// Root is always "" (user root)
	s.send(fmt.Sprintf("* QUOTAROOT %q \"\"", mailbox))

	usageKB := user.StorageUsed / 1024
	quotaKB := user.StorageQuota / 1024
	if user.StorageQuota > 0 && quotaKB == 0 {
		quotaKB = 1
	}

	response := fmt.Sprintf("* QUOTA \"\" (STORAGE %d %d)", usageKB, quotaKB)
	s.send(response)

	s.send(fmt.Sprintf("%s OK GETQUOTAROOT completed", cmd.Tag))
}

func (s *Session) handleGetQuota(cmd *Command) {
	if s.state != StateAuthenticated && s.state != StateSelected {
		s.send(fmt.Sprintf("%s NO [AUTH] Must be authenticated", cmd.Tag))
		return
	}
	if len(cmd.Args) < 1 {
		s.send(fmt.Sprintf("%s BAD Missing arguments", cmd.Tag))
		return
	}
	root := cmd.Args[0]
	if root != "" && root != "\"\"" {
		s.send(fmt.Sprintf("%s NO Quota root does not exist", cmd.Tag))
		return
	}

	ctx := context.Background()
	user, err := s.userRepo.FindByEmail(ctx, s.user.Email)
	if err != nil {
		s.logger.Error("Storage error during GETQUOTA", "error", err)
		s.send(fmt.Sprintf("%s NO Storage error", cmd.Tag))
		return
	}

	usageKB := user.StorageUsed / 1024
	quotaKB := user.StorageQuota / 1024
	if user.StorageQuota > 0 && quotaKB == 0 {
		quotaKB = 1
	}

	response := fmt.Sprintf("* QUOTA \"\" (STORAGE %d %d)", usageKB, quotaKB)
	s.send(response)
	s.send(fmt.Sprintf("%s OK GETQUOTA completed", cmd.Tag))
}

func (s *Session) handleSetQuota(cmd *Command) {
	if s.state != StateAuthenticated && s.state != StateSelected {
		s.send(fmt.Sprintf("%s NO [AUTH] Must be authenticated", cmd.Tag))
		return
	}
	if s.user == nil || s.user.Role != domain.RoleAdmin {
		s.send(fmt.Sprintf("%s NO [PERMFAIL] Permission denied", cmd.Tag))
		return
	}
	if len(cmd.Args) < 2 {
		s.send(fmt.Sprintf("%s BAD Missing arguments", cmd.Tag))
		return
	}

	root := cmd.Args[0]
	targetEmail := s.user.Email
	if root != "" && root != "\"\"" {
		targetEmail = root
	}

	// Parse values usually enclosed in parens like (STORAGE 500)
	rest := strings.Join(cmd.Args[1:], " ")
	rest = strings.Trim(rest, "()")
	parts := strings.Fields(rest)

	if len(parts) != 2 || strings.ToUpper(parts[0]) != "STORAGE" {
		s.send(fmt.Sprintf("%s BAD Unix format (expected STORAGE <limit>)", cmd.Tag))
		return
	}

	limitKB, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		s.send(fmt.Sprintf("%s BAD Invalid limit", cmd.Tag))
		return
	}

	ctx := context.Background()
	err = s.userRepo.UpdateQuota(ctx, targetEmail, limitKB*1024)
	if err != nil {
		s.logger.Error("UpdateQuota failed", "error", err)
		s.send(fmt.Sprintf("%s NO Update failed", cmd.Tag))
		return
	}

	s.send(fmt.Sprintf("%s OK SETQUOTA completed", cmd.Tag))
}
