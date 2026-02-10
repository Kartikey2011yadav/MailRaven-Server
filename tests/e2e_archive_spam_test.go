package tests

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/http/dto"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestArchiveAndSpamFeatures(t *testing.T) {
	env := setupTestEnvironment(t)
	defer env.cleanup()

	ctx := context.Background()

	// 1. Setup User
	pw, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user := &domain.User{
		Email:        "user@test.local",
		PasswordHash: string(pw),
		Role:         domain.RoleUser,
		CreatedAt:    time.Now(),
	}
	if err := env.userRepo.Create(ctx, user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// 2. Create Messages
	// Message 1: Inbox, Unstarred, Unread
	msg1 := &domain.Message{
		ID:         "msg-new-1",
		MessageID:  "<1@test.local>",
		Sender:     "sender@external.com",
		Recipient:  user.Email,
		Subject:    "Inbox Message",
		Mailbox:    "INBOX",
		IsStarred:  false,
		ReadState:  false,
		ReceivedAt: time.Now(),
		BodyPath:   "blobs/1",
	}
	// Write dummy blob
	env.blobStore.Write(ctx, "blobs/1", []byte("Message 1 Content"))

	// Message 2: Inbox, Starred, Read
	msg2 := &domain.Message{
		ID:         "msg-new-2",
		MessageID:  "<2@test.local>",
		Sender:     "sender@external.com",
		Recipient:  user.Email,
		Subject:    "Starred Message",
		Mailbox:    "INBOX",
		IsStarred:  true,
		ReadState:  true,
		ReceivedAt: time.Now().Add(-1 * time.Hour),
		BodyPath:   "blobs/2",
	}
	env.blobStore.Write(ctx, "blobs/2", []byte("Message 2 Content"))

	if err := env.emailRepo.Save(ctx, msg1); err != nil {
		t.Fatalf("Failed to save msg1: %v", err)
	}
	if err := env.emailRepo.Save(ctx, msg2); err != nil {
		t.Fatalf("Failed to save msg2: %v", err)
	}

	// Auth token
	token := env.authenticateUser(t, user.Email, "password123")

	t.Run("Star Message", func(t *testing.T) {
		// Star message 1
		reqData := dto.UpdateMessageRequest{
			IsStarred: boolPtr(true),
		}
		var resp dto.MessageSummary
		status := makeTestRequest(t, env, "PATCH", "/api/v1/messages/"+msg1.ID, reqData, &resp, token)
		assert.Equal(t, http.StatusOK, status)
		assert.True(t, resp.IsStarred)

		// Verify in DB
		updatedMsg, _ := env.emailRepo.FindByID(ctx, msg1.ID)
		assert.True(t, updatedMsg.IsStarred)
	})

	t.Run("Archive Message", func(t *testing.T) {
		// Archive message 1 (Move to Archive mailbox)
		reqData := dto.UpdateMessageRequest{
			Mailbox: stringPtr("Archive"),
		}
		var resp dto.MessageSummary
		status := makeTestRequest(t, env, "PATCH", "/api/v1/messages/"+msg1.ID, reqData, &resp, token)
		assert.Equal(t, http.StatusOK, status)
		assert.Equal(t, "Archive", resp.Mailbox)

		// Verify in DB
		updatedMsg, _ := env.emailRepo.FindByID(ctx, msg1.ID)
		assert.Equal(t, "Archive", updatedMsg.Mailbox)
	})

	t.Run("Report Spam", func(t *testing.T) {
		status := makeTestRequest(t, env, "POST", "/api/v1/messages/"+msg2.ID+"/spam", nil, nil, token)
		assert.Equal(t, http.StatusOK, status)

		// Verify moved to Junk
		updatedMsg, _ := env.emailRepo.FindByID(ctx, msg2.ID)
		assert.Equal(t, "Junk", updatedMsg.Mailbox)
	})

	t.Run("Filter Messages", func(t *testing.T) {
		// Create a new fresh message in Inbox
		msg3 := &domain.Message{
			ID:         "msg-new-3",
			MessageID:  "<3@test.local>",
			Sender:     "sender@external.com",
			Recipient:  user.Email,
			Subject:    "Fresh Inbox",
			Mailbox:    "INBOX",
			ReceivedAt: time.Now(),
			BodyPath:   "blobs/3",
		}
		env.blobStore.Write(ctx, "blobs/3", []byte("Message 3 Content"))
		env.emailRepo.Save(ctx, msg3)

		// 1. Filter Mailbox=INBOX
		var listResp dto.MessageListResponse
		code := makeTestRequest(t, env, "GET", "/api/v1/messages?mailbox=INBOX", nil, &listResp, token)
		assert.Equal(t, http.StatusOK, code)
		// Should contain msg3 only (msg1=Archive, msg2=Junk)
		// list should have 1 item.
		if assert.Equal(t, 1, len(listResp.Messages)) {
			assert.Equal(t, "msg-new-3", listResp.Messages[0].ID)
		}

		// 2. Filter Mailbox=Junk (msg2)
		code = makeTestRequest(t, env, "GET", "/api/v1/messages?mailbox=Junk", nil, &listResp, token)
		assert.Equal(t, http.StatusOK, code)
		if assert.Equal(t, 1, len(listResp.Messages)) {
			assert.Equal(t, "msg-new-2", listResp.Messages[0].ID)
		}

		// 3. Filter IsStarred=true (msg1 was starred in test 1, msg2 was created starred)
		// msg1 (Archive, Starred)
		// msg2 (Junk, Starred)
		// msg3 (INBOX, Unstarred)
		// Response should have 2 messages.

		code = makeTestRequest(t, env, "GET", "/api/v1/messages?is_starred=true", nil, &listResp, token)
		assert.Equal(t, http.StatusOK, code)
		assert.Equal(t, 2, len(listResp.Messages))
	})
}

// Helper to make requests within this test file
func makeTestRequest(t *testing.T, env *testEnvironment, method, path string, reqBody interface{}, respDest interface{}, token string) int {
	var bodyReader io.Reader
	if reqBody != nil {
		bodyReader = env.encodeJSON(t, reqBody)
	}

	req := env.newRequest(t, method, path, bodyReader, token)
	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp := env.doRequest(t, req)
	defer resp.Body.Close()

	if respDest != nil && resp.StatusCode == http.StatusOK {
		env.decodeJSON(t, resp.Body, respDest)
	}

	return resp.StatusCode
}

func stringPtr(s string) *string {
	return &s
}
