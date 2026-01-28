package imap

import (
	"fmt"
	"strings"
)

// Command represents a parsed IMAP command
type Command struct {
	Tag  string
	Name string
	Args []string
}

// ParseCommand parses a raw line into a Command struct
// Handles basic quoting: "arg with spaces"
func ParseCommand(line string) (*Command, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, fmt.Errorf("empty command")
	}

	parts := splitArgs(line)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid command format")
	}

	return &Command{
		Tag:  parts[0],
		Name: strings.ToUpper(parts[1]),
		Args: parts[2:],
	}, nil
}

// splitArgs splits string by space, respecting quotes
func splitArgs(input string) []string {
	var args []string
	var current strings.Builder
	inQuote := false

	for _, r := range input {
		if r == '"' {
			inQuote = !inQuote
			continue
		}

		if r == ' ' && !inQuote {
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		} else {
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args
}
