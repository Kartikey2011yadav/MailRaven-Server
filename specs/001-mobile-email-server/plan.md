# Implementation Plan: Modular Email Server

**Branch**: `001-mobile-email-server` | **Date**: 2026-01-22 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-mobile-email-server/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Build a modular, layered email server with clean architecture (Ports and Adapters pattern). The system implements 5 distinct layers: Listener (SMTP), Logic (routing, SPF/DMARC), Storage (repository interfaces), Search (FTS5), and API (REST/JSON). Current implementation uses SMTP + SQLite + FTS5 + REST API, with interfaces designed for future migration to IMAP/POP3 listeners, Postgres + S3 storage, and Elasticsearch search. Technical approach: Define core domain entities and port interfaces first, implement SQLite storage adapter, then integrate mox SMTP logic.

## Technical Context

**Language/Version**: Go 1.22+ (requires generics support)  
**Primary Dependencies**: 
- `modernc.org/sqlite` (CGO-free SQLite driver)
- `mox` reference for SPF/DKIM/DMARC validation logic
- `go-chi/chi` or `net/http` for REST API routing
- Go standard library for SMTP protocol, MIME parsing, crypto

**Storage**: 
- SQLite with WAL mode and FTS5 for full-text search
- File system with gzip compression for message bodies
- Repository interfaces: `EmailRepository`, `UserRepository`, `BlobStore`, `SearchIndex`

**Testing**: 
- Go standard `testing` package
- Interface mocks for unit testing
- Integration tests against real SQLite
- Durability tests (kill -9 after SMTP "250 OK")
- SMTP protocol tests using `net/smtp` client

**Target Platform**: Linux server (Ubuntu 20.04+), CGO-free binary  
**Project Type**: Backend service (single project)  

**Performance Goals**: 
- SMTP: Handle 100 concurrent connections, <50ms average processing time
- API: <500ms P95 latency for list endpoints over 4G
- Search: <200ms query time for mailboxes with 10k messages

**Constraints**: 
- Zero data loss requirement: "250 OK" = fsync complete
- CGO-free for simplified deployment
- <100MB memory for API serving 100 concurrent mobile clients
- Atomic writes (SQLite transaction + file + fsync before SMTP ack)

**Scale/Scope**: 
- MVP: <1000 users, <50,000 total messages
- Single-server deployment (HA via future Postgres+S3 migration)
- Single domain support (multi-domain is future enhancement)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Verify compliance with [MailRaven Constitution](../../.specify/memory/constitution.md):

- [x] **Reliability**: Feature ensures "250 OK" = fsync for both SQLite (PRAGMA synchronous=FULL) and file storage. All writes are atomic via transactions. Storage adapter implements durability contract.
- [x] **Protocol Parity**: SPF/DKIM/DMARC validation will use mox's implementation patterns. SMTP protocol handling follows RFC 5321 strictly. Code includes RFC cross-references.
- [x] **Mobile-First Architecture**: REST API supports pagination (cursor-based), gzip compression, delta sync (messages since timestamp). Field filtering planned for bandwidth optimization.
- [x] **Dependency Minimalism**: Uses `modernc.org/sqlite` (CGO-free), Go stdlib for SMTP/MIME/HTTP. Mox used as reference, not imported as dependency. go-chi for routing (lightweight, stdlib-compatible).
- [x] **Observability**: SMTP session logging with structured logs (session ID, remote IP, SPF/DMARC results). Prometheus metrics for messages received/rejected, API latency, storage operations.
- [x] **Testing Standards**: Test plan includes unit tests (interface mocks), integration tests (real SQLite), durability tests (kill -9 after "250 OK"), protocol tests (SMTP client), mobile API tests (pagination, latency).
- [x] **Security**: TLS for HTTPS API, STARTTLS for SMTP, JWT authentication, bcrypt passwords, rate limiting (100 req/min per IP), input validation for all endpoints.

**Additional Constitutional Principles (v1.3.0)**:
- [x] **Interface-Driven Design**: Storage layer uses repository pattern (EmailRepository, UserRepository, BlobStore, SearchIndex interfaces). Business logic depends on ports, not concrete implementations.
- [x] **Extensibility**: Logic layer designed with middleware pattern for future spam filtering. SMTP pipeline will support pluggable validators (SPF, DMARC, spam, virus scanning).
- [x] **Protocol Isolation**: Core logic separated from protocols. Listener layer (SMTP) and API layer (REST) are adapters. Future IMAP/POP3 listeners can be added without modifying storage/logic.

**Violations Requiring Justification**: None. Architecture fully aligns with constitutional principles.

## Project Structure

### Documentation (this feature)

```text
specs/001-mobile-email-server/
├── plan.md              # This file (implementation plan)
├── research.md          # Phase 0 output (technology decisions)
├── data-model.md        # Phase 1 output (domain entities, ports, schema)
├── quickstart.md        # Phase 1 output (admin setup guide)
├── contracts/           # Phase 1 output (API specifications)
│   └── rest-api.yaml    # OpenAPI 3.0 specification
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
mailraven-server/
├── cmd/
│   └── mailraven/
│       └── main.go                 # Application entrypoint
│
├── internal/                       # Private application code
│   ├── core/                       # Core domain (Clean Architecture center)
│   │   ├── domain/                 # Domain entities (pure structs)
│   │   │   ├── message.go          # Message, MessageBody
│   │   │   ├── user.go             # User, AuthToken
│   │   │   └── smtp.go             # SMTPSession
│   │   │
│   │   ├── ports/                  # Port interfaces (contracts)
│   │   │   ├── repositories.go     # EmailRepository, UserRepository
│   │   │   ├── storage.go          # BlobStore
│   │   │   └── search.go           # SearchIndex
│   │   │
│   │   └── services/               # Domain services (business logic)
│   │       ├── email_service.go    # Email processing logic
│   │       └── auth_service.go     # Authentication logic
│   │
│   ├── adapters/                   # Adapters (Clean Architecture outer layer)
│   │   ├── smtp/                   # SMTP listener adapter
│   │   │   ├── server.go           # SMTP server implementation
│   │   │   ├── handler.go          # SMTP command handler
│   │   │   ├── middleware.go       # Middleware interface & chain
│   │   │   ├── validators/         # SPF/DKIM/DMARC middleware
│   │   │   │   ├── spf.go          # SPF validation (ported from mox)
│   │   │   │   ├── dkim.go         # DKIM verification (ported from mox)
│   │   │   │   └── dmarc.go        # DMARC evaluation (ported from mox)
│   │   │   └── mime/               # MIME parsing
│   │   │       └── parser.go       # Multipart message parser
│   │   │
│   │   ├── http/                   # REST API adapter
│   │   │   ├── server.go           # HTTP server setup (go-chi router)
│   │   │   ├── routes.go           # Route definitions
│   │   │   ├── handlers/           # HTTP handlers
│   │   │   │   ├── auth.go         # /auth/login endpoint
│   │   │   │   ├── messages.go     # /messages endpoints
│   │   │   │   └── search.go       # /messages/search endpoint
│   │   │   ├── middleware/         # HTTP middleware
│   │   │   │   ├── auth.go         # JWT validation
│   │   │   │   ├── logging.go      # Request logging
│   │   │   │   ├── compression.go  # Gzip compression
│   │   │   │   └── ratelimit.go    # Rate limiting
│   │   │   └── dto/                # Data transfer objects (JSON structs)
│   │   │       └── message.go      # MessageSummary, MessageFull
│   │   │
│   │   └── storage/                # Storage adapters
│   │       ├── sqlite/             # SQLite adapter (implements ports)
│   │       │   ├── migrations/     # SQL schema migrations
│   │       │   │   └── 001_init.sql
│   │       │   ├── email_repo.go   # EmailRepository implementation
│   │       │   ├── user_repo.go    # UserRepository implementation
│   │       │   └── search_repo.go  # SearchIndex implementation (FTS5)
│   │       │
│   │       └── disk/               # File system adapter
│   │           └── blob_store.go   # BlobStore implementation (gzip)
│   │
│   ├── config/                     # Configuration management
│   │   ├── config.go               # Config struct and loader (YAML)
│   │   └── dkim.go                 # DKIM key generation
│   │
│   └── observability/              # Logging and metrics
│       ├── logger.go               # Structured logging (log/slog)
│       └── metrics.go              # Prometheus metrics
│
├── tests/                          # Integration tests
│   ├── smtp_test.go                # SMTP protocol tests
│   ├── api_test.go                 # REST API tests
│   ├── durability_test.go          # Crash recovery tests (kill -9)
│   └── testdata/                   # Test fixtures (sample emails)
│
├── deployment/                     # Deployment artifacts
│   ├── mailraven.service           # Systemd service file
│   └── config.example.yaml         # Example configuration
│
├── go.mod                          # Go module definition
├── go.sum                          # Go dependency checksums
├── Makefile                        # Build automation
└── README.md                       # Project overview
```

**Structure Decision**: Ports and Adapters (Hexagonal Architecture) with Go standard project layout.

**Key Principles**:
- **Internal package**: All application code in `internal/` (not importable by external projects)
- **Core isolation**: `internal/core/` has no dependencies on adapters or external packages
- **Dependency direction**: Adapters depend on ports (interfaces), never the reverse
- **Clear boundaries**: SMTP, HTTP, and storage are separate adapters with no cross-dependencies
- **Testability**: Core logic uses interfaces, easily mocked in tests

**Package Organization**:
- `cmd/mailraven/`: Binary entrypoint, wires adapters to ports via dependency injection
- `internal/core/domain/`: Pure domain entities (no tags, no external deps)
- `internal/core/ports/`: Interface definitions (contracts for adapters)
- `internal/core/services/`: Business logic using port interfaces
- `internal/adapters/smtp/`: SMTP protocol handling, calls email service
- `internal/adapters/http/`: REST API serving, calls email service
- `internal/adapters/storage/`: Implements repository interfaces (SQLite + disk)

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| N/A | No constitutional violations | All principles satisfied |

## Implementation Phases

Based on research findings (see [research.md](./research.md)), implementation will proceed in this sequence:

### Phase 1a: Core Domain & Ports (2-3 days)

**Goal**: Define domain entities and interface contracts

**Deliverables**:
- `internal/core/domain/message.go` - Message, MessageBody entities
- `internal/core/domain/user.go` - User, AuthToken entities
- `internal/core/ports/repositories.go` - EmailRepository, UserRepository interfaces
- `internal/core/ports/storage.go` - BlobStore interface
- `internal/core/ports/search.go` - SearchIndex interface
- Unit tests for domain validation logic

**Success Criteria**: Code compiles, interfaces are well-defined, no external dependencies in `core/`

### Phase 1b: SQLite Storage Adapter (3-5 days)

**Goal**: Implement storage layer with durability guarantees

**Deliverables**:
- `internal/adapters/storage/sqlite/migrations/001_init.sql` - Schema with indexes
- `internal/adapters/storage/sqlite/email_repo.go` - EmailRepository implementation
- `internal/adapters/storage/sqlite/user_repo.go` - UserRepository implementation
- `internal/adapters/storage/sqlite/search_repo.go` - SearchIndex with FTS5
- `internal/adapters/storage/disk/blob_store.go` - BlobStore with gzip compression
- Integration tests with real SQLite database
- Durability test: kill -9 after Save(), verify recovery

**Success Criteria**: All repository methods work correctly, PRAGMA synchronous=FULL enforced, FTS5 queries return relevant results

### Phase 1c: Mox SMTP Logic Port (5-7 days)

**Goal**: Port SPF/DKIM/DMARC validation from mox reference implementation

**Deliverables**:
- `internal/adapters/smtp/validators/spf.go` - SPF validation (RFC 7208)
- `internal/adapters/smtp/validators/dkim.go` - DKIM verification (RFC 6376)
- `internal/adapters/smtp/validators/dmarc.go` - DMARC evaluation (RFC 7489)
- `internal/adapters/smtp/mime/parser.go` - MIME multipart parsing
- Table-driven unit tests with test vectors from RFCs
- Inline RFC section comments (e.g., `// RFC 7208 section 4.6.4`)

**Success Criteria**: SPF/DKIM/DMARC validation matches mox behavior on test corpus, all RFC edge cases handled

### Phase 1d: SMTP Server Adapter (2-3 days)

**Goal**: Build SMTP listener with middleware pipeline

**Deliverables**:
- `internal/adapters/smtp/server.go` - SMTP server (net/smtp based)
- `internal/adapters/smtp/handler.go` - EHLO, MAIL FROM, RCPT TO, DATA handlers
- `internal/adapters/smtp/middleware.go` - Middleware interface and Chain function
- Integration of SPF/DKIM/DMARC validators as middleware
- Connection to EmailRepository and BlobStore ports
- Structured logging (session ID, remote IP, SPF/DMARC results)
- Integration test: send email via SMTP client, verify storage

**Success Criteria**: SMTP server accepts connections, validates SPF/DKIM, stores messages atomically, sends "250 OK" only after fsync

### Phase 1e: REST API Adapter (3-4 days)

**Goal**: Build mobile-optimized REST API

**Deliverables**:
- `internal/adapters/http/server.go` - HTTP server with go-chi router
- `internal/adapters/http/routes.go` - Route definitions (/v1/messages, etc.)
- `internal/adapters/http/handlers/auth.go` - /auth/login with JWT
- `internal/adapters/http/handlers/messages.go` - List, get, update endpoints
- `internal/adapters/http/handlers/search.go` - /messages/search endpoint
- `internal/adapters/http/middleware/` - Auth, logging, compression, rate limiting
- `internal/adapters/http/dto/` - JSON DTOs matching OpenAPI spec
- API integration tests with httptest

**Success Criteria**: All endpoints match OpenAPI spec, JWT auth works, pagination works, compression enabled, rate limiting enforced

### Phase 1f: Quickstart Command (2 days)

**Goal**: Automated server setup

**Deliverables**:
- `cmd/mailraven/main.go` - CLI with `quickstart` subcommand
- `internal/config/dkim.go` - DKIM key generation (RSA-2048)
- `internal/config/config.go` - YAML config file generation
- DNS record formatting (MX, SPF, DKIM, DMARC)
- User creation with bcrypt password hashing
- Integration test: quickstart → config → server start → send email

**Success Criteria**: Quickstart generates valid config, DKIM keys, DNS records; server starts successfully with generated config

### Phase 1g: Observability (1-2 days)

**Goal**: Logging and metrics

**Deliverables**:
- `internal/observability/logger.go` - Structured logging with log/slog
- `internal/observability/metrics.go` - Prometheus metrics (messages received/rejected, API latency, storage operations)
- SMTP middleware for logging sessions
- HTTP middleware for request logging
- Metrics endpoint at /metrics

**Success Criteria**: All SMTP sessions logged with session ID, API requests logged with latency, Prometheus metrics exported

**Total Estimated Time**: 17-24 days for full MVP implementation

## Artifacts Generated

This plan execution has generated the following artifacts in `specs/001-mobile-email-server/`:

- ✅ **plan.md** (this file) - Implementation plan with technical context, constitution check, project structure
- ✅ **research.md** - Technology decisions (modernc.org/sqlite, go-chi, FTS5, Ports and Adapters pattern)
- ✅ **data-model.md** - Domain entities, port interfaces, SQLite schema, data flows
- ✅ **contracts/rest-api.yaml** - OpenAPI 3.0 specification for mobile REST API
- ✅ **quickstart.md** - Administrator setup guide with DNS configuration

## Next Steps

1. **Review this plan**: Ensure technical decisions align with project goals
2. **Run `/speckit.tasks`**: Break down implementation phases into actionable tasks
3. **Begin Phase 1a**: Start with core domain and ports (no external dependencies)
4. **Iterate**: Build incrementally, test continuously, maintain constitutional compliance

## References

- Feature Specification: [spec.md](./spec.md)
- Constitution: [../../.specify/memory/constitution.md](../../.specify/memory/constitution.md)
- OpenAPI Spec: [contracts/rest-api.yaml](./contracts/rest-api.yaml)
