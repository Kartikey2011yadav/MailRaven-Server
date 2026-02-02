package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockEmailRepository implements ports.EmailRepository for testing
type MockEmailRepository struct {
	mock.Mock
}

func (m *MockEmailRepository) Save(ctx context.Context, msg *domain.Message) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}
func (m *MockEmailRepository) FindByID(ctx context.Context, id string) (*domain.Message, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Message), args.Error(1)
}
func (m *MockEmailRepository) FindByUser(ctx context.Context, email string, limit, offset int) ([]*domain.Message, error) {
	args := m.Called(ctx, email, limit, offset)
	return args.Get(0).([]*domain.Message), args.Error(1)
}
func (m *MockEmailRepository) UpdateReadState(ctx context.Context, id string, read bool) error {
	return m.Called(ctx, id, read).Error(0)
}
func (m *MockEmailRepository) CountByUser(ctx context.Context, email string) (int, error) {
	args := m.Called(ctx, email)
	return args.Int(0), args.Error(1)
}
func (m *MockEmailRepository) CountTotal(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}
func (m *MockEmailRepository) FindSince(ctx context.Context, email string, since time.Time, limit int) ([]*domain.Message, error) {
	args := m.Called(ctx, email, since, limit)
	return args.Get(0).([]*domain.Message), args.Error(1)
}
func (m *MockEmailRepository) GetMailbox(ctx context.Context, userID, name string) (*domain.Mailbox, error) {
	args := m.Called(ctx, userID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Mailbox), args.Error(1)
}
func (m *MockEmailRepository) CreateMailbox(ctx context.Context, userID, name string) error {
	return m.Called(ctx, userID, name).Error(0)
}
func (m *MockEmailRepository) ListMailboxes(ctx context.Context, userID string) ([]*domain.Mailbox, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*domain.Mailbox), args.Error(1)
}
func (m *MockEmailRepository) FindByUIDRange(ctx context.Context, userID, mailbox string, min, max uint32) ([]*domain.Message, error) {
	args := m.Called(ctx, userID, mailbox, min, max)
	return args.Get(0).([]*domain.Message), args.Error(1)
}
func (m *MockEmailRepository) CopyMessages(ctx context.Context, userID string, messageIDs []string, destMailbox string) error {
	return m.Called(ctx, userID, messageIDs, destMailbox).Error(0)
}
func (m *MockEmailRepository) SetACL(ctx context.Context, userID, mailboxName, identifier, rights string) error {
	return m.Called(ctx, userID, mailboxName, identifier, rights).Error(0)
}
func (m *MockEmailRepository) AddFlags(ctx context.Context, messageID string, flags ...string) error {
	return m.Called(ctx, messageID, flags).Error(0)
}
func (m *MockEmailRepository) RemoveFlags(ctx context.Context, messageID string, flags ...string) error {
	return m.Called(ctx, messageID, flags).Error(0)
}
func (m *MockEmailRepository) SetFlags(ctx context.Context, messageID string, flags ...string) error {
	return m.Called(ctx, messageID, flags).Error(0)
}
func (m *MockEmailRepository) AssignUID(ctx context.Context, messageID string, mailboxName string) (uint32, error) {
    args := m.Called(ctx, messageID, mailboxName)
    return uint32(args.Int(0)), args.Error(1)
}

func TestUpdateACL(t *testing.T) {
	mockRepo := new(MockEmailRepository)
	service := NewEmailService(mockRepo)
	ctx := context.Background()

	t.Run("Valid rights", func(t *testing.T) {
		mockRepo.On("GetMailbox", ctx, "owner", "INBOX").Return(&domain.Mailbox{}, nil)
		mockRepo.On("SetACL", ctx, "owner", "INBOX", "user1", "lr").Return(nil)

		err := service.UpdateACL(ctx, "owner", "INBOX", "user1", "lr")
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Invalid rights char", func(t *testing.T) {
		err := service.UpdateACL(ctx, "owner", "INBOX", "user1", "lz") // z is invalid
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid right: z")
	})

	t.Run("Mailbox not found", func(t *testing.T) {
		mockRepo.On("GetMailbox", ctx, "owner", "MISSING").Return(nil, errors.New("not found"))
		
		err := service.UpdateACL(ctx, "owner", "MISSING", "user1", "l")
		assert.Error(t, err)
		assert.Equal(t, "not found", err.Error())
	})
}

func TestCheckAccess(t *testing.T) {
	mockRepo := new(MockEmailRepository)
	service := NewEmailService(mockRepo)
	ctx := context.Background()

	t.Run("Owner access", func(t *testing.T) {
		err := service.CheckAccess(ctx, "owner", "INBOX", "owner", "l")
		assert.NoError(t, err)
	})

	t.Run("Specific user access", func(t *testing.T) {
		mailbox := &domain.Mailbox{
			ACL: map[string]string{
				"bob": "lr",
			},
		}
		mockRepo.On("GetMailbox", ctx, "alice", "INBOX").Return(mailbox, nil)

		// Bob has 'r' right? Yes
		err := service.CheckAccess(ctx, "alice", "INBOX", "bob", "r")
		assert.NoError(t, err)

		// Bob has 'w' right? No
		mockRepo.On("GetMailbox", ctx, "alice", "INBOX").Return(mailbox, nil)
		err = service.CheckAccess(ctx, "alice", "INBOX", "bob", "w")
		assert.Error(t, err)
		assert.Equal(t, ErrAccessDenied, err)
	})

	t.Run("Anyone access", func(t *testing.T) {
		mailbox := &domain.Mailbox{
			ACL: map[string]string{
				"anyone": "l",
			},
		}
		mockRepo.On("GetMailbox", ctx, "alice", "Public").Return(mailbox, nil)

		err := service.CheckAccess(ctx, "alice", "Public", "random", "l")
		assert.NoError(t, err)
	})
}
