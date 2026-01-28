package imap

import (
	"context"
	"fmt"
	"net"

	"github.com/Kartikey2011yadav/mailraven-server/internal/config"
	"github.com/Kartikey2011yadav/mailraven-server/internal/observability"
)

type Server struct {
	config  config.IMAPConfig
	logger  *observability.Logger
	listener net.Listener
}

func NewServer(cfg config.IMAPConfig, logger *observability.Logger) *Server {
	return &Server{
		config: cfg,
		logger: logger,
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
		
		go s.handleConnection(ctx, conn)
	}
}

func (s *Server) handleConnection(ctx context.Context, conn net.Conn) {
	session := NewSession(conn, s.config, s.logger)
	session.Serve()
}
