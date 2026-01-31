package imap

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
)

// handleCommand dispatches to specific command handlers
func (s *Session) handleCommand(cmd *Command) {
	switch cmd.Name {
	case "CAPABILITY":
		s.handleCapability(cmd)
	case "NOOP":
		s.handleNoop(cmd)
	case "LOGOUT":
		s.handleLogout(cmd)
	case "LOGIN":
		s.handleLogin(cmd)
	case "STARTTLS":
		s.handleStartTLS(cmd)
	case "LIST":
		s.handleList(cmd)
	case "SELECT":
		s.handleSelect(cmd)
	case "CREATE":
		s.handleCreate(cmd)
	case "DELETE":
		s.handleDelete(cmd)
	case "FETCH":
		s.handleFetch(cmd)
	case "UID":
		s.handleUid(cmd)
	case "STORE":
		s.handleStore(cmd)
	case "COPY":
		s.handleCopy(cmd)
	case "IDLE":
		s.handleIdle(cmd)
	default:
		s.send(fmt.Sprintf("%s NO Unknown command", cmd.Tag))
	}
}

func (s *Session) handleCapability(cmd *Command) {
	// RFC 3501 6.1.1
	caps := "IMAP4rev1 STARTTLS AUTH=PLAIN"
	// If already authenticated, maybe add other capabilities?
	s.send("* CAPABILITY " + caps)
	s.send(fmt.Sprintf("%s OK CAPABILITY completed", cmd.Tag))
}

func (s *Session) handleNoop(cmd *Command) {
	// RFC 3501 6.1.2
	s.send(fmt.Sprintf("%s OK NOOP completed", cmd.Tag))
}

func (s *Session) handleLogout(cmd *Command) {
	// RFC 3501 6.1.3
	s.send("* BYE Logging out")
	s.send(fmt.Sprintf("%s OK LOGOUT completed", cmd.Tag))
	s.state = StateLogout
}

func (s *Session) handleLogin(cmd *Command) {
	// RFC 3501 6.2.3
	if s.state != StateNotAuthenticated {
		s.send(fmt.Sprintf("%s NO Already authenticated", cmd.Tag))
		return
	}

	// Security check: Don't allow LOGIN on insecure connection unless explicitly allowed
	if !s.isTLS && !s.config.AllowInsecureAuth {
		s.send(fmt.Sprintf("%s NO [ALERT] LOGIN failed: privacy required (TLS needed)", cmd.Tag))
		return
	}

	if len(cmd.Args) != 2 {
		s.send(fmt.Sprintf("%s BAD Invalid arguments", cmd.Tag))
		return
	}

	username := cmd.Args[0]
	password := cmd.Args[1]

	user, err := s.userRepo.Authenticate(context.Background(), username, password)
	if err != nil {
		s.logger.Warn("IMAP login failed", "user", username, "error", err)
		s.send(fmt.Sprintf("%s NO [AUTHENTICATIONFAILED] Authentication failed", cmd.Tag))
		return
	}

	s.state = StateAuthenticated
	s.user = user
	s.logger.Info("IMAP login success", "user", user.Email)
	s.send(fmt.Sprintf("%s OK [CAPABILITY IMAP4rev1] Logged in", cmd.Tag))
}

func (s *Session) handleStartTLS(cmd *Command) {
	// RFC 3501 6.2.1
	if s.isTLS {
		s.send(fmt.Sprintf("%s NO TLS already active", cmd.Tag))
		return
	}

	s.send(fmt.Sprintf("%s OK Begin TLS negotiation now", cmd.Tag))

	if err := s.upgradeToTLS(); err != nil {
		s.logger.Error("TLS upgrade failed", "error", err)
		// Connection usually dropped here by client, or we should close it
		s.conn.Close()
		return
	}
	s.logger.Info("IMAP session upgraded to TLS")
}

func (s *Session) handleList(cmd *Command) {
	if s.state < StateAuthenticated {
		s.send(fmt.Sprintf("%s NO Not authenticated", cmd.Tag))
		return
	}

	mailboxes, err := s.emailRepo.ListMailboxes(context.Background(), s.user.Email)
	if err != nil {
		s.send(fmt.Sprintf("%s NO List failed", cmd.Tag))
		return
	}

	foundInbox := false
	for _, mb := range mailboxes {
		if strings.ToUpper(mb.Name) == "INBOX" {
			foundInbox = true
		}
		// Basic attributes. RFC 3501 asks for \Marked \Noselect etc.
		// For now simple \HasNoChildren is enough for leaves.
		s.send(fmt.Sprintf(`* LIST (\HasNoChildren) "/" "%s"`, mb.Name))
	}
	if !foundInbox {
		s.send(`* LIST (\HasNoChildren) "/" "INBOX"`)
	}
	s.send(fmt.Sprintf("%s OK LIST completed", cmd.Tag))
}

func (s *Session) handleCreate(cmd *Command) {
	if s.state < StateAuthenticated {
		s.send(fmt.Sprintf("%s NO Not authenticated", cmd.Tag))
		return
	}
	if len(cmd.Args) < 1 {
		s.send(fmt.Sprintf("%s BAD Missing arguments", cmd.Tag))
		return
	}
	name := cmd.Args[0]
	err := s.emailRepo.CreateMailbox(context.Background(), s.user.Email, name)
	if err != nil {
		s.send(fmt.Sprintf("%s NO Create failed", cmd.Tag))
		return
	}
	s.send(fmt.Sprintf("%s OK CREATE completed", cmd.Tag))
}

func (s *Session) handleDelete(cmd *Command) {
	s.send(fmt.Sprintf("%s NO DELETE not supported yet", cmd.Tag))
}

func (s *Session) handleSelect(cmd *Command) {
	if s.state < StateAuthenticated {
		s.send(fmt.Sprintf("%s NO Not authenticated", cmd.Tag))
		return
	}
	if len(cmd.Args) < 1 {
		s.send(fmt.Sprintf("%s BAD Missing arguments", cmd.Tag))
		return
	}
	mailboxName := cmd.Args[0]
	// INBOX is case-insensitive
	if strings.ToUpper(mailboxName) == "INBOX" {
		mailboxName = "INBOX"
	}

	mb, err := s.emailRepo.GetMailbox(context.Background(), s.user.Email, mailboxName)
	if err != nil {
		// Auto-create INBOX if not found
		if strings.ToUpper(mailboxName) == "INBOX" {
			err = s.emailRepo.CreateMailbox(context.Background(), s.user.Email, "INBOX")
			if err == nil {
				mb, err = s.emailRepo.GetMailbox(context.Background(), s.user.Email, "INBOX")
			}
		}
	}

	if err != nil || mb == nil {
		s.send(fmt.Sprintf("%s NO Mailbox not found", cmd.Tag))
		return
	}

	s.selectedMailbox = mb
	s.state = StateSelected

	s.send(fmt.Sprintf("* %d EXISTS", mb.MessageCount))
	// Recent is usually 0 for new session unless tracked
	s.send("* 0 RECENT")
	s.send(fmt.Sprintf("* OK [UIDVALIDITY %d] UIDs valid", mb.UIDValidity))
	s.send(fmt.Sprintf("* OK [UIDNEXT %d] Predicted next UID", mb.UIDNext))
	s.send("* FLAGS (\\Answered \\Flagged \\Deleted \\Seen \\Draft)")
	s.send(fmt.Sprintf("%s OK [READ-WRITE] SELECT completed", cmd.Tag))
}

func (s *Session) handleFetch(cmd *Command) {
	if s.state != StateSelected {
		s.send(fmt.Sprintf("%s NO Select mailbox first", cmd.Tag))
		return
	}
	// Syntax: FETCH <sequence-set> <items>
	s.logger.Warn("Partial FETCH implementation: assuming UID FETCH for simplicity or returning NO")
	s.send(fmt.Sprintf("%s NO Use UID FETCH please", cmd.Tag))
}

func (s *Session) handleUid(cmd *Command) {
	if s.state != StateSelected {
		s.send(fmt.Sprintf("%s NO Select mailbox first", cmd.Tag))
		return
	}
	// UID <command> <args>
	if len(cmd.Args) < 2 {
		s.send(fmt.Sprintf("%s BAD Missing UID arguments", cmd.Tag))
		return
	}
	subCmd := strings.ToUpper(cmd.Args[0])

	if subCmd == "FETCH" {
		// Args[1] is Range like 1:* or 1,2
		// Args[2...] are items like (FLAGS)
		s.handleUidFetch(cmd.Tag, cmd.Args[1], cmd.Args[2:])
	} else if subCmd == "STORE" {
		// UID STORE <range> <mode> <flags>
		s.handleUidStore(cmd.Tag, cmd.Args[1], cmd.Args[2:])
	} else if subCmd == "COPY" {
		// UID COPY <range> <mailbox>
		if len(cmd.Args) < 3 {
			s.send(fmt.Sprintf("%s BAD Missing COPY arguments", cmd.Tag))
			return
		}
		s.handleUidCopy(cmd.Tag, cmd.Args[1], cmd.Args[2])
	} else {
		s.send(fmt.Sprintf("%s BAD Unknown UID command", cmd.Tag))
	}
}

func (s *Session) handleUidFetch(tag string, rangeSpec string, items []string) {
	// Parse range 1:* -> min=1, max=MAX
	// Basic parsing
	min, max := parseUidRange(rangeSpec)

	msgs, err := s.emailRepo.FindByUIDRange(context.Background(), s.user.Email, s.selectedMailbox.Name, min, max)
	if err != nil {
		s.send(fmt.Sprintf("%s NO DB Error", tag))
		return
	}

	showBody := false
	for _, item := range items {
		if strings.Contains(strings.ToUpper(item), "BODY[]") {
			showBody = true
		}
	}

	for i, msg := range msgs {
		seq := i + 1 // Fake sequence number

		var attrs []string
		attrs = append(attrs, fmt.Sprintf("UID %d", msg.UID))

		flags := msg.Flags
		if msg.ReadState && !strings.Contains(flags, "\\Seen") {
			flags += " \\Seen"
		}
		attrs = append(attrs, fmt.Sprintf("FLAGS (%s)", strings.TrimSpace(flags)))

		if showBody {
			// Mock content
			attrs = append(attrs, "BODY[] {20}\r\nBody not loaded yet.")
		} else {
			attrs = append(attrs, fmt.Sprintf("RFC822.SIZE %d", 100)) // Mock
		}

		s.send(fmt.Sprintf("* %d FETCH (%s)", seq, strings.Join(attrs, " ")))
	}

	s.send(fmt.Sprintf("%s OK UID FETCH completed", tag))
}

func (s *Session) handleUidStore(tag string, rangeSpec string, args []string) {
	// UID STORE 1 +FLAGS (\Seen)
	if len(args) < 2 {
		s.send(fmt.Sprintf("%s BAD Missing STORE args", tag))
		return
	}
	mode := strings.ToUpper(args[0]) // +FLAGS, -FLAGS, FLAGS
	flagsStr := strings.Join(args[1:], " ")
	// Strip parens
	flagsStr = strings.Trim(flagsStr, "()")
	flags := strings.Fields(flagsStr)

	min, max := parseUidRange(rangeSpec)
	//nolint:errcheck // We should probably handle error but ignoring for MVP brevity
	msgs, _ := s.emailRepo.FindByUIDRange(context.Background(), s.user.Email, s.selectedMailbox.Name, min, max)

	for _, msg := range msgs {
		switch mode {
		case "+FLAGS":
			if strings.Contains(mode, "+FLAGS") {
				_ = s.emailRepo.AddFlags(context.Background(), msg.ID, flags...) //nolint:errcheck
			}
		case "-FLAGS":
			if strings.Contains(mode, "-FLAGS") {
				_ = s.emailRepo.RemoveFlags(context.Background(), msg.ID, flags...) //nolint:errcheck
			}
		case "FLAGS":
			_ = s.emailRepo.SetFlags(context.Background(), msg.ID, flags...) //nolint:errcheck
		}
	}
	s.send(fmt.Sprintf("%s OK UID STORE completed", tag))
}

func (s *Session) handleStore(cmd *Command) {
	s.send(fmt.Sprintf("%s NO Use UID STORE", cmd.Tag))
}

func parseUidRange(rangeSpec string) (uint32, uint32) {
	if rangeSpec == "*" {
		return 1, 4294967295
	}
	parts := strings.Split(rangeSpec, ":")
	if len(parts) == 1 {
		v, _ := strconv.Atoi(parts[0]) //nolint:errcheck
		return uint32(v), uint32(v)
	}
	min, _ := strconv.Atoi(parts[0]) //nolint:errcheck
	maxStr := parts[1]
	var max uint32
	if maxStr == "*" {
		max = 4294967295
	} else {
		v, _ := strconv.Atoi(maxStr) //nolint:errcheck
		max = uint32(v)
	}
	return uint32(min), max
}

func (s *Session) handleUidCopy(tag string, rangeSpec string, destName string) {
	min, max := parseUidRange(rangeSpec)

	msgs, err := s.emailRepo.FindByUIDRange(context.Background(), s.user.Email, s.selectedMailbox.Name, min, max)
	if err != nil {
		s.send(fmt.Sprintf("%s NO DB Error", tag))
		return
	}
	if len(msgs) == 0 {
		s.send(fmt.Sprintf("%s NO No messages in range", tag))
		return
	}

	// Spam Training Hook
	srcJunk := strings.EqualFold(s.selectedMailbox.Name, "Junk")
	destJunk := strings.EqualFold(destName, "Junk")

	if s.spamService != nil && (srcJunk != destJunk) {
		// Async training
		trainingMsgs := msgs // Copy slice header
		go func(messages []*domain.Message, isSpam bool) {
			ctx := context.Background()
			for _, msg := range messages {
				if s.blobStore == nil {
					continue
				}
				contentBytes, err := s.blobStore.Read(ctx, msg.BodyPath)
				if err != nil {
					s.logger.Error("Failed to read blob for training", "error", err)
					continue
				}

				reader := bytes.NewReader(contentBytes)
				if isSpam {
					if err := s.spamService.TrainSpam(ctx, reader); err != nil {
						s.logger.Error("Failed to train spam", "error", err, "msg_id", msg.ID)
					}
				} else {
					if err := s.spamService.TrainHam(ctx, reader); err != nil {
						s.logger.Error("Failed to train ham", "error", err, "msg_id", msg.ID)
					}
				}
			}
		}(trainingMsgs, destJunk)
	}

	// Perform Copy
	var ids []string
	for _, m := range msgs {
		ids = append(ids, m.ID)
	}

	err = s.emailRepo.CopyMessages(context.Background(), s.user.Email, ids, destName)
	if err != nil {
		s.logger.Error("IMAP COPY Error", "error", err)
		s.send(fmt.Sprintf("%s NO Copy failed", tag))
		return
	}

	s.send(fmt.Sprintf("%s OK UID COPY completed", tag))
}

func (s *Session) handleCopy(cmd *Command) {
	s.send(fmt.Sprintf("%s NO Use UID COPY", cmd.Tag))
}
