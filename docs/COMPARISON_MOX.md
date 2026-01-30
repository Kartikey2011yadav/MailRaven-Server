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
| **IMAP4**        |  Full (RFC 3501 + Extensions) | / Auth Only (Skeletal) | **CRITICAL GAP**: MailRaven supports LOGIN/STARTTLS but lacks SELECT/FETCH/IDLE. Standard clients (Outlook, Mobile) **will not work**. |
| **Mobile Push**  |  IMAP IDLE / Notifications |  None | Required for instant mobile notifications. |
| **Autodiscover** |  SRV, XML, Apple Profiles |  None | Users must manually enter Host/Port/SSL settings in apps. |
| **Security**     |  SPF, DKIM, DMARC, DANE, MTA-STS |  SPF, DKIM, DMARC | We miss advanced DANE/MTA-STS and Reporting features. |
| **TLS/ACME**     |  Built-in Automatic ACME |  Built-in Automatic ACME | Implemented via `autocert`. |
| **Spam Filter**  |  Bayesian, Grey-listing |  DNSBL + Rate Limiting | Connection-level filtering present. Content analysis (Bayesian) missing. |
| **Administration**|  Web Admin UI |  Web Admin API | Mox has a full GUI. We have a robust REST API for Admin functions. |
| **Frontend**     |  Integrated Webmail |  Separate React App | We have a specialized Frontend (Client), Mox has generic Webmail. |

## Deep Dive Analysis: Protocol Compatibility

### IMAP (Mobile & Desktop Clients)
*   **Mox**: Fully compliant. Works with iOS Mail, Gmail App, Thunderbird out-of-the-box.
*   **MailRaven**: Currently supports connection and authentication (LOGIN). **Failed** to support mailbox selection (SELECT INBOX) and message retrieval (FETCH). 
    *   *Impact*: You cannot currently use standard email apps with MailRaven. You must use our custom Frontend or build a custom Mobile App using our HTTP API.

### Autoconfiguration
*   **Mox**: Serves XML/JSON config files on `.well-known/autoconfig` and DNS SRV records.
*   **MailRaven**: No implementation.
    *   *Impact*: Friction during user onboarding. Users need technical knowledge to sign in.

## Missing Capabilities (Roadmap)

To reach parity with Mox for a "drop-in replacement" server, we need:

### 1. IMAP core (High Priority for Client Compat)
Implementation of RFC 3501 verbs:
- `SELECT` / `EXAMINE` (Open mailbox)
- `FETCH` (Read headers/body)
- `UID` (Persistent implementation)
- `IDLE` (Push notifications)

### 2. Autodiscover Service
- Implement endpoints: `/.well-known/autoconfig/mail/config-v1.1.xml` (Thunderbird)
- Implement endpoints: `/autodiscover/autodiscover.xml` (Microsoft)
- Configuration for DNS SRV records.

### 3. Advanced Security
- Implement MTA-STS policy serving.
- Implement DANE verification for outbound mail.

## Conclusion

**Mox** is a general-purpose, standards-compliant email server for *standard clients*.

**MailRaven** is currently an **API-First** Email Platform. It excels at programmatic access and custom interfaces (our React Frontend) but currently fails at interoperating with legacy/standard ecosystem tools (Outlook, Apple Mail) due to the incomplete IMAP implementation.

**Recommendation**:
If the goal is to build a *custom* mobile app, we should use the **HTTP API** (which is fully functional) rather than investing heavily in IMAP, unless generic client support is a business requirement.
