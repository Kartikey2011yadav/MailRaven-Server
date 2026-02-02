# Data Model: Protocol Completion

## Entities

### Account (Extension)

Existing entity `Account` will be extended.

| Field | Type | Description |
| :--- | :--- | :--- |
| `storage_quota` | `int64` | Maximum storage in bytes. `0` means unlimited (or default). |
| `storage_used` | `int64` | Current storage usage in bytes. Updated on atomic writes. |

### Mailbox (Extension)

Existing entity `Mailbox` will be extended to support ACLs (RFC 4314).

| Field | Type | Description |
| :--- | :--- | :--- |
| `acl` | `JSON` | Map of `identifier` (user) -> `rights` (string of chars). |

**ACL Rights Format**:
Standard RFC 4314 characters:
- `l`: lookup key
- `r`: read
- `s`: keep seen/unseen
- `w`: write flags
- `i`: insert (append/copy)
- `p`: post (send) - *not strictly used for mailbox storage but part of RFC*
- `k`: create sub-mailboxes
- `x`: delete mailbox
- `t`: delete messages
- `e`: expunge
- `a`: admin (manage ACLs)

## Database Schema (SQLite)

```sql
ALTER TABLE accounts ADD COLUMN storage_quota INTEGER DEFAULT 0;
ALTER TABLE accounts ADD COLUMN storage_used INTEGER DEFAULT 0;

ALTER TABLE mailboxes ADD COLUMN acl TEXT DEFAULT '{}'; -- JSON serialized map
```

## Validation Rules

1.  **Quota**: `storage_used + new_message_size <= storage_quota` (unless `storage_quota == 0` check global default).
2.  **ACL**:
    -   Owner always has full rights usually, or implicit `a`.
    -   `anyone` identifier maps to anonymous/public access.
    -   Circular dependencies in ACL groups (if groups implemented) must be avoided (Scope restriction: Users only for now).
