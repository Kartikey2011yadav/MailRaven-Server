# Building CLI Tools in Go (As Done in MailRaven)

---

## 1. CLI Structure

MailRaven has two binaries:
- `cmd/mailraven/` — The server (`mailraven serve`)
- `cmd/mailraven-cli/` — Admin tool (`mailraven-cli user create`)

Each uses a **subcommand pattern** (like `git commit`, `docker run`).

---

## 2. Argument Parsing with `os.Args`

The simplest approach (used in `cmd/mailraven/main.go`):

```go
package main

import (
    "fmt"
    "os"
)

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Usage: mailraven <command>")
        fmt.Println("Commands: serve, quickstart")
        os.Exit(1)
    }

    switch os.Args[1] {
    case "serve":
        if err := RunServe(); err != nil {
            fmt.Fprintf(os.Stderr, "Error: %v\n", err)
            os.Exit(1)
        }
    case "quickstart":
        if err := RunQuickstart(); err != nil {
            fmt.Fprintf(os.Stderr, "Error: %v\n", err)
            os.Exit(1)
        }
    default:
        fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
        os.Exit(1)
    }
}
```

---

## 3. Flags with `flag` Package

```go
// cmd/mailraven/serve.go
import "flag"

func RunServe() error {
    // Define flags
    fs := flag.NewFlagSet("serve", flag.ExitOnError)
    configPath := fs.String("config", "/etc/mailraven/config.yaml", "Path to config file")

    // Parse
    fs.Parse(os.Args[2:])  // Skip "mailraven" and "serve"

    // Use
    fmt.Println("Loading config from:", *configPath)
    cfg, err := config.LoadFromFile(*configPath)
    // ...
}
```

Usage: `mailraven serve --config ./my-config.yaml`

---

## 4. Interactive CLI (mailraven-cli)

```go
// cmd/mailraven-cli/users.go
func handleUserCreate(args []string) error {
    fs := flag.NewFlagSet("user-create", flag.ExitOnError)
    email := fs.String("email", "", "User email (required)")
    password := fs.String("password", "", "User password")
    role := fs.String("role", "user", "User role (user/admin)")
    fs.Parse(args)

    if *email == "" {
        return fmt.Errorf("--email is required")
    }

    // If no password provided, prompt interactively
    if *password == "" {
        fmt.Print("Password: ")
        pw, _ := term.ReadPassword(int(os.Stdin.Fd()))
        *password = string(pw)
        fmt.Println()
    }

    // Call API or DB directly
    // ...
    fmt.Printf("User %s created with role %s\n", *email, *role)
    return nil
}
```

---

## 5. Exit Codes

```go
func main() {
    if err := run(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)  // Non-zero = failure
    }
    // os.Exit(0) is implicit
}
```

Convention:
- `0` = success
- `1` = general error
- `2` = usage error (wrong args)

---

## 6. Building Multiple Binaries

From `Makefile`:
```makefile
build:
    go build -o bin/mailraven ./cmd/mailraven
    go build -o bin/mailraven-cli ./cmd/mailraven-cli
```

Each `cmd/<name>/` directory produces one binary. They share code from `internal/`.

---

## 7. Configuration Loading Pattern

```go
// internal/config/config.go
func LoadFromFile(path string) (*Config, error) {
    // 1. Read file
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("read config: %w", err)
    }

    // 2. Parse YAML
    var cfg Config
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return nil, fmt.Errorf("parse config: %w", err)
    }

    // 3. Apply defaults
    if cfg.SMTP.Port == 0 {
        cfg.SMTP.Port = 25
    }

    // 4. Override from environment
    if v := os.Getenv("MAILRAVEN_DOMAIN"); v != "" {
        cfg.Domain = v
    }

    // 5. Validate
    if cfg.Domain == "" {
        return nil, fmt.Errorf("domain is required")
    }

    return &cfg, nil
}
```

**Priority**: env vars > YAML file > defaults. This is the standard 12-factor pattern.

---

## 8. Signals and Graceful Shutdown

```go
import (
    "os"
    "os/signal"
    "syscall"
)

func RunServe() error {
    // Catch SIGTERM (docker stop) and SIGINT (Ctrl+C)
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

    // Start services...

    // Block until signal received
    sig := <-sigChan
    fmt.Printf("Received %v, shutting down...\n", sig)

    // Cleanup
    server.Stop()
    db.Close()
    return nil
}
```
