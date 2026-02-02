package domain

import "time"

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

// User represents a mailbox owner with authentication credentials
type User struct {
	Email        string    // Email address (primary key)
	PasswordHash string    // Bcrypt hash of password
	Role         Role      // Access role
	CreatedAt    time.Time // Account creation timestamp
	LastLoginAt  time.Time // Most recent successful login
	StorageQuota int64     // Max storage in bytes (0 for default/unlimited)
	StorageUsed  int64     // Current storage usage in bytes
}

// AuthToken represents a JWT token for API authentication
type AuthToken struct {
	TokenString string    // JWT token string
	UserEmail   string    // Email address of authenticated user
	IssuedAt    time.Time // When token was issued
	ExpiresAt   time.Time // When token expires (7 days from issue)
}
