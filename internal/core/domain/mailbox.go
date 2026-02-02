package domain

// Mailbox represents an IMAP folder/mailbox used to group messages
type Mailbox struct {
	Name         string            // Primary Key (Composite with UserID). e.g., "INBOX"
	UserID       string            // Owner of the mailbox
	UIDValidity  uint32            // Random non-zero integer. Changes if UIDs are reset.
	UIDNext      uint32            // Next UID to assign to a new message. Starts at 1.
	MessageCount int               // Cached count of messages in this mailbox
	ACL          map[string]string // Access Control List (Identifier -> Rights)
}
