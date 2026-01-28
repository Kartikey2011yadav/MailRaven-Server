package imap

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net"
	"strings"

	"github.com/Kartikey2011yadav/mailraven-server/internal/config"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/ports"
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
	conn     net.Conn
	state    State
	config   config.IMAPConfig
	logger   *observability.Logger
	userRepo ports.UserRepository
	reader   *bufio.Reader
	writer   *bufio.Writer
	isTLS    bool
}

func NewSession(conn net.Conn, cfg config.IMAPConfig, logger *observability.Logger, userRepo ports.UserRepository) *Session {
	return &Session{
		conn:     conn,
		state:    StateNotAuthenticated,
		config:   cfg,
		logger:   logger,
		userRepo: userRepo,
		reader:   bufio.NewReader(conn),
		writer:   bufio.NewWriter(conn),
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

		// Parse and Handle
		cmd, err := ParseCommand(line)
		if err != nil {
			s.send("* BAD syntax error")
			continue
		}

		s.handleCommand(cmd)

		if s.state == StateLogout {
			return
		}
	}
}

func (s *Session) send(msg string) {
	s.writer.WriteString(msg + "\r\n")
	s.writer.Flush()
}

func (s *Session) upgradeToTLS() error {
	cert, err := tls.LoadX509KeyPair(s.config.TLSCert, s.config.TLSKey)
	if err != nil {
		return fmt.Errorf("failed to load keypair: %w", err)
	}

	config := &tls.Config{Certificates: []tls.Certificate{cert}}
	tlsConn := tls.Server(s.conn, config)

	if err := tlsConn.Handshake(); err != nil {
		return fmt.Errorf("tls handshake failed: %w", err)
	}

	s.conn = tlsConn
	s.reader = bufio.NewReader(tlsConn)
	s.writer = bufio.NewWriter(tlsConn)
	s.isTLS = true
	return nil
}
