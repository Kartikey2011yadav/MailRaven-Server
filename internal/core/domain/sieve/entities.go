package sieve

import (
	"time"
)

// SieveScript represents a user's Sieve filter script.
type SieveScript struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	Content   string    `json:"content"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// VacationTracker tracks auto-replies to prevent loops.
type VacationTracker struct {
	UserID      string    `json:"user_id"`
	SenderEmail string    `json:"sender_email"`
	LastSentAt  time.Time `json:"last_sent_at"`
}
