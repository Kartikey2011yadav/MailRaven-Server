# Data Model: Archive and Spam

**Feature**: `014-archive-spam-handling`

## Database Schema Changes

### Table: `messages`

We are adding fields to the existing `messages` table (originally defined in `001_init.sql`).

| Field | Type | Nullable | Default | Description |
|-------|------|----------|---------|-------------|
| `is_starred` | INTEGER | No | 0 | 1 = Starred/Important, 0 = Normal. Corresponds to IMAP `\Flagged`. |

### Indexes

New indexes required for performance:
- `idx_messages_user_starred`: `(recipient, is_starred)`
- `idx_messages_user_received`: `(recipient, received_at)`

## Domain Entities

### `domain.Message`

Updated Go struct:

```go
type Message struct {
    // ... existing fields ...
    IsStarred  bool      `json:"is_starred"`
    // ... existing fields ...
}
```

### `domain.SearchFilter`

New struct for repository filtering:

```go
type StringFilter struct {
    Value string
    Valid bool
}

type DateRangeFilter struct {
    Start time.Time
    End   time.Time
    Valid bool
}

// MessageFilter defines criteria for listing messages
type MessageFilter struct {
    Limit      int
    Offset     int
    Mailbox    string // "INBOX", "Archive", "Junk", "Trash", "Sent", "Drafts"
    IsRead     *bool  // nil = all, true = read only, false = unread only
    IsStarred  *bool  // nil = all, true = starred only
    DateRange  DateRangeFilter
}
```

## API Data Transfer Objects (DTOs)

### `MessageSummary`

Updated to include `is_starred`.

```json
{
  "id": "msg_123",
  "subject": "Hello",
  "is_starred": true,
  "mailbox": "INBOX",
  ...
}
```

### Note on Mailbox Constants

Standard Mailboxes:
- `INBOX`
- `Archive`
- `Junk`
- `Trash`
- `Sent`
- `Drafts`
