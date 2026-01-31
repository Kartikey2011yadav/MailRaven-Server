package tests

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/http/handlers"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
	"github.com/Kartikey2011yadav/mailraven-server/internal/observability"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTLSRptRepo
type MockTLSRptRepo struct {
	mock.Mock
}

func (m *MockTLSRptRepo) Save(ctx context.Context, report *domain.TLSReport) error {
	args := m.Called(ctx, report)
	return args.Error(0)
}

func (m *MockTLSRptRepo) FindLatest(ctx context.Context, limit int) ([]*domain.TLSReport, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]*domain.TLSReport), args.Error(1)
}

func TestTLSRPT_Ingestion(t *testing.T) {
	// Setup
	mockRepo := &MockTLSRptRepo{}
	logger := observability.NewLogger("test", "error")
	handler := handlers.NewTLSRptHandler(mockRepo, logger)

	// Sample JSON body (from RFC example or similar)
	jsonBody := `{
		"organization-name": "Google",
		"date-range": {
			"start-datetime": "2026-01-01T00:00:00Z",
			"end-datetime": "2026-01-01T23:59:59Z"
		},
		"contact-info": "smtp-tls-reporting@google.com",
		"report-id": "2026-01-01-google-mailraven",
		"policies": [
			{
				"policy": {
					"policy-type": "sts",
					"policy-domain": "example.com",
					"mx-host": ["mx.example.com"]
				},
				"summary": {
					"total-successful-session-count": 50,
					"total-failure-session-count": 0
				}
			}
		]
	}`

	// Expectation
	mockRepo.On("Save", mock.Anything, mock.MatchedBy(func(r *domain.TLSReport) bool {
		return r.Provider == "Google" && r.TotalCount == 50
	})).Return(nil)

	// Request
	req := httptest.NewRequest("POST", "/.well-known/tlsrpt", bytes.NewBufferString(jsonBody))
	req.Header.Set("Content-Type", "application/tlsrpt+json")
	w := httptest.NewRecorder()

	handler.HandleReport(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	mockRepo.AssertExpectations(t)
}
