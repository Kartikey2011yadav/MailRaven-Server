package managesieve

import (
	"context"
	"crypto/tls"
	"net"
	"time"

	"github.com/Kartikey2011yadav/mailraven-server/internal/core/ports"
	"github.com/Kartikey2011yadav/mailraven-server/internal/observability"
)

// Server represents the ManageSieve TCP server
type Server struct {
	addr      string
	tlsConfig *tls.Config
	repo      ports.ScriptRepository
	userRepo  ports.UserRepository
	logger    *observability.Logger
	listener  net.Listener
	shutdown  chan struct{}
}

// NewServer creates a new ManageSieve server
func NewServer(
	addr string,
	tlsConfig *tls.Config,
	repo ports.ScriptRepository,
	userRepo ports.UserRepository,
	logger *observability.Logger,
) *Server {
	return &Server{
		addr:      addr,
		tlsConfig: tlsConfig,
		repo:      repo,
		userRepo:  userRepo,
		logger:    logger,
		shutdown:  make(chan struct{}),
	}
}

// Start starts the TCP listener
func (s *Server) Start() error {
	l, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	s.listener = l
	s.logger.Info("ManageSieve server started", "addr", s.addr)

	go s.serve()
	return nil
}

func (s *Server) serve() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.shutdown:
				return
			default:
				s.logger.Error("ManageSieve accept error", "error", err)
				time.Sleep(100 * time.Millisecond)
				continue
			}
		}
		go s.handleConn(conn)
	}
}

// Stop stops the server
func (s *Server) Stop(ctx context.Context) error {
	close(s.shutdown)
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

func (s *Server) handleConn(c net.Conn) {
	session := NewSession(c, s.repo, s.userRepo, s.logger, s.tlsConfig)
	session.Serve()
}
