# Tasks: Archive and Spam Handling

**Feature Branch**: `014-archive-spam-handling`
**Spec**: [specs/014-archive-spam-handling/spec.md](spec.md)
**Plan**: [specs/014-archive-spam-handling/impl-plan.md](impl-plan.md)

## Phase 1: Setup & Data Model

- [ ] T001 Create migration for `is_starred` column in `internal/adapters/storage/sqlite/migrations/011_add_starred_column.sql`
- [ ] T002 Update `Message` struct in `internal/core/domain/message.go` to include `IsStarred`
- [ ] T003 Create `MessageFilter` struct in `internal/core/domain/search.go` (or `filter.go`) for repository filtering

## Phase 2: Foundational (Repository)

- [ ] T004 Update `EmailRepository` interface in `internal/core/ports/repositories.go` with filtering support
- [ ] T005 [P] Implement `UpdateMailbox` method in `internal/adapters/storage/sqlite/email_repo.go`
- [ ] T006 [P] Implement `UpdateStarred` method in `internal/adapters/storage/sqlite/email_repo.go`
- [ ] T007 Implement dynamic query builder for `FindByUser` using `MessageFilter` in `internal/adapters/storage/sqlite/email_repo.go`

## Phase 3: API & Handlers (User Stories)

### User Story 1: Archive Message

- [ ] T008 [US1] Update `PatchMessageRequest` DTO in `internal/adapters/http/dto/message.go` to support `mailbox` field
- [ ] T009 [US1] Update `UpdateMessage` in `internal/adapters/http/handlers/messages.go` to handle mailbox changes
- [ ] T010 [US1] [P] Verify `Archive` action moves message logically (via API test or curl)

### User Story 2 & 3: Report Spam/Ham

- [ ] T011 [US2] Add `ReportSpam` handler method in `internal/adapters/http/handlers/messages.go`
- [ ] T012 [US2] Integrate `SpamFilter.TrainSpam` in `ReportSpam` handler
- [ ] T013 [US3] Add `ReportHam` handler method in `internal/adapters/http/handlers/messages.go`
- [ ] T014 [US3] Integrate `SpamFilter.TrainHam` in `ReportHam` handler
- [ ] T015 [US2] Register POST routes `/api/v1/messages/{id}/spam` and `/api/v1/messages/{id}/ham` in `internal/adapters/http/server.go`

### User Story 4: Star Message

- [ ] T016 [US4] Update `PatchMessageRequest` DTO in `internal/adapters/http/dto/message.go` to support `is_starred` field
- [ ] T017 [US4] Update `UpdateMessage` in `internal/adapters/http/handlers/messages.go` to handle starred status changes
- [ ] T018 [US4] Update `MessageSummary` DTO in `internal/adapters/http/dto/message.go` to include `is_starred` in responses

### User Story 5: Filtering (Read/Star/Date)

- [ ] T019 [US5] Update `ListMessages` handler in `internal/adapters/http/handlers/messages.go` to parse query params (mailbox, is_read, is_starred, dates)
- [ ] T020 [US5] Connect `ListMessages` query params to `EmailRepository.FindByUser` filter
- [ ] T021 [US5] [P] Verify filtering via API calls

## Final Phase: Polish

- [ ] T022 Update `contracts/openapi-partial.yaml` if any implementation details diverged
- [ ] T023 Run full integration tests to ensure no regression in existing message listing

## Dependencies

- **US2/US3 (Spam)** depends on `SpamFilter` interface (already exists).
- **US5 (Filter)** depends on **Phase 2** repository updates.

## Parallel Execution Opportunities

- T005, T006, T007 (Repository methods) can be implemented in parallel.
- T011 (Spam) and T013 (Ham) are independent handlers.
