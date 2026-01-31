package bayesian

import (
	"context"
	"strings"
	"testing"

	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/ports"
)

// Unit Test for Tokenizer
func TestTokenizer(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Basic Sentence",
			input:    "Hello World",
			expected: []string{"hello", "world"},
		},
		{
			name:     "Symbols and Numbers",
			input:    "Buy $1000 NOW!!!",
			expected: []string{"buy", "$1000", "now"}, // $ is allowed in my logic
		},
		{
			name:     "Short words",
			input:    "a an the test",
			expected: []string{"the", "test"}, // < 3 chars skipped? "the" is 3. "an" is 2.
		},
		{
			name:     "Unusual chars",
			input:    "test-drive",
			expected: []string{"test-drive"}, // - allowed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.input)
			tokens, err := Tokenize(r)
			if err != nil {
				t.Fatalf("Tokenize error: %v", err)
			}

			if len(tokens) != len(tt.expected) {
				t.Errorf("Expected %d tokens, got %d: %v", len(tt.expected), len(tokens), tokens)
			}

			for i, tok := range tokens {
				if i < len(tt.expected) && tok != tt.expected[i] {
					t.Errorf("Token %d: expected %s, got %s", i, tt.expected[i], tok)
				}
			}
		})
	}
}

// Mock Repository
type mockRepo struct {
	tokens    map[string]*domain.BayesToken
	totalSpam int
	totalHam  int
}

func (m *mockRepo) GetTokens(ctx context.Context, tokens []string) (map[string]*domain.BayesToken, error) {
	res := make(map[string]*domain.BayesToken)
	for _, t := range tokens {
		if val, ok := m.tokens[t]; ok {
			res[t] = val
		}
	}
	return res, nil
}

func (m *mockRepo) IncrementToken(ctx context.Context, token string, isSpam bool) error { return nil }

func (m *mockRepo) GetGlobalStats(ctx context.Context) (*domain.BayesGlobalStats, error) {
	return &domain.BayesGlobalStats{
		TotalSpam: m.totalSpam,
		TotalHam:  m.totalHam,
	}, nil
}

func (m *mockRepo) IncrementGlobal(ctx context.Context, isSpam bool) error { return nil }

var _ ports.BayesRepository = (*mockRepo)(nil)

// Test for Classifier
func TestClassifier(t *testing.T) {
	// Setup trained data
	// Let's say we have 100 spam emails and 100 ham emails.
	// "viagra" appears in 50 spam, 0 ham.
	// "meeting" appears in 0 spam, 50 ham.
	// "hello" appears in 50 spam, 50 ham.

	repo := &mockRepo{
		tokens: map[string]*domain.BayesToken{
			"viagra":  {Token: "viagra", SpamCount: 50, HamCount: 0},
			"meeting": {Token: "meeting", SpamCount: 0, HamCount: 50},
			"hello":   {Token: "hello", SpamCount: 50, HamCount: 50},
		},
		totalSpam: 100,
		totalHam:  100,
	}

	classifier := NewClassifier(repo)
	ctx := context.Background()

	// Case 1: Spam Content
	t.Run("Spam Content", func(t *testing.T) {
		content := strings.NewReader("Hello Viagra now")
		score, err := classifier.Classify(ctx, content)
		if err != nil {
			t.Fatalf("Classify failed: %v", err)
		}
		if score < 0.9 {
			t.Errorf("Expected high spam score for 'viagra', got %f", score)
		}
	})

	// Case 2: Ham Content
	t.Run("Ham Content", func(t *testing.T) {
		content := strings.NewReader("Hello Meeting today")
		score, err := classifier.Classify(ctx, content)
		if err != nil {
			t.Fatalf("Classify failed: %v", err)
		}
		if score > 0.1 {
			t.Errorf("Expected low spam score for 'meeting', got %f", score)
		}
	})

	// Case 3: Neutral Content
	t.Run("Neutral Content", func(t *testing.T) {
		// "Hello" is 50/50. "unknown" is ignored.
		content := strings.NewReader("Hello unknown")
		score, err := classifier.Classify(ctx, content)
		if err != nil {
			t.Fatalf("Classify failed: %v", err)
		}
		// Expect around 0.5
		if score < 0.4 || score > 0.6 {
			t.Errorf("Expected neutral score, got %f", score)
		}
	})
}
