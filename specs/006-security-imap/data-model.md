# Data Model

## 1. Configuration (YAML)

New sections in `config.yaml`:

```yaml
spam:
  enabled: true
  rspamd_url: "http://localhost:11333/checkv2"
  dnsbls:
    - "zen.spamhaus.org"
    - "bl.spamcop.net"
  reject_score: 15.0
  header_score: 6.0

imap:
  enabled: true
  port: 143
  port_tls: 993
  allow_insecure_auth: false # Default false, requires STARTTLS or Port 993
```

## 2. Internal Structures

### `internal/adapters/spam/rspamd.go`

```go
type CheckResult struct {
    Action     string  `json:"action"`
    Score      float64 `json:"score"`
    Required   float64 `json:"required_score"`
    Symbols    map[string]Symbol `json:"symbols"`
}
```

### `internal/adapters/imap/session.go`

```go
type State int

const (
    StateNotAuthenticated State = iota
    StateAuthenticated
    StateSelected // Out of scope for now
    StateLogout
)

type Session struct {
    conn    net.Conn
    state   State
    user    *domain.User // nil until auth
    isTLS   bool
}
```

## 3. Database Changes

None required for this feature.
