package ports

import (
	"context"
	"io"

	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
)

// SpamFilter defines the interface for checking if a connection or message is spam.
type SpamFilter interface {
	// CheckConnection checks if the incoming connection IP is allowed.
	// Returns an error if the connection should be rejected.
	CheckConnection(ctx context.Context, ip string) error

	// CheckContent checks the message content for spam.
	CheckContent(ctx context.Context, content io.Reader, headers map[string]string) (*domain.SpamCheckResult, error)
}
