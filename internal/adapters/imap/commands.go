package imap

import (
	"context"
	"fmt"
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
