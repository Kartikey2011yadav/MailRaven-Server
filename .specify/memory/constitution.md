<!--
SYNC IMPACT REPORT
==================
Version Change: INITIAL → 1.0.0
Created: 2026-01-22
Type: Initial Constitution

Principles Established:
- Code Quality: Standards compliance, Go idioms, RFC cross-referencing
- Testing Standards: Comprehensive test coverage, interoperability validation
- User Experience Consistency: Protocol compliance, client compatibility
- Performance Requirements: Operational metrics, resource efficiency

Templates Status:
✅ plan-template.md - Constitution Check section ready
✅ spec-template.md - Requirements section aligns
✅ tasks-template.md - Testing discipline aligned
✅ checklist-template.md - No updates needed
✅ agent-file-template.md - No updates needed

Follow-up Actions: None - all templates compatible with established principles.
-->

# Mox Constitution

## Core Principles

### I. Code Quality and Standards Compliance

Code MUST adhere to the highest quality standards with explicit traceability:

- **RFC Cross-Referencing**: All protocol implementation code MUST include inline comments
  referencing specific RFC sections (format: `// RFC 5321 section 4.5.3.1`).
- **Go Idioms**: Code MUST follow idiomatic Go practices including error handling,
  interface design, and standard library patterns.
- **Package Reusability**: Non-server packages MUST be designed as standalone, reusable
  libraries with clear boundaries and minimal dependencies.
- **Code Documentation**: All exported types, functions, and packages MUST have godoc
  comments explaining purpose, behavior, and usage.
- **Static Analysis**: Code MUST pass `go vet`, `staticcheck`, and configured linters
  without warnings before merge.

**Rationale**: Email protocols are complex and standards-heavy. RFC cross-referencing
ensures correctness and maintainability. Reusable packages enable external adoption and
focused testing.

### II. Testing Standards

Comprehensive testing is NON-NEGOTIABLE for reliability:

- **Test Coverage**: All new code MUST include tests. Minimum 80% coverage for core email
  handling logic (SMTP, IMAP, DKIM, SPF, DMARC).
- **Unit Tests**: Every package MUST have `*_test.go` files testing public APIs and critical
  internal logic.
- **Integration Tests**: SMTP, IMAP, and protocol implementations MUST have integration tests
  validating interoperability with reference implementations (e.g., Postfix, Dovecot).
- **Fuzz Testing**: Parser code (email headers, DKIM signatures, DNS records) MUST include
  fuzz tests to discover edge cases and security vulnerabilities.
- **Manual Interoperability**: New email client features MUST be manually tested against
  major clients (Thunderbird, iOS Mail, Android Mail, Outlook) before release.
- **Test-First Discipline**: For bug fixes, write a failing test that reproduces the issue
  before implementing the fix.

**Rationale**: Email infrastructure requires bulletproof reliability. Integration and fuzz
testing catch real-world issues that unit tests miss. Manual client testing ensures UX
consistency.

### III. User Experience Consistency

Email clients expect predictable, standards-compliant behavior:

- **Protocol Compliance**: MUST implement SMTP, IMAP4, and related protocols according to
  RFCs. No custom extensions that break standard clients.
- **Client Compatibility**: Features MUST work correctly with the big four ecosystems:
  Gmail, Outlook.com, Yahoo Mail, and self-hosted clients (Thunderbird, mutt).
- **Error Messages**: User-facing errors (webmail, admin interface, CLI) MUST be clear,
  actionable, and avoid technical jargon where possible.
- **Configuration Simplicity**: The quickstart process MUST get a domain operational in
  under 10 minutes for users with basic DNS knowledge.
- **Upgrade Path**: Configuration file changes MUST provide clear migration guides and
  automated upgrade tools where feasible.

**Rationale**: Email is critical infrastructure. Users cannot tolerate proprietary quirks or
complex setup. Standards compliance prevents vendor lock-in.

### IV. Performance Requirements

Operational efficiency is mandatory for self-hosted deployments:

- **Resource Efficiency**: Mox MUST run on modest hardware (2GB RAM, 2 CPU cores) serving
  100+ users with 10k+ messages per day.
- **Response Times**: IMAP operations MUST respond within 100ms for folders with <10k
  messages. SMTP delivery attempts MUST begin within 1 second of queue insertion.
- **Metrics and Observability**: All critical operations (message delivery, IMAP commands,
  authentication) MUST export Prometheus metrics for monitoring.
- **Structured Logging**: Operations MUST use structured logging (mlog package) with
  appropriate log levels to enable debugging without excessive noise.
- **Database Performance**: Queries MUST use appropriate indexes. The database schema MUST
  be tested with 100k+ messages to validate performance characteristics.

**Rationale**: Self-hosted deployments run on limited resources. Observable systems enable
operators to diagnose issues before they escalate. Performance must be validated, not assumed.

## Security Requirements

Email security is paramount:

- **TLS by Default**: All network services (SMTP, IMAP, HTTP admin) MUST use TLS.
  Let's Encrypt integration MUST be automatic via ACME.
- **Authentication**: Admin and webmail interfaces MUST require strong authentication.
  Password policies MUST enforce minimum 12 characters.
- **Input Validation**: All external inputs (email headers, DNS records, API requests)
  MUST be validated and sanitized against injection attacks.
- **Dependency Hygiene**: Dependencies MUST be reviewed for security advisories.
  Update vulnerable dependencies within 7 days of disclosure.
- **Audit Logging**: Security-sensitive operations (login attempts, configuration changes,
  message deletions) MUST be logged with sufficient detail for forensic analysis.

**Rationale**: Email is a high-value attack target. Defense in depth requires secure defaults,
proactive dependency management, and comprehensive audit trails.

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

**Version**: 1.0.0 | **Ratified**: 2026-01-22 | **Last Amended**: 2026-01-22
