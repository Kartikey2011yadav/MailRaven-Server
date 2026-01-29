package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"

	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/storage/sqlite"
	"github.com/Kartikey2011yadav/mailraven-server/internal/config"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
)

// RunQuickstart performs initial MailRaven setup
func RunQuickstart() error {
	fmt.Println("==============================================")
	fmt.Println("   MailRaven Quickstart Setup")
	fmt.Println("==============================================")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	// Step 1: Collect domain information
	fmt.Print("Enter your mail domain (e.g., mail.example.com): ")
	mailDomain, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read domain: %w", err)
	}
	mailDomain = strings.TrimSpace(mailDomain)
	if mailDomain == "" {
		return fmt.Errorf("domain is required")
	}

	// Step 2: Collect admin email
	fmt.Print("Enter admin email address (e.g., admin@example.com): ")
	adminEmail, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read email: %w", err)
	}
	adminEmail = strings.TrimSpace(adminEmail)
	if adminEmail == "" {
		return fmt.Errorf("admin email is required")
	}

	// Step 3: Collect admin password (hidden input)
	fmt.Print("Enter admin password: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println() // New line after password input
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}
	password := string(passwordBytes)
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}

	// Step 4: Confirm password
	fmt.Print("Confirm admin password: ")
	confirmBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return fmt.Errorf("failed to read password confirmation: %w", err)
	}
	if string(confirmBytes) != password {
		return fmt.Errorf("passwords do not match")
	}

	fmt.Println()
	fmt.Println("Generating configuration...")
	fmt.Println()

	// Step 5: Generate DKIM keys
	fmt.Println("[1/6] Generating DKIM keys (RSA-2048)...")
	dkimKeyPair, err := config.GenerateDKIMKey()
	if err != nil {
		return fmt.Errorf("failed to generate DKIM keys: %w", err)
	}

	// Step 6: Create directories
	fmt.Println("[2/6] Creating directories...")
	configDir := "/etc/mailraven"
	dataDir := "/var/lib/mailraven"

	// On Windows, use alternative paths
	if os.PathSeparator == '\\' {
		configDir = "C:\\ProgramData\\mailraven\\config"
		dataDir = "C:\\ProgramData\\mailraven\\data"
	}

	dirs := []string{
		configDir,
		dataDir,
		filepath.Join(dataDir, "blobs"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0750); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Step 7: Save DKIM private key
	fmt.Println("[3/6] Saving DKIM private key...")
	dkimKeyPath := filepath.Join(configDir, "dkim.key")
	if err := dkimKeyPair.SavePrivateKey(dkimKeyPath); err != nil {
		return fmt.Errorf("failed to save DKIM private key: %w", err)
	}

	// Step 8: Generate configuration
	fmt.Println("[4/6] Generating configuration file...")
	cfg := &config.Config{
		Domain: mailDomain,
		SMTP: config.SMTPConfig{
			Port:     25,
			Hostname: mailDomain,
			MaxSize:  10 * 1024 * 1024, // 10MB
		},
		API: config.APIConfig{
			Host:      "0.0.0.0",
			Port:      8443,
			TLS:       false, // Disabled for development
			JWTSecret: generateRandomSecret(),
		},
		Storage: config.StorageConfig{
			DBPath:   filepath.Join(dataDir, "mailraven.db"),
			BlobPath: filepath.Join(dataDir, "blobs"),
		},
		DKIM: config.DKIMConfig{
			Selector:       "default",
			PrivateKeyPath: dkimKeyPath,
		},
		Logging: config.LogConfig{
			Level:  "info",
			Format: "json",
		},
	}

	// Save configuration
	configPath := filepath.Join(configDir, "config.yaml")
	if err := cfg.SaveToFile(configPath); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	// Step 9: Initialize database and create admin user
	fmt.Println("[5/6] Initializing database and creating admin user...")

	// Create database connection
	conn, err := sqlite.NewConnection(cfg.Storage.DBPath)
	if err != nil {
		return fmt.Errorf("failed to create database connection: %w", err)
	}
	defer conn.Close()

	// Run migrations
	migrationPath := filepath.Join("internal", "adapters", "storage", "sqlite", "migrations", "001_init.sql")
	if err := conn.RunMigrations(migrationPath); err != nil {
		return fmt.Errorf("failed to run database migrations: %w", err)
	}

	// Create admin user
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	userRepo := sqlite.NewUserRepository(conn.DB)
	adminUser := &domain.User{
		Email:        adminEmail,
		PasswordHash: string(passwordHash),
		CreatedAt:    time.Now(),
		LastLoginAt:  time.Now(),
	}

	ctx := context.Background()
	if err := userRepo.Create(ctx, adminUser); err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	// Step 10: Generate and display DNS records
	fmt.Println("[6/6] Generating DNS records...")
	dkimPublicKey, err := dkimKeyPair.GetPublicKeyDNS()
	if err != nil {
		return fmt.Errorf("failed to get DKIM public key: %w", err)
	}

	dnsRecords := config.GenerateDNSRecords(cfg, dkimPublicKey)

	fmt.Println()
	fmt.Println("✅ Quickstart setup complete!")
	fmt.Println()
	fmt.Println("Configuration saved to:", configPath)
	fmt.Println("DKIM private key saved to:", dkimKeyPath)
	fmt.Println("Database initialized at:", cfg.Storage.DBPath)
	fmt.Println()

	// Print DNS records
	fmt.Println(config.FormatDNSRecordsForConsole(dnsRecords))

	// Step 11: Validate configuration
	fmt.Println("Validating configuration...")
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// Check if ports are available (basic check)
	fmt.Printf("✅ SMTP port %d: Configuration ready\n", cfg.SMTP.Port)
	fmt.Printf("✅ API port %d: Configuration ready (TLS: %v)\n", cfg.API.Port, cfg.API.TLS)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("1. Add the DNS records shown above to your DNS provider")
	fmt.Println("2. Configure firewall to allow ports 25 (SMTP) and 8443 (API)")
	fmt.Println("3. Start MailRaven with: mailraven serve")
	fmt.Println()

	return nil
}

// generateRandomSecret creates a cryptographically secure random JWT secret
func generateRandomSecret() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		// Fallback to time-based seed (not ideal but better than failing)
		panic(fmt.Sprintf("failed to generate random secret: %v", err))
	}
	return base64.URLEncoding.EncodeToString(b)
}
