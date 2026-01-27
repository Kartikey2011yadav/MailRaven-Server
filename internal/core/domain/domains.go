package domain

import "time"

// Domain represents a hosted email domain
type Domain struct {
	Name           string    `json:"name"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Active         bool      `json:"active"`
	DKIMSelector   string    `json:"dkim_selector"`
	DKIMPrivateKey string    `json:"-"` // Never expose private key via JSON
	DKIMPublicKey  string    `json:"dkim_public_key"`
}
