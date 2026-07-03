# Go Fundamentals (As Used in MailRaven)

This crash course covers Go concepts you'll encounter in this codebase. Each section links to real examples.

---

## 1. Project Layout

Go uses a convention-based layout. MailRaven follows the **Standard Go Project Layout**:

```
cmd/            → Entrypoints (main packages). Each subfolder = one binary.
internal/       → Private code (can't be imported by other modules).
  core/         → Business logic (domain-driven design).
  adapters/     → Infrastructure implementations.
```

**Key rule**: `internal/` is enforced by the Go compiler — no external package can import it.

**File:** `cmd/mailraven/main.go` — The server binary entry point.

---

## 2. Packages and Imports

Every Go file starts with a `package` declaration. Files in the same directory must use the same package name.

```go
package smtp  // All files in internal/adapters/smtp/ use this

import (
    "context"          // Standard library
    "fmt"

    "github.com/user/repo/internal/core/ports"  // Internal module import
    "github.com/redis/go-redis/v9"              // External dependency
)
```

**Naming convention**: Package name = last path segment. Import as `ports`, `smtp`, etc.

---

## 3. Types, Structs, and Methods

Go uses structs instead of classes:

```go
// Definition (internal/adapters/smtp/server.go)
type Server struct {
    config   *config.Config
    logger   *observability.Logger
    handler  MessageHandler
    listener net.Listener
    mu       sync.RWMutex
}

// Constructor (returns pointer to new struct)
func NewServer(cfg *config.Config, logger *observability.Logger) *Server {
    return &Server{
        config: cfg,
        logger: logger,
    }
}

// Method (receiver = pointer to Server)
func (s *Server) Start(ctx context.Context) error {
    // s is like 'this' or 'self'
    return nil
}
```

**Pointer receivers** (`*Server`): Can modify the struct. Used 99% of the time.
**Value receivers** (`Server`): Can't modify. Used for small immutable types.

---

## 4. Interfaces (The Core of Go's Design)

Interfaces define behavior without specifying implementation. This is how MailRaven achieves pluggable storage (SQLite ↔ Postgres ↔ MinIO):

```go
// Definition (internal/core/ports/storage.go)
type BlobStore interface {
    Write(ctx context.Context, messageID string, content []byte) (string, error)
    Read(ctx context.Context, path string) ([]byte, error)
    Delete(ctx context.Context, path string) error
}
```

Any struct that has these methods **automatically** implements the interface (no `implements` keyword):

```go
// internal/adapters/storage/disk/blob_store.go
type DiskBlobStore struct { basePath string }

func (d *DiskBlobStore) Write(ctx context.Context, id string, content []byte) (string, error) { ... }
func (d *DiskBlobStore) Read(ctx context.Context, path string) ([]byte, error) { ... }
func (d *DiskBlobStore) Delete(ctx context.Context, path string) error { ... }

// internal/adapters/storage/minio/blob_store.go
type MinioBlobStore struct { client *minio.Client }

func (m *MinioBlobStore) Write(ctx context.Context, id string, content []byte) (string, error) { ... }
func (m *MinioBlobStore) Read(ctx context.Context, path string) ([]byte, error) { ... }
func (m *MinioBlobStore) Delete(ctx context.Context, path string) error { ... }
```

Both implement `BlobStore` without knowing about each other. The caller just uses the interface:

```go
func StoreEmail(store ports.BlobStore, id string, body []byte) error {
    path, err := store.Write(context.Background(), id, body)
    // Works with DiskBlobStore OR MinioBlobStore
}
```

---

## 5. Error Handling

Go doesn't have exceptions. Functions return errors explicitly:

```go
func (r *EmailRepository) Save(ctx context.Context, msg *domain.Message) error {
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("failed to begin tx: %w", err)  // Wrap with context
    }
    defer tx.Rollback()  // Cleanup if we return early

    if err := tx.Commit(); err != nil {
        return fmt.Errorf("commit failed: %w", err)
    }
    return nil  // Success
}
```

**Pattern**: `if err != nil { return err }` — you'll see this thousands of times.

**`%w` wrapping**: Creates an error chain. Callers can unwrap with `errors.Is()` or `errors.As()`.

---

## 6. Context

`context.Context` flows through every function call. It carries:
- **Cancellation signal** (stop work when request is done)
- **Deadline** (timeout after X seconds)
- **Values** (request ID, user info)

```go
// Created at the top level
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// Passed down through every layer
err := smtpClient.Send(ctx, from, to, body)
// If ctx times out, Send() stops and returns ctx.Err()
```

In MailRaven: Every handler, repository method, and service call takes `ctx` as first parameter.

---

## 7. Goroutines and Concurrency

Goroutines are lightweight threads (~8KB each vs 1MB for OS threads):

```go
// Start 1000 goroutines (MailRaven: one per SMTP connection)
for {
    conn, _ := listener.Accept()
    go s.handleConnection(ctx, conn)  // 'go' keyword launches goroutine
}
```

**Channels** for communication between goroutines:

```go
// Delivery worker (internal/adapters/smtp/delivery.go)
stopChan := make(chan struct{})  // Signal channel

go func() {
    for {
        select {
        case <-stopChan:   // Received stop signal
            return
        case <-ticker.C:   // Timer fired
            processNext()
        }
    }
}()

// To stop:
close(stopChan)
```

**sync.Mutex** for shared state:

```go
type RateLimiter struct {
    mu      sync.Mutex
    windows map[string]*window
}

func (r *RateLimiter) Allow(key string) bool {
    r.mu.Lock()         // Only one goroutine at a time
    defer r.mu.Unlock() // Unlock when function returns
    // ... safe to read/write r.windows ...
}
```

---

## 8. Defer

`defer` schedules a function call to run when the enclosing function returns:

```go
func handleConnection(conn net.Conn) {
    defer conn.Close()  // Guaranteed to run, even on panic

    // ... use conn ...
    // conn.Close() runs here automatically
}
```

Used extensively in MailRaven for:
- Closing connections (`defer conn.Close()`)
- Releasing locks (`defer mu.Unlock()`)
- Rolling back transactions (`defer tx.Rollback()`)
- Panic recovery (`defer func() { recover() }()`)

---

## 9. The `init()` Function and `main()`

```go
// cmd/mailraven/main.go
package main

func main() {
    // This is the entry point
    if err := RunServe(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

`main()` is called when the binary starts. Only `package main` can have it.

---

## 10. Tags (Struct Tags)

Struct tags control how Go marshals/unmarshals data:

```go
type Config struct {
    Domain string `yaml:"domain"`  // Maps YAML key "domain" to this field
    Port   int    `yaml:"port"`
}

type Message struct {
    ID     string `json:"id"`         // JSON serialization
    Body   string `json:"body"`
    ReadAt *time.Time `json:"read_at,omitempty"`  // Omit if nil
}
```

MailRaven uses:
- `yaml:"..."` for config parsing
- `json:"..."` for API responses
- `db:"..."` (sometimes) for database column mapping
