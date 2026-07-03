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

## Distributed Docker Deployment (High Availability)

For production workloads requiring multiple backend instances:

```bash
# Create .env from template
cp .env.example .env
# Edit .env with production secrets

# Start the full distributed stack
docker compose -f docker-compose.distributed.yml up -d
```

This starts:
- **Backend** (2 replicas) — stateless Go instances
- **PostgreSQL** — primary database
- **Redis** — distributed cache, rate limiting, pub/sub
- **NATS** — JetStream message broker for async work
- **MinIO** — S3-compatible blob storage
- **Frontend** — Nginx serving React SPA

Scale backends as needed:
```bash
docker compose -f docker-compose.distributed.yml up -d --scale backend=4
```

## Kubernetes Deployment

Full Kubernetes manifests are provided in `deployment/kubernetes/`.

### Quick Start

```bash
# Edit secrets
vim deployment/kubernetes/base/secret.yaml

# Edit configmap with your domain
vim deployment/kubernetes/base/configmap.yaml

# Apply all resources
kubectl apply -k deployment/kubernetes/
```

### Architecture

- **Backend Deployment**: Stateless pods with readiness/liveness probes
- **StatefulSets**: PostgreSQL, Redis, NATS, MinIO (each with PersistentVolumeClaims)
- **Services**: LoadBalancer for SMTP (port 25), Ingress for HTTP
- **KEDA ScaledObject**: Scale-to-zero when idle, wake on SMTP connections or queue depth

### Scale-to-Zero

Requires [KEDA](https://keda.sh/) installed in your cluster:
```bash
kubectl apply -f https://github.com/kedacore/keda/releases/latest/download/keda.yaml
kubectl apply -f deployment/kubernetes/keda/scaledobject.yaml
```

Pods scale down to 0 after 5 minutes of inactivity and scale back up on:
- Incoming SMTP connections
- Queue depth > 0
- HTTP request rate increase

### Required Environment Variables

Set these in `deployment/kubernetes/base/secret.yaml`:
- `postgres.password` — database password
- `jwt.secret` — JWT signing key (32+ chars)
- `minio.access_key` / `minio.secret_key` — object storage credentials
