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
| **TLS/ACME** | 🟢 Built-in Automatic ACME (Let's Encrypt) | � Built-in Automatic ACME | Implemented via `autocert` (HTTP-01 challenges). |
| **Spam Filter** | 🟢 Bayesian Filtering, Grey-listing | 🟡 DNSBL + Rate Limiting | Connection-level filtering added. Bayesian content filter still missing. |
| **Administration**| 🟢 Web Admin UI | 🔴 None (CLI/File Config) | Mox has a full GUI for domains/accounts. We have basic Admin API endpoints. |
| **Webmail** | 🟢 Integrated | 🔴 None | We provide the API for a frontend to be built. |
| **Deployment** | 🟢 Docker, Docker Compose, Scripts | 🟢 Docker, Docker Compose | Official Dockerfile and Compose setup available. |
| **Testing** | 🟢 Huge suite, specialized test images | 🟡 Unit + E2E integration | Our T104 integration test is a good start, but Mox has `localserve`. |

## Missing Capabilities (Future Scope)

Based on the analysis of `mox`, here is the prioritized list of features we lack:

### 1. Protocol Compatibility & Security (High Priority)
- **IMAP Support**: Required for compatibility with 99% of existing email clients.
- **Advanced Security**: DANE, MTA-STS, and Reporting (TLS-RPT).
- **Spam Filtering**: Bayesian filtering (Rspamd integration).

### 2. Client Ecosystem
- **KMP Mobile Client**: A reference mobile application (Kotlin Multiplatform) consuming our custom API.

### COMPLETED (Moved from Missing)
- ~~**Storage**~~: PostgreSQL support implemented.
- ~~**Administration**~~: Web Admin UI implemented.
- ~~**Docker Support**~~: `Dockerfile` and `docker-compose.yml` implemented.
- ~~**Maintenance Scripts**~~: Backup service and scripts implemented.
- ~~**Automatic TLS**~~: ACME support implemented.
- ~~**Rate Limiting**~~: Basic rate limiting implemented.

## Conclusion

**Mox** is a mature, production-ready replacement for a traditional mail stack (Postfix+Dovecot+Rspamd). It owns the whole vertical.

**MailRaven** is currently a specialized backend for building custom email experiences (like a secure corporate messenger or a specific mobile app backend). To verify "production readiness", we should adopt Mox's **testing strategies** (fuzzing, protocol compliance) and **deployment ease** (Docker), even if we don't implement legacy protocols like IMAP immediately.
