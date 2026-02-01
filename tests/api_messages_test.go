package tests

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/http/dto"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
	"golang.org/x/crypto/bcrypt"
)

func TestMessageIsolation(t *testing.T) {
	env := setupTestEnvironment(t)
	defer env.cleanup()

	ctx := context.Background()

	// Create User A
	pwA, _ := bcrypt.GenerateFromPassword([]byte("passwordA"), bcrypt.DefaultCost)
	userA := &domain.User{
		Email:        "userA@example.com",
		PasswordHash: string(pwA),
		Role:         domain.RoleUser,
		CreatedAt:    time.Now(),
	}
	if err := env.userRepo.Create(ctx, userA); err != nil {
		t.Fatalf("Failed to create User A: %v", err)
	}

	// Create User B
	pwB, _ := bcrypt.GenerateFromPassword([]byte("passwordB"), bcrypt.DefaultCost)
	userB := &domain.User{
		Email:        "userB@example.com",
		PasswordHash: string(pwB),
		Role:         domain.RoleUser,
		CreatedAt:    time.Now(),
	}
	if err := env.userRepo.Create(ctx, userB); err != nil {
		t.Fatalf("Failed to create User B: %v", err)
	}

	// Create Message for User A
	msgA := &domain.Message{
		ID:         "msg-A",
		MessageID:  "<msgA@example.com>",
		Sender:     "someone@example.com",
		Recipient:  "userA@example.com",
		Subject:    "Message for A",
		Snippet:    "Content A",
		ReadState:  false,
		ReceivedAt: time.Now(),
		SPFResult:  "pass",
	}
	// Write blob
	pathA, _ := env.blobStore.Write(ctx, msgA.ID, []byte("Body A"))
	msgA.BodyPath = pathA
	if err := env.emailRepo.Save(ctx, msgA); err != nil {
		t.Fatalf("Failed to save msg A: %v", err)
	}

	// Create Message for User B
	msgB := &domain.Message{
		ID:         "msg-B",
		MessageID:  "<msgB@example.com>",
		Sender:     "someone@example.com",
		Recipient:  "userB@example.com",
		Subject:    "Message for B",
		Snippet:    "Content B",
		ReadState:  false,
		ReceivedAt: time.Now(),
		SPFResult:  "pass",
	}
	pathB, _ := env.blobStore.Write(ctx, msgB.ID, []byte("Body B"))
	msgB.BodyPath = pathB
	if err := env.emailRepo.Save(ctx, msgB); err != nil {
		t.Fatalf("Failed to save msg B: %v", err)
	}

	// Authenticate as User A
	tokenA := env.authenticateUser(t, "userA@example.com", "passwordA")

	// Test List: User A should only see msg-A
	t.Run("ListIsolation", func(t *testing.T) {
		req := env.newRequest(t, "GET", "/api/v1/messages", nil, tokenA)
		resp := env.doRequest(t, req)

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200 OK, got %d", resp.StatusCode)
		}

		var listResp dto.MessageListResponse
		env.decodeJSON(t, resp.Body, &listResp)

		if len(listResp.Messages) != 1 {
			t.Errorf("Expected 1 message, got %d", len(listResp.Messages))
		}
		if len(listResp.Messages) > 0 && listResp.Messages[0].ID != "msg-A" {
			t.Errorf("Expected msg-A, got %s", listResp.Messages[0].ID)
		}
	})

	// Test Get: User A trying to get msg-B
	t.Run("GetIsolation", func(t *testing.T) {
		req := env.newRequest(t, "GET", "/api/v1/messages/msg-B", nil, tokenA)
		resp := env.doRequest(t, req)

		// Expect 404 (Not Found) or 403 (Forbidden)
		// Usually implementation of GetOne in repository performs "FindByIDAndUser"
		// If it's just FindByID, then the handler must check user.
		// Let's verify what we get.
		if resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusForbidden {
			t.Errorf("Expected 404 or 403, got %d", resp.StatusCode)
		}
	})
}

func TestSendingMessages(t *testing.T) {
	env := setupTestEnvironment(t)
	defer env.cleanup()

	ctx := context.Background()

	// Create User A
	pwA, _ := bcrypt.GenerateFromPassword([]byte("passwordA"), bcrypt.DefaultCost)
	userA := &domain.User{
		Email:        "userA@example.com",
		PasswordHash: string(pwA),
		Role:         domain.RoleUser,
		CreatedAt:    time.Now(),
	}
	if err := env.userRepo.Create(ctx, userA); err != nil {
		t.Fatalf("Failed to create User A: %v", err)
	}

	// Login as User A
	tokenA := env.authenticateUser(t, "userA@example.com", "passwordA")

	// Prepare Send Request
	sendReq := map[string]string{
		"to":      "userB@example.com",
		"subject": "Hello User B",
		"body":    "This is a test message.",
	}
	body := env.encodeJSON(t, sendReq)

	// Send Message
	req := env.newRequest(t, "POST", "/api/v1/messages/send", body, tokenA)
	req.Header.Set("Content-Type", "application/json")
	resp := env.doRequest(t, req)

	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("Expected 202 Accepted, got %d", resp.StatusCode)
	}

	// Verify Queued Message
	// We use LockNextReady which picks the next pending status item
	msg, err := env.queueRepo.LockNextReady(ctx)
	if err != nil {
		t.Fatalf("Failed to fetch queued message: %v", err)
	}
	if msg == nil {
		t.Fatal("Expected queued message, got nil")
	}

	if msg.Sender != "userA@example.com" {
		t.Errorf("Expected sender userA@example.com, got %s", msg.Sender)
	}
	if msg.Recipient != "userB@example.com" {
		t.Errorf("Expected recipient userB@example.com, got %s", msg.Recipient)
	}
}
