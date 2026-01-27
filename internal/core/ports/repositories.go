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

	// FindSince retrieves messages received after a timestamp (delta sync)
	FindSince(ctx context.Context, email string, since time.Time, limit int) ([]*domain.Message, error)
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
}
