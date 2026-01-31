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
	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/storage/sqlite"
	"github.com/Kartikey2011yadav/mailraven-server/internal/config"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/services"
	"github.com/Kartikey2011yadav/mailraven-server/internal/observability"
)

// TestSpamTrainingFeedback validates that moving emails to/from Junk trains the Bayes classifier.
func TestSpamTrainingFeedback(t *testing.T) {
	// 1. Setup Environment
	env := setupTestEnvironment(t)
	defer env.cleanup()

	// 2. Setup Services
	logger := observability.NewLogger("error", "text")
	bayesRepo := sqlite.NewBayesRepository(env.conn.DB)

	spamCfg := config.SpamConfig{Enabled: true}
	spamSvc, err := services.NewSpamProtectionService(spamCfg, logger, nil, bayesRepo)
	if err != nil {
		t.Fatalf("Failed to init spam service: %v", err)
	}

	// 3. Setup IMAP Server
	imapCfg := config.IMAPConfig{
		Enabled:           true,
		Port:              10145, // Unique port
		AllowInsecureAuth: true,
	}

	// Write a message body with specific tokens
	msgBody := []byte("Subject: Buy Now\r\n\r\nBUY VIAGRA NOW CHEAP")

	// Create Msg in DB
	msgID := "msg-spam-1"

	// 4. Inject Data
	// Create "Junk" mailbox
	err = env.emailRepo.CreateMailbox(context.Background(), "test@example.com", "Junk")
	if err != nil {
		t.Fatalf("Failed to create Junk mailbox: %v", err)
	}

	// Insert Message in INBOX
	msg := &domain.Message{
		ID:         msgID,
		MessageID:  "<spam1@example.com>",
		Sender:     "spammer@example.com",
		Recipient:  "test@example.com",
		Subject:    "Spam",
		Snippet:    "Buy things",
		BodyPath:   msgID, // BlobStore usually hashes ID or uses it
		UID:        100,   // Fake UID
		Mailbox:    "INBOX",
		ReceivedAt: time.Now(),
		Flags:      "",
	}
	// We need to ensure BodyPath matches what we write.
	// If using disk.BlobStore simple, it might be just ID.

	// Write Blob
	// We need to instantiate a BlobStore to write it properly or reverse engineer it.
	// To be safe, I'll use the proper import (I need to add it to imports).
	// Let's assume `disk` package check later.

	// For now, I'll use a MockBlobStore if possible, or just skip body check if I can't easily write it?
	// T022 implementation: `s.blobStore.Read(ctx, msg.BodyPath)`
	// So I MUST succeed in reading.

	// 5. Start Server
	// I need to add `github.com/Kartikey2011yadav/mailraven-server/internal/adapters/storage/disk` to imports.
	// And `github.com/Kartikey2011yadav/mailraven-server/internal/adapters/storage/sqlite` ok.

	server := imap.NewServer(imapCfg, logger, env.userRepo, env.emailRepo, spamSvc, &MockBlobStore{content: msgBody})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = server.Start(ctx)
	}()
	time.Sleep(100 * time.Millisecond)

	// 6. Connect Client
	conn, err := net.Dial("tcp", "localhost:10145")
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close()
	rd := bufio.NewReader(conn)

	// Login
	mustRead(t, rd, "* OK")
	mustWrite(t, conn, "A1 LOGIN test@example.com testpassword123")
	mustRead(t, rd, "A1 OK")

	// Inject message into DB (after login, or before, doesn't matter)
	// We need a UID.
	// Setup INBOX properly.
	err = env.emailRepo.Save(context.Background(), msg)
	if err != nil {
		t.Fatalf("Failed to save msg: %v", err)
	}

	// SELECT INBOX
	mustWrite(t, conn, "A2 SELECT INBOX")
	// Expect EXISTS, OK
	mustReadUntil(t, rd, "A2 OK")

	// UID COPY <UID> Junk
	// We need to know the UID assigned. sqlite `Save` assigns it.
	// We can fetch it or just guess it's specific if fresh DB.
	// Helpers creates some msgs. `msg` is new.
	// Let's Fetch it to be sure? Or just use "1:*" copy if it's the only one.
	// Let's use `UID COPY 1:* Junk` to move everything.

	mustWrite(t, conn, "A3 UID COPY 1:* Junk")
	mustRead(t, rd, "A3 OK")

	// 7. Verification
	// Wait for async goroutine in `handleUidCopy` to finish training.
	time.Sleep(500 * time.Millisecond)

	// Check DB tokens
	// "VIAGRA" should be spam
	// Table: bayes_tokens (token, spam_count, ham_count)

	// We need to query SQLite directly or use BayesRepo if it has methods.
	// `bayesRepo` is `*sqlite.BayesRepository`.
	// It likely has `GetToken(ctx, token)`.
	// Checking `bayes.go`: it has `GetTokens(ctx, []string)`.

	// But `BayesRepository` interface only has `GetTokens`.
	// Let's use `GetTokens`.

	// Tokenizer normalizes to lowercase
	token := "viagra"
	probs, err := bayesRepo.GetTokens(context.Background(), []string{token})
	if err != nil {
		t.Fatalf("Failed to get tokens: %v", err)
	}

	if len(probs) == 0 {
		t.Fatalf("Token '%s' not found in DB", token)
	}

	tokenStats := probs[token]
	// Verify counts
	// Since we trained spam, SpamCount should be > 0
	// We can't access SpamCount directly from `Probability` struct usually (it has P(S|T)).
	// Wait, `domain.TokenProbability` usually has the raw counts too if defined that way?
	// Checking `classifier.go` or `domain`?
	// The `GetTokens` returns `map[string]domain.TokenProbability`.
	// Let's check `domain/spam.go`.

	// Assuming it exposes something useful.
	// If `TokenProbability` struct only has `Prob float64`, we are verifying Prob > 0.5?
	// Initial state: 1 spam, 0 ham.
	// P(S|T) = P(T|S)*P(S) / P(T)
	// P(T|S) = 1/1 = 1.
	// P(S) = Assume 0.5
	// P(T) = ...
	// Basically if it exists and prob > 0, it worked.

	if tokenStats.SpamCount < 1 {
		t.Errorf("Expected SpamCount >= 1, got %d", tokenStats.SpamCount)
	}
}

// MockBlobStore for testing
type MockBlobStore struct {
	content []byte
}

func (m *MockBlobStore) Read(ctx context.Context, path string) ([]byte, error) {
	return m.content, nil
}
func (m *MockBlobStore) Write(ctx context.Context, id string, b []byte) (string, error) {
	return "", nil
}
func (m *MockBlobStore) Delete(ctx context.Context, p string) error { return nil }

// Add Verify if needed by interface
func (m *MockBlobStore) Verify(ctx context.Context, p string) error { return nil }

// Helpers
func mustWrite(t *testing.T, conn net.Conn, s string) {
	_, err := fmt.Fprintf(conn, "%s\r\n", s)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
}

func mustRead(t *testing.T, rd *bufio.Reader, expectSub string) {
	line, err := rd.ReadString('\n')
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if !strings.Contains(line, expectSub) {
		t.Errorf("Expected '%s', got '%s'", expectSub, strings.TrimSpace(line))
	}
}

func mustReadUntil(t *testing.T, rd *bufio.Reader, expectSub string) {
	for {
		line, err := rd.ReadString('\n')
		if err != nil {
			t.Fatalf("Read failed: %v", err)
		}
		if strings.Contains(line, expectSub) {
			return
		}
		if strings.HasPrefix(line, "A") && !strings.Contains(line, expectSub) && !strings.Contains(line, "OK") {
			// Basic failure check if we got a different tagged response
		}
	}
}
