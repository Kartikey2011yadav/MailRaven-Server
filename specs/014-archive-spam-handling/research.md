# Research: Archive and Spam Mechanism

**Feature**: Archive and Spam Mechanism (`014-archive-spam-handling`)
**Date**: 2026-02-10

## Unknowns & Clarifications

### 1. Database Schema Changes
**Unknown**: How to efficiently store "Starred" status and filters?
**Research**:
- `Message` table in `001_init.sql` likely has a `mailbox` column (or `folder`).
- Need to check `001_init.sql` and `004_add_imap_fields.sql`.
- "Starred" corresponds to IMAP `\Flagged`.
- If `flags` is already stored as a space-separated string (common IMAP pattern), we can use `LIKE '%\Flagged%'` or add a dedicated boolean column `is_starred` for faster indexing/filtering.
- **Decision**: Add a dedicated `is_starred` (BOOLEAN/INTEGER) column.
- **Rationale**: Searching `flags` string is slow and harder to index. API access pattern "Show me starred emails" is frequent.
- **Migration**: `011_add_starred_column.sql`.

### 2. Spam Training Performance
**Unknown**: Is Bayesian training fast enough for synchronous API calls?
**Research**:
- Bayesian training involves tokenizing body and updating counters in DB.
- Usually fast (<100ms) for typical email processing.
- However, reading the blob body might be slow if large.
- **Decision**: Perform training Asynchronously (Goroutine) or keep it synchronous but minimal.
- **Refinement**: For MVP, lets do it synchronously to ensure user feedback ("Training failed" vs "Success"). If performance issues arise, move to Job Queue.
- **Actually**: `TrainSpam` interface might already exist.

### 3. Date Filtering
**Unknown**: Best way to filter by date? `ReceivedAt` vs `Date` header.
**Research**:
- `ReceivedAt` is when the server got it (reliable).
- `Date` header is what the sender claimed (unreliable, can be future/past).
- **Decision**: Filter by `ReceivedAt`.
- **API**: `start_date` and `end_date` (RFC3339).

### 4. Mailbox Standards
**Unknown**: What are the canonical names?
**Decision**:
- `INBOX` (case-insensitive usually, but strict in DB)
- `Archive`
- `Junk` (or `Spam`? Let's use `Junk` to match RFC 6154 commonality)
- `Trash`
- `Sent`
- `Drafts`
- **Normalization**: Backend should enforce Title Case or specific constants. `INBOX` is usually uppercase in IMAP.

## Decisions Log

1. **New Column**: Add `is_starred` (INTEGER DEFAULT 0) to `messages` table.
2. **Indexing**: Add index on `(user_email, is_starred)` and `(user_email, received_at)` for filtering performance.
3. **API Design**:
   - `PATCH /api/v1/messages/{id}` with body `{ "mailbox": "Archive" }` or `{ "is_starred": true }`.
   - `POST /api/v1/messages/{id}/spam` (Specific action for Reporting Spam + Training).
   - `POST /api/v1/messages/{id}/ham` (Specific action for Reporting Ham + Training).
   - `GET /api/v1/messages` adds params: `mailbox`, `is_starred`, `start_date`, `end_date`.

## Alternatives Considered

- **Using IMAP Flags column only**: Rejected due to poor query performance for "Get all starred".
- **Deleting Spam immediately**: Rejected. Users need "Junk" folder to review false positives.
