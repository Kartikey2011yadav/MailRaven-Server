# MailRaven Server

## Project Overview

MailRaven is an open-source, self-hosted email server designed for simplicity, robustness, and lightweight deployment. It follows a mobile-first philosophy with a unified web UI for administration and webmail access.

**Goal:** Anyone should be able to pick up this project, deploy it, and run their own full-featured mail server with minimal friction.

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
- **Auth**: golang-jwt/jwt v5
- **Sieve**: go-sieve (mail filtering)
- **Optional distributed**: Redis (cache/pubsub), NATS JetStream (broker), MinIO (blob storage)

### Frontend (client/)
- React 19 + TypeScript 5.9
- Vite 7 (build + HMR)
- Tailwind CSS 4.1
- Radix UI (accessible components)
- React Router v7, React Hook Form + Zod validation
- Axios (HTTP), Sonner (toasts), Lucide (icons)
- Playwright (e2e tests)

## Project Structure

```
cmd/
  mailraven/           # Server binary (quickstart, serve, check-config)
  mailraven-cli/       # CLI tool for domain/user management
client/                # React frontend (Web Admin + Webmail Lite)
  src/pages/           # Login, Dashboard, Domains, Users, Mail
  src/components/      # UI components (shadcn/ui pattern)
internal/
  core/
    domain/            # Entities: Message, User, Mailbox, SMTPSession
    ports/             # Interfaces: EmailRepository, BlobStore, Search, Sieve
    services/          # EmailService, UserService, BackupService, SpamProtection
  adapters/
    smtp/              # SMTP server + DKIM/SPF/DMARC/DANE validators
    imap/              # IMAP4rev1 with IDLE push support
    managesieve/       # ManageSieve protocol
    sieve/             # Sieve interpreter + vacation auto-reply
    http/              # REST API handlers + middleware + static file serving
    storage/
      sqlite/          # SQLite repository + migrations
      postgres/        # PostgreSQL repository + migrations
      disk/            # Filesystem blob storage
      minio/           # S3-compatible blob storage
    cache/
      memory/          # In-memory cache (standalone mode)
      redis/           # Redis cache + rate limiter
    broker/
      local/           # Synchronous local broker
      nats/            # NATS JetStream broker
    spam/              # Rspamd, Bayesian, Greylisting, DNSBL, Rate limiting
    backup/            # Hot backup utilities
    pubsub/            # Pub/sub adapters
    lock/              # Distributed locks
  infra/               # Infrastructure factory (wires adapters by deployment mode)
  config/              # Configuration loading and validation
  observability/       # Logging and Prometheus metrics
deployment/
  kubernetes/          # Kustomize manifests (base, local, keda overlays)
  config.example.yaml  # Full annotated config reference
  mailraven.service    # Systemd service file
docs/                  # Architecture, guides, API reference, development
tests/                 # Integration test suites
build/
  Dockerfile           # Multi-stage: Node.js -> Go -> distroless
  Dockerfile.frontend  # Node.js -> Nginx
scripts/               # Setup and integration test scripts
```

## Deployment Modes

Configured via `config.yaml` → `mode:` field:

| Mode | Database | Cache | Broker | Blob | Use Case |
|------|----------|-------|--------|------|----------|
| **standalone** | PostgreSQL or SQLite | In-memory | Local sync | Disk | Dev, single-user |
| **docker** | PostgreSQL | Redis | NATS | MinIO | Multi-user, single host |
| **kubernetes** | PostgreSQL | Redis | NATS | MinIO | Production, HA, autoscaling |

**Default database is PostgreSQL.** SQLite is available for lightweight/dev scenarios only.

## Key Commands

```bash
# Build
make build              # Build backend binaries
make ui                 # Build React frontend
make all                # Lint + test + build everything

# Development
make docker-dev         # Start dev environment with hot-reload
cd client && npm run dev  # Frontend dev server with HMR

# Testing
make test               # Unit tests with race detection + coverage
make test-integration   # Integration tests
cd client && npm test   # Playwright e2e tests

# Production
make docker-up          # Start production Docker Compose stack
make docker-build       # Build production Docker images

# Setup
mailraven quickstart    # Interactive setup: generates DKIM keys, config, admin user
mailraven serve         # Start the server
mailraven check-config  # Validate configuration

# CLI Management
mailraven-cli domain add example.com
mailraven-cli user add user@example.com
```

## Configuration

Main config: `deployment/config.example.yaml` (copy to `config.yaml`)

Environment variable overrides use `MAILRAVEN_` prefix. Key sections:
- `mode`: standalone | docker | kubernetes
- `storage.driver`: postgres (default) | sqlite
- `storage.dsn`: PostgreSQL connection string
- `smtp`: Port, hostname, TLS, max message size
- `api`: Port, JWT secret, CORS
- `dkim`: Selector, private key path
- `spam`: Rspamd URL, thresholds, DNSBL lists, rate limits
- `imap`: Ports, TLS
- `tls/acme`: Let's Encrypt auto-cert

## Email Security Features

- SPF, DKIM (2048-bit RSA signing + validation), DMARC
- MTA-STS, DANE (DNSSEC), TLS-RPT
- Autodiscover (client auto-configuration)
- Layered spam protection: Rspamd + Bayesian + Greylisting + DNSBL + Rate limiting
- Sieve filtering with vacation auto-replies
- Full-text search (PostgreSQL TSVECTOR or SQLite FTS5)

## Development Guidelines

- Pure Go, no CGO dependencies — single binary deploys anywhere
- Hexagonal architecture: all external dependencies behind port interfaces
- PostgreSQL is the default and recommended database
- Frontend builds are embedded into the Go binary for single-artifact deployment
- `internal/infra/` factory wires the correct adapters based on deployment mode
- Graceful degradation: falls back to in-memory if Redis/NATS unavailable

## Testing

- Unit tests: `go test ./...` (race detection enabled)
- Integration tests: `go test -tags=integration ./tests/...`
- Frontend e2e: `cd client && npx playwright test`
- Docker integration: `./scripts/test_integration_docker.sh`

## Documentation

- `docs/architecture/ARCHITECTURE.md` — Design deep-dive
- `docs/guides/PRODUCTION.md` — Production deployment, DNS, firewall
- `docs/guides/CONFIGURATION.md` — Full config reference
- `docs/guides/KUBERNETES.md` — K8s deployment guide
- `docs/guides/TECHNICAL_CONCEPTS.md` — Internals and advanced topics
- `docs/api/API.md` — REST API endpoints
- `docs/api/CLI.md` — CLI command reference
- `docs/guides/WebAdmin.md` — Web dashboard usage
- `docs/development/TESTING.md` — Test setup
- `docs/development/CLIENT_DEVELOPER_HANDOVER.md` — Mobile client guide
