package imap

import (
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

	if len(cmd.Args) != 2 {
		s.send(fmt.Sprintf("%s BAD Invalid arguments", cmd.Tag))
		return
	}

	username := cmd.Args[0]
	password := cmd.Args[1]

	// TODO: Validate against UserRepository
	// For groundwork, we'll accept "admin/password"
	if username == "admin" && password == "password" {
		s.state = StateAuthenticated
		s.send(fmt.Sprintf("%s OK LOGIN completed", cmd.Tag))
	} else {
		s.send(fmt.Sprintf("%s NO Authentication failed", cmd.Tag))
	}
}

func (s *Session) handleStartTLS(cmd *Command) {
	// RFC 3501 6.2.1
	// Requires refactoring Session to handle TLS upgrade on conn
	s.send(fmt.Sprintf("%s NO STARTTLS not implemented yet", cmd.Tag))
}
