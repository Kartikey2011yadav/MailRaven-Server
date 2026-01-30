package ports

import (
	"context"
	"time"

	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
)

// EmailRepository defines storage operations for email messages
type EmailRepository interface {
	// Save stores a new message (atomic with blob storage)
	// Returns error if message already exists or storage fails
	Save(ctx context.Context, msg *domain.Message) error

	// FindByID retrieves a single message by ID
	// Returns ErrNotFound if message doesn't exist
	FindByID(ctx context.Context, id string) (*domain.Message, error)

	// FindByUser retrieves paginated messages for a user
	// Results ordered by ReceivedAt DESC (newest first)
	// Returns empty slice if no messages match
	FindByUser(ctx context.Context, email string, limit, offset int) ([]*domain.Message, error)

	// UpdateReadState marks a message as read or unread
	// Returns ErrNotFound if message doesn't exist
	UpdateReadState(ctx context.Context, id string, read bool) error

	// CountByUser returns total message count for a user
	CountByUser(ctx context.Context, email string) (int, error)

	// CountTotal returns total message count in the system (admin usage)
	CountTotal(ctx context.Context) (int64, error)

	// FindSince retrieves messages received after a timestamp (delta sync)
	FindSince(ctx context.Context, email string, since time.Time, limit int) ([]*domain.Message, error)

	// IMAP Support
	GetMailbox(ctx context.Context, userID, name string) (*domain.Mailbox, error)
	CreateMailbox(ctx context.Context, userID, name string) error
	ListMailboxes(ctx context.Context, userID string) ([]*domain.Mailbox, error)

	// FindByUIDRange retrieves messages by UID range [min, max]
	FindByUIDRange(ctx context.Context, userID, mailbox string, min, max uint32) ([]*domain.Message, error)

	// Flags
	AddFlags(ctx context.Context, messageID string, flags ...string) error
	RemoveFlags(ctx context.Context, messageID string, flags ...string) error
	SetFlags(ctx context.Context, messageID string, flags ...string) error

	// UID Management
	// AssignUID assigns a UID to a message if it doesn't have one (atomic with increments)
	AssignUID(ctx context.Context, messageID string, mailbox string) (uint32, error)
}

// UserRepository defines storage operations for user accounts
type UserRepository interface {
	// Create creates a new user with hashed password
	// Returns ErrAlreadyExists if email already registered
	Create(ctx context.Context, user *domain.User) error

	// FindByEmail retrieves user by email address
	// Returns ErrNotFound if user doesn't exist
	FindByEmail(ctx context.Context, email string) (*domain.User, error)

	// Authenticate verifies email/password and returns user
	// Returns ErrInvalidCredentials if auth fails
	Authenticate(ctx context.Context, email, password string) (*domain.User, error)

	// UpdateLastLogin updates the LastLoginAt timestamp
	UpdateLastLogin(ctx context.Context, email string) error

	// New Management Methods
	List(ctx context.Context, limit, offset int) ([]*domain.User, error)
	Delete(ctx context.Context, email string) error
	UpdatePassword(ctx context.Context, email, passwordHash string) error
	UpdateRole(ctx context.Context, email string, role domain.Role) error

	// Count returns simplified user statistics (total, active, admin)
	Count(ctx context.Context) (map[string]int64, error)
}

// QueueRepository defines storage operations for outbound email queue
type QueueRepository interface {
	// Enqueue adds a message to the outbound queue
	Enqueue(ctx context.Context, msg *domain.OutboundMessage) error

	// LockNextReady finds the next message ready for delivery and marks it as PROCESSING
	// Should check for Status=PENDING/RETRYING and NextRetryAt <= now
	// Returns (nil, nil) if no messages are ready
	LockNextReady(ctx context.Context) (*domain.OutboundMessage, error)

	// UpdateStatus updates the status, retry count and next retry time
	UpdateStatus(ctx context.Context, id string, status domain.OutboundStatus, retryCount int, nextRetry time.Time, lastError string) error

	// Stats returns queue statistics (pending, processing, failed, completed)
	Stats(ctx context.Context) (pending, processing, failed, completed int64, err error)
}

// DomainRepository defines storage operations for hosted domains
type DomainRepository interface {
	// Create adds a new domain
	// Returns ErrAlreadyExists if domain exists
	Create(ctx context.Context, domain *domain.Domain) error

	// Get retrieves a domain by name
	// Returns ErrNotFound if not found
	Get(ctx context.Context, name string) (*domain.Domain, error)

	// List returns paginated domains
	List(ctx context.Context, limit, offset int) ([]*domain.Domain, error)

	// Delete removes a domain
	Delete(ctx context.Context, name string) error

	// Exists checks if a domain exists
	Exists(ctx context.Context, name string) (bool, error)
}
