# Research: Production Hardening

## 1. Docker Base Image

### Decision: `gcr.io/distroless/static-debian12`
We will use Google's Distroless Static image for the production Docker container.

### Rationale
- **Security**: Contains no shell, package manager, or unnecessary binaries, drastically reducing the attack surface.
- **Size**: Extremely small (~2MB base), ideal for a Go static binary.
- **Compatibility**: Perfect for CGO-free Go applications (MailRaven's target).

### Alternatives Considered
- **`alpine`**: Popular but includes a shell and package manager (APK), increasing vulnerability surface. Using `musl` libc can sometimes cause subtle DNS/resolution bugs in Go if not compiled carefully (though CGO_ENABLED=0 negates this).
- **`scratch`**: The smallest possible image, but lacks CA certificates (needed for outbound SMTP/TLS) and timezone data unless manually copied. Distroless includes these.

## 2. ACME / Let's Encrypt Integration

### Decision: `golang.org/x/crypto/acme/autocert`
We will use the standard Go `autocert` package.

### Implementation Pattern
- **Port 80 (HTTP)**: Run a specialized HTTP server that handles *only* the ACME "http-01" challenge and redirects everything else to HTTPS.
  - Code: `autocert.Manager.HTTPHandler(nil)`
- **Port 443 (HTTPS)**: The main `http.Server` uses `autocert.Manager.TLSConfig()`.
- **Integration**: The `autocert` manager will sit alongside our existing API router.

### Rationale
- Native Go solution, no external dependencies (like certbot).
- Automatic renewal management.
- Supports caching certificates to disk (needed for container restarts).

## 3. Spam Filtering

### Decision: DNSBL (RBL) with `github.com/mrichman/godnsbl`
We will implement an SMTP middleware that checks connecting IPs against popular Blocklists.

### Selected Providers
1.  **Zen Spamhaus** (`zen.spamhaus.org`): Gold standard, but requires acceptance of Terms (Free for non-commercial/low volume).
2.  **SpamCop** (`bl.spamcop.net`): Reliable, free.
3.  **Barracuda** (`b.barracudacentral.org`): Good alternative.

### Rationale
- **Efficiency**: DNS lookups are fast and cheap compared to content analysis (Bayesian).
- **Effectiveness**: Blocks ~80-90% of botnet traffic at the connection level before data transfer.

### Alternatives Considered
- **Rspamd**: External service. Adds significant deployment complexity (Sidecar container, Redis). Overkill for MVP production hardening.
- **Bayesian Filter (Go)**: Implementing a naive Bayesian filter requires training data and state management. Too complex for this phase.

## 4. SQLite Backup Strategy

### Decision: Application-Triggered `VACUUM INTO`
The running MailRaven server will expose an administrative mechanism to trigger a safe hot backup.

### Mechanism
- **CLI Command**: `mailraven backup --target /path`
- **Under the hood**: The CLI sends an authenticated request to the running server.
- **Server Action**: Executes `VACUUM INTO '/path/backup.db'` via the existing `modernc.org/sqlite` driver (which supports this feature).
- **Blobs**: The filesystem blobs (email bodies) are backed up via standard file copy *after* the DB snapshot.

### Rationale
- **Safety**: Copying a running SQLite file (`cp.db`) risks corruption if a transaction is active. `VACUUM INTO` creates a transactionally consistent copy.
- **Simplicity**: No need for external `sqlite3` CLI tools inside the Docker container.
