package dto

import "time"

// SieveScriptResponse represents a sieve script in API
type SieveScriptResponse struct {
	Name      string    `json:"name"`
	Content   string    `json:"content"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateSieveScriptRequest represents payload to create/update script
type CreateSieveScriptRequest struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}
