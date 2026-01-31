package bayesian

import (
	"context"
	"fmt"
	"io"
	"math"
	"sort"

	"github.com/Kartikey2011yadav/mailraven-server/internal/core/ports"
)

// Classifier implements ports.BayesClassifier
type Classifier struct {
	repo ports.BayesRepository
}

// NewClassifier creates a new Bayesian classifier
func NewClassifier(repo ports.BayesRepository) *Classifier {
	return &Classifier{repo: repo}
}

// Classify calculates the probability that the content is spam.
func (c *Classifier) Classify(ctx context.Context, content io.Reader) (float64, error) {
	// 1. Tokenize
	tokens, err := Tokenize(content)
	if err != nil {
		return 0, fmt.Errorf("tokenization failed: %w", err)
	}
	if len(tokens) == 0 {
		return 0, nil // Neutral / Ham
	}

	// 2. Fetch Token Stats & Global Stats
	// Optimization: We could fetch in parallel
	stats, err := c.repo.GetGlobalStats(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch global stats: %w", err)
	}

	// Ensure we don't divide by zero
	totalSpam := float64(stats.TotalSpam)
	totalHam := float64(stats.TotalHam)
	if totalSpam == 0 && totalHam == 0 {
		return 0.0, nil // No training data
	}

	tokenStats, err := c.repo.GetTokens(ctx, tokens)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch token stats: %w", err)
	}

	// 3. Calculate Probabilities
	// Using Gary Robinson's method or Paul Graham's?
	// Paul Graham's approach with top 15 interesting tokens.

	var probs []float64

	for _, token := range tokens {
		tData := tokenStats[token]

		spamCount := 0.0
		hamCount := 0.0

		if tData != nil {
			spamCount = float64(tData.SpamCount)
			hamCount = float64(tData.HamCount)
		}

		// Calculate P(Spam|Word)
		// P(S|W) = (spamCount / totalSpam) / ((spamCount / totalSpam) + (hamCount / totalHam))
		// Handle 0 totals

		freqSpam := 0.0
		if totalSpam > 0 {
			freqSpam = spamCount / totalSpam
		}

		freqHam := 0.0
		if totalHam > 0 {
			freqHam = hamCount / totalHam
		}

		// Skip if word is unknown (0,0) - unless we want to assign default prob?
		// Graham uses default 0.4 for unknown? No, usually ignore?
		// Or Robinson uses assumed prob 0.5 with weight x.
		if tData == nil {
			// Unknown word
			// We can ignore it or give it 0.4
			// Let's use Robinson geometric mean later, so we collect relevant probs.
			continue
		}

		if freqSpam == 0 && freqHam == 0 {
			continue
		}

		prob := freqSpam / (freqSpam + freqHam)

		// Clamp (Graham's rule: 0.01 - 0.99)
		if prob < 0.01 {
			prob = 0.01
		}
		if prob > 0.99 {
			prob = 0.99
		}

		probs = append(probs, prob)
	}

	// If no known tokens found
	if len(probs) == 0 {
		return 0.4, nil // Unknown
	}

	// 4. Combine Probabilities
	// Sort by "interestingness" - distance from 0.5
	sort.Slice(probs, func(i, j int) bool {
		return math.Abs(probs[i]-0.5) > math.Abs(probs[j]-0.5)
	})

	// Take top N (e.g. 15)
	limit := 15
	if len(probs) < limit {
		limit = len(probs)
	}

	// Combined probability: P = (abc...) / (abc... + (1-a)(1-b)(1-c)...)
	product := 1.0
	inverseProduct := 1.0

	for i := 0; i < limit; i++ {
		p := probs[i]
		product *= p
		inverseProduct *= (1 - p)
	}

	finalProb := product / (product + inverseProduct)

	return finalProb, nil
}
