# Deployment Guide

**Current Status**: MailRaven supports deployment via **Docker**, **SystemdService**, or standalone binary.

## Prerequisites

- **OS**: Linux (amd64/arm64) or Windows (amd64).
- **DNS**: properly configured MX records, SPF, DKIM, DMARC, and MTA-STS/DANE records as detailed in `PRODUCTION.md`.

## Automated Setup (Recommended)

MailRaven provides cross-platform setup scripts that build the backend, frontend, and prepare for deployment.

### Windows (PowerShell)
```powershell
.\scripts\setup.ps1
```

### Linux/macOS (Bash)
```bash
./scripts/setup.sh
```

## Docker Deployment

MailRaven includes specific Docker support with `docker-compose.yml`.

1. **Build and Run**:
   ```bash
   docker-compose up -d
   ```
   This will start:
   - MailRaven Backend
   - MailRaven Frontend (Nginx/Vite Preview)
   - PostgreSQL (if configured)

2. **Configuration**: 
   Ensure `config.yaml` is mounted or configured correctly in `docker-compose.yml`.

## Manual Deployment Steps

### 1. Building the Binary

Navigate to the project root and run:

```bash
go build -o mailraven cmd/mailraven/main.go
# Or if using Makefile
make build
```

### 2. Configuration

1. Create a directory for the application: `/opt/mailraven`
2. Copy the binary to `/opt/mailraven/bin/`
3. Copy `deployment/config.example.yaml` to `/etc/mailraven/config.yaml`
4. Edit `config.yaml`:
   - Set valid DB paths.
   - Configure TLS certificates (Paths to .pem and .key files).
   - Set Domain name.

### 3. Setting up Systemd (Linux)

Use the provided service file in `deployment/mailraven.service`.

1. Copy to systemd:
   ```bash
   cp deployment/mailraven.service /etc/systemd/system/
   ```
2. Reload daemon:
   ```bash
   systemctl daemon-reload
   ```
3. Enable and Start:
   ```bash
   systemctl enable mailraven
   systemctl start mailraven
   ```

### 4. Verification

Check logs:
```bash
journalctl -u mailraven -f
```

Check metrics:
```bash
curl http://localhost:8080/metrics
```

## Future Deployment Options

We plan to adopt the `mox` approach to containerization:
- **Dockerfile**: A multi-stage build for a minimal image.
- **Docker Compose**: Orchestration for external dependencies (though we currently use embedded SQLite, so specific DB containers aren't needed, but maybe for a separate monitoring stack like Prometheus/Grafana).
- **Scripts**: Automated release scripts (`docker-release.sh`).
