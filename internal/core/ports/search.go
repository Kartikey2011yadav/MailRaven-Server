package ports

import (
	"context"

	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
)

// SearchIndex defines full-text search operations for messages
type SearchIndex interface {
	// Index adds or updates a message in the search index
	Index(ctx context.Context, msg *domain.Message, bodyText string) error

	// Search performs full-text query and returns ranked results
	// Query syntax: plain text or operators like "from:sender@example.com"
	// Results ordered by relevance (BM25 ranking)
	Search(ctx context.Context, userEmail, query string, limit, offset int) ([]*domain.Message, error)

	// Delete removes a message from the index
	Delete(ctx context.Context, messageID string) error
}
