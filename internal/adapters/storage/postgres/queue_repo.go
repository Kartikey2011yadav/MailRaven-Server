package postgres

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
)

type QueueRepository struct {
	db *sql.DB
}

func NewQueueRepository(db *sql.DB) *QueueRepository {
	return &QueueRepository{db: db}
}

func (r *QueueRepository) Enqueue(ctx context.Context, msg *domain.OutboundMessage) error {
	query := `
		INSERT INTO queue (id, sender, recipient, blob_key, status, created_at, updated_at, next_retry_at, retry_count, last_error)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.ExecContext(ctx, query,
		msg.ID,
		msg.Sender,
		msg.Recipient,
		msg.BlobKey,
		msg.Status,
		msg.CreatedAt,
		msg.UpdatedAt,
		msg.NextRetryAt,
		msg.RetryCount,
		msg.LastError,
	)
	return err
}

func (r *QueueRepository) LockNextReady(ctx context.Context) (*domain.OutboundMessage, error) {
	// Use POSTGRES "UPDATE ... RETURNING" combined with "FOR UPDATE SKIP LOCKED"
	// This atomically finds, locks, and updates the next message.
	query := `
		UPDATE queue
		SET status = $1, updated_at = $2
		WHERE id = (
			SELECT id
			FROM queue
			WHERE status IN ('PENDING', 'RETRYING') AND next_retry_at <= $3
			ORDER BY next_retry_at ASC
			LIMIT 1
			FOR UPDATE SKIP LOCKED
		)
		RETURNING id, sender, recipient, blob_key, status, created_at, updated_at, next_retry_at, retry_count, last_error
	`

	now := time.Now().UTC()
	row := r.db.QueryRowContext(ctx, query, domain.QueueStatusProcessing, now, now)

	var msg domain.OutboundMessage
	var lastError sql.NullString
	var statusStr string

	err := row.Scan(
		&msg.ID, &msg.Sender, &msg.Recipient, &msg.BlobKey, &statusStr,
		&msg.CreatedAt, &msg.UpdatedAt, &msg.NextRetryAt, &msg.RetryCount, &lastError,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No messages ready
	}
	if err != nil {
		return nil, err
	}

	msg.Status = domain.OutboundStatus(statusStr)
	if lastError.Valid {
		msg.LastError = lastError.String
	}

	return &msg, nil
}

func (r *QueueRepository) UpdateStatus(ctx context.Context, id string, status domain.OutboundStatus, retryCount int, nextRetry time.Time, lastError string) error {
	query := `
		UPDATE queue 
		SET status = $1, retry_count = $2, next_retry_at = $3, last_error = $4, updated_at = $5
		WHERE id = $6
	`
	// lastError can be empty string, store as NULL if empty? Or just string.
	// SQLite impl stored whatever string passed.
	// If lastError is empty, pass NULL? Let's check logic. Usually empty string is fine.

	var lastErrorVal sql.NullString
	if lastError != "" {
		lastErrorVal.String = lastError
		lastErrorVal.Valid = true
	}

	_, err := r.db.ExecContext(ctx, query,
		status, retryCount, nextRetry, lastErrorVal, time.Now().UTC(), id,
	)
	return err
}

func (r *QueueRepository) Stats(ctx context.Context) (pending, processing, failed, completed int64, err error) {
	query := `
		SELECT status, COUNT(*)
		FROM queue
		GROUP BY status
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var count int64
		if err := rows.Scan(&status, &count); err != nil {
			return 0, 0, 0, 0, err
		}

		// Status strings might be upper or lower case depending on how they were inserted?
		// Usually domain constants are UPPERCASE.
		switch domain.OutboundStatus(strings.ToUpper(status)) {
		case domain.QueueStatusPending, domain.QueueStatusRetrying:
			pending += count // Retrying counts as pending usually?
			// Wait, interface asks for "pending, processing, failed, completed".
			// Retrying effectively is Pending delivery.
		case domain.QueueStatusProcessing:
			processing += count
		case domain.QueueStatusFailed: // PERMANENTLY FAILED
			failed += count
		case domain.QueueStatusSent:
			completed += count
		}
	}
	return
}
