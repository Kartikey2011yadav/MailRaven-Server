# Research: Protocol Completion & Mobile Prep

| Attribute | Details |
| :--- | :--- |
| **Feature** | Protocol Completion (Quotas, ACLs) & Mobile Context |
| **Status** | Research Complete |
| **Date** | 2026-02-02 |
| **Correction** | **MailRaven Implementation**. The `mox` directory is a **reference** only. Do not edit files in `mox/`. All changes must occur in `internal/` and `client/`. |

## Unknowns & Clarifications

### 1. Quota Implementation (RFC 2087)

-   **Reference (`mox/imapserver`)**: `mox` implements `GETQUOTA` and `GETQUOTAROOT` but lacks dynamic `SETQUOTA`.
-   **MailRaven State**: `internal/adapters/imap/server.go` exists but needs Quota extensions.
-   **Decision**: 
    -   Implement `SETQUOTA`, `GETQUOTA`, `GETQUOTAROOT` in `internal/adapters/imap`.
    -   Update `internal/core/ports/user_repository.go` (or similar) to support per-account storage limits.
    -   Enforce limits in `append` / `copy` operations.

### 2. ACL Implementation (RFC 4314)

-   **Reference (`mox/imapserver`)**: `mox` documentation suggests limited or implicitly handled ACLs.
-   **MailRaven State**: Needs full implementation for folder sharing.
-   **Decision**:
    -   Extend `internal/core/domain` (or equivalent) to support ACLs on Mailboxes.
    -   Implement `SETACL`, `DELETEACL`, `GETACL`, `LISTRIGHTS`, `MYRIGHTS` in `internal/adapters/imap`.
    -   Schema update needed in `modernc.org/sqlite` database (managed in `internal/adapters/storage`).

### 3. Mobile API & Context

-   **Reference (`mox/webapi`)**: `mox` has a web API for admin/user actions.
-   **MailRaven State**: Uses `go-chi` in `internal/adapters/http`.
-   **Decision**:
    -   Create "Mobile Agent Context" file mapping `internal/adapters/http` routes for mobile devs.
    -   Ensure `internal/adapters/http` exposes necessary endpoints for mobile app (auth, fetch, send).
    -   Do **not** use `mox`'s `webapi` directly; reference its *logic* if needed but implement in `MailRaven`'s HTTP adapter.

## Technical Approach

### Architecture
-   **Storage**: 
    -   Update `internal/adapters/storage` (SQLite) to store `Quota` (on Account) and `ACL` (on Mailbox).
-   **Logic**:
    -   **Quota**: Check `CurrentStorage` vs `MaxStorage` in `internal/core/services/email_service.go` (or `imap` adapter) before writes.
    -   **ACL**: Verify `rights` in `internal/core/services` before allowing mailbox operations.
-   **Reference Usage**:
    -   Read `mox/imapserver/quota_test.go` for test cases.
    -   Read `mox/store` to see how tracking is modeled (but implement in SQL).

### Testing Strategy
-   **Unit**: Add tests in `internal/adapters/imap/quota_test.go` and `internal/adapters/imap/acl_test.go`.
-   **Integration**: functional tests verifying User A can grant User B access (ACLs) and User C gets rejected when over quota.

## Best Practices (Constitution)
-   **Reliability**: All quota/ACL changes must be transactional in SQLite.
-   **Protocol Parity**: Strict RFC compliance (referencing `mox` behavior where correct).
-   **Structure**: Keep core domain logic in `internal/core`, adapters in `internal/adapters`.
