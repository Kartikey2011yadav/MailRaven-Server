package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/http/dto"
	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/http/handlers"
	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/http/middleware"
	domainSieve "github.com/Kartikey2011yadav/mailraven-server/internal/core/domain/sieve"
	"github.com/Kartikey2011yadav/mailraven-server/internal/observability"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockScriptRepository is a mock implementation of ports.ScriptRepository
type MockScriptRepository struct {
	mock.Mock
}

func (m *MockScriptRepository) Save(ctx context.Context, script *domainSieve.SieveScript) error {
	args := m.Called(ctx, script)
	return args.Error(0)
}

func (m *MockScriptRepository) Get(ctx context.Context, userID, name string) (*domainSieve.SieveScript, error) {
	args := m.Called(ctx, userID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainSieve.SieveScript), args.Error(1)
}

func (m *MockScriptRepository) GetActive(ctx context.Context, userID string) (*domainSieve.SieveScript, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainSieve.SieveScript), args.Error(1)
}

func (m *MockScriptRepository) List(ctx context.Context, userID string) ([]domainSieve.SieveScript, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]domainSieve.SieveScript), args.Error(1)
}

func (m *MockScriptRepository) SetActive(ctx context.Context, userID, name string) error {
	args := m.Called(ctx, userID, name)
	return args.Error(0)
}

func (m *MockScriptRepository) Delete(ctx context.Context, userID, name string) error {
	args := m.Called(ctx, userID, name)
	return args.Error(0)
}

func TestSieveHandler_ListScripts(t *testing.T) {
	mockRepo := new(MockScriptRepository)
	logger := observability.NewLogger("DEBUG", "text") // Simple logger for test
	handler := handlers.NewSieveHandler(mockRepo, logger)

	userID := "test@example.com"
	scripts := []domainSieve.SieveScript{
		{
			ID:        "1",
			UserID:    userID,
			Name:      "script1",
			Content:   "require ...",
			IsActive:  true,
			CreatedAt: time.Now(),
		},
	}

	mockRepo.On("List", mock.Anything, userID).Return(scripts, nil)

	req := httptest.NewRequest("GET", "/api/v1/sieve/scripts", nil)
	// Inject user into context
	ctx := context.WithValue(req.Context(), middleware.UserEmailKey, userID)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.ListScripts(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result []dto.SieveScriptResponse
	err := json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "script1", result[0].Name)
	assert.True(t, result[0].IsActive)
}

func TestSieveHandler_CreateScript(t *testing.T) {
	mockRepo := new(MockScriptRepository)
	logger := observability.NewLogger("DEBUG", "text")
	handler := handlers.NewSieveHandler(mockRepo, logger)

	userID := "test@example.com"
	reqPayload := dto.CreateSieveScriptRequest{
		Name:    "newscript",
		Content: "require \"fileinto\";",
	}

	mockRepo.On("Save", mock.Anything, mock.MatchedBy(func(s *domainSieve.SieveScript) bool {
		return s.UserID == userID && s.Name == "newscript" && s.Content == reqPayload.Content
	})).Return(nil)

	body, _ := json.Marshal(reqPayload)
	req := httptest.NewRequest("POST", "/api/v1/sieve/scripts", bytes.NewReader(body))
	ctx := context.WithValue(req.Context(), middleware.UserEmailKey, userID)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.CreateScript(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}
