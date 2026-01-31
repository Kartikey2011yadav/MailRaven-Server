# Phase 0: Research Findings

## Unknowns & Clarifications

### 1. Persistent UID Mapping
**Question**: How to maintain strictly ascending 32-bit UIDs for IMAP messages when the message ID is a string (UUID)?
**Findings**:
- IMAP requires `UID` to be unique and strictly ascending per mailbox.
- SQLite `ROWID` is stable and ascending usually, but relies on implementation details.
- Best Practice: Add an explicit `UID` column to the message table.
- **Decision**: Add `uid` (INTEGER) and `mailbox` (TEXT, default 'INBOX') columns to the `messages` table.
- **Mechanism**:
    - Use `sqlite_sequence` or a dedicated `mailbox_meta` table to track `UIDNEXT` and `UIDVALIDITY`.
    - On insert: Transaction -> Get `UIDNEXT` -> Assign to Message -> Increment `UIDNEXT`.

### 2. IDLE Concurrency
**Question**: How to implement `IDLE` in Go without blocking threads or leaking goroutines?
**Findings**:
- `IDLE` requires the server to wait for updates while keeping the TCP connection open.
- Go's `net.Conn` reads block.
- **Decision**:
    - Use a `chan struct{}` or `chan Event` per session.
    - When `IDLE` is received, enter a loop that `select`s on:
        1. Context Done (Client disconnect/Server stop).
        2. Client Message (Client sends `DONE` to end IDLE - requires separate goroutine to read).
        3. Event Channel (New email arrives).
    - Requires a global or user-scoped `NotificationHub` or `EventBus`.

### 3. Autodiscover XML Schemas
**Question**: What are the exact XML formats for Thunderbird and Outlook?
**Findings**:
- **Mozilla (Thunderbird)**:
    - Path: `GET /.well-known/autoconfig/mail/config-v1.1.xml`
    - Format: [Mozilla ISPDB Schema](https://wiki.mozilla.org/Thunderbird:Autoconfiguration:ConfigFileFormat)
- **Microsoft (Outlook)**:
    - Path: `POST /autodiscover/autodiscover.xml`
    - Format: [POX Autodiscover](https://learn.microsoft.com/en-us/exchange/client-developer/web-service-reference/autodiscover-xml-elements)
- **Decision**: Implement both handlers in `internal/adapters/http`. Hardcode templates with dynamic values (Hostname, encryption, ports).

## Technology Choices

| Component | Choice | Rationale | Alternatives |
|-----------|--------|-----------|--------------|
| **IMAP Parser** | Custom / Simple | Only need basic commands (SELECT, FETCH). Go stdlib string splitting/regex is sufficient for MVP. | `emersion/go-imap` (too heavy/opinionated for this stage). |
| **UID Storage** | Database Column | Easiest to query, maintain, and index. | Separate mapping table (too complex). |
| **Event Bus** | Go Channels | Simple, in-memory. Multi-node scaling not required yet. | Redis Pub/Sub (overkill). |
| **MIME Construction** | `net/textproto` | Standard lib, robust enough for creating multipart messages if needed (though we mostly serve blobs). | Third-party libs. |

## Implementation Strategy

1.  **Database Migration**: Add columns to `email` table.
2.  **Repository Update**: Add `GetUIDNext`, `ListMessagesByUID`.
3.  **HTTP Autodiscover**: Add handlers.
4.  **IMAP Handler**: Implement state machine for `SELECT`, `FETCH`, `IDLE`.
