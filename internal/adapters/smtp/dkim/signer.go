package dkim

import (
	"bufio"
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
	"time"
)

// Signer handles DKIM signing of email messages
type Signer struct {
	Domain     string
	Selector   string
	PrivateKey *rsa.PrivateKey
}

// NewSigner creates a new DKIM signer
func NewSigner(domain, selector string, privateKey *rsa.PrivateKey) *Signer {
	return &Signer{
		Domain:     domain,
		Selector:   selector,
		PrivateKey: privateKey,
	}
}

// Sign calculates the DKIM-Signature header for the given email data
// headersToSign is a list of header keys to include in the signature (e.g. "From", "To", "Subject", "Date", "Message-ID")
func (s *Signer) Sign(data []byte, headersToSign []string) (string, error) {
	// 1. Split headers and body
	parts := splitMessage(data)
	headerBytes := parts[0]
	bodyBytes := parts[1]

	// 2. Canonicalize and hash body (Relaxed)
	bh, err := bodyHash(bodyBytes)
	if err != nil {
		return "", fmt.Errorf("failed to hash body: %w", err)
	}

	// 3. Parse headers
	parsedHeaders, err := parseHeaders(headerBytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse headers: %w", err)
	}

	// 4. Select headers to sign
	var signedHeaders []string
	var headersToHash []Header

	for _, hKey := range headersToSign {
		// Find matching header (bottom-up is standard, but simple top-down is acceptable for single occurrence)
		// RFC 6376 5.4.2: "Signers usually sign the last occurrence... to prevent insertion"
		// implementation: Find ALL occurrences and sign the LAST one that isn't already used?
		// Simplifying: Find the last occurrence of the header
		var foundIdx = -1
		for i := len(parsedHeaders) - 1; i >= 0; i-- {
			if strings.EqualFold(parsedHeaders[i].Key, hKey) && !parsedHeaders[i].Used {
				foundIdx = i
				break
			}
		}

		if foundIdx != -1 {
			parsedHeaders[foundIdx].Used = true
			headersToHash = append(headersToHash, parsedHeaders[foundIdx])
			signedHeaders = append(signedHeaders, parsedHeaders[foundIdx].Key)
		}
	}

	if len(signedHeaders) == 0 {
		return "", fmt.Errorf("no headers to sign found")
	}

	// 5. Construct DKIM-Signature value
	dkimHeaderVal := fmt.Sprintf(
		"v=1; a=rsa-sha256; c=relaxed/relaxed; d=%s; s=%s; t=%d; h=%s; bh=%s; b=",
		s.Domain,
		s.Selector,
		time.Now().Unix(),
		strings.Join(signedHeaders, ":"),
		bh,
	)

	// 6. Canonicalize and hash headers
	h := sha256.New()

	for _, hdr := range headersToHash {
		canon := relaxedHeader(hdr.Key, hdr.Value)
		h.Write([]byte(canon + "\r\n"))
	}

	// Hash the DKIM-Signature header itself (key + value without b's value)
	canonDkim := relaxedHeader("DKIM-Signature", dkimHeaderVal)
	h.Write([]byte(canonDkim))

	hashed := h.Sum(nil)

	// 7. Sign the hash
	signature, err := rsa.SignPKCS1v15(rand.Reader, s.PrivateKey, crypto.SHA256, hashed)
	if err != nil {
		return "", fmt.Errorf("failed to sign: %w", err)
	}

	// 8. Base64 encode
	sigBase64 := base64.StdEncoding.EncodeToString(signature)

	return "DKIM-Signature: " + dkimHeaderVal + sigBase64, nil
}

// Helpers

type Header struct {
	Key   string
	Value string
	Used  bool
}

func splitMessage(data []byte) [][]byte {
	// Standard separation is CRLF CRLF
	parts := bytes.SplitN(data, []byte("\r\n\r\n"), 2)
	if len(parts) == 2 {
		return parts
	}
	// Fallback to LF LF
	parts = bytes.SplitN(data, []byte("\n\n"), 2)
	if len(parts) == 2 {
		return parts
	}
	// No body or header only?
	return [][]byte{data, []byte{}}
}

func parseHeaders(raw []byte) ([]Header, error) {
	var headers []Header
	reader := bufio.NewReader(bytes.NewReader(raw))

	var currentKey string
	var currentValueBuilder strings.Builder

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil && err != io.EOF {
			return nil, err
		}

		lineStr := string(line)
		trimmed := strings.TrimRight(lineStr, "\r\n")

		if trimmed == "" {
			if err == io.EOF {
				break
			}
			// Empty line usually means end of headers, but we pre-split, so this might be internal?
			// in splitMessage we took the first block. So internal empty lines shouldn't exist in valid headers.
			// Treat as end.
			break
		}

		// Check for continuation (tab or space)
		if lineStr[0] == ' ' || lineStr[0] == '\t' {
			if currentKey == "" {
				// Should not happen if headers are valid
				continue
			}
			// Append with CRLF to preserve structure for unfolding
			currentValueBuilder.WriteString("\r\n") // Synthetic CRLF for internal representation
			currentValueBuilder.WriteString(trimmed)
		} else {
			// Save previous
			if currentKey != "" {
				headers = append(headers, Header{Key: currentKey, Value: currentValueBuilder.String()})
			}

			// New header
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) != 2 {
				// Invalid header line, skip or error. Valid email headers must have colon.
				// We'll skip invalid lines to be robust
				currentKey = ""
				continue
			}

			currentKey = strings.TrimSpace(parts[0])
			currentValueBuilder.Reset()
			currentValueBuilder.WriteString(strings.TrimSpace(parts[1])) // Trim leading space of value
		}

		if err == io.EOF {
			break
		}
	}

	if currentKey != "" {
		headers = append(headers, Header{Key: currentKey, Value: currentValueBuilder.String()})
	}

	return headers, nil
}

func bodyHash(body []byte) (string, error) {
	h := sha256.New()

	scanner := bufio.NewScanner(bytes.NewReader(body))
	// Increase buffer size just in case
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 10*1024*1024)

	var lines []string
	for scanner.Scan() {
		line := scanner.Text()

		// 1. Trim trailing whitespace
		line = strings.TrimRight(line, " \t")

		// 2. Reduce WSP sequences to single SP
		var b strings.Builder
		lastSpace := false
		for _, r := range line {
			if r == ' ' || r == '\t' {
				if !lastSpace {
					b.WriteRune(' ')
					lastSpace = true
				}
			} else {
				b.WriteRune(r)
				lastSpace = false
			}
		}
		lines = append(lines, b.String())
	}

	// 3. Remove trailing empty lines
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	// 4. Hash lines with CRLF
	for _, line := range lines {
		h.Write([]byte(line + "\r\n"))
	}

	// If body is empty (was all empty lines or empty), empty body yields one CRLF
	if len(lines) == 0 {
		h.Write([]byte("\r\n"))
	}

	return base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
}

func relaxedHeader(key, value string) string {
	lowerKey := strings.ToLower(strings.TrimSpace(key))

	// Unfold: remove \r\n (we added synthetic ones in parseHeaders or they are natural)
	cleanValue := strings.ReplaceAll(value, "\r\n", "")
	cleanValue = strings.ReplaceAll(cleanValue, "\n", "")

	// Compress WSP
	var b strings.Builder
	lastSpace := false
	for _, r := range cleanValue {
		if r == ' ' || r == '\t' {
			if !lastSpace {
				b.WriteRune(' ')
				lastSpace = true
			}
		} else {
			b.WriteRune(r)
			lastSpace = false
		}
	}

	finalValue := strings.TrimSpace(b.String())

	return lowerKey + ":" + finalValue
}
