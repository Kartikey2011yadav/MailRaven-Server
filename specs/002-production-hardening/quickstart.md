# Quickstart: Production Layout

## Docker Deployment (Recommended)

### 1. Requirements
- Docker Engine 20.10+
- Docker Compose v2.0+
- DNS Records pointing to your server IP (A/AAAA)

### 2. Prepare Directory
```bash
mkdir -p /opt/mailraven/{data,config}
cd /opt/mailraven
```

### 3. Docker Compose
Create `docker-compose.yml`:

```yaml
services:
  mailraven:
    image: mailraven:latest
    container_name: mailraven
    restart: unless-stopped
    ports:
      - "25:25"   # SMTP
      - "80:80"   # HTTP (ACME Challenges)
      - "443:443" # HTTPS API
    volumes:
      - ./data:/data
      - ./config/config.yaml:/app/config.yaml
    environment:
      - MAILRAVEN_DOMAIN=mail.example.com
```

### 4. Configuration
Create `config/config.yaml`:
```yaml
domain: "mail.example.com"
tls:
  acme:
    enabled: true
    email: "postmaster@mail.example.com"
```

### 5. Launch
```bash
docker-compose up -d
docker-compose logs -f
```

## Update Agent Context
(Internal note for build process)
- This `quickstart.md` replaces the manual binary build instructions for production users.
