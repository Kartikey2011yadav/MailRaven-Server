# Concurrency Patterns in Go (As Done in MailRaven)

---

## 1. Server Per-Connection Pattern

The most common pattern in MailRaven — one goroutine per client:

```go
// internal/adapters/smtp/server.go
func (s *Server) Start(ctx context.Context) error {
    listener, _ := net.Listen("tcp", ":25")

    for {
        conn, err := listener.Accept()
        if err != nil {
            if ctx.Err() != nil { return nil }  // Shutdown
            continue
        }
        go s.handleConnection(ctx, conn)  // One goroutine per connection
    }
}
```

This handles 10,000+ concurrent connections because goroutines are cheap (~8KB each).

---

## 2. Worker Pattern (Background Processing)

```go
// internal/adapters/smtp/delivery.go
type DeliveryWorker struct {
    stopChan chan struct{}
    wg       sync.WaitGroup
    ticker   *time.Ticker
}

func (w *DeliveryWorker) Start() {
    w.wg.Add(1)
    go w.processLoop()
}

func (w *DeliveryWorker) processLoop() {
    defer w.wg.Done()
    for {
        select {
        case <-w.stopChan:
            return                    // Clean exit
        case <-w.ticker.C:
            w.ProcessNext()           // Do work every 5 seconds
        }
    }
}

func (w *DeliveryWorker) Stop() {
    close(w.stopChan)    // Signal stop
    w.wg.Wait()          // Wait for goroutine to finish
    w.ticker.Stop()
}
```

**Key patterns:**
- `chan struct{}` for signaling (zero memory)
- `sync.WaitGroup` to wait for goroutines to finish
- `select` to multiplex multiple channels

---

## 3. Pub/Sub with Channels

```go
// internal/adapters/pubsub/memory/pubsub.go
type PubSub struct {
    mu          sync.RWMutex
    subscribers map[string][]chan []byte
}

func (p *PubSub) Publish(ctx context.Context, channel string, payload []byte) error {
    p.mu.RLock()
    defer p.mu.RUnlock()

    for _, ch := range p.subscribers[channel] {
        select {
        case ch <- payload:   // Send if receiver ready
        default:              // Drop if buffer full (non-blocking)
        }
    }
    return nil
}

func (p *PubSub) Subscribe(ctx context.Context, channel string) (<-chan []byte, error) {
    ch := make(chan []byte, 64)  // Buffered channel
    p.mu.Lock()
    p.subscribers[channel] = append(p.subscribers[channel], ch)
    p.mu.Unlock()
    return ch, nil  // Return receive-only channel
}
```

---

## 4. Mutex for Shared State

```go
// internal/adapters/http/middleware/ratelimit.go
type RateLimiter struct {
    mu       sync.RWMutex          // Protects the map
    requests map[string]*ipLimit
}

// Read lock (multiple readers OK)
func (rl *RateLimiter) Count(ip string) int {
    rl.mu.RLock()
    defer rl.mu.RUnlock()
    if l, ok := rl.requests[ip]; ok {
        return l.count
    }
    return 0
}

// Write lock (exclusive)
func (rl *RateLimiter) Allow(ip string) bool {
    rl.mu.Lock()
    defer rl.mu.Unlock()
    // ... modify map ...
}
```

**`sync.Mutex`** = exclusive lock (one at a time)
**`sync.RWMutex`** = readers can share, writers are exclusive

---

## 5. Context Cancellation

```go
func (s *Server) Start(ctx context.Context) error {
    listener, _ := net.Listen("tcp", ":25")

    // When ctx is cancelled, close the listener
    go func() {
        <-ctx.Done()               // Blocks until cancelled
        listener.Close()           // Unblocks Accept() below
    }()

    for {
        conn, err := listener.Accept()
        if err != nil {
            if ctx.Err() != nil {
                return nil         // Clean shutdown
            }
            continue
        }
        go s.handleConnection(ctx, conn)
    }
}

// In serve.go:
ctx, cancel := context.WithCancel(context.Background())
go smtpServer.Start(ctx)

// On SIGTERM:
cancel()  // All servers stop
```

---

## 6. Panic Recovery

```go
func (s *Server) handleConnection(ctx context.Context, conn net.Conn) {
    defer conn.Close()
    defer func() {
        if r := recover(); r != nil {
            s.logger.Error("panic recovered", "error", r)
            // Connection closes, server keeps running
        }
    }()

    // ... handle SMTP commands ...
    // If this panics, only THIS goroutine dies, not the server
}
```

---

## 7. Timeout Pattern

```go
// Set deadline on network operations
conn.SetReadDeadline(time.Now().Add(5 * time.Minute))
line, err := reader.ReadString('\n')
if err != nil {
    // err is net.Error with Timeout() == true if deadline exceeded
    return
}

// Context timeout for operations
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
defer cancel()
err := client.Send(ctx, from, to, body)
```

---

## 8. Once Pattern (Initialization)

```go
import "sync"

var (
    instance *Database
    once     sync.Once
)

func GetDB() *Database {
    once.Do(func() {
        instance = connectToDatabase()  // Runs exactly once, thread-safe
    })
    return instance
}
```

---

## 9. Fan-Out Pattern

```go
// Send to multiple recipients in parallel
var wg sync.WaitGroup
errors := make(chan error, len(recipients))

for _, rcpt := range recipients {
    wg.Add(1)
    go func(to string) {
        defer wg.Done()
        if err := deliver(to, message); err != nil {
            errors <- err
        }
    }(rcpt)
}

wg.Wait()
close(errors)

for err := range errors {
    log.Error("delivery failed", "error", err)
}
```
