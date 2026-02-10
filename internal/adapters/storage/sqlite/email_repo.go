package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/notifications"
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
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return ports.ErrStorageFailure
	}
	defer tx.Rollback() //nolint:errcheck

	if msg.Mailbox == "" {
		msg.Mailbox = "INBOX"
	}

	// 1. Ensure mailbox exists
	// Using random validity if created here.
	uidValidity := uint32(time.Now().Unix())
	if uidValidity == 0 {
		uidValidity = 1
	}
	_, err = tx.ExecContext(ctx, "INSERT OR IGNORE INTO mailboxes (name, user_id, uid_validity, uid_next) VALUES (?, ?, ?, ?)", msg.Mailbox, msg.Recipient, uidValidity, 1)
	if err != nil {
		return ports.ErrStorageFailure
	}

	// 2. Assign UID
	err = tx.QueryRowContext(ctx, "UPDATE mailboxes SET uid_next = uid_next + 1, message_count = message_count + 1 WHERE user_id = ? AND name = ? RETURNING uid_next - 1", msg.Recipient, msg.Mailbox).Scan(&msg.UID)
	if err != nil {
		return ports.ErrStorageFailure
	}

	query := `
		INSERT INTO messages (
			id, message_id, sender, recipient, subject, snippet, body_path,
			read_state, received_at, spf_result, dkim_result, dmarc_result, dmarc_policy,
			uid, mailbox, flags, mod_seq, size, is_starred
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	readStateInt := 0
	if msg.ReadState {
		readStateInt = 1
	}

	isStarredInt := 0
	if msg.IsStarred {
		isStarredInt = 1
	}

	_, err = tx.ExecContext(ctx, query,
		msg.ID, msg.MessageID, msg.Sender, msg.Recipient, msg.Subject, msg.Snippet,
		msg.BodyPath, readStateInt, msg.ReceivedAt.Unix(), msg.SPFResult, msg.DKIMResult,
		msg.DMARCResult, msg.DMARCPolicy,
		msg.UID, msg.Mailbox, msg.Flags, msg.ModSeq, msg.Size, isStarredInt,
	)
	if err != nil {
		return ports.ErrStorageFailure
	}

	if err := tx.Commit(); err != nil {
		return ports.ErrStorageFailure
	}

	// Notify
	notifications.GlobalHub.Broadcast(notifications.Event{
		Type:    notifications.EventNewMessage,
		UserID:  msg.Recipient,
		Mailbox: msg.Mailbox,
		Data:    msg,
	})

	return nil
}

// FindByID retrieves a single message by ID
func (r *EmailRepository) FindByID(ctx context.Context, id string) (*domain.Message, error) {
	query := `
		SELECT id, message_id, sender, recipient, subject, snippet, body_path,
		       read_state, received_at, spf_result, dkim_result, dmarc_result, dmarc_policy, size,
		       uid, mailbox, flags, mod_seq, is_starred
		FROM messages
		WHERE id = ?
	`

	msg := &domain.Message{}
	var readStateInt int
	var isStarredInt int
	var receivedAtUnix int64

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&msg.ID, &msg.MessageID, &msg.Sender, &msg.Recipient, &msg.Subject, &msg.Snippet,
		&msg.BodyPath, &readStateInt, &receivedAtUnix, &msg.SPFResult, &msg.DKIMResult,
		&msg.DMARCResult, &msg.DMARCPolicy, &msg.Size,
		&msg.UID, &msg.Mailbox, &msg.Flags, &msg.ModSeq, &isStarredInt,
	)

	if err == sql.ErrNoRows {
		return nil, ports.ErrNotFound
	}
	if err != nil {
		return nil, ports.ErrStorageFailure
	}

	msg.ReadState = readStateInt == 1
	msg.IsStarred = isStarredInt == 1
	msg.ReceivedAt = time.Unix(receivedAtUnix, 0)

	return msg, nil
}

// List retrieves messages matching the filter criteria
func (r *EmailRepository) List(ctx context.Context, email string, filter domain.MessageFilter) ([]*domain.Message, error) {
	queryBuilder := strings.Builder{}
	queryBuilder.WriteString(`
		SELECT id, message_id, sender, recipient, subject, snippet, body_path,
		       read_state, received_at, spf_result, dkim_result, dmarc_result, dmarc_policy, size,
		       uid, mailbox, flags, mod_seq, is_starred
		FROM messages
		WHERE recipient = ?
	`)

	args := []interface{}{email}

	if filter.Mailbox != "" {
		queryBuilder.WriteString(" AND mailbox = ?")
		args = append(args, filter.Mailbox)
	}

	if filter.IsRead != nil {
		readVal := 0
		if *filter.IsRead {
			readVal = 1
		}
		queryBuilder.WriteString(" AND read_state = ?")
		args = append(args, readVal)
	}

	if filter.IsStarred != nil {
		starredVal := 0
		if *filter.IsStarred {
			starredVal = 1
		}
		queryBuilder.WriteString(" AND is_starred = ?")
		args = append(args, starredVal)
	}

	if filter.DateRange.Start != nil {
		queryBuilder.WriteString(" AND received_at >= ?")
		args = append(args, filter.DateRange.Start.Unix())
	}

	if filter.DateRange.End != nil {
		queryBuilder.WriteString(" AND received_at <= ?")
		args = append(args, filter.DateRange.End.Unix())
	}

	queryBuilder.WriteString(" ORDER BY received_at DESC LIMIT ? OFFSET ?")
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, queryBuilder.String(), args...)
	if err != nil {
		return nil, ports.ErrStorageFailure
	}
	defer rows.Close()

	var messages []*domain.Message
	for rows.Next() {
		msg := &domain.Message{}
		var readStateInt int
		var isStarredInt int
		var receivedAtUnix int64

		err := rows.Scan(
			&msg.ID, &msg.MessageID, &msg.Sender, &msg.Recipient, &msg.Subject, &msg.Snippet,
			&msg.BodyPath, &readStateInt, &receivedAtUnix, &msg.SPFResult, &msg.DKIMResult,
			&msg.DMARCResult, &msg.DMARCPolicy, &msg.Size,
			&msg.UID, &msg.Mailbox, &msg.Flags, &msg.ModSeq, &isStarredInt,
		)
		if err != nil {
			return nil, ports.ErrStorageFailure
		}

		msg.ReadState = readStateInt == 1
		msg.IsStarred = isStarredInt == 1
		msg.ReceivedAt = time.Unix(receivedAtUnix, 0)
		messages = append(messages, msg)
	}

	return messages, nil
}

// UpdateMailbox moves a message to a new mailbox/folder
func (r *EmailRepository) UpdateMailbox(ctx context.Context, id string, mailbox string) error {
	// 1. Ensure mailbox exists (simple approach: require it to exist or insert ignore provided user_id is known... wait, we need user_id)
	// For efficiency, we assume mailbox table contains standard folders or they are created via IDLE/IMAP.
	// But let's check recipient from message first to be safe or do a subselect?
	// Simpler: Just update. If specific mailbox creation logic needed (IMAP), it should be handled there.
	// However, we need to update UID if moving between mailboxes to be fully IMAP compliant.
	// That logic is complex (Copy + Delete).
	// Spec says "User Story 1 - Archive Message... message moves to Archive mailbox".
	// For a simple DB-based mail system without full IMAP sync requirement here, just updating mailbox col is okay.
	// But if we want IMAP compliance, it MUST be Copy+Delete or we break UID constraints.
	// Since implementing full IMAP Copy+Delete logic here is heavy and might violate "atomic" if not careful.
	// Plan says "Message moves are atomic metadata updates".
	// Implementation Plan Check: "Protocol Parity: "Starred" maps to IMAP \Flagged, "Junk" maps to \Junk...".
	// T005 says "Implement UpdateMailbox".
	// Let's implement simple Column Update for now as per "atomic metadata update".
	// NOTE: This breaks IMAP UIDs unless we re-assign UID.
	// Let's do a simple update for now, recognizing this technical debt vs IMAP.
	// Or even better: `UPDATE messages SET mailbox = ? WHERE id = ?`.

	query := `UPDATE messages SET mailbox = ? WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, mailbox, id)
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

// UpdateStarred marks a message as starred (important) or not
func (r *EmailRepository) UpdateStarred(ctx context.Context, id string, starred bool) error {
	starredInt := 0
	if starred {
		starredInt = 1
	}

	query := `UPDATE messages SET is_starred = ? WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, starredInt, id)
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

// FindByUser retrieves paginated messages for a user
func (r *EmailRepository) FindByUser(ctx context.Context, email string, limit, offset int) ([]*domain.Message, error) {
	// Deprecated implementation that wraps List for backward compatibility
	return r.List(ctx, email, domain.MessageFilter{
		Limit:   limit,
		Offset:  offset,
		Mailbox: "", // All mailboxes? Or just INBOX? Old implementation didn't filter mailbox, so it returned ALL.
		// NOTE: Original code query: "SELECT ... FROM messages WHERE recipient = ?".
		// No mailbox filter was present. So it returned all messages from all folders.
		// That matches filter with Mailbox="".
	})
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

// GetMailbox retrieves a mailbox by name for a user
func (r *EmailRepository) GetMailbox(ctx context.Context, userID, name string) (*domain.Mailbox, error) {
	query := `SELECT name, user_id, uid_validity, uid_next, message_count, acl FROM mailboxes WHERE user_id = ? AND name = ?`
	mb := &domain.Mailbox{}
	var aclStr string
	err := r.db.QueryRowContext(ctx, query, userID, name).Scan(&mb.Name, &mb.UserID, &mb.UIDValidity, &mb.UIDNext, &mb.MessageCount, &aclStr)
	if err == sql.ErrNoRows {
		return nil, ports.ErrNotFound
	}
	if err != nil {
		return nil, ports.ErrStorageFailure
	}
	if aclStr != "" {
		if err := json.Unmarshal([]byte(aclStr), &mb.ACL); err != nil {
			mb.ACL = make(map[string]string)
		}
	}
	if mb.ACL == nil {
		mb.ACL = make(map[string]string)
	}
	return mb, nil
}

// CreateMailbox creates a new mailbox
func (r *EmailRepository) CreateMailbox(ctx context.Context, userID, name string) error {
	uidValidity := uint32(time.Now().Unix())
	if uidValidity == 0 {
		uidValidity = 1
	}

	query := `INSERT INTO mailboxes (name, user_id, uid_validity) VALUES (?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, name, userID, uidValidity)
	if err != nil {
		return ports.ErrStorageFailure
	}
	return nil
}

// ListMailboxes retrieves all mailboxes for a user
func (r *EmailRepository) ListMailboxes(ctx context.Context, userID string) ([]*domain.Mailbox, error) {
	query := `SELECT name, user_id, uid_validity, uid_next, message_count, acl FROM mailboxes WHERE user_id = ?`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, ports.ErrStorageFailure
	}
	defer rows.Close()

	var mailboxes []*domain.Mailbox
	for rows.Next() {
		mb := &domain.Mailbox{}
		var aclStr string
		if err := rows.Scan(&mb.Name, &mb.UserID, &mb.UIDValidity, &mb.UIDNext, &mb.MessageCount, &aclStr); err != nil {
			return nil, ports.ErrStorageFailure
		}
		if aclStr != "" {
			//nolint:errcheck // Best effort unmarshal
			_ = json.Unmarshal([]byte(aclStr), &mb.ACL)
		}
		if mb.ACL == nil {
			mb.ACL = make(map[string]string)
		}
		mailboxes = append(mailboxes, mb)
	}
	return mailboxes, nil
}

// FindByUIDRange retrieves messages by UID range
func (r *EmailRepository) FindByUIDRange(ctx context.Context, userID, mailbox string, min, max uint32) ([]*domain.Message, error) {
	query := `
		SELECT id, message_id, sender, recipient, subject, snippet, body_path,
		       read_state, received_at, uid, mailbox, flags, mod_seq,
		       spf_result, dkim_result, dmarc_result, dmarc_policy, size
		FROM messages
		WHERE recipient = ? AND mailbox = ? AND uid >= ? AND uid <= ?
		ORDER BY uid ASC
	`
	// Note: 2^32-1 is max uint32. If max is 0 (check caller), it implies * (infinity).
	// IMAP specific: caller handles conversion of * to max uint32.

	rows, err := r.db.QueryContext(ctx, query, userID, mailbox, min, max)
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
			&msg.BodyPath, &readStateInt, &receivedAtUnix, &msg.UID, &msg.Mailbox, &msg.Flags, &msg.ModSeq,
			&msg.SPFResult, &msg.DKIMResult, &msg.DMARCResult, &msg.DMARCPolicy, &msg.Size,
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

// AddFlags adds flags to a message
func (r *EmailRepository) AddFlags(ctx context.Context, messageID string, flags ...string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	var currentFlagsStr string
	err = tx.QueryRowContext(ctx, "SELECT flags FROM messages WHERE id = ?", messageID).Scan(&currentFlagsStr)
	if err == sql.ErrNoRows {
		return ports.ErrNotFound
	}
	if err != nil {
		return err
	}

	currentFlags := strings.Fields(currentFlagsStr)
	existing := make(map[string]bool)
	for _, f := range currentFlags {
		existing[f] = true
	}

	changed := false
	for _, f := range flags {
		if !existing[f] {
			currentFlags = append(currentFlags, f)
			existing[f] = true
			changed = true
		}
	}

	if changed {
		newFlagsStr := strings.Join(currentFlags, " ")
		_, err = tx.ExecContext(ctx, "UPDATE messages SET flags = ? WHERE id = ?", newFlagsStr, messageID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// RemoveFlags removes flags from a message
func (r *EmailRepository) RemoveFlags(ctx context.Context, messageID string, flags ...string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	var currentFlagsStr string
	err = tx.QueryRowContext(ctx, "SELECT flags FROM messages WHERE id = ?", messageID).Scan(&currentFlagsStr)
	if err == sql.ErrNoRows {
		return ports.ErrNotFound
	}
	if err != nil {
		return err
	}

	currentFlags := strings.Fields(currentFlagsStr)
	toRemove := make(map[string]bool)
	for _, f := range flags {
		toRemove[f] = true
	}

	var newFlags []string
	changed := false
	for _, f := range currentFlags {
		if !toRemove[f] {
			newFlags = append(newFlags, f)
		} else {
			changed = true
		}
	}

	if changed {
		newFlagsStr := strings.Join(newFlags, " ")
		_, err = tx.ExecContext(ctx, "UPDATE messages SET flags = ? WHERE id = ?", newFlagsStr, messageID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// SetFlags sets the flags for a message
func (r *EmailRepository) SetFlags(ctx context.Context, messageID string, flags ...string) error {
	newFlagsStr := strings.Join(flags, " ")
	_, err := r.db.ExecContext(ctx, "UPDATE messages SET flags = ? WHERE id = ?", newFlagsStr, messageID)
	if err != nil {
		return ports.ErrStorageFailure
	}
	return nil
}

// AssignUID assigns a UID to a message
func (r *EmailRepository) AssignUID(ctx context.Context, messageID string, mailboxName string) (uint32, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback() //nolint:errcheck

	var recipient string
	err = tx.QueryRowContext(ctx, "SELECT recipient FROM messages WHERE id = ?", messageID).Scan(&recipient)
	if err == sql.ErrNoRows {
		return 0, ports.ErrNotFound
	}
	if err != nil {
		return 0, err
	}

	// Update Mailbox and get UID
	var assignedUID uint32

	// Check existence
	var exists bool
	err = tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM mailboxes WHERE user_id = ? AND name = ?)", recipient, mailboxName).Scan(&exists)
	if err != nil {
		return 0, err
	}

	if !exists {
		if mailboxName == "INBOX" {
			uidValidity := uint32(time.Now().Unix())
			if uidValidity == 0 {
				uidValidity = 1
			}
			_, err = tx.ExecContext(ctx, "INSERT INTO mailboxes (name, user_id, uid_validity, uid_next) VALUES (?, ?, ?, ?)", mailboxName, recipient, uidValidity, 1)
			if err != nil {
				return 0, err
			}
		} else {
			return 0, ports.ErrNotFound
		}
	}

	// Atomic increment and return (sqlite 3.35+)
	err = tx.QueryRowContext(ctx, "UPDATE mailboxes SET uid_next = uid_next + 1, message_count = message_count + 1 WHERE user_id = ? AND name = ? RETURNING uid_next - 1", recipient, mailboxName).Scan(&assignedUID)
	if err != nil {
		return 0, err
	}

	_, err = tx.ExecContext(ctx, "UPDATE messages SET uid = ?, mailbox = ? WHERE id = ?", assignedUID, mailboxName, messageID)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return assignedUID, nil
}

// SetACL updates the access rights for an identifier on a mailbox
func (r *EmailRepository) SetACL(ctx context.Context, userID, mailboxName, identifier, rights string) error {
	// 1. Get current ACL
	mb, err := r.GetMailbox(ctx, userID, mailboxName)
	if err != nil {
		return err
	}
	if mb.ACL == nil {
		mb.ACL = make(map[string]string)
	}

	// 2. Update map

	if rights == "" {
		delete(mb.ACL, identifier)
	} else {
		mb.ACL[identifier] = rights
	}

	// 3. Serialize
	data, err := json.Marshal(mb.ACL)
	if err != nil {
		return ports.ErrStorageFailure // internal error
	}

	// 4. Update DB
	query := `UPDATE mailboxes SET acl = ? WHERE user_id = ? AND name = ?`
	_, err = r.db.ExecContext(ctx, query, string(data), userID, mailboxName)
	if err != nil {
		return ports.ErrStorageFailure
	}
	return nil
}

// CopyMessages copies messages to a destination mailbox
func (r *EmailRepository) CopyMessages(ctx context.Context, userID string, messageIDs []string, destMailbox string) error {
	if len(messageIDs) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return ports.ErrStorageFailure
	}
	defer tx.Rollback() //nolint:errcheck

	// Ensure dest mailbox exists
	var exists bool
	err = tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM mailboxes WHERE user_id = ? AND name = ?)", userID, destMailbox).Scan(&exists)
	if err != nil {
		return err
	}

	if !exists {
		return ports.ErrNotFound
	}

	for _, id := range messageIDs {
		// Increment UID
		var newUID uint32
		err = tx.QueryRowContext(ctx, "UPDATE mailboxes SET uid_next = uid_next + 1, message_count = message_count + 1 WHERE user_id = ? AND name = ? RETURNING uid_next - 1", userID, destMailbox).Scan(&newUID)
		if err != nil {
			return err
		}

		newID := uuid.New().String()

		// Insert Copy
		_, err = tx.ExecContext(ctx, `
			INSERT INTO messages (
				id, message_id, sender, recipient, subject, snippet, body_path,
				read_state, received_at, spf_result, dkim_result, dmarc_result, dmarc_policy,
				uid, mailbox, flags, mod_seq
			)
			SELECT 
				?, message_id, sender, recipient, subject, snippet, body_path,
				read_state, received_at, spf_result, dkim_result, dmarc_result, dmarc_policy,
				?, ?, flags, mod_seq
			FROM messages WHERE id = ?
		`, newID, newUID, destMailbox, id)

		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
