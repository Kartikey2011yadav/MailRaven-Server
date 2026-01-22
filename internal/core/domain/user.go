package domain

import "time"

// User represents a mailbox owner with authentication credentials
type User struct {
	Email        string    // Email address (primary key)
	PasswordHash string    // Bcrypt hash of password
	CreatedAt    time.Time // Account creation timestamp
	LastLoginAt  time.Time // Most recent successful login
}

// AuthToken represents a JWT token for API authentication
type AuthToken struct {
	TokenString string    // JWT token string
	UserEmail   string    // Email address of authenticated user
	IssuedAt    time.Time // When token was issued
	ExpiresAt   time.Time // When token expires (7 days from issue)
}
