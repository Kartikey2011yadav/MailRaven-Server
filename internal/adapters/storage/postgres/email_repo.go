package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
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
			read_state, received_at, spf_result, dkim_result, dmarc_result, dmarc_policy,
			mailbox, uid, flags, modseq, is_starred
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
	`

	_, err := r.db.ExecContext(ctx, query,
		msg.ID, msg.MessageID, msg.Sender, msg.Recipient, msg.Subject, msg.Snippet,
		msg.BodyPath, msg.ReadState, msg.ReceivedAt, msg.SPFResult, msg.DKIMResult,
		msg.DMARCResult, msg.DMARCPolicy,
		msg.Mailbox, msg.UID, msg.Flags, msg.ModSeq, msg.IsStarred,
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
		       read_state, received_at, spf_result, dkim_result, dmarc_result, dmarc_policy,
		       mailbox, uid, flags, modseq, is_starred
		FROM messages
		WHERE id = $1
	`

	msg := &domain.Message{}

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&msg.ID, &msg.MessageID, &msg.Sender, &msg.Recipient, &msg.Subject, &msg.Snippet,
		&msg.BodyPath, &msg.ReadState, &msg.ReceivedAt, &msg.SPFResult, &msg.DKIMResult,
		&msg.DMARCResult, &msg.DMARCPolicy,
		&msg.Mailbox, &msg.UID, &msg.Flags, &msg.ModSeq, &msg.IsStarred,
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
	// Deprecated: wrapper around List
	return r.List(ctx, email, domain.MessageFilter{
		Limit:  limit,
		Offset: offset,
	})
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
		       read_state, received_at, spf_result, dkim_result, dmarc_result, dmarc_policy,
		       mailbox, uid, flags, modseq, is_starred
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
			&msg.Mailbox, &msg.UID, &msg.Flags, &msg.ModSeq, &msg.IsStarred,
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

// CopyMessages copies messages to a destination mailbox
func (r *EmailRepository) CopyMessages(ctx context.Context, userID string, messageIDs []string, destMailbox string) error {
	return ports.ErrStorageFailure // Not implemented
}

// SetACL updates the access rights for an identifier on a mailbox
func (r *EmailRepository) SetACL(ctx context.Context, userID, mailboxName, identifier, rights string) error {
	return ports.ErrStorageFailure // Not implemented
}

// List retrieves messages matching the filter criteria
func (r *EmailRepository) List(ctx context.Context, email string, filter domain.MessageFilter) ([]*domain.Message, error) {
	var queryBuilder strings.Builder
	queryBuilder.WriteString(`
		SELECT id, message_id, sender, recipient, subject, snippet, body_path,
		       read_state, received_at, spf_result, dkim_result, dmarc_result, dmarc_policy,
		       mailbox, uid, flags, modseq, is_starred
		FROM messages
		WHERE recipient = $1
	`)

	args := []interface{}{email}
	argIdx := 2

	if filter.Mailbox != "" {
		queryBuilder.WriteString(fmt.Sprintf(" AND mailbox = $%d", argIdx))
		args = append(args, filter.Mailbox)
		argIdx++
	}

	if filter.IsRead != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND read_state = $%d", argIdx))
		args = append(args, *filter.IsRead)
		argIdx++
	}

	if filter.IsStarred != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND is_starred = $%d", argIdx))
		args = append(args, *filter.IsStarred)
		argIdx++
	}

	if filter.DateRange.Start != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND received_at >= $%d", argIdx))
		args = append(args, *filter.DateRange.Start)
		argIdx++
	}

	if filter.DateRange.End != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND received_at <= $%d", argIdx))
		args = append(args, *filter.DateRange.End)
		argIdx++
	}

	queryBuilder.WriteString(fmt.Sprintf(" ORDER BY received_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1))
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, queryBuilder.String(), args...)
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
			&msg.Mailbox, &msg.UID, &msg.Flags, &msg.ModSeq, &msg.IsStarred,
		)
		if err != nil {
			return nil, ports.ErrStorageFailure
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// UpdateStarred marks a message as starred (important) or not
func (r *EmailRepository) UpdateStarred(ctx context.Context, id string, starred bool) error {
	query := `UPDATE messages SET is_starred = $1 WHERE id = $2`
	result, err := r.db.ExecContext(ctx, query, starred, id)
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

// UpdateMailbox moves a message to a new mailbox/folder
func (r *EmailRepository) UpdateMailbox(ctx context.Context, id string, mailbox string) error {
	query := `UPDATE messages SET mailbox = $1 WHERE id = $2`
	result, err := r.db.ExecContext(ctx, query, mailbox, id)
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
