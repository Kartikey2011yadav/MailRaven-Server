# Tasks: Archive and Spam Handling

**Feature Branch**: `014-archive-spam-handling`
**Spec**: [specs/014-archive-spam-handling/spec.md](spec.md)
**Plan**: [specs/014-archive-spam-handling/impl-plan.md](impl-plan.md)

## Phase 1: Setup & Data Model

- [x] T001 Create migration for `is_starred` column in `internal/adapters/storage/sqlite/migrations/011_add_starred_column.sql`
- [x] T002 Update `Message` struct in `internal/core/domain/message.go` to include `IsStarred`
- [x] T003 Create `MessageFilter` struct in `internal/core/domain/search.go` (or `filter.go`) for repository filtering

## Phase 2: Foundational (Repository)

- [x] T004 Update `EmailRepository` interface in `internal/core/ports/repositories.go` with filtering support
- [x] T005 [P] Implement `UpdateMailbox` method in `internal/adapters/storage/sqlite/email_repo.go`
- [x] T006 [P] Implement `UpdateStarred` method in `internal/adapters/storage/sqlite/email_repo.go`
- [x] T007 Implement dynamic query builder for `FindByUser` using `MessageFilter` in `internal/adapters/storage/sqlite/email_repo.go`

## Phase 3: API & Handlers (User Stories)

### User Story 1: Archive Message

- [x] T008 [US1] Update `PatchMessageRequest` DTO in `internal/adapters/http/dto/message.go` to support `mailbox` field
- [x] T009 [US1] Update `UpdateMessage` in `internal/adapters/http/handlers/messages.go` to handle mailbox changes
- [ ] T010 [US1] [P] Verify `Archive` action moves message logically (via API test or curl)

### User Story 2 & 3: Report Spam/Ham

- [x] T011 [US2] Add `ReportSpam` handler method in `internal/adapters/http/handlers/messages.go`
- [x] T012 [US2] Integrate `SpamFilter.TrainSpam` in `ReportSpam` handler
- [x] T013 [US3] Add `ReportHam` handler method in `internal/adapters/http/handlers/messages.go`
- [x] T014 [US3] Integrate `SpamFilter.TrainHam` in `ReportHam` handler
- [x] T015 [US2] Register POST routes `/api/v1/messages/{id}/spam` and `/api/v1/messages/{id}/ham` in `internal/adapters/http/server.go`

### User Story 4: Star Message

- [x] T016 [US4] Update `PatchMessageRequest` DTO in `internal/adapters/http/dto/message.go` to support `is_starred` field
- [x] T017 [US4] Update `UpdateMessage` in `internal/adapters/http/handlers/messages.go` to handle starred status changes
- [x] T018 [US4] Update `MessageSummary` DTO in `internal/adapters/http/dto/message.go` to include `is_starred` in responses

### User Story 5: Filtering (Read/Star/Date)

- [x] T019 [US5] Update `ListMessages` handler in `internal/adapters/http/handlers/messages.go` to parse query params (mailbox, is_read, is_starred, dates)
- [x] T020 [US5] Connect `ListMessages` query params to `EmailRepository.FindByUser` filter
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
