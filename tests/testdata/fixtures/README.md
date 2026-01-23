# Test Fixtures for MailRaven

This directory contains sample email messages and test data for integration testing.

## Email Messages

### simple-plain.eml
- Plain text email with multiple paragraphs
- Tests basic SMTP parsing and snippet extraction
- No attachments, no HTML

### multipart-html.eml
- Multipart/alternative email with both plain text and HTML versions
- Tests MIME parsing
- Useful for testing HTML rendering in mobile clients

### with-attachment.eml
- Email with a PDF attachment
- Tests multipart/mixed MIME handling
- Base64-encoded attachment
- Attachment: document.pdf (minimal test PDF)

### unicode-emoji.eml
- Tests Unicode and emoji handling
- Includes:
  - UTF-8 encoded subject line
  - Multiple languages (English, Spanish, Japanese)
  - Emoji characters
  - Special symbols (©, ®, €, etc.)
  - Mathematical symbols

## DKIM Keys

### dkim-test-private.key
- RSA private key for DKIM signing tests
- **FOR TESTING ONLY** - Never use in production
- Corresponds to public key in DNS TXT record

## Usage in Tests

```go
// Load test fixture
data, err := os.ReadFile("testdata/fixtures/simple-plain.eml")
if err != nil {
    t.Fatal(err)
}

// Parse MIME message
msg, err := mail.ReadMessage(bytes.NewReader(data))
```

## Adding New Fixtures

When adding test fixtures:
1. Use realistic email structure (RFC 5322 compliant)
2. Include Date, Message-ID, From, To headers
3. Document the purpose in this README
4. Keep attachments small (< 50KB)
5. Use example.com domains to avoid real addresses
