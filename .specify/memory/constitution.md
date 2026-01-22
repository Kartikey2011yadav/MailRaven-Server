<!--
SYNC IMPACT REPORT
==================
Version Change: 1.2.0 → 1.3.0
Amendment Date: 2026-01-22
Type: Architecture Principles Addition (MINOR)

Changes:
- NEW: Principle VI - Interface-Driven Design (repository pattern, storage abstraction)
- NEW: Principle VII - Extensibility (middleware pattern for SMTP pipeline)
- REVISED: Principle I - Reliability (clarified atomic writes requirement)
- NEW: Principle VIII - Protocol Isolation (separation of core logic from access protocols)
- Core principles expanded from 5 to 8 to support future-proof architecture

Section Updates:
- Testing Standards: Add interface mock testing requirements
- Go Code Standards: Add interface design guidelines

Templates Status:
✅ plan-template.md - Constitution Check section to be updated
✅ spec-template.md - Requirements section aligns
✅ tasks-template.md - Testing discipline aligned
✅ checklist-template.md - No updates needed
✅ agent-file-template.md - No updates needed

Rationale for Changes:
- Repository pattern enables database migration (SQLite → Postgres) without business logic changes
- Middleware pattern allows spam filtering and antivirus injection without core SMTP changes
- Protocol isolation enables supporting both IMAP and REST API simultaneously
- Interface-driven design improves testability and maintainability

Follow-up Actions: Update plan-template.md with new architecture principle checks.
-->

# MailRaven Constitution

## Core Principles

### I. Reliability - Email Acceptance is a Commitment (NON-NEGOTIABLE)

Once we reply "250 OK" to SMTP, the data MUST be durably persisted:

- **250 OK Means Synced**: Before sending "250 OK" response to SMTP, the message MUST be
  written to both SQLite database AND file storage, with `fsync()` called on both. No
  exceptions.
- **Dual Storage Strategy**: Messages stored in SQLite (metadata, headers, indexing) AND
  as files (full message body). Both must succeed atomically before acknowledging receipt.
- **Atomic Transactions**: Use SQLite transactions with PRAGMA synchronous=FULL. A message
  is either fully committed (DB + file + fsync) or rejected. All writes MUST be atomic -
  no partial states. Use database transactions and ensure rollback on any failure.
- **Crash Recovery**: On startup, verify SQLite integrity (`PRAGMA integrity_check`) and
  reconcile file storage. Orphaned files or DB entries must be logged and handled gracefully.
- **Testing for Durability**: Include tests that kill -9 the server immediately after "250 OK"
  and verify the message is recoverable on restart.
- **No Shortcuts**: Never use async writes, write-ahead logs, or delayed fsync. Email
  acceptance is a promise we MUST keep, even if it costs performance.

**Rationale**: Email is irreplaceable. Users trust us when we accept their messages. Data loss
destroys reputation and trust. Fsync overhead (10-50ms) is acceptable compared to the cost of
lost email. Atomic operations prevent data corruption.

### II. Protocol Parity - Match Mox's RFC Adherence

MailRaven follows mox's proven approach to email protocols:

- **Strict RFC Implementation**: SPF (RFC 7208), DKIM (RFC 6376), DMARC (RFC 7489), SMTP
  (RFC 5321), IMAP4 (RFC 3501) must be implemented according to specifications. Study mox's
  implementation as reference.
- **No Lenient Parsing**: Reject malformed headers, invalid DKIM signatures, and RFC
  violations. Match mox's strict parsing behavior. Better to reject than accept broken email.
- **SPF/DKIM/DMARC Checks**: All inbound email MUST pass through SPF, DKIM, and DMARC
  verification. Results stored for each message. Mobile clients can display authentication
  status.
- **DNS Validation**: Validate SPF records, DKIM keys, and DMARC policies from DNS. Cache
  results with appropriate TTLs. Handle DNSSEC where available.
- **Cross-Reference with RFCs**: Protocol implementation code MUST include inline comments
  referencing specific RFC sections (format: `// RFC 5321 section 4.5.3.1`).
- **Interoperability Testing**: Test against major mail providers (Gmail, Outlook, Yahoo)
  and verify SPF/DKIM/DMARC headers are correctly interpreted.

**Rationale**: Email protocols are complex and security-critical. Mox has battle-tested
implementations. Strict adherence prevents security vulnerabilities and deliverability issues.

### III. Mobile-First Architecture

MailRaven is a backend for mobile apps. APIs MUST optimize for mobile constraints:

- **Low-Bandwidth Design**: APIs return only essential data. Use pagination (limit/offset or
  cursor-based) for all list endpoints. Default page size: 20-50 items.
- **Delta Updates**: Support incremental sync (e.g., "messages since timestamp", "changes
  since version"). Avoid forcing clients to re-download full mailboxes.
- **Compression**: Support gzip/brotli compression for API responses. Especially important
  for message bodies and attachment metadata.
- **High-Latency Tolerance**: APIs must be idempotent. Retry-safe operations (use idempotency
  keys for non-idempotent operations like "send email").
- **Offline-First Considerations**: Mobile clients may cache data and sync later. APIs should
  support conflict resolution (last-write-wins or version vectors).
- **Minimal Payloads**: Use JSON with concise field names. Consider field filtering
  (e.g., `?fields=id,subject,from,date` to reduce response size).
- **Push Notifications**: Support webhooks or push notification integration for new message
  delivery (avoid polling where possible).

**Rationale**: Mobile networks are slow and expensive. Users expect fast, responsive apps even
on 3G/4G. Battery life matters. Bandwidth-efficient APIs improve UX and reduce infrastructure
costs.

### IV. Dependency Minimalism

Minimize external dependencies. Prefer Go standard library:

- **CGO-Free Mandate**: Use `modernc.org/sqlite` for SQLite access (pure Go, no CGO). This
  enables cross-compilation and simplifies deployment.
- **Standard Library First**: Use `net/http`, `encoding/json`, `crypto/*`, `net/smtp`,
  `net/textproto` from stdlib. Only add external dependencies when stdlib is clearly
  insufficient.
- **Justified Dependencies**: Each external dependency must be documented with rationale.
  Consider maintenance burden, security surface area, and licensing.
- **No Frameworks**: Avoid heavyweight frameworks (ORMs, web frameworks, DI containers).
  Direct use of stdlib is clearer and more maintainable.
- **Dependency Audit**: Review all dependencies quarterly for security advisories. Update
  vulnerable dependencies within 7 days of disclosure.
- **Vendoring**: Consider vendoring critical dependencies to protect against upstream
  breakage or disappearance.

**Rationale**: Dependencies increase attack surface, deployment complexity, and maintenance
burden. CGO complicates cross-compilation. Pure Go enables easy deployment to any platform
(Linux, macOS, Windows, ARM).

### V. Observability - SMTP Interaction Logging

All SMTP interactions MUST be logged structurally for debugging delivery issues:

- **Structured Logging**: Use structured logging (e.g., `log/slog` or custom structured
  logger) for all SMTP operations. Every log entry includes: timestamp, session ID,
  remote IP, sender, recipients, message ID, operation, result.
- **SMTP Session Tracing**: Log complete SMTP sessions: connection established, EHLO/HELO,
  MAIL FROM, RCPT TO, DATA, delivery result (250 OK or error), disconnection. Include
  SPF/DKIM/DMARC results.
- **Delivery Debugging**: When delivery fails, logs must contain enough context for operators
  to diagnose: rejected by which check? SPF fail? DKIM invalid? Greylisting? Quota exceeded?
- **Performance Metrics**: Log duration of expensive operations: DNS lookups, SPF/DKIM
  verification, database writes, file fsync. Alert on P95/P99 latency regressions.
- **Prometheus Metrics**: Export key metrics: messages received/delivered/rejected, SMTP
  sessions, SPF/DKIM/DMARC pass/fail counts, storage operations, API request counts.
- **Privacy Considerations**: Don't log message content or sensitive headers. Log metadata
  only (IDs, timestamps, sizes, authentication results).

**Rationale**: Email delivery issues are hard to debug without detailed logs. Operators need
visibility into why messages were accepted or rejected. Structured logs enable automated
analysis and alerting.

### VI. Interface-Driven Design

Storage and business logic MUST be decoupled through Go interfaces:

- **Repository Pattern**: Define storage operations as Go interfaces (e.g., `MessageRepository`,
  `UserRepository`). Business logic depends on interfaces, not concrete implementations.
- **Interface First**: Before implementing any storage code, define the interface contract.
  Interfaces should be minimal and focused (typically 5-10 methods per repository).
- **Implementation Independence**: SQLite implementation lives in `internal/storage/sqlite/`,
  Postgres in `internal/storage/postgres/`. Core business logic in `internal/service/` knows
  nothing about database choice.
- **Constructor Injection**: Use dependency injection via constructors. Services receive
  repository interfaces as parameters, not concrete types.
- **Mock-Friendly Testing**: Interfaces enable mock implementations for unit tests. No need
  for real database in unit tests of business logic.
- **Future Database Migration**: The repository pattern MUST allow swapping SQLite for
  Postgres without changing business logic code. Only update DI wiring in `main.go`.

**Rationale**: Tight coupling to SQLite makes future migration painful. Repository pattern
enables database evolution without rewriting business logic. Testability improves dramatically
with interface mocking.

### VII. Extensibility - Middleware Pattern

The SMTP ingestion pipeline MUST support middleware for extensibility:

- **Middleware Chain**: SMTP message processing uses a middleware/handler chain pattern.
  Each middleware can inspect, modify, or reject messages before passing to next handler.
- **Core Pipeline Unchanged**: Adding spam filtering or antivirus scanning MUST NOT require
  modifying core SMTP listener code. Middleware is registered at startup.
- **Middleware Interface**: Define `type Middleware func(ctx context.Context, msg *Message, next Handler) error`.
  Each middleware decides whether to call `next()` or short-circuit the chain.
- **Standard Middleware**: Provide built-in middleware for SPF validation, DMARC checks, size
  limits. Custom middleware (spam, antivirus) can be inserted at any point.
- **Configuration-Driven**: Middleware chain order is defined in config file, not hardcoded.
  Operators can enable/disable/reorder middleware without code changes.
- **Observable Pipeline**: Each middleware logs entry/exit with duration. Operators can see
  which middleware rejected a message and why.

**Rationale**: Email processing requirements evolve. Spam filtering and antivirus are expensive
to add if tightly coupled to SMTP listener. Middleware pattern enables zero-downtime feature
additions via configuration.

### VIII. Protocol Isolation

Core email storage logic MUST be independent of access protocol:

- **Protocol-Agnostic Core**: The `internal/service/` package handles email storage, retrieval,
  and manipulation WITHOUT knowing about IMAP, REST API, or any protocol details.
- **Protocol Adapters**: IMAP server in `internal/imapserver/` and REST API in `internal/api/`
  are thin adapters that translate protocol requests to service method calls.
- **Shared Business Logic**: Mark-as-read, delete message, fetch message body - these operations
  have single implementation in service layer, callable from any protocol adapter.
- **Simultaneous Protocols**: System MUST support both IMAP (port 143) and REST API (port 443)
  simultaneously. Both access same storage through same service interfaces.
- **Protocol-Specific Logic**: Authentication, session management, wire protocol are isolated
  in protocol adapters. Core services receive authenticated user context, not protocol details.
- **Future Protocol Addition**: Adding WebDAV or JMAP protocol adapter MUST NOT require
  changes to core service layer. Only add new adapter calling existing service methods.

**Rationale**: MailRaven starts with REST API but users may want IMAP compatibility later.
Protocol isolation enables supporting multiple access methods without core logic duplication.
Reduces testing burden - test core logic once, protocol adapters separately.

## Testing Standards

Comprehensive testing is NON-NEGOTIABLE for reliability:

- **Test Coverage**: All new code MUST include tests. Minimum 80% coverage for core email
  handling logic (SMTP, SPF/DKIM/DMARC validation, storage operations).
- **Unit Tests**: Every package MUST have `*_test.go` files testing public APIs and critical
  internal logic. Use table-driven tests with subtests for multiple scenarios.
- **Interface Mock Testing**: Business logic (services) MUST be tested with mock repository
  implementations. No real database required for unit tests. Use interfaces to inject test
  doubles.
- **Integration Tests**: SMTP protocol implementation MUST have integration tests validating
  interoperability with major mail senders (Gmail, Outlook, Yahoo).
- **Middleware Testing**: Each SMTP middleware MUST have unit tests. Test middleware chain
  composition with various orderings to ensure correct pass-through behavior.
- **Durability Tests**: Test crash recovery with kill -9 simulation immediately after "250 OK".
  Verify message is recoverable on restart with both SQLite and file storage intact.
- **SPF/DKIM/DMARC Tests**: Test all validation paths: valid signatures, invalid signatures,
  missing signatures, DNS lookup failures, policy evaluation edge cases.
- **Mobile API Tests**: Test pagination, delta updates, and bandwidth-constrained scenarios.
  Verify APIs work correctly with high-latency/flaky connections.
- **Protocol Adapter Tests**: Test IMAP and REST API adapters independently. Verify they
  correctly translate protocol requests to service method calls without business logic
  duplication.
- **CGO-Free Verification**: CI must verify `CGO_ENABLED=0` builds succeed. Prevents
  accidental CGO dependencies.

**Rationale**: Email infrastructure requires bulletproof reliability. Durability tests validate
our "250 OK" commitment. Protocol tests catch RFC compliance issues. Interface mocking enables
fast unit tests without database overhead. Protocol adapter tests ensure proper isolation.

## Protocol Compliance - Matching Mox's Standards

MailRaven adopts mox's proven approach to RFC compliance:

- **SPF Validation (RFC 7208)**: Implement complete SPF validation matching mox. Support all
  mechanisms (a, mx, ip4, ip6, include, exists) and qualifiers (+, -, ~, ?). Handle DNS
  lookup limits (10 lookups max).
- **DKIM Validation (RFC 6376)**: Verify DKIM signatures on all inbound email. Support RSA
  and Ed25519 keys. Validate signature headers, body hashes, and key retrieval from DNS.
  Match mox's strict parsing.
- **DMARC Policy Evaluation (RFC 7489)**: Retrieve and evaluate DMARC policies. Combine
  SPF and DKIM results. Support all policy actions (none, quarantine, reject). Log
  aggregate data for DMARC reporting.
- **SMTP Protocol (RFC 5321)**: Implement SMTP receiver strictly. Support ESMTP extensions
  (SIZE, STARTTLS, AUTH). Match mox's error handling and response codes.
- **Strict Parsing**: No lenient mode. Reject malformed headers, invalid signatures, and
  RFC violations. Study mox's implementation for edge cases.
- **Interoperability**: Test against Gmail, Outlook, Yahoo, and other major providers.
  Verify SPF/DKIM/DMARC results match industry behavior.

**Rationale**: Mox has battle-tested implementations refined over years. Matching their approach
inherits their expertise and avoids common pitfalls. Strict parsing prevents security issues.

## Mobile API Design Principles

APIs optimized for mobile client constraints:

- **RESTful Conventions**: Use standard HTTP methods (GET, POST, PUT, DELETE) and status
  codes. JSON request/response bodies.
- **Pagination Required**: All list endpoints MUST support pagination. Use cursor-based
  pagination for large datasets. Include `next_cursor` in responses.
- **Minimal Responses**: Return only requested data. Support field filtering via query
  params (e.g., `?fields=id,subject,from`). Default to essential fields only.
- **Versioned APIs**: API paths include version (`/v1/messages`). Breaking changes require
  new version. Maintain backward compatibility for at least 6 months.
- **Idempotency Keys**: Non-idempotent operations (send email, mark read) accept optional
  `Idempotency-Key` header. Duplicate requests with same key return cached result.
- **Delta Sync Support**: Provide endpoints for incremental updates (e.g., `/messages/changes?since=timestamp`).
  Reduce data transfer on app launch.
- **Compression**: Enable gzip compression for responses >1KB. Mobile clients can request
  compression via `Accept-Encoding: gzip`.

**Rationale**: Mobile networks are slow and expensive. Pagination prevents overwhelming clients.
Delta sync enables fast app launches. Idempotency handles retry scenarios gracefully.

## Security Requirements

Security is built into every layer:

- **TLS Mandatory**: All API endpoints MUST use HTTPS. SMTP submission MUST use STARTTLS
  or direct TLS. No plain-text protocols in production.
- **Authentication**: Mobile API uses token-based auth (JWT or opaque tokens). SMTP
  submission uses SASL AUTH. Admin endpoints require separate admin credentials.
- **Rate Limiting**: Prevent abuse with rate limits: 10 req/sec per IP for unauthenticated
  endpoints, 100 req/sec for authenticated users. Stricter limits on email sending.
- **Input Validation**: Validate all inputs (email addresses, headers, API params). Reject
  invalid data early. Use allowlists for email domains if applicable.
- **Audit Logging**: Log all security-sensitive operations: logins, password changes, email
  sends, configuration updates. Include IP address, user ID, timestamp.
- **Dependency Security**: Quarterly security audits of all dependencies. Update vulnerable
  dependencies within 7 days.

**Rationale**: Mobile backends are internet-facing and must be secure by default. Rate limiting
prevents abuse. Audit logs enable forensic analysis of security incidents.

## Performance Requirements

Optimize for mobile backend use cases:

- **Low-Latency APIs**: Mobile API endpoints MUST respond within 100ms (P95) for simple
  queries, 500ms for complex operations. Measure and alert on regressions.
- **SMTP Acceptance Speed**: "250 OK" response MUST be sent within 200ms of DATA completion
  (including fsync). This is the user-facing latency for email acceptance.
- **Database Performance**: SQLite queries MUST use appropriate indexes. Test performance
  with 100k+ messages. Use EXPLAIN QUERY PLAN to verify index usage.
- **Connection Pooling**: Reuse database connections. Connection pool size: 10-50 based on
  load. Monitor pool exhaustion.
- **Mobile Client Considerations**: APIs should minimize round trips. Batch operations where
  possible (e.g., mark multiple messages read in one request).
- **Fsync Overhead Acceptable**: Data safety takes priority over speed. 10-50ms fsync latency
  is acceptable for "250 OK" response.
- **Resource Efficiency**: Backend should run on modest hardware (2GB RAM, 2 CPU cores) serving
  100+ mobile users with 1000+ messages/day.

**Rationale**: Mobile users expect fast responses. Latency adds up across network hops. Database
indexing is critical for performance at scale. Data safety cannot be compromised for speed.

## Go Code Standards

Follow idiomatic Go practices:

- **Error Handling**: Never ignore errors. Return and wrap errors with context using
  `fmt.Errorf("context: %w", err)`. Handle errors at appropriate levels. No panic() in
  production paths.
- **Repository Pattern Interfaces**: Define repository interfaces in `internal/repository/`
  package. Keep interfaces small (5-10 methods). Each repository manages one entity type
  (MessageRepository, UserRepository).
- **Dependency Injection**: Use constructor injection. Services receive dependencies as
  interface parameters: `func NewEmailService(repo MessageRepository) *EmailService`.
- **Interface Design**: Accept interfaces, return structs. Define interfaces where they're
  consumed, not where they're implemented (consumer-driven interface design).
- **Middleware Pattern**: SMTP middleware signature: `type Middleware func(ctx context.Context, msg *Message, next Handler) error`.
  Middleware calls `next()` to continue chain or returns error to short-circuit.
- **Protocol Adapters**: Protocol-specific code lives in `internal/imapserver/` and
  `internal/api/`. Adapters translate protocol to service method calls without business logic.
- **Table-Driven Tests**: Use table-driven tests with subtests for multiple scenarios.
  Example: `t.Run(tc.name, func(t *testing.T) {...})`.
- **Godoc Comments**: All exported types, functions, packages MUST have godoc comments.
  Comments explain purpose and usage, not just restate the signature.
- **RFC Cross-References**: Protocol implementation code includes inline comments referencing
  RFC sections (e.g., `// RFC 7208 section 4.6.4`).
- **Static Analysis**: Code MUST pass `go vet`, `staticcheck`, and `golangci-lint` before
  merge. Configure linters to catch common issues.

**Rationale**: Idiomatic Go code is maintainable and benefits from tooling. Explicit error
handling prevents silent failures. Good documentation reduces onboarding time.

## Development Workflow

Process ensures quality without bureaucracy:

- **Branch Strategy**: Feature branches from main. Short-lived branches (<2 weeks).
  Direct commits to main prohibited.
- **Code Review**: All changes require review. Focus on correctness, RFC compliance, test
  coverage, and security implications.
- **Pre-Merge Checks**: CI MUST validate: `go test -race`, `go vet`, `staticcheck`,
  `golangci-lint`, integration tests pass, documentation builds.
- **Breaking Changes**: Public API changes require version bump and migration guide.
  Configuration file changes require `mox config test` validation.
- **Documentation Updates**: Feature PRs MUST update relevant docs (README, godoc,
  website/docs) before merge.

**Rationale**: Lightweight process scales with small teams. Automation catches issues early.
Documentation debt prevents feature understanding decay.

## Governance

This constitution defines non-negotiable principles:

- **Primacy**: The constitution supersedes convenience. Technical debt requires explicit
  justification and remediation plan.
- **Amendments**: Constitution changes require broad consensus (maintainer agreement).
  Version bump follows semantic versioning: MAJOR for removed principles, MINOR for added
  principles/sections, PATCH for clarifications.
- **Compliance Verification**: Code reviews MUST explicitly check constitution adherence,
  particularly testing standards and RFC cross-referencing.
- **Runtime Guidance**: The `.specify/templates/agent-file-template.md` provides agent-specific
  development guidance and MUST be consulted for active features.
- **Enforcement**: Constitution violations discovered post-merge MUST be remediated within
  one release cycle (typically 4 weeks) or explicitly documented as technical debt with
  timeline.

**Version**: 1.3.0 | **Ratified**: 2026-01-22 | **Last Amended**: 2026-01-22
