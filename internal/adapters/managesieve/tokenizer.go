package managesieve

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Tokenizer helps parsing ManageSieve commands
type Tokenizer struct {
	r *bufio.Reader
}

func NewTokenizer(r *bufio.Reader) *Tokenizer {
	return &Tokenizer{r: r}
}

// SkipWhitespace consumes spaces
func (t *Tokenizer) SkipWhitespace() error {
	for {
		b, err := t.r.Peek(1)
		if err != nil {
			return err
		}
		if b[0] == ' ' || b[0] == '\t' {
			t.r.ReadByte()
		} else {
			break
		}
	}
	return nil
}

// ReadWord reads an atom or quoted string
func (t *Tokenizer) ReadWord() (string, error) {
	if err := t.SkipWhitespace(); err != nil {
		return "", err
	}

	b, err := t.r.Peek(1)
	if err != nil {
		return "", err
	}

	if b[0] == '"' {
		return t.readQuoted()
	} else if b[0] == '{' {
		return t.readLiteral()
	}

	// Atom
	var sb strings.Builder
	for {
		b, err := t.r.Peek(1)
		if err != nil {
			if err == io.EOF && sb.Len() > 0 {
				return sb.String(), nil
			}
			return "", err
		}
		c := b[0]
		// Atoms end at space, control chars, or special delimiters
		if c <= ' ' || c == '(' || c == ')' || c == '{' || c == '"' {
			break
		}
		t.r.ReadByte()
		sb.WriteByte(c)
	}
	if sb.Len() == 0 {
		return "", fmt.Errorf("unexpected char: %c", b[0])
	}
	return sb.String(), nil
}

func (t *Tokenizer) readQuoted() (string, error) {
	t.r.ReadByte() // consume "
	var sb strings.Builder
	for {
		b, err := t.r.ReadByte()
		if err != nil {
			return "", err
		}
		if b == '"' {
			break
		}
		if b == '\\' {
			b, err = t.r.ReadByte()
			if err != nil {
				return "", err
			}
		}
		sb.WriteByte(b)
	}
	return sb.String(), nil
}

func (t *Tokenizer) readLiteral() (string, error) {
	// {123+}
	t.r.ReadByte() // {
	var lenStr strings.Builder
	for {
		b, err := t.r.ReadByte()
		if err != nil {
			return "", err
		}
		if b == '+' || b == '}' {
			// + is optionally used for sync literal? RFC 5804 uses default literal {len} or {len+}
			// {len+} means no verification necessary? No, in ManageSieve {len+} is fixed-size literal.
			// Actually RFC 5804 Section 4:
			// Strings can be "quoted" or {literal}
			// Literals: {<number>} CR LF <content>
			// ManageSieve allows {<number>+} CR LF <content>
			if b == '+' {
				t.r.ReadByte() // consume }
				break
			}
			if b == '}' {
				break
			}
			// Should be digit
			lenStr.WriteByte(b)
		} else {
			lenStr.WriteByte(b)
		}
	}

	// Read CRLF after }
	// Expect \r\n
	b, _ := t.r.ReadByte()
	if b == '\r' {
		t.r.ReadByte() // \n
	} else if b == '\n' {
		// fine
	}

	length, err := strconv.Atoi(lenStr.String())
	if err != nil {
		return "", fmt.Errorf("invalid literal length: %v", err)
	}

	buf := make([]byte, length)
	_, err = io.ReadFull(t.r, buf)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}
