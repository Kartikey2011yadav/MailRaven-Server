package sqlite

import (
	"context"
	"database/sql"
	"errors"
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
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		msg.ID,
		msg.Sender,
		msg.Recipient,
		msg.BlobKey,
		msg.Status,
		msg.CreatedAt.Unix(),
		msg.UpdatedAt.Unix(),
		msg.NextRetryAt.Unix(),
		msg.RetryCount,
		msg.LastError,
	)
	return err
}

func (r *QueueRepository) LockNextReady(ctx context.Context) (*domain.OutboundMessage, error) {
	// Start transaction to ensure atomicity
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Find next ready message
	// Status must be PENDING or RETRYING, and NextRetryAt must be in the past
	row := tx.QueryRowContext(ctx, `
		SELECT id, sender, recipient, blob_key, status, created_at, updated_at, next_retry_at, retry_count, last_error
		FROM queue
		WHERE status IN ('PENDING', 'RETRYING') AND next_retry_at <= ?
		ORDER BY next_retry_at ASC
		LIMIT 1
	`, time.Now().Unix())

	var msg domain.OutboundMessage
	var createdAt, updatedAt, nextRetryAt int64
	var statusStr string
	// sql.NullString for last_error since it can be NULL
	var lastError sql.NullString

	err = row.Scan(
		&msg.ID, &msg.Sender, &msg.Recipient, &msg.BlobKey, &statusStr,
		&createdAt, &updatedAt, &nextRetryAt, &msg.RetryCount, &lastError,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil // No messages ready
	}
	if err != nil {
		return nil, err
	}

	msg.Status = domain.OutboundStatus(statusStr)
	msg.CreatedAt = time.Unix(createdAt, 0).UTC()
	msg.UpdatedAt = time.Unix(updatedAt, 0).UTC()
	msg.NextRetryAt = time.Unix(nextRetryAt, 0).UTC()
	if lastError.Valid {
		msg.LastError = lastError.String
	}

	// Lock it by setting status to PROCESSING
	updateQuery := `UPDATE queue SET status = ?, updated_at = ? WHERE id = ?`
	_, err = tx.ExecContext(ctx, updateQuery, domain.QueueStatusProcessing, time.Now().Unix(), msg.ID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// Update returned struct
	msg.Status = domain.QueueStatusProcessing

	return &msg, nil
}

func (r *QueueRepository) UpdateStatus(ctx context.Context, id string, status domain.OutboundStatus, retryCount int, nextRetry time.Time, lastError string) error {
	query := `
		UPDATE queue 
		SET status = ?, updated_at = ?, retry_count = ?, next_retry_at = ?, last_error = ?
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query,
		status,
		time.Now().Unix(),
		retryCount,
		nextRetry.Unix(),
		lastError,
		id,
	)
	return err
}

// Stats returns queue statistics
func (r *QueueRepository) Stats(ctx context.Context) (int64, int64, int64, int64, error) {
	query := `SELECT status, COUNT(*) FROM queue GROUP BY status`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	defer rows.Close()

	var pending, processing, failed, completed int64
	for rows.Next() {
		var status string
		var count int64
		if err := rows.Scan(&status, &count); err != nil {
			return 0, 0, 0, 0, err
		}

		switch status {
		case "PENDING", "RETRYING":
			pending += count
		case "PROCESSING":
			processing += count
		case "FAILED":
			failed += count
		case "SENT":
			completed += count
		}
	}
	return pending, processing, failed, completed, rows.Err()
}
