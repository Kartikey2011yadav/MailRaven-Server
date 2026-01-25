# Tasks: Production Hardening

**Feature**: Production Hardening (Docker, ACME, Spam, Backup)
**Status**: Pending
**Spec**: [specs/002-production-hardening/spec.md](specs/002-production-hardening/spec.md)
**Plan**: [specs/002-production-hardening/plan.md](specs/002-production-hardening/plan.md)

## Phase 1: Setup & Infrastructure
**Goal**: Initialize the Docker environment and project structure for hardening tools.

- [ ] T001 Create `build` directory and `Dockerfile` using `gcr.io/distroless/static-debian12`
- [ ] T002 Create `docker-entrypoint.sh` for container initialization in `build/docker-entrypoint.sh`
- [ ] T003 Create `docker-compose.yml` defining `mailraven` service with volume mappings
- [ ] T004 Create `scripts` directory and scaffold `scripts/backup.sh` and `scripts/restore.sh`
- [ ] T005 Update `deployment/config.example.yaml` with new TLS, Spam, and Backup sections

## Phase 2: Foundational (Blocking)
**Goal**: Update core configuration structures and define interfaces to support new operational features.

- [ ] T006 Update `internal/core/config/config.go` to include `TLSConfig`, `SpamConfig`, and `BackupConfig` structs
- [ ] T007 Define `SpamFilter` interface in `internal/core/ports/spam.go`
- [ ] T008 Define `BackupService` interface in `internal/core/ports/backup.go`
- [ ] T009 [P] Define `CertificateManager` interface in `internal/core/ports/tls.go`

## Phase 3: User Story 1 - Containerized Deployment
**Goal**: Ensure the application runs correctly inside a container with persistent storage.
**Story**: [US1] Containerized Deployment

- [ ] T010 [US1] Update `Makefile` to include `docker-build` and `docker-run` targets
- [ ] T011 [US1] Implement graceful shutdown handling in `cmd/mailraven/main.go` to support Docker stop signals
- [ ] T012 [P] [US1] verify `docker-compose up` mounts `/data` and `/config` correctly via integration test
- [ ] T013 [US1] Create `E2E_Docker_test.go` in `tests/` to verify container startup and port accessibility

## Phase 4: User Story 2 - Automatic TLS (ACME)
**Goal**: Enable automatic certificate retrieval and renewal via Let's Encrypt.
**Story**: [US2] Automatic TLS Certificate Management
**Config**: `tls.acme.enabled = true`

- [ ] T014 [US2] Create `ACMEService` in `internal/core/services/acme_service.go` using `golang.org/x/crypto/acme/autocert`
- [ ] T015 [US2] Implement `DirCache` storage strategy in `ACMEService` targeting `/data/certs`
- [ ] T016 [US2] Update `internal/adapters/http/server.go` to support dual-port listening (80 for ACME challenges, 443 for API)
- [ ] T017 [US2] Wire `ACMEService` into `NewServer` in `cmd/mailraven/main.go` based on config

## Phase 5: User Story 3 - Spam & Malicious Email Protection
**Goal**: Block known bad actors using DNSBL and rate limits.
**Story**: [US3] Spam and Malicious Email Protection
**Config**: `spam.dnsbls`, `spam.rate_limit`

- [ ] T018 [P] [US3] Implement `DNSBLChecker` adapter in `internal/adapters/spam/dnsbl.go` using `godnsbl`
- [ ] T019 [P] [US3] Implement `RateLimiter` adapter in `internal/adapters/spam/ratelimit.go` using `x/time/rate`
- [ ] T020 [US3] Create `SpamProtectionService` in `internal/core/services/spam_protection.go` aggregating checkers
- [ ] T021 [US3] Implement `SpamMiddleware` in `internal/adapters/smtp/middleware.go` to reject connections early
- [ ] T022 [US3] Wire `SpamProtectionService` into the SMTP listener in `cmd/mailraven/main.go`

## Phase 6: User Story 4 - Production Backup & Recovery
**Goal**: Enable hot backups of the SQLite database and blob storage.
**Story**: [US4] Production Backup and Recovery
**Config**: `backup.location`

- [ ] T023 [P] [US4] Implement `SQLiteBackup` adapter in `internal/adapters/backup/sqlite.go` using `VACUUM INTO` query
- [ ] T024 [P] [US4] Implement `BlobBackup` adapter in `internal/adapters/backup/blob.go` (file copy)
- [ ] T025 [US4] Create `BackupService` orchestration in `internal/core/services/backup_service.go`
- [ ] T026 [US4] Implement Admin API endpoint `POST /admin/backup` in `internal/adapters/http/handlers/admin.go`
- [ ] T027 [US4] Register Admin routes in `internal/adapters/http/server.go` with auth check
- [ ] T028 [US4] Finalize `scripts/backup.sh` to trigger API and `scripts/restore.sh` to apply backup files

## Phase 7: Polish & Documentation
**Goal**: Finalize documentation and verify full system stability.

- [ ] T029 Update `README.md` with Docker Quickstart instructions
- [ ] T030 Document new config options in `docs/CONFIGURATION.md` (create if missing)
- [ ] T031 Run full regression suite `go test ./...`
- [ ] T032 Verify `tasks.md` completion and functionality check

## Dependencies
- Phase 1 & 2 must complete before Phase 3, 4, 5, 6
- Phase 3 (Docker) required for realistic testing of Phase 4 (ACME) and Phase 6 (Backup paths)
- US2, US3, US4 are largely parallelizable after Phase 2

## Parallel Execution Examples
- Developer A implements `ACMEService` (T014-T017)
- Developer B implements `SpamProtection` (T018-T022)
- Developer C implements `BackupService` (T023-T028)
