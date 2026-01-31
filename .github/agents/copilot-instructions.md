# MailRaven-Server Development Guidelines

Auto-generated from all feature plans. Last updated: 2026-01-22

## Active Technologies
- SQLite + Filesystem (persistence required via Docker Volumes) (002-production-hardening)
- Go 1.25+ (006-security-imap)
- None for this feature (Logic only). (006-security-imap)
- Go 1.21+ + Go Standard Library (`net`, `crypto/tls`, `encoding/xml`, `net/textproto`) (007-standard-client-compliance)
- SQLite (`modernc.org/sqlite`) (007-standard-client-compliance)
- Go 1.23 + Go Standard Library (`strings`, `database/sql`). Potential need for a lightweight stemmer/tokenizer (e.g., `snowball`) but prefer simple strict splitting if sufficient. (009-advanced-spam-filtering)
- SQLite / Postgres (via existing `gorm` or `sql` adapters). (009-advanced-spam-filtering)
- Go 1.22+ + `github.com/emersion/go-sieve` (likely candidate for interpreter/protocol), `github.com/emersion/go-sasl` (for ManageSieve auth). (010-sieve-filtering)
- SQLite (scripts table, vacation tracking table). (010-sieve-filtering)
- Go 1.24.0 + `github.com/emersion/go-sieve` (Interpreter/Protocol), `github.com/emersion/go-sasl` (Authentication) (010-sieve-filtering)
- SQLite (Tables: `sieve_scripts`, `vacation_trackers`) (010-sieve-filtering)

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
- 010-sieve-filtering: Added Go 1.24.0 + `github.com/emersion/go-sieve` (Interpreter/Protocol), `github.com/emersion/go-sasl` (Authentication)
- 010-sieve-filtering: Added Go 1.22+ + `github.com/emersion/go-sieve` (likely candidate for interpreter/protocol), `github.com/emersion/go-sasl` (for ManageSieve auth).
- 009-advanced-spam-filtering: Added Go 1.23 + Go Standard Library (`strings`, `database/sql`). Potential need for a lightweight stemmer/tokenizer (e.g., `snowball`) but prefer simple strict splitting if sufficient.


<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
