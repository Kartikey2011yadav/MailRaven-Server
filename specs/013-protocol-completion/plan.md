# Implementation Plan: Protocol Completion

**Branch**: `013-protocol-completion` | **Date**: 2026-02-02 | **Spec**: [specs/013-protocol-completion/spec.md](specs/013-protocol-completion/spec.md)
**Input**: Feature specification from `specs/013-protocol-completion/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Implement full IMAP RFC compliance for Quotas (RFC 2087) and ACLs (RFC 4314) in MailRaven, referencing mox logic but building on MailRaven's internal Go architecture. Additionally, prepare contexts for Mobile Client development by exposing/documenting necessary API endpoints.

## Technical Context

**Language/Version**: Go 1.24.0
**Primary Dependencies**: `github.com/go-chi/chi/v5` (HTTP), `modernc.org/sqlite` (Storage), `github.com/golang-jwt/jwt/v5` (Auth)
**Storage**: SQLite (embedded, `modernc.org/sqlite`)
**Testing**: `github.com/stretchr/testify`
**Target Platform**: Linux (Container), Windows (Dev)
**Project Type**: Server (Go) + Client (React/Vite - separate context)
**Performance Goals**: Low latency for ACL checks (cached or indexed), minimal overhead for Quota tracking
**Constraints**: Must run in container; SQLite implementation must be concurrency-safe
**Scale/Scope**: < 1000 users per instance typical, focused on self-hosting / small orgs

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Verify compliance with [MailRaven Constitution](../../.specify/memory/constitution.md):

- [x] **Reliability**: Quota/ACL updates transactionally safe in SQLite.
- [x] **Protocol Parity**: explicit goal is RFC 2087/4314 parity, using mox as reference.
- [x] **Mobile-First Architecture**: APIs will be documented for mobile consumption; lean data limits.
- [x] **Dependency Minimalism**: Using existing `modernc.org/sqlite`, no new heavy DBs.
- [x] **Observability**: Metrics for "Quota Exceeded" and "ACL Denied" will be added.
- [x] **Interface-Driven Design**: Logic in `internal/core`, impl in `internal/adapters`.
- [x] **Extensibility**: ACL system allows future roles (e.g. "Audit").
- [x] **Protocol Isolation**: ACLs enforced at service level, protecting both IMAP and API.
- [x] **Testing Standards**: Unit tests for logic, functional tests for specialized scenarios.

**Violations Requiring Justification**: None

## Project Structure

### Documentation (this feature)

```text
specs/013-protocol-completion/
 plan.md              # This file
 research.md          # Reference analysis & decisions
 data-model.md        # Entities: Quota overrides, ACL maps
 quickstart.md        # How to configure quotas/AVLs via CLI/API
 contracts/           # OpenAPI for admin management of quotas/ACLs
 tasks.md             # Implementation tasks
```

### Source Code (repository root)

```text
internal/
 core/
    domain/
       mailbox.go     # Add ACL fields
       account.go     # Add Quota fields
    ports/
       repository.go  # Update signatures
       services.go    # Add ACL/Quota checking logic
    services/
        imap_service.go # Core logic application
 adapters/
    imap/
       server.go      # Add handlers for SETQUOTA, SETACL, etc.
       session.go     # Context aware of limits
       quota.go       # New file for RFC 2087
       acl.go         # New file for RFC 4314
    storage/
       sqlite/        # SQL schema updates + repo impl
    http/
        server.go      # Endpoints for mobile context
```

**Structure Decision**: Standard "Hexagonal" (Ports & Adapters) structure used in `internal/`.
