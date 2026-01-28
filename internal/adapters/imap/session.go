package imap

import (
	"bufio"
	"net"
	"strings"

	"github.com/Kartikey2011yadav/mailraven-server/internal/config"
	"github.com/Kartikey2011yadav/mailraven-server/internal/observability"
)

type State int

const (
	StateNotAuthenticated State = iota
	StateAuthenticated
	StateSelected
	StateLogout
)

type Session struct {
	conn   net.Conn
	state  State
	config config.IMAPConfig
	logger *observability.Logger
	reader *bufio.Reader
	writer *bufio.Writer
}

func NewSession(conn net.Conn, cfg config.IMAPConfig, logger *observability.Logger) *Session {
	return &Session{
		conn:   conn,
		state:  StateNotAuthenticated,
		config: cfg,
		logger: logger,
		reader: bufio.NewReader(conn),
		writer: bufio.NewWriter(conn),
	}
}

func (s *Session) Serve() {
	defer s.conn.Close()
	
	// Greeting
	// RFC 3501 Section 2.2.1
	s.send("* OK [CAPABILITY IMAP4rev1 STARTTLS AUTH=PLAIN] MailRaven Ready")

	for {
		line, err := s.reader.ReadString('\n')
		if err != nil {
			return
		}
		
		// RFC 3501 Section 2.2.1
		// Client commands are terminated by CRLF
		line = strings.TrimRight(line, "\r\n")
		
		if line == "" {
			continue
		}

		// Simple dispatch for now (will be replaced by full parser)
		s.handleLine(line)
		
		if s.state == StateLogout {
			return
		}
	}
}

func (s *Session) send(msg string) {
	s.writer.WriteString(msg + "\r\n")
	s.writer.Flush()
}

func (s *Session) handleLine(line string) {
	// Temporary implementation
	parts := strings.SplitN(line, " ", 2)
	tag := parts[0]
	
	if len(parts) == 1 {
		// Command without args?
		s.send(tag + " BAD Missing command")
		return
	}
	
	cmd := strings.ToUpper(strings.Split(parts[1], " ")[0]) // Simplistic

	if cmd == "LOGOUT" {
		s.send("* BYE Logging out")
		s.send(tag + " OK LOGOUT completed")
		s.state = StateLogout
	} else if cmd == "CAPABILITY" {
		s.send("* CAPABILITY IMAP4rev1 STARTTLS AUTH=PLAIN")
		s.send(tag + " OK CAPABILITY completed")
	} else {
		s.send(tag + " NO Unknown command")
	}
}
