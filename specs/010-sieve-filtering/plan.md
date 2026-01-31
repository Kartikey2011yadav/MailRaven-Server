# Implementation Plan: Sieve Filtering

**Branch**: `010-sieve-filtering` | **Date**: 2026-01-31 | **Spec**: [specs/010-sieve-filtering/spec.md](spec.md)
**Input**: Feature specification from `specs/010-sieve-filtering/spec.md`

## Summary

Implement server-side email filtering using the Sieve language (RFC 5228) and ManageSieve protocol (RFC 5804). This allows users to enable robust server-side rules for folder sorting, rejection, and vacation auto-replies. The feature includes a Sieve interpreter hook in the delivery pipeline, a ManageSieve TCP server for client integration, and Web Admin APIs for management.

## Technical Context

**Language/Version**: Go 1.24.0
**Primary Dependencies**: `github.com/emersion/go-sieve` (Interpreter/Protocol), `github.com/emersion/go-sasl` (Authentication)
**Storage**: SQLite (Tables: `sieve_scripts`, `vacation_trackers`)
**Testing**: Unit tests for Sieve Engine, Integration tests for SMTP Delivery hook, Protocol tests for ManageSieve.
**Target Platform**: Pure Go Server (Cross-platform)
**Project Type**: Server Backend
**Performance Goals**: Script execution < 10ms per message.
**Constraints**: Must fail open (Implicit Keep) on runtime errors to prevent data loss.
**Scale/Scope**: Per-user scripts, script size hard limit (e.g., 32KB).

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Verify compliance with [MailRaven Constitution](../../.specify/memory/constitution.md):

- [x] **Reliability**: Scripts stored in SQLite (Durability). Execution defaults to Inbox on error (Safety).
- [x] **Protocol Parity**: Implements standard RFC 5228 and RFC 5804, matching mox capabilities.
- [x] **Mobile-First Architecture**: Script management available via REST API for mobile admin apps.
- [x] **Dependency Minimalism**: `emersion/go-sieve` is a high-quality, pure Go library, avoiding CGO.
- [x] **Observability**: Execution logs will track rule matches and actions taken.
- [x] **Interface-Driven Design**: `SieveExecutor` and `ScriptRepository` defined as interfaces.
- [x] **Extensibility**: Implemented as a post-data/pre-commit hook in the Local Delivery Agent.
- [x] **Protocol Isolation**: Logic resides in `core/services`, accessible by both SMTP and ManageSieve.
- [x] **Testing Standards**: Includes unit tests for rules and integration tests for full delivery flow.

**Violations Requiring Justification**: None.

## Project Structure

### Documentation (this feature)
- `specs/010-sieve-filtering/research.md`: Library selection and Schema design.
- `specs/010-sieve-filtering/data-model.md`: Database structure.
- `specs/010-sieve-filtering/contracts/`: REST API specifications.
- `specs/010-sieve-filtering/quickstart.md`: Testing guide.

### Modules / Packages
- `internal/core/ports/sieve.go`: Domain interfaces.
- `internal/core/domain/sieve/`: Domain entities.
- `internal/adapters/sieve/`: Engine implementation and ManageSieve server.
- `internal/adapters/storage/sqlite/`: Repository implementations.

## Phase 0: Outline & Research

1.  **Resolved Unknowns**:
    *   **Library**: Selected `emersion/go-sieve`.
    *   **Data Model**: Defined `sieve_scripts` (with active flag) and `vacation_trackers`.
    *   **Integration**: Hook into `LocalDeliveryAgent`.

2.  **Output**: [research.md](research.md) (Completed)

## Phase 1: Design & Contracts

1.  **Data Models**: Designed `sieve_scripts` and `vacation_trackers`.
2.  **Contracts**: Created `sieve_api.yaml` for Web Admin API.
3.  **Agent Context**: Updated to include Sieve capabilities.

## Phase 2: Implementation (Checklist)

See [checklists/implementation.md](checklists/implementation.md) and [tasks.md](tasks.md).

### Step 1: Core & Storage
- [ ] Define Entities and Interfaces.
- [ ] Implement SQLite Repositories.

### Step 2: Sieve Engine
- [ ] Implement Engine wrapper around `emersion/go-sieve`.
- [ ] Implement `fileinto` and `vacation` extensions.

### Step 3: Integration
- [ ] Inject Sieve hook into Local Delivery.

### Step 4: ManageSieve Server
- [ ] Implement TCP 4190 Listener.
- [ ] Implement ManageSieve commands.
- [ ] Wire up SASL.

### Step 5: Web API
- [ ] Implement REST endpoints.

### Step 6: Testing
- [ ] Verify message sorting.
- [ ] Verify auto-replies.

