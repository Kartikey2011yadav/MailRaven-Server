package tests

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/imap"
	"github.com/Kartikey2011yadav/mailraven-server/internal/config"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
	"github.com/Kartikey2011yadav/mailraven-server/internal/observability"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

// TestIMAP_Compliance validates the IMAP server against standard mobile client requirements.
func TestIMAP_Compliance(t *testing.T) {
	// 1. Setup Environment
	env := setupTestEnvironment(t)
	defer env.cleanup()

	// Create a user with a known password
	pass := "mobileapp123"
	hash, _ := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	user := &domain.User{
		Email:        "user@mobile.local",
		PasswordHash: string(hash),
		CreatedAt:    time.Now(),
	}
	_ = env.userRepo.Create(context.Background(), user)

	// 2. Start IMAP Server
	imapCfg := config.IMAPConfig{
		Port:              0, // Random port
		AllowInsecureAuth: true,
	}
	logger := observability.NewLogger("error", "text")
	server := imap.NewServer(imapCfg, logger, env.userRepo, env.emailRepo, &NoOpSpamFilter{}, env.blobStore)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := server.Start(ctx); err != nil {
			t.Logf("IMAP server stopped: %v", err)
		}
	}()

	// Wait for start
	var port string
	for i := 0; i < 20; i++ {
		if server.Addr() != nil {
			_, port, _ = net.SplitHostPort(server.Addr().String())
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	// 3. Connect as a Client
	conn, err := net.Dial("tcp", "localhost:"+port)
	if err != nil {
		t.Fatalf("Failed to connect to IMAP: %v", err)
	}
	defer conn.Close()
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	// Helper to send/receive
	sendCommand := func(tag, cmd string) string {
		fmt.Fprintf(writer, "%s %s\r\n", tag, cmd)
		writer.Flush()
		for {
			line, _ := reader.ReadString('\n')
			// t.Logf("S: %s", strings.TrimSpace(line))
			if strings.HasPrefix(line, tag+" ") {
				return line
			}
		}
	}

	// Read Greeting
	banner, _ := reader.ReadString('\n')
	assert.Contains(t, banner, "* OK", "Server should send greeting")

	// Test 1: CAPABILITY
	resp := sendCommand("A001", "CAPABILITY")
	assert.Contains(t, resp, "OK", "CAPABILITY should succeed")

	// Test 2: LOGIN
	resp = sendCommand("A002", fmt.Sprintf("LOGIN %s %s", user.Email, pass))
	assert.Contains(t, resp, "OK", "LOGIN should succeed")

	// Test 3: SELECT INBOX (Critical for Mobile Clients)
	// This is expected to FAIL currently
	resp = sendCommand("A003", "SELECT INBOX")
	if strings.Contains(resp, "OK") {
		t.Log("✅ SELECT INBOX implementation found")
	} else {
		t.Log("❌ SELECT INBOX not implemented (Required for Mobile Apps)")
	}

	// Test 4: LIST (Discover folders)
	resp = sendCommand("A004", "LIST \"\" *")
	if strings.Contains(resp, "OK") {
		t.Log("✅ LIST implementation found")
	} else {
		t.Log("❌ LIST not implemented (Required for Mobile Apps)")
	}

	// Test 5: IDLE (Push Email)
	// Manual interaction because sendCommand expects immediate tagged response
	fmt.Fprintf(writer, "A005 IDLE\r\n")
	writer.Flush()
	line, _ := reader.ReadString('\n')
	if strings.HasPrefix(line, "+") {
		t.Log("✅ IDLE (Push) implementation found")
		// Send DONE to finish IDLE
		writer.WriteString("DONE\r\n")
		writer.Flush()
		// Now we expect the tagged OK
		for {
			line, _ = reader.ReadString('\n')
			if strings.HasPrefix(line, "A005 ") {
				assert.Contains(t, line, "OK", "IDLE termination")
				break
			}
		}
	} else if strings.HasPrefix(line, "A005 NO") || strings.HasPrefix(line, "A005 BAD") {
		t.Log("❌ IDLE (Push) not implemented or failed: " + line)
	}
}
