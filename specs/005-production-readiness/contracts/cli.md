# CLI Contracts

## Setup Script

**Name**: `setup.sh` (Linux/Mac), `setup.ps1` (Windows)
**Purpose**: Automated environment setup and installation.

**Usage**:
```bash
./scripts/setup.sh [options]

Options:
  --prod        Optimizes for production (minimal logs, secure defaults)
  --dev         Development setup (installs air, sets up debug logging)
  --db=[sqlite|postgres]  Selects database backend (default: sqlite)
```

**Behavior**:
1. Check Dependencies:
   - `go` >= 1.21
   - `node` >= 18 (for frontend build)
   - `docker` (optional, warns if missing)
2. Build:
   - `go build -o bin/mailraven cmd/mailraven/main.go`
   - `cd client && npm install && npm run build`
3. Configuration:
   - Checks for `config.yaml`.
   - If missing, interactive prompt (Domain, Admin Email).
4. Database:
   - If SQLite: Ensures `data/` dir exists.
   - If Postgres: Prompts for `DSN` (Connection string) and updates `config.yaml`.

---

## Check Script

**Name**: `check.sh` (Linux/Mac), `check.ps1` (Windows)
**Purpose**: Pre-flight validation.

**Usage**:
```bash
./scripts/check.sh
```

**Checks**:
1. **Ports**: Verifies availability of configured ports (e.g., 25, 8080/8081).
2. **DNS**: Checks `hostname` resolution.
3. **Permissions**: Checks write access to `data/` and `logs/`.
4. **Database**: Attempts connection to configured DB.
5. **Certificates**: Checks if TLS certs exist and are valid (not expired).

**Output**:
- Returns `exit 0` if all PASS.
- Returns `exit 1` if Critical FAIL.
