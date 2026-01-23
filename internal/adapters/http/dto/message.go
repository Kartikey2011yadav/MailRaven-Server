package dto

import (
	"time"

	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
)

// MessageSummary represents message metadata for list views (no body content)
type MessageSummary struct {
	ID          string    `json:"id"`
	Sender      string    `json:"sender"`
	Recipient   string    `json:"recipient"`
	Subject     string    `json:"subject"`
	Snippet     string    `json:"snippet"`
	ReadState   bool      `json:"read_state"`
	ReceivedAt  time.Time `json:"received_at"`
	SPFResult   string    `json:"spf_result"`
	DKIMResult  string    `json:"dkim_result"`
	DMARCResult string    `json:"dmarc_result"`
}

// MessageFull represents complete message with body content
type MessageFull struct {
	MessageSummary
	MessageID string `json:"message_id"`
	Body      string `json:"body"`
	BodySize  int64  `json:"body_size"`
}

// SearchResult extends MessageSummary with relevance scoring
type SearchResult struct {
	MessageSummary
	Relevance float64 `json:"relevance"`
}

// MessageListResponse for GET /v1/messages endpoint
type MessageListResponse struct {
	Messages []MessageSummary `json:"messages"`
	Total    int              `json:"total"`
	Limit    int              `json:"limit"`
	Offset   int              `json:"offset"`
	HasMore  bool             `json:"has_more"`
}

// MessagesSinceResponse for GET /v1/messages/since endpoint
type MessagesSinceResponse struct {
	Messages []MessageSummary `json:"messages"`
	Count    int              `json:"count"`
	Since    time.Time        `json:"since"`
}

// SearchResponse for GET /v1/messages/search endpoint
type SearchResponse struct {
	Results      []SearchResult `json:"results"`
	Query        string         `json:"query"`
	Count        int            `json:"count"`
	TotalMatches int            `json:"total_matches"`
}

// UpdateMessageRequest for PATCH /v1/messages/{id}
type UpdateMessageRequest struct {
	ReadState *bool `json:"read_state"`
}

// LoginRequest for POST /auth/login
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse for POST /auth/login
type LoginResponse struct {
	Token     string    `json:"token"`
	Email     string    `json:"email"`
	ExpiresAt time.Time `json:"expires_at"`
}

// ErrorResponse for 4xx/5xx responses
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// RateLimitResponse for 429 Too Many Requests
type RateLimitResponse struct {
	Error      string `json:"error"`
	Message    string `json:"message"`
	RetryAfter int    `json:"retry_after"`
}

// ToMessageSummary converts domain.Message to DTO
func ToMessageSummary(msg *domain.Message) MessageSummary {
	return MessageSummary{
		ID:          msg.ID,
		Sender:      msg.Sender,
		Recipient:   msg.Recipient,
		Subject:     msg.Subject,
		Snippet:     msg.Snippet,
		ReadState:   msg.ReadState,
		ReceivedAt:  msg.ReceivedAt,
		SPFResult:   msg.SPFResult,
		DKIMResult:  msg.DKIMResult,
		DMARCResult: msg.DMARCResult,
	}
}

// ToMessageFull converts domain.Message + body to DTO
func ToMessageFull(msg *domain.Message, body string, bodySize int64) MessageFull {
	return MessageFull{
		MessageSummary: ToMessageSummary(msg),
		MessageID:      msg.MessageID,
		Body:           body,
		BodySize:       bodySize,
	}
}

// ToSearchResult converts domain.Message + relevance to DTO
func ToSearchResult(msg *domain.Message, relevance float64) SearchResult {
	return SearchResult{
		MessageSummary: ToMessageSummary(msg),
		Relevance:      relevance,
	}
}
