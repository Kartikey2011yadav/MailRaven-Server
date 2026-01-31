# Gap Analysis: MailRaven vs Mox

This document outlines the functional and architectural differences between MailRaven and [Mox](https://github.com/mjl-/mox), the reference implementation and inspiration.

## Architecture

| Feature | MailRaven | Mox | Gap / Diff |
|---------|-----------|-----|------------|
| **Core Protocol** | SMTP + IMAP + Custom REST API | SMTP + IMAP4 + JMAP | **Major**: MailRaven adds IMAP support while retaining its modern HTTP API focus. |
| **Storage** | SQLite OR PostgreSQL | SQLite (bbolt/sqlite) | **Extension**: MailRaven supports PostgreSQL for scalability. |
| **Frontend** | React (SPA) | Go Templates / Embedded JS | **Modernization**: Separated frontend allows richer UI/UX. |
| **Language** | Go (Backend) + TS (Frontend) | Go (Monolith) | **Complexity**: Increased complexity for flexibility. |

## Feature Set

| Feature | MailRaven status | Mox status | Notes |
|---------|------------------|------------|-------|
| **SMTP (Inbound)** | ✅ Implemented | ✅ Implemented | Functionally equivalent. |
| **SMTP (Outbound)** | ✅ Implemented (Queued) | ✅ Implemented | MailRaven uses `SKIP LOCKED` for Postgres queue. |
| **IMAP4** | ✅ Implemented | ✅ Full Support | Core RFC 3501 + IDLE supported. |
| **JMAP** | ❌ Custom API | ✅ Full Support | MailRaven uses a simplified REST API. |
| **DKIM/SPF/DMARC** | ✅ Implemented | ✅ Implemented | Parity achieved. |
| **MTA-STS** | ✅ Implemented | ✅ Implemented | Parity achieved (Receive). |
| **TLS-RPT** | ✅ Implemented | ✅ Implemented | Parity achieved (Receive). |
| **DANE** | ✅ Implemented | ✅ Implemented | Parity achieved (Outbound Verification). |
| **Spam Filtering** | ✅ Advanced | ✅ Advanced | Native Bayesian + DNSBL + Greylisting. |
| **Sieve Filtering**| ✅ Implemented| ✅ Implemented | RFC 5228 + ManageSieve RFC 5804 + Vacation. |
| **Webmail** | ✅ React Client (Admin) | ✅ Built-in | MailRaven Admin UI is active. |
| **Multi-Domain** | ✅ Supported | ✅ Supported | Parity achieved. |
| **Account Mgmt** | ✅ Web Admin | ✅ Web Admin | Parity achieved. |

## Conclusion

MailRaven is a specialized evolution focusing on **API-first** interaction for mobile clients, while now supporting legacy protocols (IMAP) for compatibility. While Mox is a complete drop-in replacement for traditional mail servers (Postfix/Dovecot), MailRaven bridges the gap between modern application platforms and standard email clients.
