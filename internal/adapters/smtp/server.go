package smtp

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/Kartikey2011yadav/mailraven-server/internal/config"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
	"github.com/Kartikey2011yadav/mailraven-server/internal/observability"
)

// Server represents an SMTP server
type Server struct {
	config  *config.Config
	logger  *observability.Logger
	metrics *observability.Metrics
	handler MessageHandler
	listener net.Listener
}

// NewServer creates a new SMTP server
func NewServer(cfg *config.Config, logger *observability.Logger, metrics *observability.Metrics, handler MessageHandler) *Server {
	return &Server{
		config:  cfg,
		logger:  logger,
		metrics: metrics,
		handler: handler,
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
	remoteIP := strings.Split(remoteAddr, ":")[0]
	
	sessionID := fmt.Sprintf("smtp-%d", time.Now().UnixNano())
	sessionLogger := s.logger.WithSMTPSession(sessionID, remoteIP)

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
		conn.SetReadDeadline(time.Now().Add(5 * time.Minute))

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
			s.handleEHLO(writer, args, sessionLogger)
			
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
func (s *Server) handleEHLO(writer *bufio.Writer, args string, logger *observability.Logger) {
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
		if strings.HasPrefix(line, ".") {
			line = line[1:]
		}

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
