# MailRaven-Server Development Guidelines

Auto-generated from all feature plans. Last updated: 2026-01-22

## Active Technologies
- SQLite + Filesystem (persistence required via Docker Volumes) (002-production-hardening)
- Go 1.25+ (006-security-imap)
- None for this feature (Logic only). (006-security-imap)
- Go 1.21+ + Go Standard Library (`net`, `crypto/tls`, `encoding/xml`, `net/textproto`) (007-standard-client-compliance)
- SQLite (`modernc.org/sqlite`) (007-standard-client-compliance)

- Go 1.22+ (requires generics support) (001-mobile-email-server)

## Project Structure

```text
src/
tests/
```

## Commands

# Add commands for Go 1.22+ (requires generics support)

## Code Style

Go 1.22+ (requires generics support): Follow standard conventions

## Recent Changes
- 007-standard-client-compliance: Added Go 1.21+ + Go Standard Library (`net`, `crypto/tls`, `encoding/xml`, `net/textproto`)
- 006-security-imap: Added Go 1.25+
- 002-production-hardening: Added Go 1.22


<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
