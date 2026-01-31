package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/http/handlers"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
	"github.com/stretchr/testify/assert"
)

func TestMTASTS_Serving(t *testing.T) {
	// Setup simple policy
	policy := &domain.MTASTSPolicy{
		Version: "STSv1",
		Mode:    domain.MTASTSModeTesting,
		MX:      []string{"mail.example.com"},
		MaxAge:  3600,
	}

	handler := handlers.NewMTASTSHandler(policy)

	// Setup request
	req := httptest.NewRequest("GET", "/.well-known/mta-sts.txt", nil)
	req.Host = "mta-sts.example.com"
	w := httptest.NewRecorder()

	handler.ServePolicy(w, req)

	// Assertions
	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "text/plain", resp.Header.Get("Content-Type"))

	// Check body content
	bodyString := w.Body.String()
	assert.Contains(t, bodyString, "version: STSv1")
	assert.Contains(t, bodyString, "mode: testing")
	assert.Contains(t, bodyString, "mx: mail.example.com")
	assert.Contains(t, bodyString, "max_age: 3600")
}
