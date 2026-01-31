package ports

import (
	"context"
	"time"

	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain/sieve"
)

// ScriptRepository defines storage operations for Sieve scripts.
type ScriptRepository interface {
	// Save creates or updates a script.
	Save(ctx context.Context, script *sieve.SieveScript) error
	// Get retrieves a specific script by name for a user.
	Get(ctx context.Context, userID, name string) (*sieve.SieveScript, error)
	// GetActive retrieves the currently active script for a user. Returns nil if none active.
	GetActive(ctx context.Context, userID string) (*sieve.SieveScript, error)
	// List returns all scripts for a user.
	List(ctx context.Context, userID string) ([]sieve.SieveScript, error)
	// SetActive sets a script as active and deactivates others. If name is empty, deactivates all.
	SetActive(ctx context.Context, userID, name string) error
	// Delete removes a script.
	Delete(ctx context.Context, userID, name string) error
}

// VacationRepository defines storage for vacation auto-reply tracking.
type VacationRepository interface {
	// LastReply returns the time we last replied to sender from user.
	// Returns zero time if no reply found.
	LastReply(ctx context.Context, userID, sender string) (time.Time, error)
	// RecordReply saves the timestamp of a reply.
	RecordReply(ctx context.Context, userID, sender string) error
}

// SieveExecutor defines the interface for running Sieve scripts.
type SieveExecutor interface {
	// Execute runs the active Sieve script for the user against the message.
	// Returns a slice of target mailboxes (e.g. ["INBOX", "Trash"]).
	// If the slice is empty, the message should be discarded.
	// Runtime errors should log but return ["INBOX"] (fail open).
	Execute(ctx context.Context, userID string, rawMsg []byte) (targets []string, err error)
}
