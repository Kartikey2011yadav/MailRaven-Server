# Task Breakdown: Production Readiness

## Phase 2: Implementation

### 1. Backend Logging Upgrade
- [x] Create `internal/observability/console_handler.go`.
- [x] Implement `Handle` method with ANSI color logic.
- [x] Update `cmd/mailraven/serve.go` to use the new handler in "Dev" mode.
- [x] Verify logs using `go run`.

### 2. PostgreSQL Implementation
- [x] Add `github.com/jackc/pgx/v5` dependency.
- [x] Create `internal/adapters/storage/postgres/connection.go`.
- [x] Implement `UserRepository`, `DomainRepository` in `postgres` package.
- [x] Create SQL migrations in `internal/adapters/storage/postgres/migrations/`.
- [x] Update `internal/config/config.go` to support `Storage.Driver` ("sqlite" | "postgres").
- [x] Update `cmd/mailraven/serve.go` to switch factory based on config.

### 3. Setup Scripts (Cross-Platform)
- [x] Create `scripts/setup.sh` (Bash) with:
    - OS detection (Linux/Mac).
    - Go/Node checks.
    - Build commands.
- [x] Create `scripts/setup.ps1` (PowerShell) for Windows.
- [x] Create `scripts/check.sh` and `scripts/check.ps1` for environment validation.
- [x] Verify scripts on local environment.

### 4. Docker Improvements
- [x] Refactor `docker-compose.yml`:
    - Service: `backend` (Go)
    - Service: `frontend` (Nginx/Vite preview)
    - Service: `db` (Postgres - optional profile)
- [x] Update `Dockerfile` to be multi-stage (Build Go + Build React -> Distroless image).

### 5. Documentation & Analysis
- [x] Create `docs/PRODUCTION.md` (Cleanup `quickstart.md` content).
- [x] Perform Gap Analysis vs Mox and write `docs/MOX_GAP_ANALYSIS.md`.
- [x] Write `docs/BACKUP_ANALYSIS_PROMPT.md` for the user.

### 6. Verification
- [ ] Run `check.sh`.
- [ ] Test Postgres connection.
- [ ] Verify Docker Compose startup.
