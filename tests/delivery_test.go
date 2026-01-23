package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/smtp"
	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/storage/disk"
	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/storage/sqlite"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
	"github.com/Kartikey2011yadav/mailraven-server/internal/observability"
	"github.com/google/uuid"
)

// MockSender for testing DeliveryWorker
type MockSender struct {
	ShouldFail bool
	Calls      int
	LastData   []byte
}

func (m *MockSender) Send(ctx context.Context, from, to string, data []byte) error {
	m.Calls++
	m.LastData = data
	if m.ShouldFail {
		return errors.New("simulated network failure")
	}
	return nil
}

// TestDeliveryRetry tests T090: Exponential backoff
func TestDeliveryRetry(t *testing.T) {
	// Setup dependencies
	dbDir := t.TempDir()
	env := setupTestEnvironment(t) // Uses api_test.go helper but need queue repo
	defer env.cleanup()

	// Create real QueueRepository
	queueRepo := sqlite.NewQueueRepository(env.conn.DB)

	// Create real BlobStore
	blobStore, err := disk.NewBlobStore(dbDir)
	if err != nil {
		t.Fatalf("Failed to create blob store: %v", err)
	}

	mockSender := &MockSender{ShouldFail: true}
	logger := observability.NewLogger("debug", "text")
	metrics := observability.NewMetrics()

	worker := smtp.NewDeliveryWorker(queueRepo, blobStore, mockSender, logger, metrics)

	// Create a test message in queue
	msgID := uuid.New().String()
	// Write dummy blob
	storedPath, err := blobStore.Write(context.Background(), msgID, []byte("Content"))
	if err != nil {
		t.Fatalf("Failed to write blob: %v", err)
	}

	outMsg := &domain.OutboundMessage{
		ID:          msgID,
		Sender:      "sender@example.com",
		Recipient:   "recipient@example.com",
		BlobKey:     storedPath,
		Status:      domain.QueueStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		NextRetryAt: time.Now().Add(-1 * time.Minute), // Ready immediately
		RetryCount:  0,
	}

	if err := queueRepo.Enqueue(context.Background(), outMsg); err != nil {
		t.Fatalf("Failed to enqueue: %v", err)
	}

	// 1. First Attempt (Should Fail)
	worker.ProcessNext()

	if mockSender.Calls != 1 {
		t.Errorf("Expected 1 call, got %d", mockSender.Calls)
	}

	// Check status
	// We need to fetch from DB
	// Check queue table directly or via LockNextReady (but it won't be ready)
	// I'll add `FindByID` to QueueRepository later? Or just use raw sql here?
	// `sqlite` package exposes `QueueRepository` struct but fields are private.
	// I can use `env.conn.DB` to query.

	row := env.conn.DB.QueryRow("SELECT status, retry_count, next_retry_at FROM queue WHERE id = ?", msgID)
	var status string
	var retryCount int
	var nextRetryAtUnix int64
	if err := row.Scan(&status, &retryCount, &nextRetryAtUnix); err != nil {
		t.Fatalf("Failed to query queue: %v", err)
	}

	if status != "RETRYING" {
		t.Errorf("Expected status RETRYING, got %s", status)
	}
	if retryCount != 1 {
		t.Errorf("Expected retry_count 1, got %d", retryCount)
	}

	// Check backoff (1st retry = 1 minute)
	nextRetryAt := time.Unix(nextRetryAtUnix, 0)
	diff := nextRetryAt.Sub(time.Now())
	// Should be around ~1 minute from now
	if diff < 55*time.Second || diff > 65*time.Second {
		// Note: execution time might affect this, but rough check ok.
		// Actually nextRetryAt is set to time.Now().Add(delay).
		// So `nextRetryAt` should be roughly Now + 60s.
		// diff = nextRetryAt - Now ~= 60s.
		t.Errorf("Expected ~1m backoff, got %v", diff)
	}

	// 2. Second Attempt (Simulate time pass)
	// Update next_retry_at to past
	_, err = env.conn.DB.Exec("UPDATE queue SET next_retry_at = ? WHERE id = ?", time.Now().Add(-1*time.Minute).Unix(), msgID)
	if err != nil {
		t.Fatalf("Failed to update retry time: %v", err)
	}

	worker.ProcessNext()

	if mockSender.Calls != 2 {
		t.Errorf("Expected 2 calls, got %d", mockSender.Calls)
	}

	row = env.conn.DB.QueryRow("SELECT status, retry_count, next_retry_at FROM queue WHERE id = ?", msgID)
	if err := row.Scan(&status, &retryCount, &nextRetryAtUnix); err != nil {
		t.Fatalf("Failed to query queue: %v", err)
	}

	if retryCount != 2 {
		t.Errorf("Expected retry_count 2, got %d", retryCount)
	}

	// Check backoff (2nd retry = 5 minutes)
	nextRetryAt = time.Unix(nextRetryAtUnix, 0)
	diff = nextRetryAt.Sub(time.Now())
	if diff < 4*time.Minute || diff > 6*time.Minute {
		t.Errorf("Expected ~5m backoff, got %v", diff)
	}
}
