package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/ports"
)

// EmailRepository implements ports.EmailRepository using PostgreSQL
type EmailRepository struct {
	db *sql.DB
}

// NewEmailRepository creates a new PostgreSQL email repository
func NewEmailRepository(db *sql.DB) *EmailRepository {
	return &EmailRepository{db: db}
}

// Save stores a new message (atomic with transaction)
func (r *EmailRepository) Save(ctx context.Context, msg *domain.Message) error {
	query := `
		INSERT INTO messages (
			id, message_id, sender, recipient, subject, snippet, body_path,
			read_state, received_at, spf_result, dkim_result, dmarc_result, dmarc_policy
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err := r.db.ExecContext(ctx, query,
		msg.ID, msg.MessageID, msg.Sender, msg.Recipient, msg.Subject, msg.Snippet,
		msg.BodyPath, msg.ReadState, msg.ReceivedAt, msg.SPFResult, msg.DKIMResult,
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
		WHERE id = $1
	`

	msg := &domain.Message{}

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&msg.ID, &msg.MessageID, &msg.Sender, &msg.Recipient, &msg.Subject, &msg.Snippet,
		&msg.BodyPath, &msg.ReadState, &msg.ReceivedAt, &msg.SPFResult, &msg.DKIMResult,
		&msg.DMARCResult, &msg.DMARCPolicy,
	)

	if err == sql.ErrNoRows {
		return nil, ports.ErrNotFound
	}
	if err != nil {
		return nil, ports.ErrStorageFailure
	}

	return msg, nil
}

// FindByUser retrieves paginated messages for a user
func (r *EmailRepository) FindByUser(ctx context.Context, email string, limit, offset int) ([]*domain.Message, error) {
	query := `
		SELECT id, message_id, sender, recipient, subject, snippet, body_path,
		       read_state, received_at, spf_result, dkim_result, dmarc_result, dmarc_policy
		FROM messages
		WHERE recipient = $1
		ORDER BY received_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, email, limit, offset)
	if err != nil {
		return nil, ports.ErrStorageFailure
	}
	defer rows.Close()

	var messages []*domain.Message
	for rows.Next() {
		msg := &domain.Message{}

		err := rows.Scan(
			&msg.ID, &msg.MessageID, &msg.Sender, &msg.Recipient, &msg.Subject, &msg.Snippet,
			&msg.BodyPath, &msg.ReadState, &msg.ReceivedAt, &msg.SPFResult, &msg.DKIMResult,
			&msg.DMARCResult, &msg.DMARCPolicy,
		)
		if err != nil {
			return nil, ports.ErrStorageFailure
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// UpdateReadState marks a message as read or unread
func (r *EmailRepository) UpdateReadState(ctx context.Context, id string, read bool) error {
	query := `UPDATE messages SET read_state = $1 WHERE id = $2`
	result, err := r.db.ExecContext(ctx, query, read, id)
	if err != nil {
		return ports.ErrStorageFailure
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return ports.ErrStorageFailure
	}
	if rows == 0 {
		return ports.ErrNotFound
	}

	return nil
}

// CountByUser returns total message count for a user
func (r *EmailRepository) CountByUser(ctx context.Context, email string) (int, error) {
	query := `SELECT COUNT(*) FROM messages WHERE recipient = $1`
	var count int
	err := r.db.QueryRowContext(ctx, query, email).Scan(&count)
	if err != nil {
		return 0, ports.ErrStorageFailure
	}
	return count, nil
}

// CountTotal returns total message count in the system
func (r *EmailRepository) CountTotal(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM messages`
	var count int64
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, ports.ErrStorageFailure
	}
	return count, nil
}

// FindSince retrieves messages received after a timestamp
func (r *EmailRepository) FindSince(ctx context.Context, email string, since time.Time, limit int) ([]*domain.Message, error) {
	query := `
		SELECT id, message_id, sender, recipient, subject, snippet, body_path,
		       read_state, received_at, spf_result, dkim_result, dmarc_result, dmarc_policy
		FROM messages
		WHERE recipient = $1 AND received_at > $2
		ORDER BY received_at DESC
		LIMIT $3
	`

	rows, err := r.db.QueryContext(ctx, query, email, since, limit)
	if err != nil {
		return nil, ports.ErrStorageFailure
	}
	defer rows.Close()

	var messages []*domain.Message
	for rows.Next() {
		msg := &domain.Message{}

		err := rows.Scan(
			&msg.ID, &msg.MessageID, &msg.Sender, &msg.Recipient, &msg.Subject, &msg.Snippet,
			&msg.BodyPath, &msg.ReadState, &msg.ReceivedAt, &msg.SPFResult, &msg.DKIMResult,
			&msg.DMARCResult, &msg.DMARCPolicy,
		)
		if err != nil {
			return nil, ports.ErrStorageFailure
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// IMAP Support Stubs (TODO: Implement for Postgres)

func (r *EmailRepository) GetMailbox(ctx context.Context, userID, name string) (*domain.Mailbox, error) {
return nil, nil // Not implemented
}

func (r *EmailRepository) CreateMailbox(ctx context.Context, userID, name string) error {
return nil // Not implemented
}

func (r *EmailRepository) ListMailboxes(ctx context.Context, userID string) ([]*domain.Mailbox, error) {
return nil, nil // Not implemented
}

func (r *EmailRepository) FindByUIDRange(ctx context.Context, userID, mailbox string, min, max uint32) ([]*domain.Message, error) {
return nil, nil // Not implemented
}

func (r *EmailRepository) AddFlags(ctx context.Context, messageID string, flags ...string) error {
return nil // Not implemented
}

func (r *EmailRepository) RemoveFlags(ctx context.Context, messageID string, flags ...string) error {
return nil // Not implemented
}

func (r *EmailRepository) SetFlags(ctx context.Context, messageID string, flags ...string) error {
return nil // Not implemented
}

func (r *EmailRepository) AssignUID(ctx context.Context, messageID string, mailbox string) (uint32, error) {
return 0, nil // Not implemented
}

