package bayesian

import (
	"context"
	"fmt"
	"io"

	"github.com/Kartikey2011yadav/mailraven-server/internal/core/ports"
)

// Trainer implements ports.BayesTrainer
type Trainer struct {
	repo ports.BayesRepository
}

// NewTrainer creates a new Bayesian trainer
func NewTrainer(repo ports.BayesRepository) *Trainer {
	return &Trainer{repo: repo}
}

// TrainSpam learns from spam content
func (t *Trainer) TrainSpam(ctx context.Context, content io.Reader) error {
	return t.train(ctx, content, true)
}

// TrainHam learns from ham (non-spam) content
func (t *Trainer) TrainHam(ctx context.Context, content io.Reader) error {
	return t.train(ctx, content, false)
}

func (t *Trainer) train(ctx context.Context, content io.Reader, isSpam bool) error {
	tokens, err := Tokenize(content)
	if err != nil {
		return fmt.Errorf("tokenization failed: %w", err)
	}

	if len(tokens) == 0 {
		return nil
	}

	// 1. Update Global Stats
	if err := t.repo.IncrementGlobal(ctx, isSpam); err != nil {
		return fmt.Errorf("failed to update global stats: %w", err)
	}

	// 2. Update Token Stats
	// Optimization: Batch update? Repo interface only has single IncrementToken.
	// For MVP, loop is fine (SQLite is fast, or we add Batch method later).
	// But `bayes_repo.go` (SQLite) might not handle connection contention well if we do 100s of writes per email.
	// However, usually training is rare (user action).

	// Dedup tokens per email?
	// Usually, we count presence (Bernoulli) or frequency (Multinomial).
	// Paul Graham uses presence (unique tokens per email).
	// `Tokenize` function I wrote ALREADY deduplicates!
	// "unique := make(map[string]struct{})" in tokenizer.go.
	// So `tokens` are unique. Good.

	for _, token := range tokens {
		if err := t.repo.IncrementToken(ctx, token, isSpam); err != nil {
			return fmt.Errorf("failed to update token %s: %w", token, err)
		}
	}

	return nil
}

// Check implementation
var _ ports.BayesTrainer = (*Trainer)(nil)
