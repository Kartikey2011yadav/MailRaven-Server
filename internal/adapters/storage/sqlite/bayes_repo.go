package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/ports"
)

// BayesRepository implements ports.BayesRepository using SQLite
type BayesRepository struct {
	db *sql.DB
}

// NewBayesRepository creates a new SQLite bayes repository
func NewBayesRepository(db *sql.DB) *BayesRepository {
	return &BayesRepository{db: db}
}

// GetTokens fetches multiple tokens in a single operation
func (r *BayesRepository) GetTokens(ctx context.Context, tokens []string) (map[string]*domain.BayesToken, error) {
	result := make(map[string]*domain.BayesToken)
	if len(tokens) == 0 {
		return result, nil
	}

	// Batch in chunks of 500 to avoid SQLite variable limit (usually 999 or 32000)
	chunkSize := 500
	for i := 0; i < len(tokens); i += chunkSize {
		end := i + chunkSize
		if end > len(tokens) {
			end = len(tokens)
		}
		chunk := tokens[i:end]

		if err := r.getTokensBatch(ctx, chunk, result); err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (r *BayesRepository) getTokensBatch(ctx context.Context, tokens []string, result map[string]*domain.BayesToken) error {
	queryBuilder := strings.Builder{}
	queryBuilder.WriteString("SELECT token, spam_count, ham_count FROM bayes_tokens WHERE token IN (")

	args := make([]interface{}, len(tokens))
	for i, token := range tokens {
		if i > 0 {
			queryBuilder.WriteString(",")
		}
		queryBuilder.WriteString("?")
		args[i] = token
	}
	queryBuilder.WriteString(")")

	rows, err := r.db.QueryContext(ctx, queryBuilder.String(), args...)
	if err != nil {
		return fmt.Errorf("bayes tokens query failed: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var t domain.BayesToken
		if err := rows.Scan(&t.Token, &t.SpamCount, &t.HamCount); err != nil {
			return fmt.Errorf("scan failed: %w", err)
		}
		result[t.Token] = &t
	}
	return rows.Err()
}

// IncrementToken adds to the spam or ham count of a token (Upsert)
func (r *BayesRepository) IncrementToken(ctx context.Context, token string, isSpam bool) error {
	spamInc := 0
	hamInc := 0
	if isSpam {
		spamInc = 1
	} else {
		hamInc = 1
	}

	query := `
        INSERT INTO bayes_tokens (token, spam_count, ham_count)
        VALUES (?, ?, ?)
        ON CONFLICT(token) DO UPDATE SET
            spam_count = spam_count + excluded.spam_count,
            ham_count = ham_count + excluded.ham_count
    `
	_, err := r.db.ExecContext(ctx, query, token, spamInc, hamInc)
	if err != nil {
		return fmt.Errorf("increment token failed: %w", err)
	}
	return nil
}

// GetGlobalStats retrieves the total spam/ham counts
func (r *BayesRepository) GetGlobalStats(ctx context.Context) (*domain.BayesGlobalStats, error) {
	// Initialize defaults
	stats := &domain.BayesGlobalStats{TotalSpam: 0, TotalHam: 0}

	query := `SELECT key, value FROM bayes_global WHERE key IN ('spam_total', 'ham_total')`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get global stats failed: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var key string
		var val int
		if err := rows.Scan(&key, &val); err != nil {
			return nil, err
		}
		if key == "spam_total" {
			stats.TotalSpam = val
		} else if key == "ham_total" {
			stats.TotalHam = val
		}
	}
	return stats, nil
}

// IncrementGlobal adds to the global spam/ham counters
func (r *BayesRepository) IncrementGlobal(ctx context.Context, isSpam bool) error {
	key := "ham_total"
	if isSpam {
		key = "spam_total"
	}

	query := `
        INSERT INTO bayes_global (key, value)
        VALUES (?, 1)
        ON CONFLICT(key) DO UPDATE SET value = value + 1
    `
	_, err := r.db.ExecContext(ctx, query, key)
	if err != nil {
		return fmt.Errorf("increment global stats failed: %w", err)
	}
	return nil
}

// Check compliance
var _ ports.BayesRepository = (*BayesRepository)(nil)
