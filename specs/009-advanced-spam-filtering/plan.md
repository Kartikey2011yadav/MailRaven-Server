# Implementation Plan: Advanced Spam Filtering

**Branch**: `009-advanced-spam-filtering` | **Date**: 2026-01-31 | **Spec**: [specs/009-advanced-spam-filtering/spec.md](../spec.md)
**Input**: Feature specification from `specs/009-advanced-spam-filtering/spec.md`

## Summary

Implement a native Advanced Spam Filtering system componsed of two main layers: a Greylisting middleware to block bounce-less bots, and a naive Bayesian classifier for content analysis. This includes database storage for greylist tuples and Bayesian tokens, SMTP pipeline middleware integration, and a feedback loop system for user training (Move-to-Junk).

## Technical Context

**Language/Version**: Go 1.23
**Primary Dependencies**: Go Standard Library (`strings`, `database/sql`). Potential need for a lightweight stemmer/tokenizer (e.g., `snowball`) but prefer simple strict splitting if sufficient.
**Storage**: SQLite / Postgres (via existing `gorm` or `sql` adapters).
**Testing**: Unit tests for classifier math, integration tests for DB storage, middleware chain tests.
**Target Platform**: Architecture agnostic (Pure Go).
**Project Type**: Server Backend
**Performance Goals**: <100ms classification latency per message.
**Constraints**: "Dependency Minimalism" (Constitution IV) - avoid heavy ML frameworks. "Reliability" (Constitution I) - DB writes for tokens need not be strictly `fsync` critical path per message (can be batched or async if justified, but default to safe).
**Scale/Scope**: ~1M tokens, single-tenant or SMB scale.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Verify compliance with [MailRaven Constitution](../../.specify/memory/constitution.md):

- [x] **Reliability**: Greylist state must be durable. Bayesian token updates are critical for accuracy but loss just means "dumber" filter, not data loss.
- [x] **Protocol Parity**: "Batteries included" spam filtering matches Mox's approach.
- [x] **Mobile-First Architecture**: N/A for core logic, but Training API (Move to Junk) supports mobile clients via IMAP.
- [x] **Dependency Minimalism**: Will implement Naive Bayes logic manually (simple math) to avoid dependencies.
- [x] **Observability**: `X-Spam-Status` headers added. Ops logs for "Greylisted" events.
- [x] **Interface-Driven Design**: `SpamFilter` and `GreylistStore` will be interfaces.
- [x] **Extensibility**: Uses `Middleware` pattern for SMTP injection.
- [x] **Protocol Isolation**: Core spam logic resides in `internal/core/services` or `internal/adapters/spam`, independent of SMTP/IMAP.
- [x] **Testing Standards**: Unit tests for math, mock tests for stores.

## Project Structure

### Documentation (this feature)
- `specs/009-advanced-spam-filtering/research.md`: Research findings (Algo, DB Schema)
- `specs/009-advanced-spam-filtering/data-model.md`: DB Schema for tokens/greylist
- `specs/009-advanced-spam-filtering/contracts/`: Not adding new external APIs, but internal interfaces.

### Modules / Packages
- `internal/core/ports/spam.go`: Interfaces (`Classifier`, `Greylister`)
- `internal/adapters/spam/bayesian/`: Naive Bayes Implementation
- `internal/adapters/spam/greylist/`: Greylisting Logic
- `internal/core/services/spam_service.go`: Orchestrator (Feedback loop)
- `internal/adapters/smtp/middleware_spam.go`: SMTP Middleware hooks

## Phase 0: Outline & Research

1.  **Unknowns & Research Tasks**:
    *   [ ] **Research Task**: Evaluate simple tokenization strategies for Go (whitespace vs regex vs stemmer) for best accuracy/perf balance without heavy deps.
    *   [ ] **Research Task**: Determine optimal storage schema for high-write Token counts (Bayes). Key-Value table vs specialized structure.
    *   [ ] **Research Task**: Pruning strategy for Greylist records (Expiration).
    *   [ ] **Integration Task**: Verify how to hook into the exact point of "MAIL FROM" vs "DATA" in the existing `middleware.go` chain (Greylisting needs to happen early, Bayes at DATA end).

2.  **Output**: `specs/009-advanced-spam-filtering/research.md`

## Phase 1: Design & Contracts

1.  **Data Model (`data-model.md`)**:
    *   `greylist_entries` table.
    *   `bayes_tokens` table (Word, SpamCount, HamCount).
    *   `global_stats` (TotalSpam, TotalHam).

2.  **Interfaces**:
    *   Define `Classifier` interface.
    *   Define `Trainer` interface (`LearnSpam`, `LearnHam`).

3.  **Agent Context**:
    *   Update `spam` related context instructions.

## Phase 2: Implementation (Checklist)

### Step 1: Core Domain & Ports
- [ ] Define `SpamClassifier` and `GreylistService` interfaces in `internal/core/ports`.
- [ ] Create domain entities (`Token`, `GreylistRecord`).

### Step 2: Storage Adapters
- [ ] Implement `GreylistRepository` (SQLite/PG).
- [ ] Implement `BayesRepository` (SQLite/PG) with atomic increment.

### Step 3: Logic Implementation
- [ ] Implement `internal/adapters/spam/bayesian` (Classify, Train).
- [ ] Implement `internal/adapters/spam/greylist` (Check, Allow).

### Step 4: Middleware Integration
- [ ] Create `GreylistMiddleware` (Early reject).
- [ ] Create `BayesMiddleware` (Header stamping).
- [ ] Register middlewares in `smtpserver`.

### Step 5: Feedback Loop
- [ ] Update `IMAP` "Move" or "Copy" handler to detect moves to/from Junk.
- [ ] Trigger `Train` async job.

### Step 6: Testing
- [ ] Unit tests for Bayes math.
- [ ] Integration test for Greylist retry behavior.
