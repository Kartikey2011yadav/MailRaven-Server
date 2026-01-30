package domain

import "time"

// Message represents an email message with metadata and authentication results
type Message struct {
	ID         string    // Unique message ID (generated, e.g., UUID)
	MessageID  string    // Email Message-ID header value
	Sender     string    // MAIL FROM address
	Recipient  string    // RCPT TO address (single recipient for MVP)
	Subject    string    // Email subject line
	Snippet    string    // First 200 chars of body for list view
	BodyPath   string    // Path to compressed body file in blob store
	ReadState  bool      // Has user read this message?
	ReceivedAt time.Time // When server accepted the message

	// IMAP Support
	UID     uint32 // IMAP UID (Unique, Monotonic per Mailbox)
	Mailbox string // Mailbox name (default "INBOX")
	Flags   string // Space-separated list of flags (e.g., "\Seen \Flagged")
	ModSeq  uint64 // Modification Sequence (for CONDSTORE)

	// Email authentication results (from SPF/DKIM/DMARC validation)
	SPFResult   string // "pass", "fail", "softfail", "neutral", "none"
	DKIMResult  string // "pass", "fail", "none"
	DMARCResult string // "pass", "fail", "none"
	DMARCPolicy string // "none", "quarantine", "reject"
}

// MessageBody represents the full raw email content stored in blob storage
type MessageBody struct {
	MessageID      string // References Message.ID
	RawContent     []byte // Full MIME message (after decompression)
	CompressedSize int64  // Size of compressed file
	OriginalSize   int64  // Size before compression
}
