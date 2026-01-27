# Implementation Plan: Production Hardening

**Branch**: `002-production-hardening` | **Date**: 2026-01-25 | **Spec**: [specs/002-production-hardening/spec.md](specs/002-production-hardening/spec.md)
**Input**: Feature specification from `specs/002-production-hardening/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

This feature aims to bring MailRaven to production readiness by implementing critical DevOps and security infrastructure. Key deliverables include a containerized deployment pipeline (Docker), automated TLS certificate management (ACME/Let's Encrypt), basic spam protection (DNSBL, Rate Limiting), and operational scripts for backup/recovery. This aligns MailRaven's operational capabilities with the `mox` reference implementation.

## Technical Context

**Language/Version**: Go 1.22
**Primary Dependencies**: 
- `golang.org/x/crypto/acme/autocert` (TLS)
- `golang.org/x/time/rate` (Rate Limiting)
- SQLite (modernc.org/sqlite - existing)
**Storage**: SQLite + Filesystem (persistence required via Docker Volumes)
**Testing**: `go test` for logic, `docker-compose` for E2E integration, `api_test.go` refactoring.
**Target Platform**: Linux containers (Docker/Podman), standard Linux VMs.
**Project Type**: Backend Server (Go) + DevOps Artifacts
**Performance Goals**: Docker image < 50MB, ACME renewal non-blocking, Spam checks < 500ms.
**Constraints**: Zero-downtime reloads (nice to have), robust data persistence across container updates.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Verify compliance with [MailRaven Constitution](../../.specify/memory/constitution.md):

- [x] **Reliability**: Docker volumes mapped correctly to ensure atomic writes (SQLite+FS) persist. Backup script uses safe copy mechanisms (sqlite backup API).
- [x] **Protocol Parity**: ACME is standard. Spam protection (DNSBL) follows RFC usage.
- [x] **Mobile-First Architecture**: API remains primary. DevOps improvements don't regress API payload sizes.
- [x] **Dependency Minimalism**: Using standard `autocert` and `rate` packages. No heavy external spam filter (e.g. SpamAssassin) embedded yet—starting with efficient Go-native checks.
- [x] **Observability**: New spam rejection metrics and ACME status logs to be added.
- [x] **Interface-Driven Design**: `SpamFilter` and `CertificateManager` will be defined as interfaces.
- [x] **Extensibility**: Spam checks implemented as middleware/chainable handlers.
- [x] **Protocol Isolation**: Operational logic (ACME, Backup) is separate from Email Protocol handling.
- [x] **Testing Standards**: Integration tests will verify Docker container startup and volume persistence.

**Violations Requiring Justification**: None.

## Project Structure

### Documentation (this feature)

```text
specs/002-production-hardening/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
└── tasks.md             # Phase 2 output
```

### Source Code (repository root)

```text
src/ (Repository Root)
├── cmd/
│   └── mailraven/
├── internal/
│   ├── core/
│   │   ├── ports/              # New interfaces: SpamFilter, BackupService
│   │   └── services/           # New services: ACMEService, SpamProtection
│   ├── adapters/
│   │   ├── spam/               # DNSBL, RateLimit implementation
│   │   └── backup/             # Backup implementation
├── build/                      # New directory for Docker context
│   ├── Dockerfile
│   └── docker-entrypoint.sh
├── scripts/                    # New operational scripts
│   ├── backup.sh
│   └── restore.sh
└── docker-compose.yml          # Root level
```
