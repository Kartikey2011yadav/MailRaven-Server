package managesieve

import (
	"bufio"
	"context"
	"strings"
	"testing"

	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain" // needed for User
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain/sieve"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) Save(ctx context.Context, script *sieve.SieveScript) error {
	return m.Called(ctx, script).Error(0)
}
func (m *MockRepo) Get(ctx context.Context, userID, name string) (*sieve.SieveScript, error) {
	args := m.Called(ctx, userID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sieve.SieveScript), args.Error(1)
}
func (m *MockRepo) GetActive(ctx context.Context, userID string) (*sieve.SieveScript, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sieve.SieveScript), args.Error(1)
}
func (m *MockRepo) List(ctx context.Context, userID string) ([]sieve.SieveScript, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]sieve.SieveScript), args.Error(1)
}
func (m *MockRepo) SetActive(ctx context.Context, userID, name string) error {
	return m.Called(ctx, userID, name).Error(0)
}
func (m *MockRepo) Delete(ctx context.Context, userID, name string) error {
	return m.Called(ctx, userID, name).Error(0)
}

type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) Create(ctx context.Context, user *domain.User) error { return nil }
func (m *MockUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	return nil, nil
}
func (m *MockUserRepo) Authenticate(ctx context.Context, email, pass string) (*domain.User, error) {
	args := m.Called(ctx, email, pass)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}
func (m *MockUserRepo) UpdateLastLogin(ctx context.Context, email string) error { return nil }
func (m *MockUserRepo) List(ctx context.Context, limit, offset int) ([]*domain.User, error) {
	return nil, nil
}
func (m *MockUserRepo) Delete(ctx context.Context, email string) error            { return nil }
func (m *MockUserRepo) UpdatePassword(ctx context.Context, email, h string) error { return nil }
func (m *MockUserRepo) UpdateRole(ctx context.Context, email string, role domain.Role) error {
	return nil
}
func (m *MockUserRepo) Count(ctx context.Context) (map[string]int64, error) { return nil, nil }

func TestTokenizer(t *testing.T) {
	input := "\"test\" {4}\r\nabcd active"
	r := bufio.NewReader(strings.NewReader(input))
	tokenizer := NewTokenizer(r)

	w1, err := tokenizer.ReadWord()
	assert.NoError(t, err)
	assert.Equal(t, "test", w1)

	w2, err := tokenizer.ReadWord()
	assert.NoError(t, err)
	assert.Equal(t, "abcd", w2)

	w3, err := tokenizer.ReadWord()
	assert.NoError(t, err)
	assert.Equal(t, "active", w3)
}
