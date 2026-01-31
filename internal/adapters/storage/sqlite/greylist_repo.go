package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/ports"
)

// GreylistRepository implements ports.GreylistRepository using SQLite
type GreylistRepository struct {
	db *sql.DB
}

// NewGreylistRepository creates a new SQLite greylist repository
func NewGreylistRepository(db *sql.DB) *GreylistRepository {
	return &GreylistRepository{db: db}
}

// Get retrieves a greylist entry by tuple
func (r *GreylistRepository) Get(ctx context.Context, tuple domain.GreylistTuple) (*domain.GreylistEntry, error) {
	query := `
        SELECT first_seen_unix, last_seen_unix, blocked_count
        FROM greylist
        WHERE ip_net = ? AND sender = ? AND recipient = ?
    `
	row := r.db.QueryRowContext(ctx, query, tuple.IPNet, tuple.Sender, tuple.Recipient)

	entry := &domain.GreylistEntry{
		Tuple: tuple,
	}

	err := row.Scan(&entry.FirstSeenAt, &entry.LastSeenAt, &entry.BlockedCount)
	if err == sql.ErrNoRows {
		// Return nil if not found, consistent with other lookup methods that might return nil
		// However, standard repo pattern often returns ErrNotFound.
		// For Check(), receiving nil entry is a valid "not found" state.
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("greylist lookup failed: %w", err)
	}

	return entry, nil
}

// Upsert creates or updates a greylist entry
func (r *GreylistRepository) Upsert(ctx context.Context, entry *domain.GreylistEntry) error {
	query := `
        INSERT INTO greylist (ip_net, sender, recipient, first_seen_unix, last_seen_unix, blocked_count)
        VALUES (?, ?, ?, ?, ?, ?)
        ON CONFLICT(ip_net, sender, recipient) DO UPDATE SET
            last_seen_unix = excluded.last_seen_unix,
            blocked_count = excluded.blocked_count
    `
	// Note: We deliberately do NOT update first_seen_unix on conflict,
	// to preserve the original first contact time.

	_, err := r.db.ExecContext(ctx, query,
		entry.Tuple.IPNet,
		entry.Tuple.Sender,
		entry.Tuple.Recipient,
		entry.FirstSeenAt,
		entry.LastSeenAt,
		entry.BlockedCount,
	)
	if err != nil {
		return fmt.Errorf("greylist upsert failed: %w", err)
	}
	return nil
}

// DeleteOlderThan prunes entries that haven't been seen since timestamp
func (r *GreylistRepository) DeleteOlderThan(ctx context.Context, timestamp int64) (int64, error) {
	query := `DELETE FROM greylist WHERE last_seen_unix < ?`
	result, err := r.db.ExecContext(ctx, query, timestamp)
	if err != nil {
		return 0, fmt.Errorf("greylist pruning failed: %w", err)
	}
	return result.RowsAffected()
}

// Check compliance
var _ ports.GreylistRepository = (*GreylistRepository)(nil)
