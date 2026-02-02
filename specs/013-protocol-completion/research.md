# Research: Protocol Completion & Mobile Prep

| Attribute | Details |
| :--- | :--- |
| **Feature** | Protocol Completion (Quotas, ACLs) & Mobile Context |
| **Status** | Research Complete |
| **Date** | 2026-02-02 |

## Unknowns & Clarifications

### 1. Current Quota Code (RFC 2087)

-   **Finding**: `mox/imapserver/quota_test.go` confirms `GETQUOTA` and `GETQUOTAROOT` are implemented but `SETQUOTA` returns `BAD` (Not Implemented).
-   **Storage**: Usage is tracked. Limits are currently configuration-based or test-injected, not dynamically settable via IMAP.
-   **Decision**: Implement `SETQUOTA` command in `imapserver/server.go`. Persist quota limits in `store/account.go` (or similar).

### 2. Current ACL Code (RFC 4314)

-   **Finding**: No evidence of `SETACL`, `GETACL`, etc. in `imapserver`. `server.go` implementation notes do not mention RFC 4314.
-   **Decision**: Full implementation required.
    -   Add `ACL` field to `store.Mailbox` struct.
    -   Implement `SETACL`, `DELETEACL`, `GETACL`, `LISTRIGHTS`, `MYRIGHTS` in `imapserver`.
    -   Enforce permissions in `select`, `append`, `copy`, `expunge`, etc.

### 3. Mobile API & Context

-   **Finding**: `mox` exposes `webapi`. `webauth` likely handles login.
-   **Requirement**: The "Mobile Agent Context" file needs to map these existing endpoints for the mobile developer agent.
-   **Decision**: Document `webapi` endpoints, JWT flow, and "Headless" usage (using `mox` as a backend only).

## Technical Approach

### Architecture
-   **Storage**: Extend `mox/store` schemas (using `bstore` or existing DB mechanism).
    -   `Account`: Add `StorageLimit` (int64).
    -   `Mailbox`: Add `ACL` (map/list of identifiers and rights).
-   **Logic**:
    -   `SETQUOTA` updates `Account` record.
    -   `SETACL` updates `Mailbox` record.
    -   Middleware/Check logic in `imapserver` to verify ACLs before actions.

### Testing Strategy
-   **Unit**: Extend `quota_test.go`. Create `acl_test.go` in `imapserver`.
-   **Integration**: Functional tests in `spec.md` (e.g., User A shares to User B).

## Best Practices (Constitution)
-   **Reliability**: ACL/Quota changes must be atomic DB updates.
-   **Parity**: Strict RFC 2087/4314 compliance.
-   **Mobile**: The Context file is the key enabler here.
