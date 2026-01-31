package sieve

import (
	"context"
	"net/mail"
	"testing"
	"time"

	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
	"github.com/stretchr/testify/mock"
)

// Mock repos - Use types from engine_test.go where possible, but here we define MockQueueRepo and MockBlobStore
// MockVacationRepo is also in engine_test.go, but Go test compiles package-wide?
// If execution fails due to duplicate, we need to remove one.
// Let's delete MockVacationRepo from here if it conflicts.
// The error was "redeclarion".

type MockQueueRepo struct {
	mock.Mock
}

func (m *MockQueueRepo) Enqueue(ctx context.Context, msg *domain.OutboundMessage) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}
func (m *MockQueueRepo) LockNextReady(ctx context.Context) (*domain.OutboundMessage, error) {
	return nil, nil
}
func (m *MockQueueRepo) UpdateStatus(ctx context.Context, id string, status domain.OutboundStatus, retryCount int, nextRetry time.Time, lastError string) error {
	return nil
}
func (m *MockQueueRepo) Stats(ctx context.Context) (pending, processing, failed, completed int64, err error) {
	return 0, 0, 0, 0, nil
}

type MockBlobStore struct {
	mock.Mock
}

func (m *MockBlobStore) Write(ctx context.Context, messageID string, content []byte) (string, error) {
	args := m.Called(ctx, messageID, content)
	return args.String(0), args.Error(1)
}
func (m *MockBlobStore) Read(ctx context.Context, path string) ([]byte, error) { return nil, nil }
func (m *MockBlobStore) Delete(ctx context.Context, path string) error         { return nil }
func (m *MockBlobStore) Verify(ctx context.Context, path string) error         { return nil }

func TestVacationManager_ProcessVacation(t *testing.T) {
	// Setup mocks
	vRepo := new(MockVacationRepo)
	qRepo := new(MockQueueRepo)
	bStore := new(MockBlobStore)

	vm := NewVacationManager(vRepo, qRepo, bStore)

	ctx := context.Background()
	recipient := "user@example.com"
	sender := "sender@example.com"

	msg := &mail.Message{
		Header: mail.Header{
			"From":        []string{sender},
			"Subject":     []string{"Hello"},
			"Return-Path": []string{sender},
		},
	}

	// Case 1: New sender, should reply
	reason := "I am on vacation"
	opts := map[string]interface{}{
		"days":    7,
		"subject": "Out of Office",
	}

	// Expectations
	vRepo.On("LastReply", ctx, recipient, sender).Return(time.Time{}, nil)
	bStore.On("Write", ctx, mock.Anything, mock.Anything).Return("", nil)
	qRepo.On("Enqueue", ctx, mock.MatchedBy(func(m *domain.OutboundMessage) bool {
		return m.Sender == recipient && m.Recipient == sender
	})).Return(nil)
	vRepo.On("RecordReply", ctx, recipient, sender).Return(nil)

	err := vm.ProcessVacation(ctx, recipient, msg, opts, reason)
	if err != nil {
		t.Errorf("ProcessVacation failed: %v", err)
	}

	vRepo.AssertExpectations(t)
	qRepo.AssertExpectations(t)

	// Case 2: Recently replied, should NOT reply
	vRepo = new(MockVacationRepo)
	qRepo = new(MockQueueRepo) // Reset
	vm = NewVacationManager(vRepo, qRepo, bStore)

	vRepo.On("LastReply", ctx, recipient, sender).Return(time.Now().Add(-1*time.Hour), nil)
	// No Enqueue expected

	err = vm.ProcessVacation(ctx, recipient, msg, opts, reason)
	if err != nil {
		t.Errorf("ProcessVacation failed: %v", err)
	}
	vRepo.AssertExpectations(t)
	qRepo.AssertNotCalled(t, "Enqueue")
}
