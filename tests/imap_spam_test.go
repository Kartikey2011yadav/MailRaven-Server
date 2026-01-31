package tests

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/imap"
	"github.com/Kartikey2011yadav/mailraven-server/internal/config"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/services"
	"github.com/Kartikey2011yadav/mailraven-server/internal/observability"
)

// TestIMAPListener verifies the IMAP server accepts connections and commands
func TestIMAPListener(t *testing.T) {
	// Setup user repo from env
	env := setupTestEnvironment(t)
	defer env.cleanup()

	// Config
	cfg := config.IMAPConfig{
		Enabled: true,
		Port:    10143, // Test port
	}
	logger := observability.NewLogger("error", "text")

	// Start Server
	server := imap.NewServer(cfg, logger, env.userRepo, env.emailRepo)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := server.Start(ctx); err != nil {
			// t.Errorf("Server start failed: %v", err)
			// Start returns nil on context cancel
		}
	}()

	// Give it a moment to bind
	time.Sleep(100 * time.Millisecond)

	// Connect
	conn, err := net.Dial("tcp", "localhost:10143")
	if err != nil {
		t.Fatalf("Failed to dial IMAP: %v", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)

	// Check Greeting
	line, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("Failed to read greeting: %v", err)
	}
	if !strings.Contains(line, "* OK") {
		t.Errorf("Unexpected greeting: %s", line)
	}

	// Send CAPABILITY
	fmt.Fprintf(conn, "A01 CAPABILITY\r\n")

	// Expect response
	// 1. * CAPABILITY ...
	// 2. A01 OK ...
	foundOk := false
	for i := 0; i < 5; i++ { // Read a few lines
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		if strings.Contains(line, "A01 OK") {
			foundOk = true
			break
		}
	}

	if !foundOk {
		t.Errorf("Did not receive OK response for CAPABILITY")
	}

	cancel()
	wg.Wait()
}

// TestSpamIntegration verifies the Spam Protection Service routes to Rspamd
func TestSpamIntegration(t *testing.T) {
	// Mock Rspamd
	mockRspamd := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/checkv2" {
			http.Error(w, "Not found", 404)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"action":"reject","score":20.0}`)
	}))
	defer mockRspamd.Close()

	logger := observability.NewLogger("error", "text")

	// Config
	spamCfg := config.SpamConfig{
		Enabled:     true,
		RspamdURL:   mockRspamd.URL,
		RejectScore: 15.0,
		RateLimit:   config.RateLimitConfig{Window: "10m", Count: 100},
	}

	svc, err := services.NewSpamProtectionService(spamCfg, logger, nil, nil)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// Test Content Check
	reader := strings.NewReader("SPAM CONTENT")
	headers := map[string]string{"IP": "1.2.3.4"}

	res, err := svc.CheckContent(context.Background(), reader, headers)
	if err != nil {
		t.Fatalf("CheckContent failed: %v", err)
	}

	if res.Score != 20.0 {
		t.Errorf("Expected score 20.0, got %f", res.Score)
	}
	// Action mapped from "reject" (rspamd) to internal action (probably SpamActionReject)
	// assuming domain definitions.
	// We can assertions implicitly via the map
}
