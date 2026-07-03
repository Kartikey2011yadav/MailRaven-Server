# Database Patterns in Go (As Done in MailRaven)

---

## 1. The `database/sql` Package

Go's standard library provides a generic SQL interface. You plug in a driver (SQLite, Postgres, etc.) and the API stays the same:

```go
import (
    "database/sql"
    _ "modernc.org/sqlite"          // SQLite driver (imported for side effect)
    _ "github.com/jackc/pgx/v5"     // PostgreSQL driver
)

db, err := sql.Open("sqlite", "file:mailraven.db?_journal_mode=WAL")
// or
db, err := sql.Open("pgx", "postgres://user:pass@localhost/mailraven")
```

The `_` import registers the driver without using it directly.

---

## 2. Connection Pooling

`sql.DB` is NOT a single connection — it's a **connection pool**:

```go
db.SetMaxOpenConns(25)       // Max simultaneous connections
db.SetMaxIdleConns(10)       // Keep 10 idle for reuse
db.SetConnMaxLifetime(5*time.Minute)  // Recycle old connections
```

In MailRaven: SQLite uses 25 (limited by WAL), PostgreSQL uses 50+ per pod.

---

## 3. Queries

```go
// Single row
var email string
err := db.QueryRowContext(ctx,
    "SELECT email FROM users WHERE id = $1", userID,
).Scan(&email)

// Multiple rows
rows, err := db.QueryContext(ctx,
    "SELECT id, subject, sender FROM messages WHERE recipient = $1 LIMIT $2",
    recipient, limit,
)
defer rows.Close()

var messages []Message
for rows.Next() {
    var m Message
    if err := rows.Scan(&m.ID, &m.Subject, &m.Sender); err != nil {
        return nil, err
    }
    messages = append(messages, m)
}
```

**Key rules**:
- Always use `Context` variants (`QueryContext`, `ExecContext`)
- Always `defer rows.Close()`
- Use `$1, $2` placeholders (PostgreSQL) or `?` (SQLite) — NEVER string interpolation (SQL injection)

---

## 4. Transactions

```go
func (r *EmailRepository) Save(ctx context.Context, msg *domain.Message) error {
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()  // No-op if committed

    // Insert into mailboxes
    _, err = tx.ExecContext(ctx,
        "INSERT INTO mailboxes (name, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING",
        msg.Mailbox, msg.Recipient,
    )
    if err != nil {
        return err
    }

    // Insert message
    _, err = tx.ExecContext(ctx,
        "INSERT INTO messages (id, sender, recipient, subject) VALUES ($1, $2, $3, $4)",
        msg.ID, msg.Sender, msg.Recipient, msg.Subject,
    )
    if err != nil {
        return err  // tx.Rollback() runs via defer
    }

    return tx.Commit()  // Both inserts succeed or neither does
}
```

---

## 5. Migrations

MailRaven runs SQL migration files on startup:

```go
// Read migration files from directory
files, _ := os.ReadDir("migrations/")
for _, file := range files {
    sql, _ := os.ReadFile(filepath.Join("migrations", file.Name()))
    _, err := db.ExecContext(ctx, string(sql))
}
```

Migration file example (`001_initial.sql`):
```sql
CREATE TABLE IF NOT EXISTS messages (
    id TEXT PRIMARY KEY,
    sender TEXT NOT NULL,
    recipient TEXT NOT NULL,
    subject TEXT DEFAULT '',
    body_path TEXT NOT NULL,
    received_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_messages_recipient ON messages(recipient);
```

---

## 6. Repository Pattern

MailRaven uses the **Repository Pattern** — each domain entity gets a repository interface:

```go
// internal/core/ports/repository.go (interface)
type EmailRepository interface {
    Save(ctx context.Context, msg *domain.Message) error
    FindByID(ctx context.Context, id string) (*domain.Message, error)
    List(ctx context.Context, filter domain.MessageFilter) ([]*domain.Message, int, error)
    Delete(ctx context.Context, id string) error
}

// internal/adapters/storage/sqlite/email_repo.go (implementation)
type EmailRepository struct {
    db *sql.DB
}

func (r *EmailRepository) FindByID(ctx context.Context, id string) (*domain.Message, error) {
    row := r.db.QueryRowContext(ctx, "SELECT * FROM messages WHERE id = ?", id)
    // ... scan into struct ...
}
```

**Why this pattern?**
- Swap databases without changing business logic
- Test with mock repositories (no real DB needed)
- Clear separation of SQL from domain logic

---

## 7. PostgreSQL-Specific Features

```go
// FOR UPDATE SKIP LOCKED (used in delivery queue)
row := tx.QueryRowContext(ctx, `
    UPDATE outbound_queue 
    SET status = 'PROCESSING' 
    WHERE id = (
        SELECT id FROM outbound_queue 
        WHERE status = 'PENDING' AND next_retry_at <= NOW()
        ORDER BY created_at 
        FOR UPDATE SKIP LOCKED 
        LIMIT 1
    )
    RETURNING id, sender, recipient, blob_key
`)

// Full-text search with TSVECTOR
rows, _ := db.QueryContext(ctx, `
    SELECT id, subject, ts_rank(search_vector, query) as rank
    FROM messages, plainto_tsquery('english', $1) query
    WHERE search_vector @@ query
    ORDER BY rank DESC
    LIMIT $2
`, searchTerm, limit)
```

---

## 8. Testing with Repositories

```go
// In tests, use a real in-memory SQLite DB
func setupTestDB(t *testing.T) *sql.DB {
    db, _ := sql.Open("sqlite", ":memory:")
    // Run migrations
    db.Exec(migrationSQL)
    return db
}

func TestEmailSave(t *testing.T) {
    db := setupTestDB(t)
    repo := sqlite.NewEmailRepository(db, nil)

    msg := &domain.Message{ID: "test-1", Sender: "a@b.com"}
    err := repo.Save(context.Background(), msg)
    assert.NoError(t, err)

    found, err := repo.FindByID(context.Background(), "test-1")
    assert.Equal(t, "a@b.com", found.Sender)
}
```
