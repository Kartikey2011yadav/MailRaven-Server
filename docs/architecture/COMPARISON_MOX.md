# Comparison: MailRaven vs. Mox

This document analyzes the differences between our project (**MailRaven**) and the reference implementation (**Mox**).

## Philosophy and Architecture

| Feature | Mox | MailRaven |
|---------|-----|-----------|
| **Core Goal** | "Full-featured modern email server" (Replace Postfix/Dovecot + extras) | "Mobile-first, API-centric email platform" |
| **Architecture** | Monolithic, Protocol-heavy (IMAP/SMTP/Web) | Layered (Ports & Adapters), API-heavy (REST/JSON) |
| **Language** | Go (Custom packages for everything) | Go (Standard Lib + standard patterns) |
| **Storage** | Custom Key-Value / BoltDB variant | SQLite / Postgres (ORM-based) |

## Feature Matrix

| Feature Category | Mox Implementation | MailRaven Status | Notes |
|------------------|--------------------|------------------|-------|
| **SMTP Delivery**|  Full (Inbound/Outbound) |  Full (Inbound/Outbound) | Both support SPF/DKIM/DMARC checks. |
| **IMAP4**        |  Full (RFC 3501 + Extensions) |  Core RFC 3501 Compliance | Supports LOGIN, LIST, SELECT, FETCH, UID, STORE. Compatible with Standard Clients. |
| **Mobile Push**  |  IMAP IDLE / Notifications |  IMAP IDLE Supported | Real-time notifications for standard clients. |
| **Autodiscover** |  SRV, XML, Apple Profiles |  XML (MS/Mozilla) | Supports Outlook and Thunderbird auto-config. |
| Security | SPF, DKIM, DMARC, DANE, MTA-STS | SPF, DKIM, DMARC, DANE, MTA-STS, TLS-RPT | **Gap Closed**. We now support MTA-STS (Receive), TLS-RPT (Receive), and DANE (Send). |
| TLS/ACME         | Built-in Automatic ACME | Built-in Automatic ACME | Implemented via `autocert`. |
| Spam Filter      | Bayesian, Grey-listing | Bayesian, Grey-listing, DNSBL | **Gap Closed**. Added Native Bayesian filter (with IMAP feedback loop) and Greylisting. |
| **Sieve Scripts**| Full (RFC 5228)    | Full (RFC 5228)  | **Gap Closed**. Added Sieve engine RFC 5228 + ManageSieve RFC 5804 + Vacation RFC 5230. |
| **Quotas**       | Full (RFC 2087)    | None             | No enforcement of storage limits per user. |
| Administration   | Web Admin UI       | Web Admin API + UI | **Gap Closed**. MailRaven now includes a Unified "User Portal" (Admin + User Webmail). |
| **Frontend**     |  Integrated Webmail |  Integrated Webmail | **Gap Closed**. React App is now bundled into the binary. |

## Deep Dive Analysis: Protocol Compatibility

### IMAP (Mobile & Desktop Clients)
*   **Mox**: Fully compliant. Works with iOS Mail, Gmail App, Thunderbird out-of-the-box.
*   **MailRaven**: **Now Compatible**. We have implemented the core IMAP4rev1 features required by Outlook, Thunderbird, and iOS Mail.
    *   *Status*: Functional. Supports Authentication (TLS), Folder listing, Message retrieval, and IDLE (Push).

### Autoconfiguration
*   **Mox**: Serves XML/JSON config files on `.well-known/autoconfig` and DNS SRV records.
*   **MailRaven**: Implemented XML-based Autodiscover.
    *   *Status*: Supports `autodiscover.xml` (Microsoft) and `config-v1.1.xml` (Mozilla).

## Missing Capabilities (Roadmap)

To reach parity with Mox for a "drop-in replacement" server, we need to address these gaps.

### 1. Protocol Feature Gaps (Functional Gap)
These are standard email server features present in Mox (and Dovecot) that we have not yet implemented:
- **Storage Quotas (RFC 2087)**: We cannot limit user mailbox sizes, which is critical for hosted environments.
- **IMAP ACLs (RFC 4314)**: No support for shared mailboxes or delegation (e.g., "secretary accesses boss's inbox").

### Completed Items
- [x] **IMAP Core**: SELECT, FETCH, UID, STORE implemented.
- [x] **IMAP IDLE**: Real-time push notification support.
- [x] **Autodiscover**: XML configuration for Outlook and Thunderbird.
- [x] **Modern Security**: MTA-STS (Serve), TLS-RPT (Serve), DANE (Send Verification).
- [x] **Spam Protection**: Native Bayesian Filtering, Greylisting, and IMAP retraining hooks.
- [x] **Sieve Filtering**: RFC 5228 Engine, RFC 5804 ManageSieve Protocol, and Vacation extension.
- [x] **Integrated UI**: Web Admin and User Webmail (Read/Write/Vacation settings) are now available and bundled.

## Conclusion

**Mox** is a general-purpose, standards-compliant email server for *standard clients*.

**MailRaven** is an **API-First** Email Platform. It now supports standard clients (Outlook, iOS) via our new IMAP implementation, but philosophically prioritizes programmatic access and flexibility.

**Recommendation**:
With the completion of **Feature 012 (User Portal / Webmail)**, MailRaven has effectively closed the "Headless" gap. It now offers a full-stack experience (Backend + Frontend) out of the box.

The update strategy should now shift away from "Catch up with Mox" to "Beyond Mox" features, such as:
1.  **Mobile Push Notifications** (Native APNS/FCM integration, which Mox lacks).
2.  **AI-assisted Organization** (Categorization).

