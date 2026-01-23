package ports

import (
	"context"

	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
)

// SearchResult represents a search result with relevance score
type SearchResult struct {
	MessageID string  // Message ID
	Relevance float64 // BM25 relevance score (0-1, higher = more relevant)
}

// SearchIndex defines full-text search operations for messages
type SearchIndex interface {
	// Index adds or updates a message in the search index
	Index(ctx context.Context, msg *domain.Message, bodyText string) error

	// Search performs full-text query and returns ranked results
	// Query syntax: plain text or operators like "from:sender@example.com"
	// Results ordered by relevance (BM25 ranking)
	Search(ctx context.Context, userEmail, query string, limit, offset int) ([]SearchResult, error)

	// Delete removes a message from the index
	Delete(ctx context.Context, messageID string) error
}
