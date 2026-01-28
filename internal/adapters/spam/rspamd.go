package spam

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// CheckResult represents the response from Rspamd
type CheckResult struct {
	IsSkipped     bool              `json:"is_skipped"`
	Score         float64           `json:"score"`
	RequiredScore float64           `json:"required_score"`
	Action        string            `json:"action"`
	Symbols       map[string]Symbol `json:"symbols"`
	Urls          []string          `json:"urls"`
	Emails        []string          `json:"emails"`
	MessageID     string            `json:"message-id"`
}

// Symbol represents a single spam symbol result
type Symbol struct {
	Name        string  `json:"name"`
	Score       float64 `json:"score"`
	MetricScore float64 `json:"metric_score"`
	Description string  `json:"description"`
}

// Client interacts with Rspamd API
type Client struct {
	url        string
	httpClient *http.Client
}

// NewClient creates a new Rspamd client
func NewClient(url string) *Client {
	return &Client{
		url: url,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Check scans a message stream
// headers can include: IP, Helo, User, From, Rcpt, Queue-Id
func (c *Client) Check(r io.Reader, headers map[string]string) (*CheckResult, error) {
	req, err := http.NewRequest("POST", c.url+"/checkv2", r)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers (Queue-ID, IP, From, Rcpt, etc)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// Rspamd expects the message body as POST body
	req.Header.Set("Content-Type", "application/x-mime")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("rspamd request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("rspamd error %d: %s", resp.StatusCode, string(body))
	}

	var result CheckResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}
