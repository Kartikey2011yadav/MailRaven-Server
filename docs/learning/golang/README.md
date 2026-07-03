# Go Crash Course (MailRaven Edition)

Learn Go by understanding how this codebase works. Each doc covers one topic with real examples from MailRaven.

## Reading Order

| # | File | What You'll Learn |
|---|------|-------------------|
| 1 | [Fundamentals](01-fundamentals.md) | Packages, structs, interfaces, error handling, context, goroutines, defer |
| 2 | [Building APIs](02-building-apis.md) | HTTP handlers, middleware, JWT auth, JSON, routing, graceful shutdown |
| 3 | [CLI Tools](03-cli-tools.md) | Subcommands, flags, config loading, signals, exit codes |
| 4 | [Database Patterns](04-database-patterns.md) | SQL, transactions, migrations, repository pattern, connection pooling |
| 5 | [Concurrency](05-concurrency-patterns.md) | Goroutines, channels, mutexes, worker pools, pub/sub, panic recovery |

## Quick Reference

| Go Concept | Where in MailRaven |
|---|---|
| Interface polymorphism | `internal/core/ports/` — every adapter implements a port |
| Dependency injection | `cmd/mailraven/serve.go` — wires all dependencies |
| Context propagation | Every function takes `ctx context.Context` |
| Goroutine-per-connection | `smtp/server.go`, `imap/server.go` |
| Background workers | `smtp/delivery.go` — ticker + stop channel |
| Mutex-protected maps | `http/middleware/ratelimit.go` |
| Config from YAML + env | `internal/config/config.go` |
| Builder/factory pattern | `internal/infra/factory.go` |
| Graceful shutdown | `cmd/mailraven/serve.go` — SIGTERM handling |

## How to Explore

```bash
# See all interfaces (the "contracts")
grep -rn "type.*interface" internal/core/ports/

# See all constructors
grep -rn "func New" internal/adapters/

# See how everything is wired
cat cmd/mailraven/serve.go
```
