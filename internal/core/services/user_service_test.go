package services

import (
	"context"
	"testing"

	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a manual mock for testing
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}
func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}
func (m *MockUserRepository) Authenticate(ctx context.Context, email, password string) (*domain.User, error) {
	args := m.Called(ctx, email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}
func (m *MockUserRepository) UpdateLastLogin(ctx context.Context, email string) error {
	return m.Called(ctx, email).Error(0)
}
func (m *MockUserRepository) List(ctx context.Context, limit, offset int) ([]*domain.User, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*domain.User), args.Error(1)
}
func (m *MockUserRepository) Delete(ctx context.Context, email string) error {
	return m.Called(ctx, email).Error(0)
}
func (m *MockUserRepository) UpdatePassword(ctx context.Context, email, passwordHash string) error {
	return m.Called(ctx, email, passwordHash).Error(0)
}
func (m *MockUserRepository) UpdateRole(ctx context.Context, email string, role domain.Role) error {
	return m.Called(ctx, email, role).Error(0)
}
func (m *MockUserRepository) Count(ctx context.Context) (map[string]int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(map[string]int64), args.Error(1)
}
func (m *MockUserRepository) UpdateQuota(ctx context.Context, email string, bytes int64) error {
	return m.Called(ctx, email, bytes).Error(0)
}

func (m *MockUserRepository) IncrementStorageUsed(ctx context.Context, email string, delta int64) error {
	return m.Called(ctx, email, delta).Error(0)
}

func TestUpdateQuota(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	// Test Case 1: Negative Quota
	err := service.UpdateQuota(context.Background(), "user@example.com", -1)
	assert.Error(t, err)
	assert.Equal(t, "quota cannot be negative", err.Error())

	// Test Case 2: Valid Quota
	mockRepo.On("UpdateQuota", mock.Anything, "user@example.com", int64(1024)).Return(nil)
	err = service.UpdateQuota(context.Background(), "user@example.com", 1024)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}
