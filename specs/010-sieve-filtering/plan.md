# Implementation Plan: Sieve Filtering

**Branch**: `010-sieve-filtering` | **Date**: 2026-01-31 | **Spec**: [specs/010-sieve-filtering/spec.md](spec.md)
**Input**: Feature specification from `specs/010-sieve-filtering/spec.md`

## Summary

Implement server-side email filtering using the Sieve language (RFC 5228) to allow users to automatically sort emails into folders and set up vacation auto-replies. This includes a Sieve interpreter integration, database storage for scripts, a ManageSieve (RFC 5804) protocol listener for desktop clients, and REST APIs for web administration.

## Technical Context

**Language/Version**: Go 1.22+
**Primary Dependencies**: `github.com/emersion/go-sieve` (likely candidate for interpreter/protocol), `github.com/emersion/go-sasl` (for ManageSieve auth).
**Storage**: SQLite (scripts table, vacation tracking table).
**Testing**: Unit tests for interpreter hooks, Integration tests for SMTP+Sieve delivery, Protocol tests for ManageSieve.
**Target Platform**: Pure Go (Cross-platform).
**Project Type**: Server Backend.
**Performance Goals**: <50ms overhead per message.
**Constraints**: Must fail open (deliver to Inbox) on script errors.
**Scale/Scope**: Per-user scripts, typical script size <32KB.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Verify compliance with [MailRaven Constitution](../../.specify/memory/constitution.md):

- [x] **Reliability**: Scripts stored in SQLite (Durability). Execution failure defaults to Inbox (Safety).
- [x] **Protocol Parity**: Implements RFC 5228 and RFC 5804, matching Mox/Dovecot capabilities.
- [x] **Mobile-First Architecture**: N/A for Sieve engine, but Management API is REST-based for web/mobile admin.
- [x] **Dependency Minimalism**: Will evaluate `emersion/go-sieve` vs writing from scratch. Writing a Sieve interpreter from scratch is likely error-prone and non-maintenance friendly, so a robust library is justified if CGO-free.
- [x] **Observability**: Execution logs will include applied rules and filing actions.
- [x] **Interface-Driven Design**: `ScriptRepository` and `SieveEngine` will be interfaces.
- [x] **Extensibility**: Sieve will be implemented as a hook in the `LocalDelivery` phase, compatible with the middleware concept.
- [x] **Protocol Isolation**: Sieve logic resides in `core/services`, accessible by SMTP (execution) and ManageSieve/REST (management).
- [x] **Testing Standards**: Standard unit/integration tests planned.

## Project Structure

### Documentation (this feature)
- `specs/010-sieve-filtering/research.md`: Research findings (Library choice, DB Schema)
- `specs/010-sieve-filtering/data-model.md`: DB Schema for scripts/vacation
- `specs/010-sieve-filtering/contracts/`: ManageSieve logic does not need new contracts (Standard RFC), REST API spec.

### Modules / Packages
- `internal/core/ports/sieve.go`: Interfaces (`SieveExecutor`, `ScriptRepository`)
- `internal/core/domain/sieve.go`: Domain entities
- `internal/adapters/sieve/manager.go`: ManageSieve Protocol Server
- `internal/adapters/sieve/engine.go`: Interpreter integration
- `internal/adapters/storage/sqlite/sieve_repo.go`: Storage implementation

## Phase 0: Outline & Research

1.  **Unknowns & Research Tasks**:
    *   [ ] **Research Task**: Evaluate `github.com/emersion/go-sieve` for CGO dependencies and extensibility (Vacation/FileInto support).
    *   [ ] **Research Task**: Determine schema for `sieve_scripts` (Active vs Inactive) and `vacation_replies` (Auto-expiry).
    *   [ ] **Research Task**: Identify the exact hook point in `internal/core/services/email_service.go` or `LocalDelivery` flow.
    *   [ ] **Research Task**: Review `go-sasl` interaction with existing `AuthService` for ManageSieve.

2.  **Output**: `specs/010-sieve-filtering/research.md`

## Phase 1: Design & Contracts

1.  **Data Model (`data-model.md`)**:
    *   `sieve_scripts` table.
    *   `vacation_tracking` table.

2.  **Interfaces**:
    *   `ScriptRepository`
    *   `SieveExecutor`

3.  **Contracts**:
    *   REST API for Script Management (`modules/openapi.yaml` update or documented in `contracts/`).

## Phase 2: Implementation (Checklist)

### Step 1: Core & Storage
- [ ] Define Domain & Ports (`internal/core/{domain,ports}/sieve.go`).
- [ ] Implement `SqliteScriptRepository` & Migrations.

### Step 2: Interpreter Engine
- [ ] Integrate `go-sieve`.
- [ ] Implement `FileInto` and `Vacation` extensions.

### Step 3: SMTP Integration
- [ ] Hook Sieve into Delivery Pipeline.

### Step 4: Management Protocols
- [ ] Implement ManageSieve Server (TCP 4190).
- [ ] Implement REST API handlers.

### Step 5: Testing
- [ ] Unit tests for engine.
- [ ] Integration test (SMTP -> Sieve -> Mailbox).
