# Implementation Plan: Protocol Completion

**Branch**: `013-protocol-completion` | **Date**: 2026-02-02 | **Spec**: [specs/013-protocol-completion/spec.md](specs/013-protocol-completion/spec.md)
**Input**: Feature specification from `/specs/013-protocol-completion/spec.md`

## Summary

Implement missing RFC 2087 (Quota) and RFC 4314 (ACL) features in the `mox` IMAP server. Additionally, formalize the API contract for the future Mobile App.

## Technical Context

**Language/Version**: Go 1.24
**Primary Dependencies**: `mox` codebase, `modernc.org/sqlite`
**Storage**: SQLite (via `mox/store` abstraction or direct)
**Testing**: `go test` (standard lib)
**Target Platform**: Linux/Windows Server
**Project Type**: Backend Server
**Performance Goals**: Minimal overhead on IMAP command processing.
**Constraints**: Must match existing `mox` patterns.
**Scale/Scope**: Support standard IMAP clients and future Mobile App.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Verify compliance with [MailRaven Constitution](../../.specify/memory/constitution.md):

- [x] **Reliability**: ACL/Quota writes will use `bstore` / SQLite transactions.
- [x] **Protocol Parity**: Strict adherence to RFC 2087 and RFC 4314.
- [x] **Mobile-First Architecture**: `MOBILE_AGENT_CONTEXT.md` serves this.
- [x] **Dependency Minimalism**: No new external dependencies required.
- [x] **Observability**: New commands will emit existing metrics (duration/count).
- [x] **Interface-Driven Design**: `store` updates will be method-based.
- [x] **Extensibility**: N/A (Core feature).
- [x] **Protocol Isolation**: Logic resides in `imapserver` and `store`.
- [x] **Testing Standards**: New `acl_test.go` and updated `quota_test.go`.

**Violations Requiring Justification**: None

## Project Structure

### Documentation (this feature)

```text
specs/013-protocol-completion/
├── plan.md              # This file
├── research.md          # Research findings
├── data-model.md        # DB Schema changes
├── quickstart.md        # How to run/test
├── contracts/           # API definitions
└── tasks.md             # Implementation tasks
```

### Source Code

```text
mox/
├── imapserver/
│   ├── quota_test.go    # Updated
│   ├── acl.go           # NEW: ACL command implementation
│   ├── acl_test.go      # NEW: ACL tests
├── store/
│   ├── account.go       # Updated: Add Quota Limit
│   ├── mailbox.go       # Updated: Add ACL
docs/
└── development/
    └── MOBILE_AGENT_CONTEXT.md # NEW
```
