package tests

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/http/dto"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
)

// TestAPIIntegration tests the REST API endpoints (T069)
func TestAPIIntegration(t *testing.T) {
	// Setup test environment
	env := setupTestEnvironment(t)
	defer env.cleanup()

	// Create test user and authenticate
	token := env.authenticateUser(t, "test@example.com", "testpassword123")

	// Test 1: List messages (GET /api/v1/messages)
	t.Run("ListMessages", func(t *testing.T) {
		req := env.newRequest(t, "GET", "/api/v1/messages", nil, token)
		resp := env.doRequest(t, req)

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200 OK, got %d", resp.StatusCode)
		}

		var listResp dto.MessageListResponse
		env.decodeJSON(t, resp.Body, &listResp)

		if len(listResp.Messages) != 3 {
			t.Errorf("Expected 3 messages, got %d", len(listResp.Messages))
		}
		if listResp.Total != 3 {
			t.Errorf("Expected total=3, got %d", listResp.Total)
		}
		if listResp.Limit != 20 {
			t.Errorf("Expected limit=20, got %d", listResp.Limit)
		}
		if listResp.HasMore {
			t.Errorf("Expected has_more=false with only 3 messages")
		}
	})

	// Test 2: Get specific message (GET /api/v1/messages/{id})
	t.Run("GetMessage", func(t *testing.T) {
		messageID := env.messages[0].ID
		req := env.newRequest(t, "GET", fmt.Sprintf("/api/v1/messages/%s", messageID), nil, token)
		resp := env.doRequest(t, req)

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200 OK, got %d", resp.StatusCode)
		}

		var fullMsg dto.MessageFull
		env.decodeJSON(t, resp.Body, &fullMsg)

		if fullMsg.ID != messageID {
			t.Errorf("Expected message ID %s, got %s", messageID, fullMsg.ID)
		}
		if fullMsg.Body == "" {
			t.Error("Expected message body, got empty string")
		}
		if fullMsg.BodySize == 0 {
			t.Error("Expected body_size > 0")
		}
	})

	// Test 3: Update message read state (PATCH /api/v1/messages/{id})
	t.Run("UpdateMessageReadState", func(t *testing.T) {
		messageID := env.messages[0].ID

		// Mark as read
		updateReq := dto.UpdateMessageRequest{
			ReadState: boolPtr(true),
		}
		body := env.encodeJSON(t, updateReq)
		req := env.newRequest(t, "PATCH", fmt.Sprintf("/api/v1/messages/%s", messageID), body, token)
		req.Header.Set("Content-Type", "application/json")
		resp := env.doRequest(t, req)

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200 OK, got %d", resp.StatusCode)
		}

		var updatedMsg dto.MessageSummary
		env.decodeJSON(t, resp.Body, &updatedMsg)

		if !updatedMsg.ReadState {
			t.Error("Expected read_state=true after update")
		}

		// Verify persistence by fetching again
		req = env.newRequest(t, "GET", fmt.Sprintf("/api/v1/messages/%s", messageID), nil, token)
		resp = env.doRequest(t, req)
		var fullMsg dto.MessageFull
		env.decodeJSON(t, resp.Body, &fullMsg)

		if !fullMsg.ReadState {
			t.Error("Expected read_state to persist as true")
		}
	})

	// Test 4: Unauthorized access (no token)
	t.Run("UnauthorizedAccess", func(t *testing.T) {
		req := env.newRequest(t, "GET", "/api/v1/messages", nil, "")
		resp := env.doRequest(t, req)

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected 401 Unauthorized, got %d", resp.StatusCode)
		}

		var errResp dto.ErrorResponse
		env.decodeJSON(t, resp.Body, &errResp)

		if errResp.Error != "Unauthorized" {
			t.Errorf("Expected error='Unauthorized', got '%s'", errResp.Error)
		}
	})

	// Test 5: Message not found
	t.Run("MessageNotFound", func(t *testing.T) {
		req := env.newRequest(t, "GET", "/api/v1/messages/nonexistent-id", nil, token)
		resp := env.doRequest(t, req)

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected 404 Not Found, got %d", resp.StatusCode)
		}
	})
}

// TestPagination tests message pagination (T070)
func TestPagination(t *testing.T) {
	env := setupTestEnvironment(t)
	defer env.cleanup()

	// Create 25 messages for pagination testing (3 already exist from setup)
	for i := 4; i <= 25; i++ {
		msg := &domain.Message{
			ID:          fmt.Sprintf("msg-%d", i),
			MessageID:   fmt.Sprintf("<test%d@example.com>", i),
			Sender:      "sender@example.com",
			Recipient:   "test@example.com",
			Subject:     fmt.Sprintf("Test Message %d", i),
			Snippet:     "Test content",
			BodyPath:    fmt.Sprintf("test/path/msg-%d.eml.gz", i), // Dummy path for pagination test
			ReadState:   false,
			ReceivedAt:  time.Now().Add(-time.Duration(i) * time.Hour),
			SPFResult:   "pass",
			DKIMResult:  "pass",
			DMARCResult: "pass",
		}
		if err := env.emailRepo.Save(context.Background(), msg); err != nil {
			t.Fatalf("Failed to create test message: %v", err)
		}
	}

	token := env.authenticateUser(t, "test@example.com", "testpassword123")

	// Test 1: First page (default limit=20)
	t.Run("FirstPage", func(t *testing.T) {
		req := env.newRequest(t, "GET", "/api/v1/messages", nil, token)
		resp := env.doRequest(t, req)

		var listResp dto.MessageListResponse
		env.decodeJSON(t, resp.Body, &listResp)

		if len(listResp.Messages) != 20 {
			t.Errorf("Expected 20 messages on first page, got %d", len(listResp.Messages))
		}
		if listResp.Total != 25 {
			t.Errorf("Expected total=25, got %d", listResp.Total)
		}
		if !listResp.HasMore {
			t.Error("Expected has_more=true when more pages exist")
		}
	})

	// Test 2: Second page with offset
	t.Run("SecondPage", func(t *testing.T) {
		req := env.newRequest(t, "GET", "/api/v1/messages?limit=20&offset=20", nil, token)
		resp := env.doRequest(t, req)

		var listResp dto.MessageListResponse
		env.decodeJSON(t, resp.Body, &listResp)

		if len(listResp.Messages) != 5 {
			t.Errorf("Expected 5 messages on second page, got %d", len(listResp.Messages))
		}
		if listResp.Offset != 20 {
			t.Errorf("Expected offset=20, got %d", listResp.Offset)
		}
		if listResp.HasMore {
			t.Error("Expected has_more=false on last page")
		}
	})

	// Test 3: Custom page size
	t.Run("CustomPageSize", func(t *testing.T) {
		req := env.newRequest(t, "GET", "/api/v1/messages?limit=10", nil, token)
		resp := env.doRequest(t, req)

		var listResp dto.MessageListResponse
		env.decodeJSON(t, resp.Body, &listResp)

		if len(listResp.Messages) != 10 {
			t.Errorf("Expected 10 messages with limit=10, got %d", len(listResp.Messages))
		}
		if listResp.Limit != 10 {
			t.Errorf("Expected limit=10, got %d", listResp.Limit)
		}
	})

	// Test 4: Invalid parameters
	t.Run("InvalidLimit", func(t *testing.T) {
		req := env.newRequest(t, "GET", "/api/v1/messages?limit=2000", nil, token)
		resp := env.doRequest(t, req)

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected 400 Bad Request for limit>1000, got %d", resp.StatusCode)
		}
	})
}

// TestCompression tests gzip compression (T071)
func TestCompression(t *testing.T) {
	env := setupTestEnvironment(t)
	defer env.cleanup()

	token := env.authenticateUser(t, "test@example.com", "testpassword123")

	// Test 1: With gzip encoding
	t.Run("WithGzipEncoding", func(t *testing.T) {
		req := env.newRequest(t, "GET", "/api/v1/messages", nil, token)
		req.Header.Set("Accept-Encoding", "gzip")
		resp := env.doRequest(t, req)

		// Check Content-Encoding header
		if resp.Header.Get("Content-Encoding") != "gzip" {
			t.Error("Expected Content-Encoding: gzip header")
		}

		// Decompress and verify JSON
		gzReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			t.Fatalf("Failed to create gzip reader: %v", err)
		}
		defer gzReader.Close()

		var listResp dto.MessageListResponse
		if err := json.NewDecoder(gzReader).Decode(&listResp); err != nil {
			t.Fatalf("Failed to decode gzipped JSON: %v", err)
		}

		if len(listResp.Messages) != 3 {
			t.Errorf("Expected 3 messages after decompression, got %d", len(listResp.Messages))
		}
	})

	// Test 2: Without gzip encoding
	t.Run("WithoutGzipEncoding", func(t *testing.T) {
		req := env.newRequest(t, "GET", "/api/v1/messages", nil, token)
		// Don't set Accept-Encoding header
		resp := env.doRequest(t, req)

		// Should not have Content-Encoding header
		if resp.Header.Get("Content-Encoding") == "gzip" {
			t.Error("Should not have gzip encoding without Accept-Encoding header")
		}

		// Should be able to decode directly
		var listResp dto.MessageListResponse
		env.decodeJSON(t, resp.Body, &listResp)

		if len(listResp.Messages) != 3 {
			t.Errorf("Expected 3 messages, got %d", len(listResp.Messages))
		}
	})
}

// TestDeltaSync tests incremental sync by timestamp (T072)
func TestDeltaSync(t *testing.T) {
	env := setupTestEnvironment(t)
	defer env.cleanup()

	token := env.authenticateUser(t, "test@example.com", "testpassword123")

	// Get initial sync time - use a time before all setup messages
	// Find the earliest message and set syncTime 1 second before it
	var earliestTime time.Time
	for _, msg := range env.messages {
		if earliestTime.IsZero() || msg.ReceivedAt.Before(earliestTime) {
			earliestTime = msg.ReceivedAt
		}
	}
	syncTime := earliestTime.Add(-1 * time.Second)

	// Test 1: Query messages since sync time
	t.Run("MessagesSinceTimestamp", func(t *testing.T) {
		sinceParam := syncTime.UTC().Format(time.RFC3339)
		req := env.newRequest(t, "GET", fmt.Sprintf("/api/v1/messages/since?since=%s", sinceParam), nil, token)
		resp := env.doRequest(t, req)

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected 200 OK, got %d. Body: %s", resp.StatusCode, string(bodyBytes))
		}

		var syncResp dto.MessagesSinceResponse
		env.decodeJSON(t, resp.Body, &syncResp)

		// All 3 test messages were created recently, should all be returned
		if syncResp.Count != 3 {
			t.Errorf("Expected 3 messages since sync time, got %d", syncResp.Count)
		}
		if len(syncResp.Messages) != 3 {
			t.Errorf("Expected 3 messages in array, got %d", len(syncResp.Messages))
		}
	})

	// Test 2: Add new message and verify incremental sync
	t.Run("IncrementalSync", func(t *testing.T) {
		// Record sync point
		syncTime := time.Now()
		time.Sleep(100 * time.Millisecond) // Ensure timestamp difference

		// Add new message
		newMsg := &domain.Message{
			ID:          "new-msg",
			MessageID:   "<new@example.com>",
			Sender:      "sender@example.com",
			Recipient:   "test@example.com",
			Subject:     "New Message After Sync",
			Snippet:     "This is new",
			BodyPath:    "test/path/new-msg.eml.gz", // Dummy path
			ReadState:   false,
			ReceivedAt:  time.Now(),
			SPFResult:   "pass",
			DKIMResult:  "pass",
			DMARCResult: "pass",
		}
		if err := env.emailRepo.Save(context.Background(), newMsg); err != nil {
			t.Fatalf("Failed to create new message: %v", err)
		}

		// Query messages since sync point
		sinceParam := syncTime.UTC().Format(time.RFC3339)
		req := env.newRequest(t, "GET", fmt.Sprintf("/api/v1/messages/since?since=%s", sinceParam), nil, token)
		resp := env.doRequest(t, req)

		var syncResp dto.MessagesSinceResponse
		env.decodeJSON(t, resp.Body, &syncResp)

		// Should only return the new message
		if syncResp.Count != 1 {
			t.Errorf("Expected 1 new message, got %d. Messages: %+v", syncResp.Count, syncResp.Messages)
		}
		if len(syncResp.Messages) > 0 && syncResp.Messages[0].Subject != "New Message After Sync" {
			t.Error("Expected to receive the new message")
		}
	})

	// Test 3: Invalid timestamp format
	t.Run("InvalidTimestamp", func(t *testing.T) {
		req := env.newRequest(t, "GET", "/api/v1/messages/since?since=invalid-timestamp", nil, token)
		resp := env.doRequest(t, req)

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected 400 Bad Request for invalid timestamp, got %d", resp.StatusCode)
		}
	})
}

// TestExpiredToken verifies that expired JWT tokens are rejected (T080)
func TestExpiredToken(t *testing.T) {
	env := setupTestEnvironment(t)
	defer env.cleanup()

	// Create expired token manually (token that expired 1 hour ago)
	expiredToken := generateExpiredToken(t, "test@example.com", time.Now().Add(-1*time.Hour))

	// Attempt to access protected endpoint with expired token
	req := env.newRequest(t, "GET", "/api/v1/messages", nil, expiredToken)
	resp := env.doRequest(t, req)

	if resp.StatusCode != http.StatusUnauthorized {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Errorf("Expected 401 Unauthorized for expired token, got %d. Body: %s", resp.StatusCode, string(bodyBytes))
	}

	var errorResp dto.ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
		// JWT library returns "token is expired" in the error, but our API returns it in Message field
		errorMsg := strings.ToLower(errorResp.Error + " " + errorResp.Message)
		if !strings.Contains(errorMsg, "expired") && !strings.Contains(errorMsg, "invalid") {
			t.Errorf("Expected error message to mention token expiration/invalid, got Error: '%s', Message: '%s'",
				errorResp.Error, errorResp.Message)
		}
	}
}
