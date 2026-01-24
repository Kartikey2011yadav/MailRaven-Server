package mime

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/mail"
	"strings"
)

// ParsedMessage represents a parsed MIME message
type ParsedMessage struct {
	From        string
	To          string
	Subject     string
	MessageID   string
	PlainText   string
	HTML        string
	Snippet     string
	Attachments []Attachment
}

// Attachment represents an email attachment
type Attachment struct {
	Filename    string
	ContentType string
	Size        int64
}

// ParseMessage parses a raw MIME message
// Implements RFC 2045 - Multipurpose Internet Mail Extensions
func ParseMessage(rawMessage []byte) (*ParsedMessage, error) {
	// RFC 5322: Parse email headers
	msg, err := mail.ReadMessage(bytes.NewReader(rawMessage))
	if err != nil {
		return nil, fmt.Errorf("failed to parse message: %w", err)
	}

	parsed := &ParsedMessage{
		From:      msg.Header.Get("From"),
		To:        msg.Header.Get("To"),
		Subject:   msg.Header.Get("Subject"),
		MessageID: msg.Header.Get("Message-ID"),
	}

	// RFC 2045 Section 5: Content-Type header
	contentType := msg.Header.Get("Content-Type")
	if contentType == "" {
		// RFC 2045 Section 5.2: Default is text/plain
		contentType = "text/plain; charset=us-ascii"
	}

	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Content-Type: %w", err)
	}

	// RFC 2046 Section 5.1: Multipart messages
	if strings.HasPrefix(mediaType, "multipart/") {
		boundary := params["boundary"]
		if boundary == "" {
			return nil, fmt.Errorf("multipart message missing boundary")
		}

		if err := parseMultipart(msg.Body, boundary, parsed); err != nil {
			return nil, fmt.Errorf("failed to parse multipart: %w", err)
		}
	} else if mediaType == "text/plain" {
		// Simple text message
		bodyBytes, err := io.ReadAll(msg.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read body: %w", err)
		}
		parsed.PlainText = string(bodyBytes)
	} else if mediaType == "text/html" {
		// HTML message
		bodyBytes, err := io.ReadAll(msg.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read body: %w", err)
		}
		parsed.HTML = string(bodyBytes)
	}

	// Generate snippet (first 200 characters of plaintext)
	parsed.Snippet = generateSnippet(parsed.PlainText)

	return parsed, nil
}

// parseMultipart recursively parses multipart message parts
func parseMultipart(body io.Reader, boundary string, parsed *ParsedMessage) error {
	// RFC 2046 Section 5.1.1: Multipart boundary
	mr := multipart.NewReader(body, boundary)

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read part: %w", err)
		}

		contentType := part.Header.Get("Content-Type")
		mediaType, params, err := mime.ParseMediaType(contentType)
		if err != nil {
			continue
		}

		// RFC 2046 Section 5.1: Handle nested multipart
		if strings.HasPrefix(mediaType, "multipart/") {
			nestedBoundary := params["boundary"]
			if nestedBoundary != "" {
				if err := parseMultipart(part, nestedBoundary, parsed); err != nil {
					return err
				}
			}
			continue
		}

		// Read part content
		partBytes, err := io.ReadAll(part)
		if err != nil {
			continue
		}

		// RFC 2046 Section 5.1.4: text/plain parts
		if mediaType == "text/plain" {
			parsed.PlainText += string(partBytes)
		}

		// RFC 2046 Section 5.1.4: text/html parts
		if mediaType == "text/html" {
			parsed.HTML += string(partBytes)
		}

		// Handle attachments
		contentDisposition := part.Header.Get("Content-Disposition")
		if strings.HasPrefix(contentDisposition, "attachment") {
			_, dispParams, _ := mime.ParseMediaType(contentDisposition)
			filename := dispParams["filename"]
			if filename == "" {
				filename = "attachment"
			}

			parsed.Attachments = append(parsed.Attachments, Attachment{
				Filename:    filename,
				ContentType: mediaType,
				Size:        int64(len(partBytes)),
			})
		}
	}

	return nil
}

// generateSnippet creates a 200-character preview from text
func generateSnippet(text string) string {
	// Remove extra whitespace
	text = strings.TrimSpace(text)
	text = strings.ReplaceAll(text, "\r\n", " ")
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\t", " ")

	// Collapse multiple spaces
	for strings.Contains(text, "  ") {
		text = strings.ReplaceAll(text, "  ", " ")
	}

	// Truncate to 200 characters
	if len(text) > 200 {
		return text[:197] + "..."
	}

	return text
}

// ExtractPlainTextFromHTML extracts text content from HTML (basic implementation)
func ExtractPlainTextFromHTML(html string) string {
	// For MVP: Simple tag stripping
	// Production should use proper HTML parser
	text := html

	// Remove script and style tags with their content
	text = removeTagsWithContent(text, "script")
	text = removeTagsWithContent(text, "style")

	// Remove all HTML tags
	text = stripHTMLTags(text)

	return text
}

// removeTagsWithContent removes HTML tags and their content
func removeTagsWithContent(html, tag string) string {
	for {
		start := strings.Index(html, "<"+tag)
		if start == -1 {
			break
		}
		end := strings.Index(html[start:], "</"+tag+">")
		if end == -1 {
			break
		}
		html = html[:start] + html[start+end+len(tag)+3:]
	}
	return html
}

// stripHTMLTags removes all HTML tags
func stripHTMLTags(html string) string {
	inTag := false
	var result strings.Builder

	for _, char := range html {
		if char == '<' {
			inTag = true
		} else if char == '>' {
			inTag = false
		} else if !inTag {
			result.WriteRune(char)
		}
	}

	return result.String()
}
