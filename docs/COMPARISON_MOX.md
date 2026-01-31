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
| Administration   | Web Admin UI | Web Admin API | Mox has a full GUI. We have a robust REST API for Admin functions. |
| **Frontend**     |  Integrated Webmail |  Separate React App | We have a specialized Frontend (Client), Mox has generic Webmail. |

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

### 1. User Interface & Administration (Philosophical Gap)
This difference is intentional but noteworthy. Mox includes full Webmail and Admin UIs in the single binary.
- **Webmail**: While we have a specialized React Client, `mox` enables a self-contained deployment. We might consider bundling our client assets into the Go binary for "single file" deployment parity.
- **Admin UI**: We currently offer a comprehensive REST API. Parity would require building a GUI consuming this API, potentially embedded in the binary.

### Completed Items
- [x] **IMAP Core**: SELECT, FETCH, UID, STORE implemented.
- [x] **IMAP IDLE**: Real-time push notification support.
- [x] **Autodiscover**: XML configuration for Outlook and Thunderbird.
- [x] **Modern Security**: MTA-STS (Serve), TLS-RPT (Serve), DANE (Send Verification).
- [x] **Spam Protection**: Native Bayesian Filtering, Greylisting, and IMAP retraining hooks.

## Conclusion

**Mox** is a general-purpose, standards-compliant email server for *standard clients*.

**MailRaven** is an **API-First** Email Platform. It now supports standard clients (Outlook, iOS) via our new IMAP implementation, but philosophically prioritizes programmatic access and flexibility.

**Recommendation**:
With Basic Client Compliance, Modern Security, and now **Advanced Spam Filtering** gaps closed, MailRaven is nearly feature-complete as a backend server. The remaining major difference is the included UI/Frontend (Mox has it, we separate it). We must decide if we want to abandon the "Headless/API" philosophy to build bundled UIs.
