package sqlite

import (
	"context"
	"database/sql"

	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/ports"
)

// SearchRepository implements ports.SearchIndex using SQLite FTS5
type SearchRepository struct {
	db *sql.DB
}

// NewSearchRepository creates a new SQLite FTS5 search repository
func NewSearchRepository(db *sql.DB) *SearchRepository {
	return &SearchRepository{db: db}
}

// Index adds or updates a message in the search index
func (r *SearchRepository) Index(ctx context.Context, msg *domain.Message, bodyText string) error {
	// Update FTS table with body text (trigger handles initial insert)
	query := `
		UPDATE messages_fts
		SET body_text = ?
		WHERE message_id = ?
	`

	_, err := r.db.ExecContext(ctx, query, bodyText, msg.ID)
	if err != nil {
		return ports.ErrStorageFailure
	}

	return nil
}

// Search performs full-text query and returns ranked results
func (r *SearchRepository) Search(ctx context.Context, userEmail, query string, limit, offset int) ([]ports.SearchResult, error) {
	// FTS5 query with recipient filter and BM25 ranking
	ftsQuery := `
		SELECT m.id, bm25(messages_fts) as relevance
		FROM messages_fts fts
		JOIN messages m ON fts.message_id = m.id
		WHERE fts.recipient = ? AND messages_fts MATCH ?
		ORDER BY bm25(messages_fts) DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, ftsQuery, userEmail, query, limit, offset)
	if err != nil {
		return nil, ports.ErrStorageFailure
	}
	defer rows.Close()

	var results []ports.SearchResult
	for rows.Next() {
		var result ports.SearchResult
		if err := rows.Scan(&result.MessageID, &result.Relevance); err != nil {
			return nil, ports.ErrStorageFailure
		}
		// Normalize relevance to 0-1 range (BM25 scores are typically negative)
		result.Relevance = 1.0 / (1.0 - result.Relevance)
		results = append(results, result)
	}

	return results, nil
}

// Delete removes a message from the index
func (r *SearchRepository) Delete(ctx context.Context, messageID string) error {
	// Trigger handles deletion when message is deleted from messages table
	// This method is for manual cleanup if needed
	query := `DELETE FROM messages_fts WHERE message_id = ?`

	_, err := r.db.ExecContext(ctx, query, messageID)
	if err != nil {
		return ports.ErrStorageFailure
	}

	return nil
}
