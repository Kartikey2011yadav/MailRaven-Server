# Tasks: Sieve Filtering

**Reviewer**: Copilot | **Feature Phase**: Implementation | **Total Tasks**: 26

**Feature**: Sieve Filtering (Feature 010)
**Branch**: `010-sieve-filtering`

## Phase 1: Setup
**Goal**: Initialize dependencies and project structure.

- [ ] T001 Add `github.com/emersion/go-sieve` dependency to `go.mod`
- [ ] T002 Add `github.com/emersion/go-sasl` dependency to `go.mod`
- [ ] T003 Create directory structure `internal/core/sieve` and `internal/adapters/sieve`

## Phase 2: Foundational (Blocking)
**Goal**: Core data models and database storage.
**Interactive Test**: Manual verify tables created in SQLite.

- [ ] T004 Define `SieveScript` and `VacationTracker` entities in `internal/core/domain/sieve/entities.go`
- [ ] T005 Define `ScriptRepository` and `VacationRepository` interfaces in `internal/core/ports/sieve.go`
- [ ] T006 Create SQL migration for `sieve_scripts` and `vacation_trackers` in `internal/adapters/storage/sqlite/migrations/010_sieve.sql`
- [ ] T007 [P] Implement `SqliteScriptRepository` in `internal/adapters/storage/sqlite/sieve_repo.go`
- [ ] T008 [P] Implement `SqliteVacationRepository` in `internal/adapters/storage/sqlite/vacation_repo.go`

## Phase 3: Rules Engine & FileInto (User Story 1 - P1)
**Goal**: Core Sieve engine integration delivering emails to folders.
**Independent Test**: Inject email with "Subject: Bills", verify it appears in "Finance" folder.

- [ ] T009 [US1] Create `SieveEngine` struct and `Run` method in `internal/adapters/sieve/engine.go`
- [ ] T010 [US1] Implement `fileinto` extension hook in `engine.go` connected to `MailboxRepository`
- [ ] T011 [US1] Implement `Implicit Keep` and Error handling behavior in `engine.go`
- [ ] T012 [US1] Inject `SieveExecutor` into `LocalDeliveryAgent` in `internal/core/smtp/delivery.go`
- [ ] T013 [US1] Unit Test: Verify `Run` executes simple script correctly in `internal/adapters/sieve/engine_test.go`
- [ ] T014 [US1] Integration: Update `LocalDelivery` to fetch active script and execute before storage in `internal/core/smtp/delivery.go`

## Phase 4: Vacation Auto-Reply (User Story 2 - P2)
**Goal**: Automatic out-of-office replies with rate limiting.
**Independent Test**: Send email to user with vacation active, receive reply. Second email gets no reply.

- [ ] T015 [US2] Implement `vacation` extension hook in `internal/adapters/sieve/vacation.go`
- [ ] T016 [US2] Implement rate limiting check using `VacationRepository` in `vacation.go`
- [ ] T017 [US2] Implement email generation and injection into Outgoing Queue in `vacation.go`
- [ ] T018 [US2] Unit Test: Verify vacation logic respects `:days` and prevents loops in `internal/adapters/sieve/vacation_test.go`

## Phase 5: Script Management API (User Story 3 - P3)
**Goal**: REST API for Web Admin.
**Independent Test**: `curl POST /api/v1/sieve/scripts` updates the DB.

- [ ] T019 [P] [US3] Implement HTTP handlers for Scripts (List, Create, Get, Delete) in `internal/adapters/http/sieve_handlers.go`
- [ ] T020 [P] [US3] Implement HTTP handler for Activate Script in `internal/adapters/http/sieve_handlers.go`
- [ ] T021 [US3] Register Sieve routes in `internal/adapters/http/router.go`

## Phase 6: ManageSieve Protocol (User Story 4 - P3)
**Goal**: Support desktop clients (Thunderbird).
**Independent Test**: Connect via `openssl s_client -connect localhost:4190`.

- [ ] T022 [US4] Create `ManageSieveServer` struct in `internal/adapters/managesieve/server.go`
- [ ] T023 [US4] Implement TCP Listener and ConnectionLoop in `internal/adapters/managesieve/listener.go`
- [ ] T024 [US4] Implement SASL Authentication integration in `internal/adapters/managesieve/auth.go`
- [ ] T025 [US4] Implement Commands (PUTSCRIPT, SETACTIVE, etc.) using `emersion/go-manage-sieve` (or custom parser) in `internal/adapters/managesieve/commands.go`

## Phase 7: Polish & Cross-Cutting
**Goal**: Logging, Metrics, and Config.

- [ ] T026 Add configuration for Sieve (MaxScriptSize, VacationMinDays) in `internal/config/config.go`

## Dependencies
- Phase 1 & 2 blocks ALL
- Phase 3 blocks Phase 4
- Phase 3 blocks Phase 6 (Need engine to validate scripts)
- Phase 5 and Phase 6 are parallelizable

## Implementation Strategy
Start with the Engine (US1) as it provides immediate value (server-side filtering). The API (US3) is needed to upload scripts to *test* the engine easily, although manual DB insertion can work for dev. Vacation (US2) is complex due to loop prevention. ManageSieve (US4) is a separate protocol server.
