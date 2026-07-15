# MailRaven Server

## Project Overview

MailRaven is an open-source, self-hosted email server designed for simplicity, robustness, and lightweight deployment. It follows a mobile-first philosophy with a unified web UI for administration and webmail access.

**Goal:** Anyone should be able to pick up this project, deploy it, and run their own full-featured mail server with minimal friction.

**Role Model:** All users get webmail access. Admins get webmail + admin panel. Not all users are admins, but all admins are users.

## Architecture

Hexagonal (Ports & Adapters) architecture with 5 layers:
- **Listener**: SMTP (RFC 5321), IMAP4rev1 (RFC 3501), ManageSieve (RFC 5804)
- **Logic**: SPF/DKIM/DMARC validation, email routing, Sieve filtering
- **Storage**: Repository interfaces (database-agnostic via ports)
- **Search**: FTS5 (SQLite) or TSVECTOR (PostgreSQL)
- **API**: REST/JSON with JWT authentication

**Data Integrity Guarantee:** "250 OK" means data is fsync'd to disk before SMTP acknowledgment.

## Tech Stack

### Backend (Go 1.25+, CGO-free)
- **Router**: go-chi/chi v5
- **Database**: PostgreSQL (default, recommended) via pgx/v5 | SQLite via modernc.org/sqlite
- **DNS**: miekg/dns (SPF/DKIM/DMARC lookups)
- **Auth**: golang-jwt/jwt v5 (7-day expiry, HMAC-SHA256, 32+ char secret required)
- **Sieve**: go-sieve (mail filtering, 5s execution timeout, 100KB script limit)
- **Optional distributed**: Redis (cache/pubsub), NATS JetStream (broker), MinIO (blob storage)

### Frontend (client/)
- React 19 + TypeScript 5.9
- Vite 7 (build + HMR), output → `internal/adapters/http/static/dist/`
- Tailwind CSS 4.1 with dark-first glassmorphism theme
- shadcn/ui + Radix UI, Framer Motion (animations), Recharts (charts)
- React Router v7, React Hook Form + Zod validation
- Axios (HTTP), Sonner (toasts), Lucide (icons)

### Design System
- **Theme**: Dark-first glassmorphism (frosted glass cards, gradient accents, glow effects)
- **Layout**: Unified sidebar — Mail section for all users, Admin section for admins only
- **Components**: `glass-card.tsx` for cards, `unified-sidebar.tsx` for navigation
- **Login**: Animated glassmorphism card with setup detection

## Project Structure

```
cmd/
  mailraven/           # Server binary (quickstart, serve, check-config)
  mailraven-cli/       # CLI tool for domain/user management
client/                # React frontend (Web Admin + Webmail Lite)
  src/pages/           # Login, Dashboard, Domains, Users, Mail (Inbox/Sent/Drafts/Archive/Trash)
  src/pages/setup/     # First-launch setup wizard (5 steps)
  src/components/      # UI components (shadcn/ui pattern + unified-sidebar)
  src/layout/          # UnifiedLayout.tsx (single layout for all routes)
internal/
  core/
    domain/            # Entities: Message, User, Mailbox, SMTPSession
    ports/             # Interfaces: EmailRepository, BlobStore, Search, Sieve
    services/          # EmailService, UserService, BackupService, SpamProtection
  adapters/
    smtp/              # SMTP server (500 conn limit, temp file streaming, DKIM/SPF/DMARC)
    imap/              # IMAP4rev1 with IDLE push, 500 conn limit, session context
    managesieve/       # ManageSieve protocol with session context
    sieve/             # Sieve interpreter + vacation (5s timeout, 100KB limit)
    http/              # REST API handlers + middleware + static file serving
    storage/
      postgres/        # PostgreSQL repo + migrations (50 max conns, 5min lifetime)
      sqlite/          # SQLite repo + migrations
      disk/            # Filesystem blob storage
      minio/           # S3-compatible blob storage
    cache/memory/      # In-memory cache (standalone mode)
    cache/redis/       # Redis cache + rate limiter
    broker/local/      # Synchronous local broker
    broker/nats/       # NATS JetStream broker
    spam/              # Rspamd, Bayesian, Greylisting, DNSBL (3s timeout), Rate limiting
    backup/            # Hot backup utilities
  infra/               # Infrastructure factory (wires adapters by deployment mode)
  config/              # Configuration (validates 32+ char JWT secret)
  observability/       # Logging and Prometheus metrics
deployment/
  kubernetes/          # Kustomize manifests (base, local, keda overlays)
  config.example.yaml  # Full annotated config reference
build/
  Dockerfile           # Multi-stage: Node 22 → Go 1.25 → distroless (go mod download, not vendor)
  Dockerfile.frontend  # Node 22 → Nginx
tests/                 # Integration tests (all pass in ~14s)
```

## API Endpoints

### Public (no auth)
- `GET /api/v1/setup/status` — check if setup needed (user count = 0)
- `POST /api/v1/setup/complete` — initial setup (mutex-protected, creates domain + admin)
- `POST /api/v1/auth/login` — JWT login (7-day token)
- `GET /health`, `GET /healthz`, `GET /readyz` — health probes
- `GET /metrics` — Prometheus metrics

### Protected (JWT required)
- Messages: `GET/PATCH /messages`, `GET /messages/{id}`, `POST /messages/send`, `GET /messages/search`
- Sieve: CRUD at `/sieve/scripts/`
- Self: `PUT /users/self/password`

### Admin (JWT + admin role)
- Stats: `GET /admin/stats`
- Users: CRUD at `/admin/users`
- Domains: CRUD at `/admin/domains`
- System: `GET/POST /admin/system/update`, `POST /admin/backup`

## Security Hardening (Completed)

- CORS: credentials only with specific origins (not wildcard)
- SMTP: CRLF injection prevention, 500-connection semaphore, temp file streaming
- IMAP: 500-connection semaphore, session context for cancellation
- Rate limiting: extracts real IP from X-Forwarded-For/X-Real-IP
- JWT: 32+ character secret enforced in config validation
- PostgreSQL: connection pool (50 open, 10 idle, 5min lifetime)
- DNSBL: 3-second timeout per lookup
- Setup: mutex prevents TOCTOU race condition
- SPF: 10 DNS lookup limit, loop detection, include error → temperror, ip6 + redirect support
- DKIM verification: proper RFC 6376 relaxed/simple canonicalization, h= and c= tag parsing
- DMARC: domain alignment checking (relaxed/strict via aspf=/adkim= tags)
- Sieve: 5-second execution timeout, 100KB script size limit
- Delivery: jitter on retry backoff (prevents thundering herd)
- Context: session contexts in IMAP/ManageSieve, r.Context() in HTTP handlers

## CI/CD

- **CI** (`.github/workflows/ci.yml`): Go 1.25, golangci-lint v2.12.2 (only-new-issues), Node 22, TypeScript type-check, Docker build with GHA cache
- **CD** (`.github/workflows/cd.yml`): Triggered by CI success or v* tags, pushes to GHCR, multi-platform (amd64/arm64)
- **Release** (`.github/workflows/release.yml`): Cross-platform Go binaries on GitHub Release
- **Dependabot**: weekly for gomod, npm, github-actions, docker
- **Pre-commit hook**: `scripts/pre-commit` — go vet, go build, tsc, eslint

## Key Commands

```bash
# Build & Test
make build                    # Build backend binaries
make ui                       # Build React frontend
go test ./internal/... ./cmd/...  # Unit tests (~5s)
go test -timeout 60s ./tests/...  # Integration tests (~14s)
cd client && npx tsc -b      # TypeScript check
cd client && npx eslint .     # Frontend lint

# Development
make docker-dev               # Docker dev environment with hot-reload
cd client && npm run dev      # Frontend dev server (port 5173)

# Lint (local)
golangci-lint run --timeout=5m ./...  # Go lint (v2.12.2 required)
golangci-lint config verify           # Validate .golangci.yml

# Docker
docker build -f build/Dockerfile -t mailraven:latest .
```

## Configuration

Main config: `deployment/config.example.yaml` (copy to `config.yaml`)

Key validation rules:
- `api.jwt_secret`: minimum 32 characters (enforced)
- `storage.driver`: defaults to "postgres"
- `smtp.max_size`: defaults to 10MB

Environment variable overrides use `MAILRAVEN_` prefix.

## Coding Conventions

- No comments unless WHY is non-obvious
- No CGO — pure Go only
- All external deps behind port interfaces
- Assign ignored error returns to `_` (e.g., `_ = conn.Close()`) to satisfy errcheck
- Use `//nolint:gosec` with explanation when suppressing security linter
- Frontend: `eslint-disable-line` for intentional setState-in-effect patterns
- Commits: conventional format (`feat:`, `fix:`, `security:`, `docs:`)
- No `--no-verify` on commits in production (pre-commit hook runs checks)
