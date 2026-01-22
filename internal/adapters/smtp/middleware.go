package smtp

import (
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
)

// MessageHandler is a function that processes an SMTP message
type MessageHandler func(session *domain.SMTPSession, rawMessage []byte) error

// Middleware wraps a MessageHandler with additional processing
type Middleware func(MessageHandler) MessageHandler

// Chain combines multiple middleware into a single middleware
func Chain(middlewares ...Middleware) Middleware {
	return func(final MessageHandler) MessageHandler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			final = middlewares[i](final)
		}
		return final
	}
}
