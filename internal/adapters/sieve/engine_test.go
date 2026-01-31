package sieve

import (
	"context"
	"testing"
	"time"

	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain/sieve"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type Mocks struct {
	ScriptRepo   *MockScriptRepo
	MailboxRepo  *MockMailboxRepo
	VacationRepo *MockVacationRepo
}

func (m *Mocks) AssertExpectations(t *testing.T) {
	m.ScriptRepo.AssertExpectations(t)
	m.MailboxRepo.AssertExpectations(t)
	m.VacationRepo.AssertExpectations(t)
}

func TestSieveEngine_Execute(t *testing.T) {
	tests := []struct {
		name            string
		scriptContent   string
		msg             string
		mockSetup       func(*Mocks)
		expectedTargets []string
	}{
		{
			name:          "Implicit Keep",
			scriptContent: `if header :contains "Subject" "spam" { discard; }`,
			msg:           "Subject: Hello\r\n\r\nBody",
			mockSetup: func(m *Mocks) {
				m.ScriptRepo.On("GetActive", mock.Anything, "u1").Return(&sieve.SieveScript{
					Content: `if header :contains "Subject" "spam" { discard; }`,
				}, nil)
			},
			expectedTargets: []string{"INBOX"},
		},
		{
			name:          "Discard",
			scriptContent: `discard;`,
			msg:           "Subject: spam\r\n\r\nBody",
			mockSetup: func(m *Mocks) {
				m.ScriptRepo.On("GetActive", mock.Anything, "u1").Return(&sieve.SieveScript{
					Content: `discard;`,
				}, nil)
			},
			expectedTargets: []string{}, // Empty = discard
		},
		{
			name:          "FileInto",
			scriptContent: `require "fileinto"; fileinto "Junk";`,
			msg:           "Subject: spam\r\n\r\nBody",
			mockSetup: func(m *Mocks) {
				m.ScriptRepo.On("GetActive", mock.Anything, "u1").Return(&sieve.SieveScript{
					Content: `require "fileinto"; fileinto "Junk";`,
				}, nil)
				m.MailboxRepo.On("CreateMailbox", mock.Anything, "u1", "Junk").Return(nil)
			},
			expectedTargets: []string{"Junk"},
		},
		{
			name:          "Keep Explicit",
			scriptContent: `keep;`,
			msg:           "Subject: hi\r\n\r\nBody",
			mockSetup: func(m *Mocks) {
				m.ScriptRepo.On("GetActive", mock.Anything, "u1").Return(&sieve.SieveScript{
					Content: `keep;`,
				}, nil)
			},
			expectedTargets: []string{"INBOX"},
		},
		{
			name:          "Error Runtime (Fail Open)",
			scriptContent: `if header :matches "Subject" "*[" { discard; }`, // Invalid regex might not fail parse in go-sieve but runtime?
			// go-sieve parser is strict. Let's use something valid but runtime failing...
			// Hard to induce runtime error in pure sieve without bad extensions.
			// Let's assume script parser failure.
			msg: "Subject: hi",
			mockSetup: func(m *Mocks) {
				m.ScriptRepo.On("GetActive", mock.Anything, "u1").Return(&sieve.SieveScript{
					Content: `invalid syntax ???`,
				}, nil)
			},
			expectedTargets: []string{"INBOX"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mocks := &Mocks{
				ScriptRepo:   new(MockScriptRepo),
				MailboxRepo:  new(MockMailboxRepo),
				VacationRepo: new(MockVacationRepo),
			}
			if tt.mockSetup != nil {
				tt.mockSetup(mocks)
			}

			engine := NewSieveEngine(mocks.ScriptRepo, mocks.MailboxRepo, mocks.VacationRepo, nil, nil)
			targets, err := engine.Execute(context.TODO(), "u1", []byte(tt.msg))

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedTargets, targets)
			mocks.AssertExpectations(t)
		})
	}
}

// Mocks Implementation

type MockScriptRepo struct {
	mock.Mock
}

func (m *MockScriptRepo) GetActive(ctx context.Context, userID string) (*sieve.SieveScript, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sieve.SieveScript), args.Error(1)
}

func (m *MockScriptRepo) Save(ctx context.Context, script *sieve.SieveScript) error { return nil }
func (m *MockScriptRepo) Get(ctx context.Context, userID, name string) (*sieve.SieveScript, error) {
	return nil, nil
}
func (m *MockScriptRepo) List(ctx context.Context, userID string) ([]sieve.SieveScript, error) {
	return nil, nil
}
func (m *MockScriptRepo) SetActive(ctx context.Context, userID, name string) error { return nil }
func (m *MockScriptRepo) Delete(ctx context.Context, userID, name string) error    { return nil }

type MockMailboxRepo struct {
	mock.Mock
}

func (m *MockMailboxRepo) CreateMailbox(ctx context.Context, userID, name string) error {
	args := m.Called(ctx, userID, name)
	return args.Error(0)
}

func (m *MockMailboxRepo) Save(ctx context.Context, msg *domain.Message) error { return nil }
func (m *MockMailboxRepo) FindByID(ctx context.Context, id string) (*domain.Message, error) {
	return nil, nil
}
func (m *MockMailboxRepo) FindByUser(ctx context.Context, email string, limit, offset int) ([]*domain.Message, error) {
	return nil, nil
}
func (m *MockMailboxRepo) UpdateReadState(ctx context.Context, id string, read bool) error {
	return nil
}
func (m *MockMailboxRepo) CountByUser(ctx context.Context, email string) (int, error) { return 0, nil }
func (m *MockMailboxRepo) CountTotal(ctx context.Context) (int64, error)              { return 0, nil }
func (m *MockMailboxRepo) FindSince(ctx context.Context, email string, since time.Time, limit int) ([]*domain.Message, error) {
	return nil, nil
}
func (m *MockMailboxRepo) GetMailbox(ctx context.Context, userID, name string) (*domain.Mailbox, error) {
	return nil, nil
}
func (m *MockMailboxRepo) ListMailboxes(ctx context.Context, userID string) ([]*domain.Mailbox, error) {
	return nil, nil
}
func (m *MockMailboxRepo) FindByUIDRange(ctx context.Context, userID, mailbox string, min, max uint32) ([]*domain.Message, error) {
	return nil, nil
}
func (m *MockMailboxRepo) CopyMessages(ctx context.Context, userID string, messageIDs []string, destMailbox string) error {
	return nil
}
func (m *MockMailboxRepo) AddFlags(ctx context.Context, messageID string, flags ...string) error {
	return nil
}
func (m *MockMailboxRepo) RemoveFlags(ctx context.Context, messageID string, flags ...string) error {
	return nil
}
func (m *MockMailboxRepo) SetFlags(ctx context.Context, messageID string, flags ...string) error {
	return nil
}
func (m *MockMailboxRepo) AssignUID(ctx context.Context, messageID string, mailbox string) (uint32, error) {
	return 0, nil
}

type MockVacationRepo struct {
	mock.Mock
}

func (m *MockVacationRepo) LastReply(ctx context.Context, userID, sender string) (time.Time, error) {
	args := m.Called(ctx, userID, sender)
	return args.Get(0).(time.Time), args.Error(1)
}

func (m *MockVacationRepo) RecordReply(ctx context.Context, userID, sender string) error {
	args := m.Called(ctx, userID, sender)
	return args.Error(0)
}
