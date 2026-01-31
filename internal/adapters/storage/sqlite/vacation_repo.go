package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type SqliteVacationRepository struct {
	db *sql.DB
}

func NewSqliteVacationRepository(db *sql.DB) *SqliteVacationRepository {
	return &SqliteVacationRepository{db: db}
}

func (r *SqliteVacationRepository) LastReply(ctx context.Context, userID, sender string) (time.Time, error) {
	var lastSent time.Time
	err := r.db.QueryRowContext(ctx, "SELECT last_sent_at FROM vacation_trackers WHERE user_id = ? AND sender_email = ?", userID, sender).Scan(&lastSent)
	if errors.Is(err, sql.ErrNoRows) {
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, err
	}
	return lastSent, nil
}

func (r *SqliteVacationRepository) RecordReply(ctx context.Context, userID, sender string) error {
	query := `
		INSERT INTO vacation_trackers (user_id, sender_email, last_sent_at)
		VALUES (?, ?, ?)
		ON CONFLICT(user_id, sender_email) DO UPDATE SET last_sent_at = excluded.last_sent_at
	`
	_, err := r.db.ExecContext(ctx, query, userID, sender, time.Now())
	return err
}
