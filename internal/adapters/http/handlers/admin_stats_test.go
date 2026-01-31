package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/http/handlers"
	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/http/middleware"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
	"github.com/Kartikey2011yadav/mailraven-server/internal/observability"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock Repositories
type MockUserRepo struct{ mock.Mock }

func (m *MockUserRepo) Create(ctx context.Context, user *domain.User) error { return nil }
func (m *MockUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	return nil, nil
}
func (m *MockUserRepo) Authenticate(ctx context.Context, email, password string) (*domain.User, error) {
	return nil, nil
}
func (m *MockUserRepo) UpdateLastLogin(ctx context.Context, email string) error { return nil }
func (m *MockUserRepo) List(ctx context.Context, limit, offset int) ([]*domain.User, error) {
	return nil, nil
}
func (m *MockUserRepo) Delete(ctx context.Context, email string) error { return nil }
func (m *MockUserRepo) UpdatePassword(ctx context.Context, email, passwordHash string) error {
	return nil
}
func (m *MockUserRepo) UpdateRole(ctx context.Context, email string, role domain.Role) error {
	return nil
}
func (m *MockUserRepo) Count(ctx context.Context) (map[string]int64, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]int64), args.Error(1)
}

type MockEmailRepo struct{ mock.Mock }

func (m *MockEmailRepo) Save(ctx context.Context, msg *domain.Message) error { return nil }
func (m *MockEmailRepo) FindByID(ctx context.Context, id string) (*domain.Message, error) {
	return nil, nil
}
func (m *MockEmailRepo) FindByUser(ctx context.Context, email string, limit, offset int) ([]*domain.Message, error) {
	return nil, nil
}
func (m *MockEmailRepo) UpdateReadState(ctx context.Context, id string, read bool) error { return nil }
func (m *MockEmailRepo) CountByUser(ctx context.Context, email string) (int, error) {
	return 0, nil
}
func (m *MockEmailRepo) FindSince(ctx context.Context, email string, since time.Time, limit int) ([]*domain.Message, error) {
	return nil, nil
}
func (m *MockEmailRepo) CountTotal(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

// IMAP Support Stubs
func (m *MockEmailRepo) GetMailbox(ctx context.Context, userID, name string) (*domain.Mailbox, error) {
	return nil, nil
}
func (m *MockEmailRepo) CreateMailbox(ctx context.Context, userID, name string) error { return nil }
func (m *MockEmailRepo) ListMailboxes(ctx context.Context, userID string) ([]*domain.Mailbox, error) {
	return nil, nil
}
func (m *MockEmailRepo) FindByUIDRange(ctx context.Context, userID, mailbox string, min, max uint32) ([]*domain.Message, error) {
	return nil, nil
}
func (m *MockEmailRepo) AddFlags(ctx context.Context, messageID string, flags ...string) error {
	return nil
}
func (m *MockEmailRepo) RemoveFlags(ctx context.Context, messageID string, flags ...string) error {
	return nil
}
func (m *MockEmailRepo) SetFlags(ctx context.Context, messageID string, flags ...string) error {
	return nil
}
func (m *MockEmailRepo) AssignUID(ctx context.Context, messageID string, mailbox string) (uint32, error) {
	return 0, nil
}
func (m *MockEmailRepo) CopyMessages(ctx context.Context, userID string, msgIDs []string, destMailbox string) error {
	return nil
}

type MockQueueRepo struct{ mock.Mock }

func (m *MockQueueRepo) Enqueue(ctx context.Context, msg *domain.OutboundMessage) error { return nil }
func (m *MockQueueRepo) LockNextReady(ctx context.Context) (*domain.OutboundMessage, error) {
	return nil, nil
}
func (m *MockQueueRepo) UpdateStatus(ctx context.Context, id string, status domain.OutboundStatus, retryCount int, nextRetry time.Time, lastError string) error {
	return nil
}
func (m *MockQueueRepo) Stats(ctx context.Context) (int64, int64, int64, int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Get(1).(int64), args.Get(2).(int64), args.Get(3).(int64), args.Error(4)
}

func TestGetSystemStats_Success(t *testing.T) {
	// Setup
	mockUserRepo := new(MockUserRepo)
	mockEmailRepo := new(MockEmailRepo)
	mockQueueRepo := new(MockQueueRepo)
	logger := observability.NewLogger("test", "v0.0.0")

	// Mock expectations
	userStats := map[string]int64{"total": 10, "active": 8, "admin": 2}
	mockUserRepo.On("Count", mock.Anything).Return(userStats, nil)
	mockEmailRepo.On("CountTotal", mock.Anything).Return(int64(100), nil)
	mockQueueRepo.On("Stats", mock.Anything).Return(int64(5), int64(2), int64(1), int64(50), nil)

	handler := handlers.NewAdminStatsHandler(mockUserRepo, mockEmailRepo, mockQueueRepo, logger)

	// Create Request
	req := httptest.NewRequest("GET", "/api/v1/admin/stats", nil)
	// Inject Context with Admin Role
	ctx := context.WithValue(req.Context(), middleware.UserRoleKey, "ADMIN")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	// Execute
	handler.GetSystemStats(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var resp handlers.SystemStatsResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)

	assert.Equal(t, int64(10), resp.Users.Total)
	assert.Equal(t, int64(8), resp.Users.Active)
	assert.Equal(t, int64(2), resp.Users.Admin)
	assert.Equal(t, int64(100), resp.Emails.Total)
	assert.Equal(t, int64(5), resp.Queue.Pending)
	assert.Equal(t, int64(2), resp.Queue.Processing)
	assert.Equal(t, int64(1), resp.Queue.Failed)
	assert.Equal(t, int64(50), resp.Queue.Completed)
}

func TestGetSystemStats_Forbidden(t *testing.T) {
	// Setup
	mockUserRepo := new(MockUserRepo)
	mockEmailRepo := new(MockEmailRepo)
	mockQueueRepo := new(MockQueueRepo)
	logger := observability.NewLogger("test", "v0.0.0")

	handler := handlers.NewAdminStatsHandler(mockUserRepo, mockEmailRepo, mockQueueRepo, logger)

	// Create Request with USER role
	req := httptest.NewRequest("GET", "/api/v1/admin/stats", nil)
	ctx := context.WithValue(req.Context(), middleware.UserRoleKey, "USER")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	// Execute
	handler.GetSystemStats(w, req)

	// Assert
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestGetSystemStats_RepoError(t *testing.T) {
	// Setup
	mockUserRepo := new(MockUserRepo)
	mockEmailRepo := new(MockEmailRepo)
	mockQueueRepo := new(MockQueueRepo)
	logger := observability.NewLogger("test", "v0.0.0")

	// Mock failure
	mockUserRepo.On("Count", mock.Anything).Return(nil, errors.New("db error"))

	handler := handlers.NewAdminStatsHandler(mockUserRepo, mockEmailRepo, mockQueueRepo, logger)

	req := httptest.NewRequest("GET", "/api/v1/admin/stats", nil)
	ctx := context.WithValue(req.Context(), middleware.UserRoleKey, "ADMIN")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	handler.GetSystemStats(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
