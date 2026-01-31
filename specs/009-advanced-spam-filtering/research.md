# Research: Advanced Spam Filtering

**Feature**: 009-advanced-spam-filtering
**Date**: 2026-01-31

## 1. Tokenization Strategy

To adhere to "Dependency Minimalism" and "High Performance", we will implement a custom Unicode scanner rather than using `regexp` or external NLP libraries.

### Algorithm
1.  **Input**: Raw text user body (prioritizing plain text part of multipart emails).
2.  **HTML**: If HTML detection is positive, use `golang.org/x/net/html` to extract text nodes (skipping scripts/styles).
3.  **Normalization**: `strings.ToLower`.
4.  **Segmentation**: Iterate over runes.
    *   Effective Tokens: Sequence of `unicode.IsLetter` or `unicode.IsNumber`.
    *   Separators: Any other character.
    *   Min Length: 3 chars (skips "a", "of", "in" naturally).
    *   Strict Max Length: 20 chars (avoids huge garbage tokens).
5.  **Stopwords**: Minimal stopword list (en) is acceptable but Paul Graham's original "Plan for Spam" suggests *not* removing stopwords increases accuracy (e.g., "click here" context). We will skip stopword filtering v1.

## 2. Bayesian Storage Schema (SQLite/Postgres)

We require a schema that supports high-speed "Check" (Select many) and "Train" (Upsert).

### Table: `bayes_tokens`
| Column | Type | Notes |
|Struture|------|-------|
| `token` | TERM (Text) | Primary Key. Case-folded. |
| `spam_count` | INT | Number of times seen in Spam. |
| `ham_count` | INT | Number of times seen in Ham. |

*Storage Optimization*: Use `WITHOUT ROWID` in SQLite to save space and lookups.

### Table: `bayes_global`
| Column | Type | Notes |
|Struture|------|-------|
| `key` | TEXT | PK (e.g., 'total_spam', 'total_ham') |
| `value` | INT | Count |

### Query Strategy
- **Classify**: `SELECT token, spam, ham FROM bayes_tokens WHERE token IN (?,?,?...)` (Batched in chunks of 500).
- **Train**: `INSERT INTO bayes_tokens ... ON CONFLICT(token) DO UPDATE SET spam_count = spam_count + 1`.

## 3. Greylisting Schema

Greylisting tracks "Triplets": (Sender IP Subnet, Sender Email, Recipient Email).

### Table: `greylist`
| Column | Type | Notes |
|Struture|------|-------|
| `ip_net` | TEXT | /24 (IPv4) or /64 (IPv6) CIDR. |
| `sender` | TEXT | |
| `recipient` | TEXT | |
| `first_seen` | INT | Unix Timestamp |
| `last_seen` | INT | Unix Timestamp |
| `blocked_count` | INT | Metrics |

*Primary Key*: `(ip_net, sender, recipient)`

### Logic
1.  **Check**: `SELECT first_seen FROM greylist WHERE pk = ?`.
2.  **If Miss**: Insert `first_seen=now, last_seen=now`. Return 451.
3.  **If Hit**:
    *   If `now - first_seen < RetryDelay`: Return 451 (Too early).
    *   If `now - first_seen >= RetryDelay`: Pass. Update `last_seen`.

## 4. Integration Point

The current `server.go` implementation uses a switch statement.
To support Greylisting properly (saving bandwidth), we must reject at `RCPT TO`.

**Pattern**:
1.  Extend `ports.SpamFilter` with `CheckRecipient(ctx, ip, sender, recipient) error`.
2.  Update `server.go` to call this new method in the `RCPT` case.
3.  Implement `CheckRecipient` in the Spam Adapter to hit the Greylist DB.

**Middleware**:
The `Server` struct has a `handler MessageHandler`. This is valid for Body/DATA checks (Bayes).
Greylisting will live in the Protocol Adapter (`server.go` -> `SpamFilter`), while Bayes lives in the Content Pipeline (`middleware.go` -> `MessageHandler`).
