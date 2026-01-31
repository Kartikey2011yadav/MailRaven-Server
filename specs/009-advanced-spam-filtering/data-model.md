# Data Model: Advanced Spam Filtering

**Feature**: 009-advanced-spam-filtering

## Entities

### 1. BayesToken
Stores the statistical data for each word found in training emails.

- **Table**: `bayes_tokens`
- **Storage**: SQLite / Postgres

| Field | Type | Description |
|-------|------|-------------|
| `token` | Text (PK) | The word/token (lowercase). Limit 64 chars. |
| `spam_count` | Int | Occurrences in Spam messages. |
| `ham_count` | Int | Occurrences in Ham messages. |
| `updated_at` | Timestamp | For pruning old tokens (optional future) |

### 2. BayesGlobal
Stores global training counters needed for probability normalization ($P(Spam)$).

- **Table**: `bayes_global`

| Field | Type | Description |
|-------|------|-------------|
| `key` | Text (PK) | 'total_spam', 'total_ham' |
| `value` | Int | The count. |

### 3. GreylistEntry
Tracks delivery attempts from unknown triplets.

- **Table**: `greylist`

| Field | Type | Description |
|-------|------|-------------|
| `ip_net` | Text (PK) | Network CIDR string (e.g., "1.2.3.0/24"). |
| `sender` | Text (PK) | Normalize envelope sender. |
| `recipient` | Text (PK) | Normalize envelope recipient. |
| `first_seen` | Int64 | Unix epoch of first attempt. |
| `last_seen` | Int64 | Unix epoch of last attempt/update. |
| `count` | Int | Number of attempts blocked. |

## Interfaces

### `ports.SpamFilter` (Update)

```go
type SpamFilter interface {
    // Existing
    CheckConnection(ctx context.Context, ip string) error
    
    // NEW: Generic content check (Bayes)
    CheckContent(ctx context.Context, content io.Reader, headers map[string]string) (*domain.SpamCheckResult, error)

    // NEW: Recipient check (Greylist)
    CheckRecipient(ctx context.Context, ip, sender, recipient string) error
}
```

### `ports.BayesTrainer` (New)

```go
type BayesTrainer interface {
    TrainSpam(ctx context.Context, content io.Reader) error
    TrainHam(ctx context.Context, content io.Reader) error
}
```
