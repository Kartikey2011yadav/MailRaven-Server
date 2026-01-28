# Gap Analysis: MailRaven vs Mox

This document outlines the functional and architectural differences between MailRaven and [Mox](https://github.com/mjl-/mox), the reference implementation and inspiration.

## Architecture

| Feature | MailRaven | Mox | Gap / Diff |
|---------|-----------|-----|------------|
| **Core Protocol** | SMTP + Custom REST API | SMTP + IMAP4 + JMAP | **Major**: MailRaven removes IMAP in favor of a modern HTTP API for mobile clients. |
| **Storage** | SQLite OR PostgreSQL | SQLite (bbolt/sqlite) | **Extension**: MailRaven supports PostgreSQL for scalability. |
| **Frontend** | React (SPA) | Go Templates / Embedded JS | **Modernization**: Separated frontend allows richer UI/UX. |
| **Language** | Go (Backend) + TS (Frontend) | Go (Monolith) | **Complexity**: Increased complexity for flexibility. |

## Feature Set

| Feature | MailRaven status | Mox status | Notes |
|---------|------------------|------------|-------|
| **SMTP (Inbound)** | ✅ Implemented | ✅ Implemented | Functionally equivalent. |
| **SMTP (Outbound)** | ✅ Implemented (Queued) | ✅ Implemented | MailRaven uses `SKIP LOCKED` for Postgres queue. |
| **IMAP4** | ❌ Not Planned | ✅ Full Support | Intentional omission (Mobile-first). |
| **JMAP** | ❌ Custom API | ✅ Full Support | MailRaven uses a simplified REST API. |
| **DKIM/SPF/DMARC** | ✅ Implemented | ✅ Implemented | Parity achieved. |
| **MTA-STS** | ❌ Missing | ✅ Implemented | Future roadmap item. |
| **DANE** | ❌ Missing | ✅ Implemented | Future roadmap item. |
| **Webmail** | ✅ React Client | ✅ Built-in | MailRaven offers a dedicated SPA. |
| **Multi-Domain** | ✅ Supported | ✅ Supported | Parity achieved. |
| **Account Mgmt** | ✅ Admin API | ✅ Web Admin | Parity achieved. |

## Conclusion

MailRaven is a specialized fork/evolution focusing on **API-first** interaction for mobile clients, deliberately sacrificing legacy protocol support (IMAP) for a streamlined architecture. While Mox is a complete drop-in replacement for traditional mail servers (Postfix/Dovecot), MailRaven is an application platform for modern email apps.
