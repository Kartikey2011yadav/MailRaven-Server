# Data Model Changes

## Entities

### Message (Updated)

| Field | Type | Description |
|-------|------|-------------|
| `UID` | `INTEGER` | **[NEW]** IMAP UID (Unique, Monotonic per Mailbox). Indexed. |
| `Mailbox` | `TEXT` | **[NEW]** Mailbox name (default "INBOX"). Indexed. |
| `Flags` | `TEXT` | **[NEW]** Space-separated list of flags (e.g., "\Seen \Flagged"). |
| `ModSeq` | `INTEGER` | **[NEW]** Modification Sequence (for CONDSTORE extension, optional for MVP but good for sync). |

### Mailbox (New)

Represents a folder state, primarily to track UID validity and next UID.

| Field | Type | Description |
|-------|------|-------------|
| `Name` | `TEXT` | Primary Key (Composite with UserID). e.g., "INBOX". |
| `UserID` | `TEXT` | Owner. |
| `UIDValidity` | `INTEGER` | Random non-zero integer. Changes if UIDs are reset. |
| `UIDNext` | `INTEGER` | Next UID to assign. Starts at 1. |
| `MessageCount`| `INTEGER` | Cached count. |

## Storage Interface Updates

### EmailRepository

```go
type EmailRepository interface {
    // ... existing ...

    // IMAP Support
    GetMailbox(ctx context.Context, userID, name string) (*domain.Mailbox, error)
    CreateMailbox(ctx context.Context, userID, name string) error
    
    // Fetch messages by UID range
    FindByUIDRange(ctx context.Context, userID, mailbox string, min, max uint32) ([]*domain.Message, error)
    
    // Flags
    AddFlags(ctx context.Context, messageID string, flags ...string) error
    RemoveFlags(ctx context.Context, messageID string, flags ...string) error
    SetFlags(ctx context.Context, messageID string, flags ...string) error

    // UID Management
    // AssignUID assigns a UID to a message if it doesn't have one (atomic with increments)
    AssignUID(ctx context.Context, messageID string, mailbox string) (uint32, error)
}
```
