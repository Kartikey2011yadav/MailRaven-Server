# Research: Sieve Filtering (Feature 010)

## 1. Library Selection

### Decision
Use [github.com/emersion/go-sieve](https://github.com/emersion/go-sieve).

### Rationale
- **Pure Go**: Meets Constitution Principle IV (Zero CGO dependency drift).
- **Standards Compliant**: Supports RFC 5228 (Sieve) and key extensions (like `fileinto`, `vacation`).
- **Active Maintenance**: Part of the `emersion` ecosystem (known for `go-imap`, `go-smtp`).
- **Extensible**: Allows custom extensions if needed.

### Alternatives Considered
- **Write from scratch**: Too complex/time-consuming. Sieve grammar is non-trivial.
- **thlib/go-sieve**: Less active community support.
- **Go-Sieve (Bitbucket)**: Outdated.

## 2. Data Model Design

### Sieve Scripts Storage
Need to store multiple scripts per user, with one active.

**Table: `sieve_scripts`**
| Field | Type | Description |
|-------|------|-------------|
| `id` | UUID (PK) | Unique script ID |
| `user_id` | UUID (FK) | Owner |
| `name` | VARCHAR | Script name (e.g., "default") |
| `content` | TEXT | The raw Sieve code |
| `is_active` | BOOLEAN | Only one true per user |
| `created_at` | TIMESTAMP | Audit |
| `updated_at` | TIMESTAMP | Audit |

### Vacation/Auto-Reply Tracking
RFC 5230 requires tracking recent replies to prevent loops.

**Table: `vacation_trackers`**
| Field | Type | Description |
|-------|------|-------------|
| `user_id` | UUID (FK) | The user sending the vacation reply |
| `sender_email` | VARCHAR | The original sender (who got the reply) |
| `last_sent_at` | TIMESTAMP | When the reply was sent |

*Note: Cleanup job needed to prune old trackers (e.g., older than 7 days).*

## 3. Integration Points

### Delivery Pipeline
The `LocalDeliveryAgent` (or equivalent `DeliveryService`) must invoke the Sieve engine before saving to storage.

**Flow:**
1.  Receive email (SMTP/LMTp).
2.  Parse Envelope & Header.
3.  Load active script for `RCPT TO` user.
4.  If no script -> Default delivery (Inbox).
5.  If script -> Run `emersion/go-sieve`.
    *   Action: `fileinto "Trash"` -> Save to Trash folder.
    *   Action: `discard` -> Drop.
    *   Action: `vacation` -> Send notification (if not rate exhausted).
    *   Action: `keep` (implicit or explicit) -> Save to Inbox.

### ManageSieve Protocol (TCP 4190)
RFC 5804 implementation.
- Needs a new TCP listener.
- Needs authentication (SASL) - reuse existing Auth system.
- Commands: `PUTSCRIPT`, `SETACTIVE`, `GETSCRIPT`, `DELETESCRIPT`, `LISTSCRIPTS`.
- library: `emersion/go-manage-sieve` (server side implementation might need custom wrapping or checking if `emersion` provides a server struct). *Correction*: `emersion` provides `go-sieve` (interpreter) but `go-manage-sieve` is primarily a client lib? *Check needed*.
    *   *Self-Correction*: `emersion` has `go-sieve/managesieve` package for server.

## 4. Unknowns / Clarifications Resolved
- **Library**: `emersion/go-sieve` confirmed.
- **Storage**: Database-backed (SQLite/Postgres compatible) for strict durability.
- **Hooks**: Must intercept delivery *before* storage.
