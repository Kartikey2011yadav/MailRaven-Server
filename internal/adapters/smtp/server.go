package smtp

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/Kartikey2011yadav/mailraven-server/internal/config"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/ports"
	"github.com/Kartikey2011yadav/mailraven-server/internal/observability"
)

// Server represents an SMTP server
type Server struct {
	config     *config.Config
	logger     *observability.Logger
	metrics    *observability.Metrics
	handler    MessageHandler
	spamFilter ports.SpamFilter
	listener   net.Listener
}

// NewServer creates a new SMTP server
func NewServer(cfg *config.Config, logger *observability.Logger, metrics *observability.Metrics, handler MessageHandler, spamFilter ports.SpamFilter) *Server {
	return &Server{
		config:     cfg,
		logger:     logger,
		metrics:    metrics,
		handler:    handler,
		spamFilter: spamFilter,
	}
}

// Start begins listening for SMTP connections
// Implements RFC 5321 - Simple Mail Transfer Protocol
func (s *Server) Start(ctx context.Context) error {
	addr := fmt.Sprintf(":%d", s.config.SMTP.Port)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}
	s.listener = listener

	s.logger.Info("SMTP server started", "address", addr)

	go func() {
		<-ctx.Done()
		s.listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil {
				// Context cancelled, shutting down
				return nil
			}
			s.logger.Error("failed to accept connection", "error", err)
			continue
		}

		s.metrics.IncrementSMTPConnections()
		go s.handleConnection(ctx, conn)
	}
}

// handleConnection processes a single SMTP connection
func (s *Server) handleConnection(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	remoteAddr := conn.RemoteAddr().String()
	remoteIP, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		s.logger.Warn("failed to split host port", "addr", remoteAddr, "error", err)
		return
	}
	if remoteIP == "" {
		remoteIP = remoteAddr // fallback
	}

	sessionID := fmt.Sprintf("smtp-%d", time.Now().UnixNano())
	sessionLogger := s.logger.WithSMTPSession(sessionID, remoteIP)

	// Check Spam Filter (DNSBL / Rate Limit)
	if s.spamFilter != nil {
		if err := s.spamFilter.CheckConnection(ctx, remoteIP); err != nil {
			sessionLogger.Warn("connection rejected by spam filter", "error", err)
			// Return 554 No SMTP service here
			_ = conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
			fmt.Fprintf(conn, "554 Service unavailable: %v\r\n", err)
			return
		}
	}

	sessionLogger.Info("new SMTP connection")

	session := &domain.SMTPSession{
		SessionID:   sessionID,
		RemoteIP:    remoteIP,
		ConnectedAt: time.Now(),
	}

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	// RFC 5321 Section 4.2: SMTP greeting
	s.send(writer, "220 %s ESMTP MailRaven", s.config.SMTP.Hostname)

	for {
		// Set read deadline
		_ = conn.SetReadDeadline(time.Now().Add(5 * time.Minute))

		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				sessionLogger.Error("read error", "error", err)
			}
			return
		}

		line = strings.TrimSpace(line)
		sessionLogger.Debug("received command", "command", line)

		// Parse SMTP command
		parts := strings.SplitN(line, " ", 2)
		command := strings.ToUpper(parts[0])
		args := ""
		if len(parts) > 1 {
			args = parts[1]
		}

		// RFC 5321 Section 4.1: SMTP commands
		switch command {
		case "EHLO", "HELO":
			s.handleEHLO(writer, args, session, sessionLogger)

		case "MAIL":
			if !strings.HasPrefix(strings.ToUpper(args), "FROM:") {
				s.send(writer, "501 Syntax error in parameters")
				continue
			}
			sender := extractEmailAddress(strings.TrimPrefix(strings.ToUpper(args), "FROM:"))
			session.Sender = sender
			sessionLogger.Info("mail from", "sender", sender)
			s.send(writer, "250 OK")

		case "RCPT":
			if !strings.HasPrefix(strings.ToUpper(args), "TO:") {
				s.send(writer, "501 Syntax error in parameters")
				continue
			}
			recipient := extractEmailAddress(strings.TrimPrefix(strings.ToUpper(args), "TO:"))
			session.Recipients = append(session.Recipients, recipient)
			sessionLogger.Info("rcpt to", "recipient", recipient)
			s.send(writer, "250 OK")

		case "DATA":
			s.handleDATA(ctx, reader, writer, session, sessionLogger)

		case "QUIT":
			s.send(writer, "221 Bye")
			return

		case "RSET":
			// Reset session
			session.Sender = ""
			session.Recipients = nil
			s.send(writer, "250 OK")

		case "NOOP":
			s.send(writer, "250 OK")

		default:
			sessionLogger.Warn("unknown command", "command", command)
			s.send(writer, "502 Command not implemented")
		}
	}
}

// handleEHLO responds to EHLO/HELO command
func (s *Server) handleEHLO(writer *bufio.Writer, args string, session *domain.SMTPSession, logger *observability.Logger) {
	session.HeloName = args
	// RFC 5321 Section 4.1.1.1: EHLO response
	s.send(writer, "250-%s", s.config.SMTP.Hostname)
	s.send(writer, "250-SIZE %d", s.config.SMTP.MaxSize)
	s.send(writer, "250 8BITMIME")
}

// handleDATA processes the DATA command and message content
func (s *Server) handleDATA(ctx context.Context, reader *bufio.Reader, writer *bufio.Writer, session *domain.SMTPSession, logger *observability.Logger) {
	if session.Sender == "" || len(session.Recipients) == 0 {
		s.send(writer, "503 Bad sequence of commands")
		return
	}

	s.send(writer, "354 End data with <CR><LF>.<CR><LF>")

	// RFC 5321 Section 4.1.1.4: Read message until "."
	var messageData []byte
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			logger.Error("failed to read message data", "error", err)
			s.send(writer, "451 Error reading message")
			return
		}

		// Check for end of message
		if line == ".\r\n" || line == ".\n" {
			break
		}

		// RFC 5321 Section 4.5.2: Transparency (remove leading dot)
		line = strings.TrimPrefix(line, ".")

		messageData = append(messageData, []byte(line)...)
		session.BytesRecv += int64(len(line))

		// Check size limit
		if session.BytesRecv > s.config.SMTP.MaxSize {
			logger.Warn("message too large", "size", session.BytesRecv)
			s.send(writer, "552 Message size exceeds maximum")
			return
		}
	}

	logger.Info("received message", "size", session.BytesRecv)

	// Check Spam Content
	if s.spamFilter != nil {
		headers := map[string]string{
			"IP":       session.RemoteIP,
			"Helo":     session.HeloName,
			"Queue-ID": session.SessionID,
			"From":     session.Sender,
		}
		res, err := s.spamFilter.CheckContent(ctx, bytes.NewReader(messageData), headers)
		if err != nil {
			logger.Error("spam check failed", "error", err)
			// Fail open
		} else {
			if res.Action == domain.SpamActionReject {
				s.metrics.IncrementMessagesRejected()
				logger.Warn("message rejected as spam", "score", res.Score)
				s.send(writer, "554 Message rejected due to spam content")
				return
			} else if res.Action == domain.SpamActionSoftReject {
				s.send(writer, "451 Temporary failure, please try again")
				return
			}

			// Add headers if needed
			if len(res.Headers) > 0 {
				var sb bytes.Buffer
				for k, v := range res.Headers {
					sb.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
				}
				messageData = append(sb.Bytes(), messageData...)
			}
		}
	}

	// Process message through middleware pipeline
	if err := s.handler(session, messageData); err != nil {
		logger.Error("failed to process message", "error", err)
		s.metrics.IncrementMessagesRejected()
		s.send(writer, "451 Temporary failure: %v", err)
		return
	}

	// RFC 5321 Section 4.1.1.4: Success response
	// "250 OK" indicates message has been durably saved
	s.metrics.IncrementMessagesReceived()
	logger.Info("message accepted")
	s.send(writer, "250 OK: Message accepted for delivery")
}

// send writes a formatted response to the client
func (s *Server) send(writer *bufio.Writer, format string, args ...interface{}) {
	fmt.Fprintf(writer, format+"\r\n", args...)
	writer.Flush()
}

// extractEmailAddress extracts email from SMTP address format
func extractEmailAddress(input string) string {
	// Remove angle brackets and whitespace
	input = strings.TrimSpace(input)
	input = strings.Trim(input, "<>")
	return strings.ToLower(input)
}

// Stop gracefully shuts down the SMTP server
func (s *Server) Stop() error {
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

// Addr returns the listener address
func (s *Server) Addr() net.Addr {
	if s.listener != nil {
		return s.listener.Addr()
	}
	return nil
}
