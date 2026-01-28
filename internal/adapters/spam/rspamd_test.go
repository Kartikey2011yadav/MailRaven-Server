package spam

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRspamdCheck(t *testing.T) {
	// Mock Rspamd Server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify expected headers (optional)
		if r.Header.Get("IP") != "1.2.3.4" {
			t.Errorf("Expected IP header 1.2.3.4, got %s", r.Header.Get("IP"))
		}

		// Return mock response
		response := `{
			"is_skipped": false,
			"score": 10.5,
			"required_score": 15.0,
			"action": "add header",
			"symbols": {
				"GTUBE": {"name": "GTUBE", "score": 100.0}
			}
		}`
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(response))
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	headers := map[string]string{"IP": "1.2.3.4"}
	content := bytes.NewReader([]byte("test email content"))

	result, err := client.Check(content, headers)
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}

	if result.Score != 10.5 {
		t.Errorf("Expected score 10.5, got %f", result.Score)
	}
	if result.Action != "add header" {
		t.Errorf("Expected action 'add header', got %s", result.Action)
	}
}
