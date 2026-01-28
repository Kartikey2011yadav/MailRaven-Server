# Implementation Plan: Production Readiness & Polish

**Branch**: `005-production-readiness` | **Date**: 2026-01-28 | **Spec**: [specs/005-production-readiness/spec.md](specs/005-production-readiness/spec.md)
**Input**: Feature specification from `/specs/005-production-readiness/spec.md`

## Summary

This feature focuses on hardening the MailRaven server for production use. It includes enhancing backend logging for better observability, creating comprehensive documentation for deployment (including a gap analysis vs. Mox), and providing cross-platform setup scripts. Additionally, it introduces flexibility in database selection (PostgreSQL vs SQLite) and container orchestration.

## Technical Context

**Language/Version**: Go 1.21+ (Backend), Typescript/React (Frontend), Shell/PowerShell (Scripts)
**Primary Dependencies**: 
- `log/slog` (Standard Go Logger) or `zap` (for structured logging) - NEEDS CLARIFICATION on current usage.
- `modernc.org/sqlite` (Current DB)
- `lib/pq` or `pgx` (PostgreSQL Driver) - NEEDS CLARIFICATION on preference.
- Docker & Docker Compose
**Storage**: SQLite (default), PostgreSQL (new option)
**Testing**: Go `testing` package, Playwright (E2E)
**Target Platform**: Linux (primary production), Windows (dev), Mac (supported via scripts), Docker
**Project Type**: Full-stack Mail Server (Web + Backend)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Verify compliance with [MailRaven Constitution](../../.specify/memory/constitution.md):

- [ ] **Reliability**: Postgres selection must ensure ACID properties similar to the existing SQLite implementation.
- [ ] **Protocol Parity**: Logs must capture SMTP protocol details clearly.
- [ ] **Mobile-First Architecture**: N/A for this infrastructure feature.
- [ ] **Dependency Minimalism**: Evaluate if adding a Postgres driver violates minimalism (usually justified for production scale).
- [ ] **Observability**: This feature *directly implements* the Observability principle (structured logs).
- [ ] **Interface-Driven Design**: The Storage layer must abstract both SQLite and Postgres behind the existing repository interfaces.
- [ ] **Extensibility**: Setup scripts should be modular.
- [ ] **Protocol Isolation**: N/A.
- [ ] **Testing Standards**: New scripts must be tested (manually or via CI).

**Violations Requiring Justification**: None. Adding Postgres increases dependencies but is a standard requirement for "production readiness" at scale.

## Project Structure

### Documentation (this feature)

```text
specs/005-production-readiness/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output (DB schema for Postgres)
├── quickstart.md        # Phase 1 output (Production guide draft)
└── contracts/           # Phase 1 output (N/A)
```

### Source Code (repository root)

```text
f:/MailRaven-Server/
├── internal/
│   ├── adapters/
│   │   ├── storage/
│   │   │   ├── postgres/        # NEW: Postgres implementation
│   │   │   └── sqlite/          # Existing
│   │   └── logger/              # NEW/UPDATED: Structured logging adapter
│   └── config/                  # Updated for DB choice
├── scripts/
│   ├── setup.sh                 # NEW: Linux/Mac setup
│   ├── setup.ps1                # NEW: Windows setup
│   ├── check.sh                 # NEW: Environment check
│   └── check.ps1                # NEW: Windows check
├── docker/                      # NEW: Organized docker files
│   ├── Dockerfile.backend
│   └── Dockerfile.frontend
├── docker-compose.yml           # Updated
└── docs/
    ├── PRODUCTION.md            # NEW
    ├── GAP_ANALYSIS.md          # NEW
    └── BACKUP_ANALYSIS_PROMPT.md# NEW
```

## Phases

### Phase 0: Outline & Research

1. **Extract unknowns**:
   - Current logging implementation (std `log` vs `slog`).
   - Repository interface compatibility with Postgres.
   - Mox's specific script capabilities.
   - `DB-BACKUP` project structure (for prompt generation).

2. **Research Tasks**:
   - "Analyze `internal/observability` (if any) or logging usage to plan color/structure upgrade."
   - "Analyze `internal/core/ports/repository.go` to ensure Postgres can satisfy interfaces."
   - "Fetch Mox repo file list (if possible) or infer script types from query."
   - "Research best practices for Go cross-platform setup scripts."

### Phase 1: Design & Contracts

1. **Entities**: Define the Postgres schema (should mirror SQLite but with Postgres types).
2. **Contracts**: None (Internal feature).
3. **Agent Context**: Update context to include Postgres and Docker details.

### Phase 2: Implementation Breakdown

(To be generated in `tasks.md`)
