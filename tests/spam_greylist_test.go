package tests

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/textproto"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/smtp"
	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/spam/greylist"
	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/storage/sqlite"
	"github.com/Kartikey2011yadav/mailraven-server/internal/config"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/services"
	"github.com/Kartikey2011yadav/mailraven-server/internal/observability"
)

// T013: Integration test for Greylisting
func TestGreylisting_SMTP_Interaction(t *testing.T) {
	// 1. Setup Environment (DB)
	env := setupTestEnvironment(t) // Creates temp dir, db, basic logic
	// Note: setupTestEnvironment closes DB/server on cleanup?
	// It doesn't close DB, but we get the connection.

	// 2. Apply Greylist Migration manually (since helpers might not have it yet)
	migrationPath := "../internal/adapters/storage/sqlite/migrations/006_add_greylist.sql"
	// Verify path exists relative to test execution
	absMigPath, _ := filepath.Abs(migrationPath)
	if _, err := filepath.Abs(migrationPath); err != nil {
		t.Logf("Warning: could not verify migration path: %v", err)
	}
	if err := env.conn.RunMigrations(migrationPath); err != nil {
		t.Logf("Migration warning (006): %v. Trying absolute path: %s", err, absMigPath)
		// Try absolute path if relative fails
		if err := env.conn.RunMigrations(absMigPath); err != nil {
			t.Fatalf("Failed to run Greylist migration: %v", err)
		}
	}

	// 3. Initialize Greylist Service
	repo := sqlite.NewGreylistRepository(env.conn.DB)
	greylistCfg := config.GreylistConfig{
		Enabled:    true,
		RetryDelay: "2s", // Short delay for test speed
		Expiration: "1m",
	}

	greylistSvc, err := greylist.NewService(repo, greylistCfg)
	if err != nil {
		t.Fatalf("Failed to init greylist service: %v", err)
	}

	// 4. Initialize Spam Protection Service
	logger := observability.NewLogger("debug", "text")
	spamCfg := config.SpamConfig{
		Enabled:  true,
		Greylist: greylistCfg,
		RateLimit: config.RateLimitConfig{
			Window: "10m",
			Count:  1000,
		},
		// Disable other checks to isolate greylist
		RspamdURL: "",
		DNSBLs:    nil,
	}

	// Using nil for Bayes repo as it's not needed for Greylist check
	spamSvc, err := services.NewSpamProtectionService(spamCfg, logger, greylistSvc, nil)
	if err != nil {
		t.Fatalf("Failed to init spam service: %v", err)
	}

	// 5. Start SMTP Server
	smtpCfg := &config.Config{
		Domain: "greylist.test",
		SMTP: config.SMTPConfig{
			Hostname: "greylist.test",
			Port:     0, // Random port
			MaxSize:  1024 * 1024,
		},
		Spam: spamCfg,
	}

	metrics := observability.NewMetrics()
	// Dummy handler
	handler := func(session *domain.SMTPSession, message []byte) error {
		return nil
	}

	server := smtp.NewServer(smtpCfg, logger, metrics, handler, spamSvc)

	// Context for server
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server in goroutine
	// Server.Start blocks, so separate goroutine
	// Wait for listener to be ready?
	// The Start() method sets s.listener under lock, but we can't easily wait on it without a channel or sleep.
	// We'll trust it starts quickly or retry dial.
	go func() {
		if err := server.Start(ctx); err != nil {
			// t.Logf("Server stopped: %v", err)
		}
	}()

	// Wait for server to bind
	var addr net.Addr
	for i := 0; i < 20; i++ {
		addr = server.Addr()
		if addr != nil {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if addr == nil {
		t.Fatal("Server failed to start (no address)")
	}
	t.Logf("SMTP Server listening on %s", addr.String())

	// 6. Test Scenario
	// Helper to simulate SMTP client
	connectAndSendRCPT := func() string {
		conn, err := net.Dial("tcp", addr.String())
		if err != nil {
			t.Fatalf("Failed to dial: %v", err)
		}
		defer conn.Close()

		reader := textproto.NewReader(bufio.NewReader(conn))

		// Greeting
		line, _ := reader.ReadLine()
		if !strings.HasPrefix(line, "220") {
			t.Fatalf("Expected 220, got %s", line)
		}

		// EHLO
		fmt.Fprintf(conn, "EHLO client.test\r\n")
		// Read EHLO response
		for {
			line, err := reader.ReadLine()
			if err != nil {
				t.Fatalf("EHLO read error: %v", err)
			}
			if strings.HasPrefix(line, "250 ") {
				break
			}
		}

		// MAIL FROM
		fmt.Fprintf(conn, "MAIL FROM:<sender@remote.com>\r\n")
		line, _ = reader.ReadLine()
		if !strings.HasPrefix(line, "250") {
			t.Fatalf("MAIL FROM failed: %s", line)
		}

		// RCPT TO
		fmt.Fprintf(conn, "RCPT TO:<recipient@greylist.test>\r\n")
		line, err = reader.ReadLine()
		if err != nil {
			t.Fatalf("RCPT TO read error: %v", err)
		}
		return line
	}

	// A. First Attempt -> Should be Greylisted (451)
	resp := connectAndSendRCPT()
	if !strings.HasPrefix(resp, "451") {
		t.Errorf("Expected 451 Greylisted on first attempt, got: %s", resp)
	} else {
		t.Logf("First attempt correctly rejected: %s", resp)
	}

	// B. Immediate Retry -> Should be Greylisted (451)
	resp = connectAndSendRCPT()
	if !strings.HasPrefix(resp, "451") {
		t.Errorf("Expected 451 Greylisted on immediate retry, got: %s", resp)
	} else {
		t.Logf("Immediate retry correctly rejected: %s", resp)
	}

	// C. Wait for RetryDelay
	t.Log("Waiting for retry delay (2.1s)...")
	time.Sleep(2100 * time.Millisecond)

	// D. Post-Delay Attempt -> Should be Allowed (250)
	resp = connectAndSendRCPT()
	if !strings.HasPrefix(resp, "250") {
		t.Errorf("Expected 250 OK after delay, got: %s", resp)
	} else {
		t.Logf("Retry after delay correctly accepted: %s", resp)
	}
}
