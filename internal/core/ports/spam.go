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

	// CheckRecipient checks if the sender/recipient pair allows delivery (Greylisting).
	// Returns an error (typically transient) if the recipient is currently greylisted.
	CheckRecipient(ctx context.Context, ip, sender, recipient string) error

	// TrainSpam learns from spam content
	TrainSpam(ctx context.Context, content io.Reader) error

	// TrainHam learns from ham (non-spam) content
	TrainHam(ctx context.Context, content io.Reader) error
}

// Greylister defines the core logic for the grey-listing mechanism.
type Greylister interface {
	// Check determines if a tuple should be allowed or temporarily rejected.
	// Returns nil if allowed, or an error (indicating failure/blocking).
	Check(ctx context.Context, tuple domain.GreylistTuple) error

	// Prune removes expired entries.
	Prune(ctx context.Context) (int64, error)
}

// BayesClassifier defines the interface for content classification.
type BayesClassifier interface {
	// Classify returns the probability (0.0 - 1.0) that the content is spam.
	Classify(ctx context.Context, content io.Reader) (float64, error)
}

// BayesTrainer defines the interface for teaching the filter.
type BayesTrainer interface {
	TrainSpam(ctx context.Context, content io.Reader) error
	TrainHam(ctx context.Context, content io.Reader) error
}
