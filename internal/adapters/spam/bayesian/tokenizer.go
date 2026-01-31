package bayesian

import (
	"bufio"
	"io"
	"strings"
	"unicode"
)

// Tokenize splits text into tokens for Bayesian analysis.
// It normalizes to lowercase and ignores distinct non-letter characters.
// We keep it simple: Split by any non-letter, keep tokens length > 2, max length 20.
func Tokenize(r io.Reader) ([]string, error) {
	scanner := bufio.NewScanner(r)
	// Set split function to split into words
	scanner.Split(splitWords)

	unique := make(map[string]struct{})
	var tokens []string

	for scanner.Scan() {
		word := strings.ToLower(scanner.Text())

		// Filter
		if len(word) < 3 || len(word) > 20 {
			continue
		}

		// Skip words starting with digit (usually noise)
		if unicode.IsDigit(rune(word[0])) {
			continue
		}

		if _, exists := unique[word]; !exists {
			unique[word] = struct{}{}
			tokens = append(tokens, word)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return tokens, nil
}

// splitWords is a custom split function that splits on non-letter characters
func splitWords(data []byte, atEOF bool) (advance int, token []byte, err error) {
	start := 0
	for ; start < len(data); start++ {
		r := rune(data[start])
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '$' || r == '-' { // Allow letters, digits, $, -
			break
		}
	}

	for i := start; i < len(data); i++ {
		r := rune(data[i])
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '$' && r != '-' {
			return i + 1, data[start:i], nil
		}
	}

	if atEOF && len(data) > start {
		return len(data), data[start:], nil
	}

	return start, nil, nil
}
