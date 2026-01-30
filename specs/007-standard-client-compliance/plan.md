# Implementation Plan: [FEATURE]

**Branch**: `[###-feature-name]` | **Date**: [DATE] | **Spec**: [link]
**Input**: Feature specification from `/specs/[###-feature-name]/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Implement core IMAP functionality (SELECT, FETCH, UID, IDLE) and Autodiscover endpoints to support standard email clients like Outlook and Thunderbird, bridging the compatibility gap between MailRaven and reference implementations like Mox.

## Technical Context

**Language/Version**: Go 1.21+
**Primary Dependencies**: Go Standard Library (`net`, `crypto/tls`, `encoding/xml`, `net/textproto`)
**Storage**: SQLite (`modernc.org/sqlite`)
**Testing**: Go `testing` package
**Target Platform**: Linux/Windows/Mac Server
**Project Type**: Single Backend Server
**Performance Goals**: Support 100+ concurrent IDLE connections.
**Constraints**: Must run on low-resource VPS. Must be thread-safe.

**Unknowns**:
- [Resolved in research.md] UID Mapping strategy.
- [Resolved in research.md] IDLE concurrency pattern.
- [Resolved in research.md] Autodiscover XML schemas.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Verify compliance with [MailRaven Constitution](../../.specify/memory/constitution.md):

- [x] **Reliability**: Feature identifies data write operations and ensures "250 OK" = fsync (SQLite + File), all writes are atomic
- [x] **Protocol Parity**: RFC compliance verified, matches mox's SPF/DKIM/DMARC implementation approach
- [x] **Mobile-First Architecture**: APIs support pagination, delta updates, compression; optimized for bandwidth/latency
- [x] **Dependency Minimalism**: Uses Go stdlib where possible, CGO-free (modernc.org/sqlite), dependencies justified
- [x] **Observability**: SMTP interaction logging planned (structured logs), Prometheus metrics identified
- [x] **Interface-Driven Design**: Storage uses repository pattern, business logic depends on interfaces not concrete implementations
- [x] **Extensibility**: SMTP pipeline uses middleware pattern for spam/antivirus injection without core changes
- [x] **Protocol Isolation**: Core logic (email storage) separated from access protocols (IMAP vs REST API)
- [x] **Testing Standards**: Test strategy includes unit, integration, durability (kill -9), interface mocks, middleware tests, protocol adapter tests

**Violations Requiring Justification**: None.

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
internal/
├── adapters/
│   ├── imap/           # IMAP Protocol Adapter
│   │   ├── commands.go # Command handlers
│   │   ├── server.go   # Connection management
│   │   ├── session.go  # Session state machine
│   │   ├── idle.go     # [NEW] IDLE implementation
│   │   └── uid.go      # [NEW] UID Mapping logic
│   └── http/           # HTTP Adapter
│       └── autodiscover.go # [NEW] Autoconfig/Autodiscover handlers
└── core/
    ├── domain/
    │   └── message.go  # Update for Flags/UID
    └── ports/
        └── repositories.go # Update EmailRepository interface
```

**Structure Decision**: Single Project (Backend Server). Standard Internal/Adapter layout.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |
