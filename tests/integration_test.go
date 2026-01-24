package tests

import (
	"context"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/http/dto"
	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/smtp"
	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/storage/disk"
	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/storage/sqlite"
	"github.com/Kartikey2011yadav/mailraven-server/internal/config"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
	"github.com/Kartikey2011yadav/mailraven-server/internal/observability"
)

// T104: Integration test suite runner
// TestEndToEnd_Loopback performs a full cycle:
// 1. Send email via API (US5)
// 2. Queue & Delivery Worker picks it up
// 3. Delivers to local SMTP Server (US2) via loopback
// 4. SMTP Server stores it
// 5. API lists received message (US3)
func TestEndToEnd_Loopback(t *testing.T) {
	// 1. Setup Environment (DB, Repos, HTTP Server)
	env := setupTestEnvironment(t)
	defer env.cleanup()

	logger := observability.NewLogger("error", "text")
	metrics := observability.NewMetrics()

	// Re-create dependencies not exposed by env
	blobStore, err := disk.NewBlobStore(env.tempDir + "/blobs")
	if err != nil {
		t.Fatalf("Failed to open blob store: %v", err)
	}
	searchIdx := sqlite.NewSearchRepository(env.conn.DB)
	queueRepo := sqlite.NewQueueRepository(env.conn.DB)

	// 2. Start SMTP Server (US2) on random port
	smtpCfg := &config.Config{
		Domain: "loopback.local",
		SMTP: config.SMTPConfig{
			Port:    0, // Random port
			MaxSize: 10 * 1024 * 1024,
		},
	}

	smtpHandler := smtp.NewHandler(env.emailRepo, blobStore, searchIdx, env.conn.DB, logger, metrics)
	messageHandler := smtpHandler.BuildMiddlewarePipeline()
	smtpServer := smtp.NewServer(smtpCfg, logger, metrics, messageHandler)

	// Start server in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := smtpServer.Start(ctx); err != nil {
			// Don't fail test from goroutine, just log
			logger.Error("SMTP server stopped", "error", err)
		}
	}()
	defer smtpServer.Stop()

	// Wait for server to start and retrieve port
	var addr net.Addr
	for i := 0; i < 20; i++ {
		addr = smtpServer.Addr()
		if addr != nil {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	if addr == nil {
		t.Fatal("SMTP server failed to start (addr is nil)")
	}
	_, portStr, _ := net.SplitHostPort(addr.String())
	t.Logf("SMTP Server listening on port %s", portStr)

	// 3. Setup Delivery Worker (US5)
	// Create SMTP Client pointing to our server
	client := smtp.NewClient(logger)
	client.Port = portStr // Force client to connect to our test port

	// Mock MX lookup to direct "loopback.local" to localhost
	client.LookupMX = func(name string) ([]*net.MX, error) {
		return []*net.MX{{Host: "127.0.0.1", Pref: 10}}, nil
	}

	worker := smtp.NewDeliveryWorker(queueRepo, blobStore, client, logger, metrics)
	worker.Start()
	defer worker.Stop()

	// 4. Send Email via API (US5)
	// Login first to get token
	token := env.authenticateUser(t, "test@example.com", "testpassword123")

	sendReq := dto.SendRequest{
		From:    "test@example.com",
		To:      "recipient@loopback.local",
		Subject: "End-to-End Test Message",
		Body:    "This message traveled through the entire system!",
	}
	body := env.encodeJSON(t, sendReq)
	req := env.newRequest(t, "POST", "/api/v1/messages/send", body, token)
	req.Header.Set("Content-Type", "application/json")

	resp := env.doRequest(t, req)
	if resp.StatusCode != http.StatusAccepted {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("Send failed: %d %s", resp.StatusCode, string(respBody))
	}

	// 5. Verify Delivery (US3)
	// Poll API until message appears in INBOX (stored by SMTP server)
	// Note: The Recipient is "recipient@loopback.local".
	// The API lists messages for logged-in user "test@example.com".
	// Wait! We sent TO "recipient@loopback.local".
	// The SMTP server receives it. It stores it in EmailRepository.
	// But it belongs to "recipient@loopback.local".
	// To see it in API, we need to login as "recipient@loopback.local".

	// Create the recipient user so they can login
	recipientUser := &domain.User{
		Email:        "recipient@loopback.local",
		PasswordHash: "$2a$10$......................................................", // dummy hash
		CreatedAt:    time.Now(),
	}
	// We need to create this user in DB.
	// But `env.authenticateUser` does LOGIN. It needs valid password.
	// Let's create the user properly.

	// We need bcrypt hash for "password"
	// helpers.go has code for this, but not exposed well.
	// We can't access `bcrypt` from here easily without import.
	// Or we can just create user with SAME hash as test user (testpassword123).
	// We can use `env.userRepo.FindByEmail` to get test user and copy hash?
	// Or use helper `setupTestEnvironment` created user.

	testUser, _ := env.userRepo.FindByEmail(context.Background(), "test@example.com")
	recipientUser.PasswordHash = testUser.PasswordHash
	if err := env.userRepo.Create(context.Background(), recipientUser); err != nil {
		t.Fatalf("Failed to create recipient user: %v", err)
	}

	recipientToken := env.authenticateUser(t, "recipient@loopback.local", "testpassword123")

	// Poll loop
	deadline := time.Now().Add(15 * time.Second)
	found := false
	for time.Now().Before(deadline) {
		req := env.newRequest(t, "GET", "/api/v1/messages", nil, recipientToken)
		resp := env.doRequest(t, req)

		var listResp dto.MessageListResponse
		env.decodeJSON(t, resp.Body, &listResp)

		for _, msg := range listResp.Messages {
			if msg.Subject == "End-to-End Test Message" {
				found = true
				break
			}
		}
		if found {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	if !found {
		t.Fatal("Message was not received by recipient within timeout")
	}

	t.Log("End-to-End Test Passed!")
}
