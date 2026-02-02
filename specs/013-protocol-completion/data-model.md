# Data Model: Protocol Completion

## Storage Schema Changes

### Account (mox/store/account.go)

| Field | Type | Description |
| :--- | :--- | :--- |
| `QuotaStorageLimit` | `int64` | Max storage in bytes. 0 or -1 for unlimited (logic dependent). |

### Mailbox (mox/store/mailbox.go)

| Field | Type | Description |
| :--- | :--- | :--- |
| `ACL` | `map[string]string` | Map of Identifier -> Rights string. <br>Key: user/group ID (e.g. "mjl@mox.example" or "anyone"). <br>Value: Rights (e.g., "lrswipcda"). |

## Entities

### QuotaResource

-   **Name**: "STORAGE"
-   **Usage**: Calculated from `Account.DiskUsage` (existing).
-   **Limit**: From `QuotaStorageLimit`.

### AccessControlList

-   **Identifier**: "anyone", "authuser", "admin", or email address.
-   **Rights**:
    -   `l`: lookup (mailbox visible)
    -   `r`: read (select, examine)
    -   `s`: keep seen (seen flag)
    -   `w`: write (flags, keywords)
    -   `i`: insert (append, copy)
    -   `p`: post (not used usually, but RFC 4314)
    -   `c`: create (sub-mailboxes)
    -   `d`: delete (mailbox)
    -   `a`: admin (setacl)
