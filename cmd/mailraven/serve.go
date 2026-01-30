package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"database/sql"

	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/backup"
	httpAdapter "github.com/Kartikey2011yadav/mailraven-server/internal/adapters/http"
	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/imap"
	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/smtp"
	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/storage/disk"
	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/storage/postgres"
	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/storage/sqlite"
	"github.com/Kartikey2011yadav/mailraven-server/internal/config"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/ports"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/services"
	"github.com/Kartikey2011yadav/mailraven-server/internal/observability"
)

// RunServe starts the SMTP and API servers
func RunServe() error {
	// Load configuration
	configPath := "/etc/mailraven/config.yaml"
	if os.PathSeparator == '\\' {
		configPath = "C:\\ProgramData\\mailraven\\config\\config.yaml"
	}

	// Check for custom config path
	if len(os.Args) > 2 && os.Args[2] == "--config" && len(os.Args) > 3 {
		configPath = os.Args[3]
	}

	fmt.Printf("Loading configuration from: %s\n", configPath)
	cfg, err := config.LoadFromFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Initialize logger
	logger := observability.NewLogger(cfg.Logging.Level, cfg.Logging.Format)
	logger.Info("MailRaven starting", "version", "0.1.0-alpha")

	// Initialize metrics
	metrics := observability.NewMetrics()

	// Initialize repositories
	var (
		dbConn     *sql.DB
		emailRepo  ports.EmailRepository
		userRepo   ports.UserRepository
		domainRepo ports.DomainRepository
		queueRepo  ports.QueueRepository
		searchIdx  ports.SearchIndex
		dbBackup   ports.DatabaseBackup
	)

	if cfg.Storage.Driver == "postgres" {
		logger.Info("connecting to postgres database")
		if cfg.Storage.DSN == "" {
			return fmt.Errorf("postgres driver selected but dsn is empty")
		}
		conn, err := postgres.NewConnection(cfg.Storage.DSN)
		if err != nil {
			return fmt.Errorf("failed to connect to postgres: %w", err)
		}
		defer conn.Close()
		dbConn = conn.DB

		logger.Info("running postgres migrations")
		if err := conn.RunMigrations(); err != nil {
			logger.Warn("migration warning", "error", err)
		}

		emailRepo = postgres.NewEmailRepository(conn.DB)
		userRepo = postgres.NewUserRepository(conn.DB)
		domainRepo = postgres.NewDomainRepository(conn.DB)
		queueRepo = postgres.NewQueueRepository(conn.DB)
		searchIdx = postgres.NewSearchRepository(conn.DB)
		dbBackup = backup.NewPostgresBackup(cfg.Storage.DSN)

	} else {
		// Initialize database connection
		logger.Info("connecting to database", "path", cfg.Storage.DBPath)
		conn, err := sqlite.NewConnection(cfg.Storage.DBPath)
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		defer conn.Close()

		// Check database integrity
		logger.Info("checking database integrity")
		if err := conn.CheckIntegrity(); err != nil {
			return fmt.Errorf("database integrity check failed: %w", err)
		}

		// Run migrations
		migrationsDir := filepath.Join("internal", "adapters", "storage", "sqlite", "migrations")
		files, err := os.ReadDir(migrationsDir)
		if err == nil {
			logger.Info("running database migrations")
			for _, file := range files {
				if filepath.Ext(file.Name()) == ".sql" {
					migrationPath := filepath.Join(migrationsDir, file.Name())
					logger.Info("applying migration", "file", file.Name())
					if err := conn.RunMigrations(migrationPath); err != nil {
						logger.Warn("migration warning", "file", file.Name(), "error", err)
					}
				}
			}
		}

		dbConn = conn.DB
		logger.Info("initializing storage adapters")
		emailRepo = sqlite.NewEmailRepository(conn.DB)
		userRepo = sqlite.NewUserRepository(conn.DB)
		domainRepo = sqlite.NewDomainRepository(conn.DB)
		queueRepo = sqlite.NewQueueRepository(conn.DB)
		searchIdx = sqlite.NewSearchRepository(conn.DB)
		dbBackup = backup.NewSQLiteBackup(conn.DB)
	}

	// Initialize blob store
	blobStore, err := disk.NewBlobStore(cfg.Storage.BlobPath)
	if err != nil {
		return fmt.Errorf("failed to initialize blob store: %w", err)
	}

	// Initialize SMTP handler
	smtpHandler := smtp.NewHandler(emailRepo, blobStore, searchIdx, dbConn, logger, metrics)
	messageHandler := smtpHandler.BuildMiddlewarePipeline()

	// Initialize Spam Protection
	spamService, err := services.NewSpamProtectionService(cfg.Spam, logger)
	if err != nil {
		logger.Warn("Failed to initialize spam protection", "error", err)
	}

	// Initialize SMTP server
	smtpServer := smtp.NewServer(cfg, logger, metrics, messageHandler, spamService)

	// Initialize ACME service
	acmeService, err := services.NewACMEService(cfg.TLS.ACME)
	if err != nil {
		return fmt.Errorf("failed to initialize ACME service: %w", err)
	}

	// Initialize Backup Service
	blobBackup := backup.NewBlobBackup(cfg.Storage.BlobPath)
	backupService := services.NewBackupService(cfg.Backup, dbBackup, blobBackup, logger)

	// Initialize HTTP server
	httpServer := httpAdapter.NewServer(cfg, emailRepo, userRepo, queueRepo, domainRepo, blobStore, searchIdx, acmeService, backupService, logger, metrics)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("received shutdown signal")

		// Stop HTTP server first
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()
		if err := httpServer.Stop(shutdownCtx); err != nil {
			logger.Error("HTTP server shutdown error", "error", err)
		}

		cancel() // Stop SMTP server
	}()

	// Start HTTP server in background
	go func() {
		logger.Info("starting HTTP server", "host", cfg.API.Host, "port", cfg.API.Port, "tls", cfg.API.TLS)
		if err := httpServer.Start(ctx); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server error", "error", err)
		}
	}()

	// Start IMAP server in background (if enabled)
	if cfg.IMAP.Enabled {
		go func() {
			imapServer := imap.NewServer(cfg.IMAP, logger, userRepo, emailRepo)
			logger.Info("starting IMAP server", "port", cfg.IMAP.Port)
			if err := imapServer.Start(ctx); err != nil {
				logger.Error("IMAP server error", "error", err)
			}
		}()
	}

	// Start SMTP server (blocking)
	logger.Info("starting SMTP server", "port", cfg.SMTP.Port)
	fmt.Printf("\n")
	fmt.Printf("==============================================\n")
	fmt.Printf("   MailRaven Server Running\n")
	fmt.Printf("==============================================\n")
	fmt.Printf("SMTP Port:   %d\n", cfg.SMTP.Port)
	fmt.Printf("HTTP Port:   %d (TLS: %v)\n", cfg.API.Port, cfg.API.TLS)
	if cfg.IMAP.Enabled {
		fmt.Printf("IMAP Port:   %d\n", cfg.IMAP.Port)
	}
	fmt.Printf("Domain:      %s\n", cfg.Domain)
	fmt.Printf("Log Level:   %s\n", cfg.Logging.Level)
	fmt.Printf("\n")
	fmt.Printf("Press Ctrl+C to stop\n")
	fmt.Printf("\n")

	if err := smtpServer.Start(ctx); err != nil {
		return fmt.Errorf("SMTP server error: %w", err)
	}

	logger.Info("MailRaven stopped")
	return nil
}

// RunCheckConfig validates the configuration file
func RunCheckConfig() error {
	// Load configuration
	configPath := "/etc/mailraven/config.yaml"
	if os.PathSeparator == '\\' {
		configPath = "C:\\ProgramData\\mailraven\\config\\config.yaml"
	}

	// Check for custom config path
	// Args[0] is program, [1] command, [2] flag, [3] value
	if len(os.Args) > 2 && os.Args[2] == "--config" && len(os.Args) > 3 {
		configPath = os.Args[3]
	}

	fmt.Printf("Checking configuration at: %s\n", configPath)
	cfg, err := config.LoadFromFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Check DB connection validity?
	if cfg.Storage.Driver == "postgres" && cfg.Storage.DSN == "" {
		return fmt.Errorf("postgres driver selected but DSN is empty")
	}

	return nil
}
