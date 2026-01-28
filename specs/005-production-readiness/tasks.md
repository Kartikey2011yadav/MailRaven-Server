# Task Breakdown: Production Readiness

## Phase 2: Implementation

### 1. Backend Logging Upgrade
- [ ] Create `internal/observability/console_handler.go`.
- [ ] Implement `Handle` method with ANSI color logic.
- [ ] Update `cmd/mailraven/serve.go` to use the new handler in "Dev" mode.
- [ ] Verify logs using `go run`.

### 2. PostgreSQL Implementation
- [ ] Add `github.com/jackc/pgx/v5` dependency.
- [ ] Create `internal/adapters/storage/postgres/connection.go`.
- [ ] Implement `UserRepository`, `DomainRepository` in `postgres` package.
- [ ] Create SQL migrations in `internal/adapters/storage/postgres/migrations/`.
- [ ] Update `internal/config/config.go` to support `Storage.Driver` ("sqlite" | "postgres").
- [ ] Update `cmd/mailraven/serve.go` to switch factory based on config.

### 3. Setup Scripts (Cross-Platform)
- [ ] Create `scripts/setup.sh` (Bash) with:
    - OS detection (Linux/Mac).
    - Go/Node checks.
    - Build commands.
- [ ] Create `scripts/setup.ps1` (PowerShell) for Windows.
- [ ] Create `scripts/check.sh` and `scripts/check.ps1` for environment validation.
- [ ] Verify scripts on local environment.

### 4. Docker Improvements
- [ ] Refactor `docker-compose.yml`:
    - Service: `backend` (Go)
    - Service: `frontend` (Nginx/Vite preview)
    - Service: `db` (Postgres - optional profile)
- [ ] Update `Dockerfile` to be multi-stage (Build Go + Build React -> Distroless image).

### 5. Documentation & Analysis
- [ ] Create `docs/PRODUCTION.md` (Cleanup `quickstart.md` content).
- [ ] Perform Gap Analysis vs Mox and write `docs/MOX_GAP_ANALYSIS.md`.
- [ ] Write `docs/BACKUP_ANALYSIS_PROMPT.md` for the user.

### 6. Verification
- [ ] Run `check.sh`.
- [ ] Test Postgres connection.
- [ ] Verify Docker Compose startup.
