# Tasks: Protocol Completion & Mobile Prep

**Feature**: Protocol Completion (Quotas, ACLs) & Mobile Context
**Status**: Pending
**Branch**: `013-protocol-completion`

## Dependencies

1.  **Phase 1 (Setup)**: Must run first.
2.  **Phase 2 (Foundational)**: Blocks all User Stories.
3.  **Phase 3 (US1 - Quotas)**: Independent.
4.  **Phase 4 (US3 - Mobile Context)**: Independent, can run parallel with US1/US2.
5.  **Phase 5 (US2 - ACLs)**: Independent, but technically easier after Quota foundation.
6.  **Phase 6 (Polish)**: Final cleanup.

## Implementation Strategy

We will adopt a "backend-first" strategy. We will first extend the database and domain models to support Quotas and ACLs. Then we will implement User Story 1 (Quotas) as it affects the fundamental ability to accept mail (SMTP/IMAP Append). Next, we will document the system for the Mobile Agent (US3). Finally, we will implement the complex ACL logic (US2).

## Phase 1: Setup

*Goal: Initialize documentation and prepare database migration scripts.*

- [ ] T001 Create `docs/development/MOBILE_AGENT_CONTEXT.md` generic structure [docs/development/MOBILE_AGENT_CONTEXT.md](docs/development/MOBILE_AGENT_CONTEXT.md)
- [ ] T002 [P] Create migration file to add `storage_quota`, `storage_used` to `accounts` table in `internal/adapters/storage/sqlite/migrations/00X_add_quotas.sql` [internal/adapters/storage/sqlite/migrations/00X_add_quotas.sql](internal/adapters/storage/sqlite/migrations/00X_add_quotas.sql)
- [ ] T003 [P] Create migration file to add `acl` column to `mailboxes` table in `internal/adapters/storage/sqlite/migrations/00Y_add_acls.sql` [internal/adapters/storage/sqlite/migrations/00Y_add_acls.sql](internal/adapters/storage/sqlite/migrations/00Y_add_acls.sql)

## Phase 2: Foundational

*Goal: Update Core Domain and Storage Adapters to support new fields. Blocking for all stories.*

- [ ] T004 Update `Account` struct with `StorageQuota` and `StorageUsed` fields in `internal/core/domain/account.go` [internal/core/domain/account.go](internal/core/domain/account.go)
- [ ] T005 Update `Mailbox` struct with `ACL` field (map/json) in `internal/core/domain/mailbox.go` [internal/core/domain/mailbox.go](internal/core/domain/mailbox.go)
- [ ] T006 Update `UserRepository` interface to include methods for Quota management in `internal/core/ports/user_repository.go` [internal/core/ports/user_repository.go](internal/core/ports/user_repository.go)
- [ ] T007 Update `MailboxRepository` interface to include methods for ACL management in `internal/core/ports/email_repository.go` (or mailbox_repository.go) [internal/core/ports/email_repository.go](internal/core/ports/email_repository.go)
- [ ] T008 [P] Implement SQLite storage logic for Quota fields in `internal/adapters/storage/sqlite/user_repo.go` [internal/adapters/storage/sqlite/user_repo.go](internal/adapters/storage/sqlite/user_repo.go)
- [ ] T009 [P] Implement SQLite storage logic for ACL JSON serialization/deserialization in `internal/adapters/storage/sqlite/mailbox_repo.go` [internal/adapters/storage/sqlite/mailbox_repo.go](internal/adapters/storage/sqlite/mailbox_repo.go)

## Phase 3: User Story 1 - Admin Control via Quotas (P1)

*Goal: Enforce storage limits to prevent abuse.*
*Independent Test: `mailraven users set-quota` CLI works, and SMTP rejects mail when over quota.*

- [ ] T010 [US1] Implement `UpdateQuota` service method in `internal/core/services/user_service.go` [internal/core/services/user_service.go](internal/core/services/user_service.go)
- [ ] T011 [US1] Create unit tests for Quota logic (checking limits) in `internal/core/services/user_service_test.go` [internal/core/services/user_service_test.go](internal/core/services/user_service_test.go)
- [ ] T012 [P] [US1] Implement `GETQUOTA` and `GETQUOTAROOT` IMAP handlers in `internal/adapters/imap/quota.go` [internal/adapters/imap/quota.go](internal/adapters/imap/quota.go)
- [ ] T013 [P] [US1] Implement `SETQUOTA` IMAP handler (admin only) in `internal/adapters/imap/quota.go` [internal/adapters/imap/quota.go](internal/adapters/imap/quota.go)
- [ ] T014 [US1] Register Quota handlers in `internal/adapters/imap/server.go` [internal/adapters/imap/server.go](internal/adapters/imap/server.go)
- [ ] T015 [US1] Enforce Quota checks in IMAP `APPEND` and `COPY` handlers in `internal/adapters/imap/session.go` [internal/adapters/imap/session.go](internal/adapters/imap/session.go)
- [ ] T016 [US1] Enforce Quota checks in SMTP `DATA` handler in `internal/adapters/smtp/session.go` [internal/adapters/smtp/session.go](internal/adapters/smtp/session.go)
- [ ] T017 [US1] Implement HTTP Admin Endpoint `PUT /users/{username}/quota` in `internal/adapters/http/user_handler.go` [internal/adapters/http/user_handler.go](internal/adapters/http/user_handler.go)

## Phase 4: User Story 3 - Mobile Dev Agent Context (P1)

*Goal: Enable the Mobile Agent to build the client independently.*
*Independent Test: `curl` commands from the context file work against a running local server.*

- [ ] T018 [US3] Document Authentication flows (JWT) and endpoints in `docs/development/MOBILE_AGENT_CONTEXT.md` [docs/development/MOBILE_AGENT_CONTEXT.md](docs/development/MOBILE_AGENT_CONTEXT.md)
- [ ] T019 [US3] Document Mailbox Sync endpoints (REST/IMAP) and params in `docs/development/MOBILE_AGENT_CONTEXT.md` [docs/development/MOBILE_AGENT_CONTEXT.md](docs/development/MOBILE_AGENT_CONTEXT.md)
- [ ] T020 [US3] Document "Send Email" endpoints and payload formats in `docs/development/MOBILE_AGENT_CONTEXT.md` [docs/development/MOBILE_AGENT_CONTEXT.md](docs/development/MOBILE_AGENT_CONTEXT.md)
- [ ] T021 [US3] Verify functionality of documented endpoints and mark as "Verified" in `docs/development/MOBILE_AGENT_CONTEXT.md` [docs/development/MOBILE_AGENT_CONTEXT.md](docs/development/MOBILE_AGENT_CONTEXT.md)

## Phase 5: User Story 2 - Mailbox Sharing via ACLs (P2)

*Goal: Allow users to share folders with specific permissions.*
*Independent Test: User B cannot access User A's folder until `SETACL` is called.*

- [ ] T022 [US2] Implement `UpdateACL` service method logic in `internal/core/services/email_service.go` [internal/core/services/email_service.go](internal/core/services/email_service.go)
- [ ] T023 [US2] Add ACL enforcement checks (Lookup, Read, Write) to `internal/core/services/email_service.go` [internal/core/services/email_service.go](internal/core/services/email_service.go)
- [ ] T024 [P] [US2] Create unit tests for ACL enforcement logic in `internal/core/services/email_service_test.go` [internal/core/services/email_service_test.go](internal/core/services/email_service_test.go)
- [ ] T025 [P] [US2] Implement `SETACL`, `DELETEACL`, `GETACL` handlers in `internal/adapters/imap/acl.go` [internal/adapters/imap/acl.go](internal/adapters/imap/acl.go)
- [ ] T026 [P] [US2] Implement `LISTRIGHTS`, `MYRIGHTS` handlers in `internal/adapters/imap/acl.go` [internal/adapters/imap/acl.go](internal/adapters/imap/acl.go)
- [ ] T027 [US2] Register ACL handlers in `internal/adapters/imap/server.go` [internal/adapters/imap/server.go](internal/adapters/imap/server.go)
- [ ] T028 [US2] Integrate ACL checks into IMAP `SELECT`, `LIST`, `FETCH` handlers in `internal/adapters/imap/session.go` [internal/adapters/imap/session.go](internal/adapters/imap/session.go)
- [ ] T029 [US2] Implement HTTP Admin Endpoint `PUT /mailboxes/{id}/acl` in `internal/adapters/http/mailbox_handler.go` [internal/adapters/http/mailbox_handler.go](internal/adapters/http/mailbox_handler.go)

## Phase 6: Polish

*Goal: Final verification and cleanup.*

- [ ] T030 Ensure all new IMAP extensions are advertised in `CAPABILITY` response in `internal/adapters/imap/server.go` [internal/adapters/imap/server.go](internal/adapters/imap/server.go)
- [ ] T031 Run full integration test suite ensuring no regression in existing features in `tests/integration/protocol_test.go` [tests/integration/protocol_test.go](tests/integration/protocol_test.go)
