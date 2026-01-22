# Research: Modular Email Server

**Feature**: 001-mobile-email-server  
**Date**: 2026-01-22  
**Phase**: 0 - Technology Research & Decision Documentation

## Overview

This document consolidates research on key technology decisions for the modular email server. All technical unknowns from the planning phase have been resolved.

## Research Areas

### 1. CGO-Free SQLite: modernc.org/sqlite

**Decision**: Use `modernc.org/sqlite` (pure Go SQLite implementation)

**Rationale**:
- **CGO-Free Requirement**: Constitution mandates CGO-free for simplified deployment and cross-compilation
- **Performance**: Benchmarks show ~10-20% slower than cgo-based mattn/go-sqlite3, but acceptable for <1000 users
- **WAL Mode Support**: Full support for Write-Ahead Logging (improves concurrent read/write performance)
- **FTS5 Support**: Native support for SQLite FTS5 virtual tables (required for search)
- **Maintenance**: Actively maintained, regular updates tracking SQLite releases
- **Production Usage**: Used by several Go projects in production

**Alternatives Considered**:
- `mattn/go-sqlite3`: Requires CGO (rejected per constitution)
- `cznic/sqlite`: Predecessor to modernc.org/sqlite (deprecated, use modernc.org)
- Postgres: Overkill for MVP, complicates deployment (planned for future HA via interface migration)

**Best Practices**:
- Enable WAL mode: `PRAGMA journal_mode=WAL`
- Set synchronous mode: `PRAGMA synchronous=FULL` (required for durability)
- Use prepared statements to avoid SQL injection
- Connection pool size: 10-50 connections based on load
- Regular `PRAGMA optimize` for query planner statistics

### 2. SMTP Implementation: Porting from mox

**Decision**: Port SMTP logic from mox repository as reference implementation

**Rationale**:
- **Proven RFC Compliance**: mox has battle-tested SPF/DKIM/DMARC validation
- **Strict Parsing**: Follows RFC 5321 strictly, rejects malformed input
- **Educational Value**: Well-structured code with RFC cross-references in comments
- **License Compatibility**: MIT/Apache 2.0 compatible with MailRaven
- **Not a Dependency**: Use as reference/inspiration, not imported directly (avoids unnecessary dependencies)

**Key Components to Port**:
- SPF validation logic (RFC 7208): DNS lookups, mechanism evaluation, policy enforcement
- DKIM signature verification (RFC 6376): DNS key retrieval, signature validation, body hash verification
- DMARC policy evaluation (RFC 7489): Policy retrieval, alignment checks, action decisions
- SMTP protocol handling: EHLO, MAIL FROM, RCPT TO, DATA command processing
- MIME parsing: Multipart messages, header extraction, body part processing

**Alternatives Considered**:
- `emersion/go-smtp`: Solid SMTP server library, but lacks SPF/DKIM/DMARC validation
- `jordan-wright/email`: Basic SMTP sending, not suitable for server implementation
- `gomail`: Client-side only
- Build from scratch: High risk of RFC compliance bugs, security issues

**Best Practices from mox**:
- Inline RFC section comments (e.g., `// RFC 5321 section 4.5.3.1`)
- Comprehensive error messages citing RFC violations
- Separate protocol parsing from business logic
- DNS lookup caching with appropriate TTLs
- Connection timeouts and resource limits

### 3. Ports and Adapters Architecture (Hexagonal Architecture)

**Decision**: Implement Clean Architecture with Ports and Adapters pattern

**Rationale**:
- **Testability**: Core domain logic isolated from external dependencies (easy to mock)
- **Swappability**: Storage/search implementations are interfaces (SQLite → Postgres migration path)
- **Protocol Independence**: Business logic doesn't know about SMTP vs IMAP vs REST (protocol isolation principle)
- **Constitutional Alignment**: Directly supports Interface-Driven Design and Protocol Isolation principles
- **Future-Proofing**: Can add IMAP listener, Elasticsearch search, S3 storage without touching core logic

**Architecture Layers**:

```
┌──────────────────────────────────────────────────┐
│  Adapters (External)                             │
│  - internal/adapters/smtp (SMTP listener)        │
│  - internal/adapters/http (REST API)             │
│  - internal/adapters/storage/sqlite (DB impl)    │
│  - internal/adapters/storage/disk (File impl)    │
└─────────────────┬────────────────────────────────┘
                  │ depends on
                  ↓
┌──────────────────────────────────────────────────┐
│  Ports (Interfaces)                              │
│  - internal/core/ports/repositories.go           │
│    • EmailRepository                             │
│    • UserRepository                              │
│    • BlobStore                                   │
│    • SearchIndex                                 │
└─────────────────┬────────────────────────────────┘
                  │ used by
                  ↓
┌──────────────────────────────────────────────────┐
│  Core Domain (Pure Business Logic)               │
│  - internal/core/domain (Email, User, Mailbox)   │
│  - internal/core/services (EmailService, etc.)   │
└──────────────────────────────────────────────────┘
```

**Alternatives Considered**:
- **Layered Architecture**: Traditional N-tier (presentation → business → data). Less flexible, harder to swap implementations.
- **Monolithic**: All code in one package. Simple initially, but becomes unmaintainable. No testability.
- **Microservices**: Overkill for <1000 users. Adds complexity (distributed transactions, network calls). Not aligned with single-server MVP scope.

**Best Practices**:
- Keep domain models pure Go structs (no database tags, no JSON tags)
- Ports define interfaces as small as possible (1-3 methods per interface)
- Adapters depend on ports, never the reverse (Dependency Inversion Principle)
- Use dependency injection at application startup (wire adapters → ports → services)
- Domain services only depend on port interfaces, never concrete implementations

### 4. REST API Routing: go-chi/chi vs net/http

**Decision**: Use `go-chi/chi` v5 for REST API routing

**Rationale**:
- **Stdlib-Compatible**: Built on net/http, uses standard http.Handler interface
- **Lightweight**: Minimal abstraction over stdlib (~1000 LOC), no magic
- **Idiomatic**: Follows Go conventions, doesn't hide stdlib concepts
- **Middleware Support**: Clean middleware chaining (logging, auth, compression)
- **Path Parameters**: Better path param extraction than stdlib (`chi.URLParam(r, "id")`)
- **Submuxes**: Supports nested routers for `/api/v1` versioning
- **Constitutional Alignment**: Justifiable dependency (stdlib router is verbose, chi adds value without complexity)

**Alternatives Considered**:
- **net/http stdlib**: Verbose routing, no built-in path params, manual middleware chaining. Workable but tedious.
- **gorilla/mux**: Heavier than chi (~5000 LOC), more features than needed
- **gin**: Framework-style (too opinionated), doesn't feel like stdlib
- **echo**: Similar to gin, more magic than needed

**Best Practices with chi**:
- Use chi.Router for main router and submuxes
- Group versioned endpoints: `r.Route("/v1", func(r chi.Router) {...})`
- Apply middleware at appropriate levels (global for compression, route-specific for auth)
- Use chi.URLParam for path parameters instead of regex
- Keep handlers thin (call service layer, return JSON)

### 5. SQLite FTS5 for Full-Text Search

**Decision**: Use SQLite FTS5 virtual tables for initial search implementation

**Rationale**:
- **Zero Additional Dependencies**: FTS5 is built into SQLite
- **Good Performance**: Sub-200ms search for 10k messages (meets success criteria)
- **Simple Integration**: SQL-based queries, no separate search infrastructure
- **Tokenization**: Built-in tokenizers (unicode61, porter stemming)
- **Relevance Ranking**: BM25 ranking algorithm built-in
- **Migration Path**: SearchIndex interface allows future upgrade to Elasticsearch/Bleve

**Alternatives Considered**:
- **Elasticsearch**: Overkill for MVP (<50k messages). Requires separate service, JVM, clustering complexity.
- **Bleve**: Pure Go full-text search. Good option, but adds dependency. FTS5 sufficient for MVP.
- **PostgreSQL Full-Text**: Would require switching from SQLite (not aligned with MVP scope)
- **Manual LIKE queries**: Too slow, no relevance ranking

**FTS5 Design**:
```sql
-- Virtual table for full-text search
CREATE VIRTUAL TABLE messages_fts USING fts5(
    message_id UNINDEXED,  -- foreign key, not searched
    sender,                 -- searchable
    recipient,              -- searchable
    subject,                -- searchable
    body_text,              -- searchable (extracted from MIME)
    tokenize='porter unicode61'
);

-- Query with relevance ranking
SELECT message_id, rank
FROM messages_fts
WHERE messages_fts MATCH 'invoice'
ORDER BY rank
LIMIT 20 OFFSET 0;
```

**Best Practices**:
- Keep FTS5 table in sync with main messages table (triggers or application-level consistency)
- Index plaintext only (strip HTML from body parts)
- Limit indexed body length (first 10KB) to control index size
- Use `MATCH` operator for full-text queries (not LIKE)
- Prefix search with `*`: `MATCH 'invo*'` for autocomplete
- Combine with date/sender filters: `WHERE MATCH 'invoice' AND sender = 'billing@example.com'`

### 6. Configuration Management: YAML

**Decision**: Use YAML for configuration with `gopkg.in/yaml.v3`

**Rationale**:
- **Human-Readable**: Easier for sysadmins than JSON or TOML
- **Comments**: YAML supports comments for inline documentation
- **Multi-Line Strings**: Good for embedding DNS records in quickstart output
- **Stdlib-Compatible**: `gopkg.in/yaml.v3` is well-maintained, widely used
- **Validation**: Can unmarshal into Go structs with validation tags

**Config Structure**:
```yaml
# mailraven.yaml
domain: example.com
ports:
  smtp: 25
  https: 443

storage:
  type: sqlite  # future: postgres
  sqlite:
    path: /var/lib/mailraven/mail.db
    wal_mode: true
  blob_store:
    path: /var/lib/mailraven/messages
    compression: gzip

search:
  type: fts5  # future: elasticsearch

dkim:
  key_path: /etc/mailraven/dkim.key
  selector: mail

logging:
  level: info  # debug, info, warn, error
  format: json  # json or text

prometheus:
  enabled: true
  port: 9090
```

**Alternatives Considered**:
- **JSON**: No comments, less human-readable
- **TOML**: Good format, but less familiar to sysadmins than YAML
- **Environment Variables**: Works for 12-factor apps, but complex config gets messy
- **Go flags**: Tedious for many config options

**Best Practices**:
- Use struct tags for validation: `yaml:"domain" validate:"required,fqdn"`
- Provide defaults in code, override from file
- Support environment variable overrides for sensitive values (e.g., `MAILRAVEN_DKIM_KEY_PATH`)
- Validate config at startup (fail fast if misconfigured)
- Include example config in repository

### 7. Middleware Pattern for SMTP Pipeline

**Decision**: Implement chainable middleware for SMTP message processing

**Rationale**:
- **Extensibility Principle**: Constitutional requirement for pluggable spam/antivirus
- **Clean Separation**: Each validator (SPF, DKIM, DMARC) is independent middleware
- **Future-Proofing**: Can add spam filter, virus scanner without modifying core SMTP code
- **Testing**: Each middleware unit testable in isolation

**Middleware Interface**:
```go
// MessageHandler processes a message and optionally rejects it
type MessageHandler func(ctx context.Context, msg *Message) error

// Middleware wraps a MessageHandler
type Middleware func(MessageHandler) MessageHandler

// Chain multiple middleware
func Chain(handler MessageHandler, middleware ...Middleware) MessageHandler {
    for i := len(middleware) - 1; i >= 0; i-- {
        handler = middleware[i](handler)
    }
    return handler
}
```

**Middleware Examples**:
- **SPFValidator**: Checks SPF records, adds authentication result to message
- **DKIMValidator**: Verifies DKIM signatures
- **DMARCValidator**: Evaluates DMARC policy based on SPF/DKIM results
- **SpamFilter** (future): Scores message, rejects if spam score too high
- **VirusScanner** (future): Scans attachments, rejects if malware detected
- **Logger**: Logs message processing (session ID, sender, result)

**Best Practices**:
- Return error to reject message (SMTP 5xx response)
- Enrich message with metadata (SPF result, DKIM status) for downstream middleware
- Keep middleware stateless (no shared state between messages)
- Order matters: SPF/DKIM before DMARC, validators before storage

## Implementation Strategy

Based on research, the implementation will follow this sequence:

1. **Phase 1a: Define Core Domain & Ports** (2-3 days)
   - Create domain entities: `Email`, `User`, `Mailbox` structs
   - Define port interfaces: `EmailRepository`, `UserRepository`, `BlobStore`, `SearchIndex`
   - Pure Go code, no external dependencies

2. **Phase 1b: Implement SQLite Storage Adapter** (3-5 days)
   - SQLite schema with migrations
   - Implement `EmailRepository` (save, find, update read state)
   - Implement `UserRepository` (create, authenticate)
   - Implement `BlobStore` (write compressed message files)
   - Implement `SearchIndex` using FTS5
   - Integration tests with real SQLite database

3. **Phase 1c: Port mox SMTP Logic** (5-7 days)
   - SPF validation (DNS lookups, policy evaluation)
   - DKIM verification (signature parsing, DNS key retrieval, validation)
   - DMARC evaluation (policy retrieval, alignment checks)
   - SMTP protocol handler (EHLO, MAIL FROM, RCPT TO, DATA)
   - MIME parsing (multipart messages, body extraction)
   - Unit tests with table-driven test cases

4. **Phase 1d: Build SMTP Adapter** (2-3 days)
   - SMTP server using net/smtp and mox logic
   - Middleware pipeline (SPF → DKIM → DMARC → storage)
   - Connect SMTP listener to `EmailRepository` port
   - Durability guarantee: fsync before "250 OK"
   - Integration tests with SMTP client

5. **Phase 1e: Build REST API Adapter** (3-4 days)
   - go-chi router with /v1 endpoints
   - Authentication middleware (JWT)
   - Pagination, compression, field filtering
   - Handlers call domain services via ports
   - API integration tests

6. **Phase 1f: Quickstart Command** (2 days)
   - DKIM key generation
   - Config file generation
   - DNS record output (MX, SPF, DKIM, DMARC)
   - Integration test: quickstart → config → server start

**Total Estimated Implementation**: 17-24 days for full MVP

## Open Questions & Future Research

None. All technical requirements are clarified. Future enhancements (Postgres, Elasticsearch, IMAP) will require additional research but are out of MVP scope.

## References

- [modernc.org/sqlite](https://gitlab.com/cznic/sqlite) - CGO-free SQLite driver
- [mox GitHub](https://github.com/mjl-/mox) - Reference SMTP implementation
- [go-chi/chi](https://github.com/go-chi/chi) - HTTP router
- [SQLite FTS5 Documentation](https://www.sqlite.org/fts5.html) - Full-text search
- [Hexagonal Architecture](https://alistair.cockburn.us/hexagonal-architecture/) - Ports and Adapters pattern
- RFC 5321 (SMTP), RFC 7208 (SPF), RFC 6376 (DKIM), RFC 7489 (DMARC)
