package imap

import (
	"context"
	"fmt"
	"net"

	"github.com/Kartikey2011yadav/mailraven-server/internal/config"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/ports"
	"github.com/Kartikey2011yadav/mailraven-server/internal/observability"
)

const maxIMAPConnections = 500

type Server struct {
	config          config.IMAPConfig
	logger          *observability.Logger
	metrics         *observability.Metrics
	userRepo        ports.UserRepository
	emailRepo       ports.EmailRepository
	spamService     ports.SpamFilter
	blobStore       ports.BlobStore
	notificationBus ports.NotificationBus
	listener        net.Listener
	connSem         chan struct{}
}

func NewServer(cfg config.IMAPConfig, logger *observability.Logger, metrics *observability.Metrics, userRepo ports.UserRepository, emailRepo ports.EmailRepository, spamService ports.SpamFilter, blobStore ports.BlobStore, notificationBus ports.NotificationBus) *Server {
	return &Server{
		config:          cfg,
		logger:          logger,
		metrics:         metrics,
		userRepo:        userRepo,
		emailRepo:       emailRepo,
		spamService:     spamService,
		blobStore:       blobStore,
		notificationBus: notificationBus,
		connSem:         make(chan struct{}, maxIMAPConnections),
	}
}

func (s *Server) Start(ctx context.Context) error {
	addr := fmt.Sprintf(":%d", s.config.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("IMAP listen failed: %w", err)
	}
	s.listener = listener
	s.logger.Info("IMAP server started", "addr", addr)

	go func() {
		<-ctx.Done()
		s.listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			s.logger.Error("IMAP accept failed", "error", err)
			continue
		}

		select {
		case s.connSem <- struct{}{}:
			go func() {
				defer func() { <-s.connSem }()
				s.handleConnection(ctx, conn)
			}()
		default:
			s.logger.Warn("IMAP connection limit reached", "remote", conn.RemoteAddr())
			_, _ = fmt.Fprintf(conn, "* BYE Too many connections\r\n")
			_ = conn.Close()
		}
	}
}

// Addr returns the listener address
func (s *Server) Addr() net.Addr {
	if s.listener != nil {
		return s.listener.Addr()
	}
	return nil
}

func (s *Server) handleConnection(ctx context.Context, conn net.Conn) {
	if s.metrics != nil {
		s.metrics.IncrementActiveIMAP()
		defer s.metrics.DecrementActiveIMAP()
	}
	session := NewSession(conn, s.config, s.logger, s.userRepo, s.emailRepo, s.spamService, s.blobStore, s.notificationBus)
	session.Serve()
}
