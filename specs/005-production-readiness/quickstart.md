# Quickstart / Production Guide

## Introduction
MailRaven is designed to be a "install and forget" mail server. This guide covers deploying MailRaven on a Linux VPS (Ubuntu/Debian) for production use.

## Prerequisites
1. **VPS**: 1 CPU, 512MB RAM minimum (Standard drops: DigitalOcean, Hetzner, Linode).
2. **Domain**: You must own a domain name (e.g., `example.com`).
3. **Ports**: Ensure your firewall allows: `25` (SMTP), `80/443` (HTTP/HTTPS), `587` (Submission).

## Step 1: DNS Setup
Before installing, configure your DNS records:
- **A Record**: `mail.example.com` -> `[Your VPS IP]`
- **MX Record**: `example.com` -> `mail.example.com` (Priority 10)
- **SPF (TXT)**: `example.com` -> `v=spf1 mx -all`

## Step 2: Installation

```bash
# Clone the repository
git clone https://github.com/Kartikey2011yadav/MailRaven-Server.git /opt/mailraven
cd /opt/mailraven

# Run the Setup Script
# This will install dependencies, build the binary, and generate config.
sudo ./scripts/setup.sh --prod --db=sqlite
```

## Step 3: HTTPS & TLS
MailRaven includes AutoTLS (via Let's Encrypt).
1. Ensure port 80 is open.
2. Edit `config.yaml`:
   ```yaml
   tls:
     enabled: true
     provider: "letsencrypt"
     domain: "mail.example.com"
   ```

## Step 4: Running as a Service
The setup script generates a Systemd unit file.
```bash
sudo cp scripts/mailraven.service /etc/systemd/system/
sudo systemctl enable mailraven
sudo systemctl start mailraven
```

## Step 5: Post-Install Verification
Run the check script to verify everything is healthy.
```bash
./scripts/check.sh
# Output: [PASS] All systems go.
```

Access the Admin UI at `https://mail.example.com/admin`.

## Database Options
MailRaven supports SQLite (default) and PostgreSQL.
To use PostgreSQL:
1. Provision a Postgres DB.
2. Update `config.yaml`:
   ```yaml
   storage:
     driver: "postgres"
     dsn: "postgres://user:pass@localhost:5432/mailraven?sslmode=disable"
   ```
