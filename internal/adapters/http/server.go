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
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/services"
	"github.com/Kartikey2011yadav/mailraven-server/internal/observability"
	"github.com/go-chi/chi/v5"
)

// Server represents the HTTP server
type Server struct {
	router      *chi.Mux
	httpServer  *http.Server
	cfg         *config.Config
	logger      *observability.Logger
	acmeService *services.ACMEService
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
	queueRepo ports.QueueRepository,
	domainRepo ports.DomainRepository,
	blobStore ports.BlobStore,
	searchIdx ports.SearchIndex,
	acmeService *services.ACMEService,
	backupService ports.BackupService,
	logger *observability.Logger,
	metrics *observability.Metrics,
) *Server {
	router := chi.NewRouter()

	// Create handlers
	authHandler := handlers.NewAuthHandler(userRepo, cfg.API.JWTSecret, logger, metrics)
	messageHandler := handlers.NewMessageHandler(emailRepo, blobStore, searchIdx, logger, metrics)
	searchHandler := handlers.NewSearchHandler(emailRepo, searchIdx, logger, metrics)
	adminBackupHandler := handlers.NewAdminHandler(backupService, logger, metrics)
	adminUserHandler := handlers.NewAdminUserHandler(userRepo, domainRepo, logger)
	adminDomainHandler := handlers.NewAdminDomainHandler(domainRepo, logger)
	adminStatsHandler := handlers.NewAdminStatsHandler(userRepo, emailRepo, queueRepo, logger)
	sendHandler, err := handlers.NewSendHandler(
		queueRepo,
		blobStore,
		logger,
		metrics,
		cfg.Domain,
		cfg.DKIM.Selector,
		cfg.DKIM.PrivateKeyPath,
	)
	if err != nil {
		// Log error but don't fail server startup if DKIM keys missing?
		// Or fail? Handler returns error if key invalid.
		// If fails, we can't send.
		logger.Error("failed to create send handler (DKIM init failed)", "error", err)
	}

	// Apply global middleware (order matters: first applied = outermost)
	router.Use(middleware.Logging(logger))
	// Allow localhost:5173 for now. In production this should be config driven.
	router.Use(middleware.CORS("http://localhost:5173"))
	router.Use(middleware.Compression())
	router.Use(middleware.RateLimit(100)) // 100 req/min per IP

	// Public routes (no auth required)
	router.Post("/api/v1/auth/login", authHandler.Login)

	// Metrics endpoint
	router.Get("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		metrics.WritePrometheus(w)
	})

	// Protected routes (require JWT auth)
	router.Group(func(r chi.Router) {
		r.Use(middleware.Auth(cfg.API.JWTSecret))

		// Message endpoints
		r.Get("/api/v1/messages", messageHandler.ListMessages)
		r.Get("/api/v1/messages/since", messageHandler.GetMessagesSince)
		r.Get("/api/v1/messages/search", searchHandler.SearchMessages)
		r.Get("/api/v1/messages/{id}", messageHandler.GetMessage)
		r.Patch("/api/v1/messages/{id}", messageHandler.UpdateMessage)
		// Outbound
		if sendHandler != nil {
			r.Post("/api/v1/messages/send", sendHandler.Send)
		}

		// Admin endpoints
		r.Route("/api/v1/admin", func(r chi.Router) {
			r.Use(middleware.RequireAdmin)

			// Stats
			r.Get("/stats", adminStatsHandler.GetSystemStats)

			// Backup
			if backupService != nil {
				r.Post("/backup", adminBackupHandler.TriggerBackup)
			}

			// User Management
			r.Get("/users", adminUserHandler.ListUsers)
			r.Post("/users", adminUserHandler.CreateUser)
			r.Delete("/users/{email}", adminUserHandler.DeleteUser)
			r.Put("/users/{email}/role", adminUserHandler.UpdateRole)

			// Domain Management
			r.Get("/domains", adminDomainHandler.ListDomains)
			r.Post("/domains", adminDomainHandler.CreateDomain)
			r.Delete("/domains/{domain}", adminDomainHandler.DeleteDomain)
		})
	})

	// Health check endpoint (no auth)
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	return &Server{
		router:      router,
		cfg:         cfg,
		logger:      logger,
		acmeService: acmeService,
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
	if s.acmeService != nil {
		go func() {
			s.logger.Info("Starting ACME HTTP-01 challenge listener on :80")
			// Use http.Server with timeouts for security
			srv := &http.Server{
				Addr:         ":80",
				Handler:      s.acmeService.HTTPHandler(nil),
				ReadTimeout:  30 * time.Second,
				WriteTimeout: 30 * time.Second,
				IdleTimeout:  120 * time.Second,
			}
			if err := srv.ListenAndServe(); err != nil {
				s.logger.Error("ACME listener failed", "error", err)
			}
		}()
		s.httpServer.TLSConfig = s.acmeService.TLSConfig()
		// ListenAndServeTLS with empty strings uses certificates from TLSConfig
		return s.httpServer.ListenAndServeTLS("", "")
	}

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
