package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/ports"
)

// EmailRepository implements ports.EmailRepository using SQLite
type EmailRepository struct {
	db *sql.DB
}

// NewEmailRepository creates a new SQLite email repository
func NewEmailRepository(db *sql.DB) *EmailRepository {
	return &EmailRepository{db: db}
}

// Save stores a new message (atomic with transaction)
func (r *EmailRepository) Save(ctx context.Context, msg *domain.Message) error {
	query := `
		INSERT INTO messages (
			id, message_id, sender, recipient, subject, snippet, body_path,
			read_state, received_at, spf_result, dkim_result, dmarc_result, dmarc_policy
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	readStateInt := 0
	if msg.ReadState {
		readStateInt = 1
	}

	_, err := r.db.ExecContext(ctx, query,
		msg.ID, msg.MessageID, msg.Sender, msg.Recipient, msg.Subject, msg.Snippet,
		msg.BodyPath, readStateInt, msg.ReceivedAt.Unix(), msg.SPFResult, msg.DKIMResult,
		msg.DMARCResult, msg.DMARCPolicy,
	)
	if err != nil {
		return ports.ErrStorageFailure
	}

	return nil
}

// FindByID retrieves a single message by ID
func (r *EmailRepository) FindByID(ctx context.Context, id string) (*domain.Message, error) {
	query := `
		SELECT id, message_id, sender, recipient, subject, snippet, body_path,
		       read_state, received_at, spf_result, dkim_result, dmarc_result, dmarc_policy
		FROM messages
		WHERE id = ?
	`

	msg := &domain.Message{}
	var readStateInt int
	var receivedAtUnix int64

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&msg.ID, &msg.MessageID, &msg.Sender, &msg.Recipient, &msg.Subject, &msg.Snippet,
		&msg.BodyPath, &readStateInt, &receivedAtUnix, &msg.SPFResult, &msg.DKIMResult,
		&msg.DMARCResult, &msg.DMARCPolicy,
	)

	if err == sql.ErrNoRows {
		return nil, ports.ErrNotFound
	}
	if err != nil {
		return nil, ports.ErrStorageFailure
	}

	msg.ReadState = readStateInt == 1
	msg.ReceivedAt = time.Unix(receivedAtUnix, 0)

	return msg, nil
}

// FindByUser retrieves paginated messages for a user
func (r *EmailRepository) FindByUser(ctx context.Context, email string, limit, offset int) ([]*domain.Message, error) {
	query := `
		SELECT id, message_id, sender, recipient, subject, snippet, body_path,
		       read_state, received_at, spf_result, dkim_result, dmarc_result, dmarc_policy
		FROM messages
		WHERE recipient = ?
		ORDER BY received_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, email, limit, offset)
	if err != nil {
		return nil, ports.ErrStorageFailure
	}
	defer rows.Close()

	var messages []*domain.Message
	for rows.Next() {
		msg := &domain.Message{}
		var readStateInt int
		var receivedAtUnix int64

		err := rows.Scan(
			&msg.ID, &msg.MessageID, &msg.Sender, &msg.Recipient, &msg.Subject, &msg.Snippet,
			&msg.BodyPath, &readStateInt, &receivedAtUnix, &msg.SPFResult, &msg.DKIMResult,
			&msg.DMARCResult, &msg.DMARCPolicy,
		)
		if err != nil {
			return nil, ports.ErrStorageFailure
		}

		msg.ReadState = readStateInt == 1
		msg.ReceivedAt = time.Unix(receivedAtUnix, 0)
		messages = append(messages, msg)
	}

	return messages, nil
}

// UpdateReadState marks a message as read or unread
func (r *EmailRepository) UpdateReadState(ctx context.Context, id string, read bool) error {
	readStateInt := 0
	if read {
		readStateInt = 1
	}

	query := `UPDATE messages SET read_state = ? WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, readStateInt, id)
	if err != nil {
		return ports.ErrStorageFailure
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return ports.ErrStorageFailure
	}
	if rowsAffected == 0 {
		return ports.ErrNotFound
	}

	return nil
}

// CountByUser returns total message count for a user
func (r *EmailRepository) CountByUser(ctx context.Context, email string) (int, error) {
	query := `SELECT COUNT(*) FROM messages WHERE recipient = ?`

	var count int
	err := r.db.QueryRowContext(ctx, query, email).Scan(&count)
	if err != nil {
		return 0, ports.ErrStorageFailure
	}

	return count, nil
}

// FindSince retrieves messages received after a timestamp (delta sync)
func (r *EmailRepository) FindSince(ctx context.Context, email string, since time.Time, limit int) ([]*domain.Message, error) {
	query := `
		SELECT id, message_id, sender, recipient, subject, snippet, body_path,
		       read_state, received_at, spf_result, dkim_result, dmarc_result, dmarc_policy
		FROM messages
		WHERE recipient = ? AND received_at >= ?
		ORDER BY received_at DESC
		LIMIT ?
	`

	rows, err := r.db.QueryContext(ctx, query, email, since.Unix(), limit)
	if err != nil {
		return nil, ports.ErrStorageFailure
	}
	defer rows.Close()

	var messages []*domain.Message
	for rows.Next() {
		msg := &domain.Message{}
		var readStateInt int
		var receivedAtUnix int64

		err := rows.Scan(
			&msg.ID, &msg.MessageID, &msg.Sender, &msg.Recipient, &msg.Subject, &msg.Snippet,
			&msg.BodyPath, &readStateInt, &receivedAtUnix, &msg.SPFResult, &msg.DKIMResult,
			&msg.DMARCResult, &msg.DMARCPolicy,
		)
		if err != nil {
			return nil, ports.ErrStorageFailure
		}

		msg.ReadState = readStateInt == 1
		msg.ReceivedAt = time.Unix(receivedAtUnix, 0)
		messages = append(messages, msg)
	}

	return messages, nil
}

// CountTotal returns total message count in the system
func (r *EmailRepository) CountTotal(ctx context.Context) (int64, error) {
	query := "SELECT COUNT(*) FROM messages"
	var count int64
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

