# Comparison: MailRaven vs. Mox

This document analyzes the differences between our project (**MailRaven**) and the reference implementation (**Mox**).

## Philosophy and Architectue

| Feature | Mox | MailRaven |
|---------|-----|-----------|
| **Core Goal** | "Full-featured modern email server" (Replace Postfix/Dovecot + extras) | "Mobile-first, API-centric email platform" |
| **Architecture** | Monolithic, Protocol-heavy (IMAP/SMTP/Web) | Layered (Ports & Adapters), API-heavy (REST/JSON) |
| **Language** | Go (Custom packages for everything) | Go (Standard Lib + standard patterns) |
| **Storage** | Custom Key-Value / BoltDB variant | SQLite + Filesstem |

## Feature Matrix

| Feature Category | Mox Implementation | MailRaven Status | Notes |
|------------------|--------------------|------------------|-------|
| **Protocols** | 🟢 SMTP, IMAP4, WEBMAIL | 🟡 SMTP, JSON API | We deliberately skip IMAP for the MVP in favor of a custom API. |
| **Security** | 🟢 SPF, DKIM, DMARC, DANE, MTA-STS | 🟡 SPF, DKIM, DMARC | We miss advanced DANE/MTA-STS and Reporting features. |
| **TLS/ACME** | 🟢 Built-in Automatic ACME (Let's Encrypt) | 🔴 Manual Config | We rely on external cert management or reverse proxies. |
| **Spam Filter** | 🟢 Bayesian Filtering, Grey-listing | 🔴 None / Basic SPF rejection | **Major Gap**: We accept all underlying valid email. |
| **Administration**| 🟢 Web Admin UI | 🔴 None (CLI/File Config) | Mox has a full GUI for domains/accounts. |
| **Webmail** | 🟢 Integrated | 🔴 None | We provide the API for a frontend to be built. |
| **Deployment** | 🟢 Docker, Docker Compose, Scripts | 🟠 Systemd only | We miss containerization. |
| **Testing** | 🟢 Huge suite, specialized test images | 🟡 Unit + E2E integration | Our T104 integration test is a good start, but Mox has `localserve`. |

## Missing Capabilities (Future Scope)

Based on the analysis of `mox`, here is the prioritized list of features we lack:

### 1. Operational Tooling (High Priority)
- **Docker Support**: We lack `Dockerfile` and `docker-compose.yml`. Mox has split images for testing (`Dockerfile.imaptest`) and release.
- **Maintenance Scripts**: Mox includes `apidiff.sh` (API compatibility), `genlicenses.sh` (Dependency compliance), and `backup.go` (Hot backups). We need a backup strategy.

### 2. Security Enhancements
- **Automatic TLS**: Implementing ACME (via `golang.org/x/crypto/acme/autocert`) would simplify setup significantly.
- **Spam Filtering**: Implement a Junk filter. We can study `mox/junk` and `mox/dnsbl` for inspiration.
- **Rate Limiting**: Mox has `ratelimit/`. We likely have basic or no rate limiting on the API/SMTP, making us vulnerable to DOS.

### 3. Feature Gaps
- **IMAP Support**: While not our focus, lacking IMAP makes us incompatible with 99% of existing email clients. Standard adoption requires this.
- **Web Admin**: Managing users via SQL or Config files is error-prone. A simple `/admin` React app talking to our API would be beneficial.
- **Observability**: We have `/metrics`, but Mox has structured logging (`mlog`) deep integration.

## Conclusion

**Mox** is a mature, production-ready replacement for a traditional mail stack (Postfix+Dovecot+Rspamd). It owns the whole vertical.

**MailRaven** is currently a specialized backend for building custom email experiences (like a secure corporate messenger or a specific mobile app backend). To verify "production readiness", we should adopt Mox's **testing strategies** (fuzzing, protocol compliance) and **deployment ease** (Docker), even if we don't implement legacy protocols like IMAP immediately.
