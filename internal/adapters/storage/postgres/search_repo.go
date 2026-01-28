package postgres

import (
	"context"
	"database/sql"

	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/ports"
)

// SearchRepository implements ports.SearchIndex using PostgreSQL TSVECTOR
type SearchRepository struct {
	db *sql.DB
}

// NewSearchRepository creates a new PostgreSQL search repository
func NewSearchRepository(db *sql.DB) *SearchRepository {
	return &SearchRepository{db: db}
}

// Index adds or updates a message in the search index
func (r *SearchRepository) Index(ctx context.Context, msg *domain.Message, bodyText string) error {
	// Insert into search table with weight (Subject: A, Body: B)
	query := `
		INSERT INTO messages_search (message_id, tsv)
		VALUES ($1, setweight(to_tsvector('english', coalesce($2, '')), 'A') || 
		            setweight(to_tsvector('english', coalesce($3, '')), 'B'))
		ON CONFLICT (message_id) DO UPDATE 
		SET tsv = EXCLUDED.tsv
	`

	_, err := r.db.ExecContext(ctx, query, msg.ID, msg.Subject, bodyText)
	if err != nil {
		return ports.ErrStorageFailure
	}
	return nil
}

// Search performs full-text query and returns ranked results
func (r *SearchRepository) Search(ctx context.Context, userEmail, queryText string, limit, offset int) ([]ports.SearchResult, error) {
	// Full text search with rank
	query := `
		SELECT m.id, ts_rank(s.tsv, plainto_tsquery('english', $2)) as rank
		FROM messages m
		JOIN messages_search s ON m.id = s.message_id
		WHERE m.recipient = $1 AND s.tsv @@ plainto_tsquery('english', $2)
		ORDER BY rank DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.QueryContext(ctx, query, userEmail, queryText, limit, offset)
	if err != nil {
		return nil, ports.ErrStorageFailure
	}
	defer rows.Close()

	var results []ports.SearchResult
	for rows.Next() {
		var res ports.SearchResult
		if err := rows.Scan(&res.MessageID, &res.Relevance); err != nil {
			return nil, ports.ErrStorageFailure
		}
		results = append(results, res)
	}
	return results, nil
}
