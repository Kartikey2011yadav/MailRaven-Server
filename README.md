# MailRaven Server

A modern, modular email server built with mobile-first architecture. MailRaven implements clean architecture (Ports and Adapters) with a layered design optimized for mobile email clients.

## Features

- **Mobile-First API**: RESTful JSON API with pagination, compression, and delta sync
- **Web Admin UI**: React-based dashboard for managing domains, users, and system stats
- **Reliable Email Reception**: SMTP server with SPF/DKIM/DMARC validation
- **Production Ready**: Docker support, Postgres or SQLite backend, Automatic HTTPS, and Hot Backups
- **Spam Protection**: Rspamd integration, DNSBL checking, and Connection Rate Limiting
- **IMAP4rev1 Support**: Standard IMAP listener for desktop/mobile client compatibility
- **Full-Text Search**: SQLite FTS5 or Postgres TSVECTOR for fast message search
- **Zero Data Loss**: Atomic writes with fsync before SMTP acknowledgment
- **CGO-Free**: Pure Go implementation for simple deployment

## Development

### Prerequisites
- Go 1.22+
- Node.js 20+
- Docker & Docker Compose
- Make (optional)

### Quick Start (Docker)
```bash
# Start development environment (hot-reload backend & frontend)
make docker-dev

# Run integration tests in container
make test-docker

# Build production images
make docker-build

# Start production stack
make docker-up
```

### Manual Setup
See [docs/CONFIGURATION.md](docs/CONFIGURATION.md) for detailed configuration options.

## Architecture

MailRaven follows the Ports and Adapters (Hexagonal) pattern with 5 distinct layers:

```
┌─────────────────────────────────────────────────┐
│  Listener Layer (SMTP Protocol)                │
├─────────────────────────────────────────────────┤
│  Logic Layer (SPF/DKIM/DMARC, Routing)         │
├─────────────────────────────────────────────────┤
│  Storage Layer (Repository Interfaces)          │
├─────────────────────────────────────────────────┤
│  Search Layer (FTS5 / TSVECTOR)                 │
├─────────────────────────────────────────────────┤
│  API Layer (REST/JSON for Mobile Clients)      │
└─────────────────────────────────────────────────┘
```

**Current Implementation**:
- **Listener**: SMTP (RFC 5321), IMAP4rev1 (RFC 3501)
- **Storage**: SQLite (default) or PostgreSQL
- **Search**: SQLite FTS5 or Postgres TSVECTOR with BM25 ranking
- **API**: REST/JSON with JWT authentication
- **Frontend**: React + Vite (Web Admin)

**Designed for Future Migration**:
- Listener: IMAP/POP3 support
- Storage: S3 (Object Storage)
- Search: Elasticsearch (Optional for massive scale)

## Quick Start

### Prerequisites

- Go 1.22 or higher
- Linux server (Ubuntu 20.04+, Debian 11+, CentOS 8+, or Windows 10/11)
- Domain with DNS access (for production email)

### Installation

#### Option 1: Development Setup (Recommended)

MailRaven includes cross-platform setup scripts that build the Backend (Go) and Frontend (React).

**Windows (PowerShell):**
```powershell
.\scripts\setup.ps1
```

**Linux/macOS (Bash):**
```bash
./scripts/setup.sh
```

#### Option 2: Build from Source

```bash
# Clone the repository
git clone https://github.com/Kartikey2011yadav/mailraven-server.git
cd mailraven-server

# Build the binary
go build -o mailraven ./cmd/mailraven

# (Optional) Install to system path
sudo cp mailraven /usr/local/bin/
```

#### Option 3: Docker Information

We provide a multi-stage `Dockerfile` and `docker-compose.yml` for easy deployment.

```bash
# Start Backend, Frontend, and PostgreSQL
docker-compose up -d
```

### Configuration

MailRaven uses a `config.yaml` file. By default, it uses SQLite. To switch to PostgreSQL:

```yaml
storage:
  driver: "postgres"
  dsn: "postgres://user:pass@localhost:5432/mailraven?sslmode=disable"
```

### Run Quickstart Setup

```bash
# Linux/macOS
sudo mailraven quickstart

# Windows (run as Administrator)
mailraven.exe quickstart
```

This interactive command will:
1. Generate DKIM keys (2048-bit RSA)
2. Create configuration file with crypto-secure JWT secret
3. Display DNS records to configure (MX, SPF, DKIM, DMARC)
4. Create initial admin user account
```

For detailed setup instructions, see [Production Guide](docs/PRODUCTION.md).

### Start Server

```bash
# Linux/macOS with default config
sudo mailraven serve

# Custom config path
mailraven serve --config /path/to/config.yaml

# Windows (run as Administrator)
mailraven.exe serve
```

The server will start:
- **SMTP Server**: Port 2525 (or 25 in production)
- **HTTP API**: Port 8443 with TLS (or 8080 without TLS)

### Test Email Reception

```bash
# Using swaks (SMTP test tool)
swaks --to user@yourdomain.com \
  --from sender@example.com \
  --server localhost:2525 \
  --body "Test message from MailRaven"

# Using mail command
echo "Test message" | mail -s "Test" user@yourdomain.com
```

### Test Mobile API

```bash
# 1. Login to get JWT token
curl -X POST https://localhost:8443/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@yourdomain.com","password":"your_password"}' \
  --insecure

# Response: {"token":"eyJhbGc...","expires_at":"2026-01-30T..."}

# 2. List messages
curl -X GET https://localhost:8443/api/v1/messages \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  --insecure

# 3. Get specific message
curl -X GET https://localhost:8443/api/v1/messages/msg-123 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  --insecure

# 4. Mark message as read
curl -X PATCH https://localhost:8443/api/v1/messages/msg-123 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"read_state":true}' \
  --insecure

# 5. Search messages
curl -X GET 'https://localhost:8443/api/v1/messages/search?q=invoice' \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  --insecure
```

### Deployment

#### Systemd Service (Linux)

```bash
# Copy service file
sudo cp deployment/mailraven.service /etc/systemd/system/

# Create mailraven user
sudo useradd -r -s /bin/false mailraven

# Create directories
sudo mkdir -p /var/lib/mailraven /var/log/mailraven
sudo chown mailraven:mailraven /var/lib/mailraven /var/log/mailraven

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable mailraven
sudo systemctl start mailraven

# Check status
sudo systemctl status mailraven
sudo journalctl -u mailraven -f
```

## Development

### Run Tests

```bash
make test
```

### Run Integration Tests

```bash
make test-integration
```

### Run Linters

```bash
make lint
```

### Generate Coverage Report

```bash
make coverage
```

## Project Structure

```
cmd/mailraven/          # Binary entrypoint
internal/
  core/
    domain/             # Domain entities (Message, User)
    ports/              # Interface definitions
    services/           # Business logic
  adapters/
    smtp/               # SMTP protocol handling
    http/               # REST API server
    storage/
      sqlite/           # SQLite repository implementation
      disk/             # File system blob storage
  config/               # Configuration management
  observability/        # Logging and metrics
tests/                  # Integration tests
deployment/             # Deployment artifacts
```

## Configuration

Configuration is stored in `config.yaml`. Example:

```yaml
domain: mail.example.com
smtp:
  port: 25
  hostname: mail.example.com
api:
  port: 8080
  jwt_secret: "generate_random_secret"
storage:
  db_path: /var/lib/mailraven/mailraven.db
  blob_path: /var/lib/mailraven/blobs
dkim:
  selector: default
  private_key_path: /etc/mailraven/dkim.key
```

## Documentation

- [Feature Specification](docs/ARCHITECTURE.md) - High level design.
- [Production Guide](docs/PRODUCTION.md) - Postgres, Docker, and deployment.
- [Web Admin Guide](docs/WebAdmin.md) - Using the dashboard.
- [Mobile API Spec](docs/API.md) - Endpoints for client developers.
- [Architecture](docs/ARCHITECTURE.md) - Deep dive into internal design.
- [Configuration](docs/CONFIGURATION.md) - Full config reference.
- [CLI Reference](docs/CLI.md) - Command line tools.
- [Testing](docs/TESTING.md) - How to run tests.

## Constitution

MailRaven follows a strict development constitution focusing on:

1. **Reliability**: "250 OK" means data is fsync'd to disk
2. **Protocol Parity**: Strict adherence to email RFCs
3. **Mobile-First**: Optimized for bandwidth and latency
4. **Dependency Minimalism**: CGO-free, minimal external dependencies
5. **Observability**: Comprehensive logging and metrics
6. **Interface-Driven Design**: Clean boundaries between layers
7. **Extensibility**: Middleware pattern for custom processing
8. **Protocol Isolation**: Independent protocol implementations

See [.specify/memory/constitution.md](.specify/memory/constitution.md) for details.

## Performance Goals

- **SMTP**: 100 concurrent connections, <50ms processing time
- **API**: <500ms P95 latency over 4G networks
- **Search**: <200ms query time for 10k message mailboxes
- **Memory**: <100MB for 100 concurrent mobile clients

## License

Dual-licensed under:
- **Mozilla Public License 2.0** (MPLv2.0) - See [LICENSE.MPLv2.0](LICENSE.MPLv2.0)
- **MIT License** - See [LICENSE.MIT](LICENSE.MIT)

You may use, distribute, and modify this code under the terms of either license.

## Contributing

Contributions are welcome! Please:
1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes following the project's code style
4. Write tests for new functionality
5. Ensure all tests pass (`go test ./...`)
6. Commit your changes (`git commit -m 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

### Development Guidelines

- Follow the [Constitution](.specify/memory/constitution.md) principles
- Write tests for all new features
- Document public APIs with godoc comments
- Keep dependencies minimal (CGO-free preferred)
- Use semantic commit messages

## Documentation

- [Production Guide](docs/PRODUCTION.md) - Postgres, Docker, and deployment.
- [Web Admin Guide](docs/WebAdmin.md) - Using the dashboard.
- [Mobile API Spec](docs/API.md) - Endpoints for client developers.
- [Architecture](docs/ARCHITECTURE.md) - Deep dive into internal design.

## Verification

Run the environment check script to ensure prerequisites are met:

```powershell
.\scripts\check.ps1  # Windows
./scripts/check.sh   # Linux/macOS
```

## Acknowledgments

- Inspired by [Mox](https://github.com/mjl-/mox) email server
- Uses modernc.org/sqlite for CGO-free SQLite
- Built with go-chi/chi v5 HTTP router
