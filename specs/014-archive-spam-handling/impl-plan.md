# Implementation Plan: Archive and Spam Mechanism

**Branch**: `014-archive-spam-handling` | **Date**: 2026-02-10 | **Spec**: [specs/014-archive-spam-handling/spec.md](../specs/014-archive-spam-handling/spec.md)
**Input**: Feature specification from `/specs/014-archive-spam-handling/spec.md`

## Summary

This feature implements core email organization workflows: Archiving (moving to Archive folder), Spam Reporting (moving to Junk + Bayesian Training), and Spam Reversal (moving to Inbox + Ham Training), along with "Starred" functionality and advanced message filtering (Read/Unread, Starred, Date Range).

## Technical Context

**Language/Version**: Go 1.25+
**Primary Dependencies**:
- `internal/core/ports`: For repository interfaces.
- `internal/adapters/storage/sqlite`: For DB implementation.
- `internal/adapters/http`: For API handlers.
- `github.com/Kartikey2011yadav/mailraven-server/internal/core/domain`: For domain entities.
**Storage**: SQLite (persisting `Mailbox` and `IsStarred` state).
**Testing**: Go standard testing package + Integration tests for repository interactions.
**Target Platform**: Linux/Windows server, API consumed by Mobile/Web clients.
**Performance Goals**: API response < 500ms for actions.

**Dependencies**:
- **Spam Filter**: Needs interface connection to `SpamFilter.TrainSpam` and `SpamFilter.TrainHam`.
- **Database Schema**: Needs migration to add `is_starred` column (or flags) and support `mailbox` updates efficiently.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Verify compliance with [MailRaven Constitution](../../.specify/memory/constitution.md):

- [x] **Reliability**: Message moves are atomic metadata updates (SQLite transactions). No payload data is moved, avoiding data loss risk.
- [x] **Protocol Parity**: "Starred" maps to IMAP `\Flagged`, "Junk" maps to `\Junk`, "Trash" maps to `\Trash` or `\Deleted`.
- [x] **Mobile-First Architecture**: APIs support delta sync (`since`) and filtering params (read/starred/date) to minimize bandwidth.
- [x] **Dependency Minimalism**: Uses internal repository interfaces, no new external libs.
- [x] **Observability**: Metrics for "Mark as Spam" actions vs "Mark as Ham" (false positive tracking).
- [x] **Interface-Driven Design**: `EmailRepository` extended with specific methods (`SetMailbox`, `SetFlag`) rather than raw SQL in handlers.
- [x] **Extensibility**: Spam training hooks into existing `SpamFilter` interface.
- [x] **Protocol Isolation**: Logic resides in `core/services` or `adapters/http` handlers, agnostic of IMAP/SMTP details.
- [x] **Testing Standards**: Integration tests required for repository filtering and movement logic.

**Violations Requiring Justification**: None.

## Project Structure

### Documentation (this feature)

```text
specs/014-archive-spam-handling/
├── impl-plan.md         # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output (API Usage Guide)
├── contracts/           # Phase 1 output (OpenAPI partials)
└── tasks.md             # Phase 2 output
```

### Source Code

```text
internal/
├── core/
│   ├── ports/
│   │   └── repositories.go          # Update EmailRepository (Move, Star, Filter)
│   └── domain/
│       └── message.go               # Add IsStarred field
├── adapters/
│   ├── http/
│   │   ├── handlers/
│   │   │   └── messages.go          # Update ListMessages (filtering) + Add Move/Star actions
│   │   └── dto/
│   │       └── message.go           # Update DTOs
│   └── storage/
│       └── sqlite/
│           ├── email_repo.go        # Implement filtering & updates
│           └── migrations/          # Add is_starred column
└── ...
```
