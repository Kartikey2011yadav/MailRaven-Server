# MailRaven Server

A modern, modular email server built with mobile-first architecture. MailRaven implements clean architecture (Ports and Adapters) with a layered design optimized for mobile email clients.

## Features

- **Mobile-First API**: RESTful JSON API with pagination, compression, and delta sync
- **Reliable Email Reception**: SMTP server with SPF/DKIM/DMARC validation
- **Full-Text Search**: SQLite FTS5 for fast message search on mobile devices
- **Zero Data Loss**: Atomic writes with fsync before SMTP acknowledgment
- **CGO-Free**: Pure Go implementation for simple deployment

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
│  Search Layer (FTS5)                            │
├─────────────────────────────────────────────────┤
│  API Layer (REST/JSON for Mobile Clients)      │
└─────────────────────────────────────────────────┘
```

**Current Implementation**:
- **Listener**: SMTP (RFC 5321)
- **Storage**: SQLite + File system with gzip compression
- **Search**: SQLite FTS5 with BM25 ranking
- **API**: REST/JSON with JWT authentication

**Designed for Future Migration**:
- Listener: IMAP/POP3 support
- Storage: PostgreSQL + S3
- Search: Elasticsearch

## Quick Start

### Prerequisites

- Go 1.22 or higher
- Linux server (Ubuntu 20.04+)
- Domain with DNS access

### Build

```bash
make build
```

### Run Quickstart Setup

```bash
sudo ./bin/mailraven quickstart
```

This interactive command will:
1. Generate DKIM keys
2. Configure DNS records (MX, SPF, DKIM, DMARC)
3. Set up TLS certificates
4. Create initial configuration

For detailed setup instructions, see [specs/001-mobile-email-server/quickstart.md](specs/001-mobile-email-server/quickstart.md).

### Start Server

```bash
sudo ./bin/mailraven serve
```

### Test Email Reception

```bash
echo "Test message" | mail -s "Test" user@yourdomain.com
```

### Test Mobile API

```bash
# Login
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@yourdomain.com","password":"your_password"}'

# List messages
curl -X GET http://localhost:8080/messages \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
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

- [Feature Specification](specs/001-mobile-email-server/spec.md)
- [Implementation Plan](specs/001-mobile-email-server/plan.md)
- [Architecture Research](specs/001-mobile-email-server/research.md)
- [Data Model](specs/001-mobile-email-server/data-model.md)
- [REST API Contract](specs/001-mobile-email-server/contracts/rest-api.yaml)
- [Quickstart Guide](specs/001-mobile-email-server/quickstart.md)

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

[License information to be added]

## Contributing

[Contributing guidelines to be added]

## Acknowledgments

- Inspired by [Mox](https://github.com/mjl-/mox) email server
- Uses modernc.org/sqlite for CGO-free SQLite
- Built with go-chi/chi v5 HTTP router
