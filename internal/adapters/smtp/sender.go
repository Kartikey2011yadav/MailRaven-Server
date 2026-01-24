package smtp

import (
	"context"
)

// Sender interface allows mocking the SMTP client
type Sender interface {
	Send(ctx context.Context, from string, recipient string, data []byte) error
}
