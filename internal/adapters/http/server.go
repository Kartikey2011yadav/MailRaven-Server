package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/http/handlers"
	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/http/middleware"
	"github.com/Kartikey2011yadav/mailraven-server/internal/config"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/ports"
	"github.com/Kartikey2011yadav/mailraven-server/internal/observability"
	"github.com/go-chi/chi/v5"
)

// Server represents the HTTP server
type Server struct {
	router     *chi.Mux
	httpServer *http.Server
	cfg        *config.Config
	logger     *observability.Logger
}

// Router returns the chi router (for testing)
func (s *Server) Router() *chi.Mux {
	return s.router
}

// NewServer creates a new HTTP server
func NewServer(
	cfg *config.Config,
	emailRepo ports.EmailRepository,
	userRepo ports.UserRepository,
	blobStore ports.BlobStore,
	searchIdx ports.SearchIndex,
	logger *observability.Logger,
	metrics *observability.Metrics,
) *Server {
	router := chi.NewRouter()

	// Create handlers
	authHandler := handlers.NewAuthHandler(userRepo, cfg.API.JWTSecret, logger, metrics)
	messageHandler := handlers.NewMessageHandler(emailRepo, blobStore, searchIdx, logger, metrics)
	searchHandler := handlers.NewSearchHandler(emailRepo, searchIdx, logger, metrics)

	// Apply global middleware (order matters: first applied = outermost)
	router.Use(middleware.Logging(logger))
	router.Use(middleware.Compression())
	router.Use(middleware.RateLimit(100)) // 100 req/min per IP

	// Public routes (no auth required)
	router.Post("/api/v1/auth/login", authHandler.Login)

	// Protected routes (require JWT auth)
	router.Group(func(r chi.Router) {
		r.Use(middleware.Auth(cfg.API.JWTSecret))

		// Message endpoints
		r.Get("/api/v1/messages", messageHandler.ListMessages)
		r.Get("/api/v1/messages/since", messageHandler.GetMessagesSince)
		r.Get("/api/v1/messages/search", searchHandler.SearchMessages)
		r.Get("/api/v1/messages/{id}", messageHandler.GetMessage)
		r.Patch("/api/v1/messages/{id}", messageHandler.UpdateMessage)
	})

	// Health check endpoint (no auth)
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	return &Server{
		router: router,
		cfg:    cfg,
		logger: logger,
	}
}

// Start starts the HTTP server
func (s *Server) Start(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", s.cfg.API.Host, s.cfg.API.Port)

	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	s.logger.Info("HTTP server starting", "addr", addr, "tls", s.cfg.API.TLS)

	// Start server (TLS or plain HTTP)
	if s.cfg.API.TLS {
		if s.cfg.API.TLSCert == "" || s.cfg.API.TLSKey == "" {
			return fmt.Errorf("TLS enabled but cert/key paths not configured")
		}
		return s.httpServer.ListenAndServeTLS(s.cfg.API.TLSCert, s.cfg.API.TLSKey)
	}

	return s.httpServer.ListenAndServe()
}

// Stop gracefully stops the HTTP server
func (s *Server) Stop(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}

	s.logger.Info("HTTP server stopping")
	return s.httpServer.Shutdown(ctx)
}
