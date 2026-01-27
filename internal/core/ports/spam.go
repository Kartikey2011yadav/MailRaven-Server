package ports

import "context"

// SpamFilter defines the interface for checking if a connection or message is spam.
type SpamFilter interface {
	// CheckConnection checks if the incoming connection IP is allowed.
	// Returns and error if the connection should be rejected.
	CheckConnection(ctx context.Context, ip string) error
}
