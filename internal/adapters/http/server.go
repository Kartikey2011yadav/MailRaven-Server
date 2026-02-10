package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"strings"

	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/http/handlers"
	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/http/middleware"
	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/http/static"
	"github.com/Kartikey2011yadav/mailraven-server/internal/config"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/ports"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/services"
	"github.com/Kartikey2011yadav/mailraven-server/internal/observability"
	"github.com/go-chi/chi/v5"
)

// Server represents the HTTP server
type Server struct {
	router        *chi.Mux
	httpServer    *http.Server
	cfg           *config.Config
	logger        *observability.Logger
	acmeService   *services.ACMEService
	mtaStsHandler *handlers.MTASTSHandler
}

// Router returns the chi router (for testing)
func (s *Server) Router() *chi.Mux {
	return s.router
}

// ServeHTTP implements http.Handler with host-based routing interceptor
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Host-based routing for MTA-STS
	if strings.HasPrefix(r.Host, "mta-sts.") && s.mtaStsHandler != nil {
		s.mtaStsHandler.ServePolicy(w, r)
		return
	}
	s.router.ServeHTTP(w, r)
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
	// Add new repo
	tlsRptRepo ports.TLSRptRepository,
	sieveRepo ports.ScriptRepository,
	updateManager ports.UpdateManager,
	spamFilter ports.SpamFilter,
	logger *observability.Logger,
	metrics *observability.Metrics,
) *Server {
	router := chi.NewRouter()

	// Create handlers
	authHandler := handlers.NewAuthHandler(userRepo, cfg.API.JWTSecret, logger, metrics)
	messageHandler := handlers.NewMessageHandler(emailRepo, blobStore, searchIdx, spamFilter, logger, metrics)
	searchHandler := handlers.NewSearchHandler(emailRepo, searchIdx, logger, metrics)
	adminBackupHandler := handlers.NewAdminHandler(backupService, logger, metrics)
	adminUserHandler := handlers.NewAdminUserHandler(userRepo, domainRepo, logger)
	adminDomainHandler := handlers.NewAdminDomainHandler(domainRepo, logger)
	adminStatsHandler := handlers.NewAdminStatsHandler(userRepo, emailRepo, queueRepo, logger)
	adminSystemHandler := handlers.NewSystemHandler(updateManager, logger)
	tlsRptHandler := handlers.NewTLSRptHandler(tlsRptRepo, logger)
	sieveHandler := handlers.NewSieveHandler(sieveRepo, logger)
	userSelfHandler := handlers.NewUserSelfHandler(userRepo, logger)

	// Create EmailService
	emailService := services.NewEmailService(emailRepo)
	mailboxHandler := handlers.NewMailboxHandler(emailService, logger)

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

	// Create Autodiscover handler
	autodiscoverHandler := handlers.NewAutodiscoverHandler(cfg, logger)

	// Create MTA-STS Handler
	// Using hardcoded config derived from system config for now
	mtaStsPolicy := &domain.MTASTSPolicy{
		Version: "STSv1",
		Mode:    domain.MTASTSModeEnforce,
		MX:      []string{cfg.Domain}, // Assume MX is same as domain for simple setup
		MaxAge:  86400,
	}
	mtaStsHandler := handlers.NewMTASTSHandler(mtaStsPolicy)

	// Public routes (no auth required)
	router.Post("/api/v1/auth/login", authHandler.Login)

	// Autodiscover endpoints
	router.Get("/.well-known/autoconfig/mail/config-v1.1.xml", autodiscoverHandler.HandleMozillaAutoconfig)
	router.Post("/autodiscover/autodiscover.xml", autodiscoverHandler.HandleMicrosoftAutodiscover)

	// TLS-RPT Endpoint
	router.Post("/.well-known/tlsrpt", tlsRptHandler.HandleReport)

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
		r.Post("/api/v1/messages/{id}/spam", messageHandler.ReportSpam)
		r.Post("/api/v1/messages/{id}/ham", messageHandler.ReportHam)
		// Outbound
		if sendHandler != nil {
			r.Post("/api/v1/messages/send", sendHandler.Send)
		}

		// User Self-Management
		r.Put("/api/v1/users/self/password", userSelfHandler.ChangePassword)

		// Sieve Scripts
		r.Route("/api/v1/sieve/scripts", func(r chi.Router) {
			r.Get("/", sieveHandler.ListScripts)
			r.Post("/", sieveHandler.CreateScript)
			r.Get("/{name}", sieveHandler.GetScript)
			r.Delete("/{name}", sieveHandler.DeleteScript)
			r.Put("/{name}/active", sieveHandler.ActivateScript)
		})

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
			r.Put("/users/{email}/quota", adminUserHandler.UpdateQuota)
			// ACL Management
			r.Put("/users/{userID}/mailboxes/{mailboxName}/acl", mailboxHandler.UpdateACL)

			// Domain Management
			r.Get("/domains", adminDomainHandler.ListDomains)
			r.Post("/domains", adminDomainHandler.CreateDomain)
			r.Delete("/domains/{domain}", adminDomainHandler.DeleteDomain)

			// System Management (Updates)
			if updateManager != nil {
				r.Get("/system/update", adminSystemHandler.CheckUpdate)
				r.Post("/system/update", adminSystemHandler.PerformUpdate)
			}
		})
	})

	// Health check endpoint (no auth)
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		//nolint:errcheck // Health check simple write
		_, _ = w.Write([]byte("OK"))
	})

	// Mount Static Assets (SPA)
	if fs, err := static.GetFS(); err == nil {
		router.Handle("/*", static.Handler(fs))
	} else {
		logger.Warn("Failed to load static assets (SPA)", "error", err)
	}

	return &Server{
		router:        router,
		cfg:           cfg,
		logger:        logger,
		acmeService:   acmeService,
		mtaStsHandler: mtaStsHandler,
	}
}

// Start starts the HTTP server
func (s *Server) Start(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", s.cfg.API.Host, s.cfg.API.Port)

	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      s, // Use s (ServeHTTP) instead of s.router directly
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
