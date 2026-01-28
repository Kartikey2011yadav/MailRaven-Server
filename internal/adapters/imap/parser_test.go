package imap

import (
	"reflect"
	"testing"
)

func TestParseCommand(t *testing.T) {
	tests := []struct {
		input    string
		expected *Command
		hasError bool
	}{
		{
			input:    "A01 CAPABILITY",
			expected: &Command{Tag: "A01", Name: "CAPABILITY", Args: []string{}},
			hasError: false,
		},
		{
			input:    "A02 LOGIN user pass",
			expected: &Command{Tag: "A02", Name: "LOGIN", Args: []string{"user", "pass"}},
			hasError: false,
		},
		{
			input:    "A03 LOGIN \"user name\" \"pass word\"",
			expected: &Command{Tag: "A03", Name: "LOGIN", Args: []string{"user name", "pass word"}},
			hasError: false,
		},
		{
			input:    "  A04   NOOP  ",
			expected: &Command{Tag: "A04", Name: "NOOP", Args: []string{}},
			hasError: false,
		},
		{
			input:    "",
			expected: nil,
			hasError: true,
		},
	}

	for _, tt := range tests {
		cmd, err := ParseCommand(tt.input)
		if tt.hasError {
			if err == nil {
				t.Errorf("ParseCommand(%q) expected error, got nil", tt.input)
			}
			continue
		}

		if err != nil {
			t.Errorf("ParseCommand(%q) unexpected error: %v", tt.input, err)
			continue
		}

		if !reflect.DeepEqual(cmd, tt.expected) {
			t.Errorf("ParseCommand(%q) = %+v, expected %+v", tt.input, cmd, tt.expected)
		}
	}
}
