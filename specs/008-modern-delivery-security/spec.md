# Feature Specification: Modern Delivery Security

**Feature ID**: 008-modern-delivery-security
**Status**: Draft
**Priority**: High (Critical for Deliverability)

## Overview

Implement advanced email security standards (MTA-STS, TLS-RPT, and DANE) to improve domain reputation, prevent Man-in-the-Middle (MITM) attacks, and ensure reliable delivery to major providers like Gmail and Microsoft. This brings MailRaven into parity with `mox`'s security suite.

## Problem Statement

Currently, MailRaven supports basic SMTP security (TLS, SPF, DKIM, DMARC-check). However, it lacks:
1.  **MTA-STS**: A mechanism to tell senders "You MUST use TLS to talk to me" (preventing downgrade attacks).
2.  **TLS-RPT**: A way to receive reports when senders fail to connect securely.
3.  **DANE**: A stronger, DNSSEC-based method for authenticating destination servers during outbound delivery.

Without these, MailRaven domains may be perceived as "less secure" by strict receivers, and we lack visibility into connectivity issues.

## Functional Requirements

### 1. MTA-STS (Receive Side)
- **Constraint**: `user-can-configure-mta-sts`
    - The server MUST serve the MTA-STS policy file at `https://mta-sts.<domain>/.well-known/mta-sts.txt`.
    - The content MUST be dynamically generated based on config (mode: testing/enforce, max_age, mx).
    - Checks MUST be performed to ensure the `mta-sts` subdomain is correctly configured in DNS (optional helper).

### 2. TLS Reporting (Receive Side)
- **Constraint**: `user-can-receive-tls-reports`
    - The server MUST accept HTTP POST requests at a configured report endpoint (e.g., `/report/tlsrpt`).
    - It MUST support the `application/tlsrpt+json` (or `application/json`) content type.
    - It MUST store received reports for admin inspection (database or log).
    - *Note*: We are not implementing the *sending* of reports yet, only receiving them for our domain.

### 3. DANE Verification (Send Side)
- **Constraint**: `system-verifies-dane-on-send`
    - When sending email, the SMTP client MUST query for TLSA records if DNSSEC is validated.
    - If TLSA records exist, the server certificate MUST match the DANE record constraint.
    - If DANE fails, the connection MUST be terminated (if policy requires).

## User Stories

1.  **As an Admin**, I want to enable MTA-STS "Enforce" mode so that attackers cannot strip TLS from incoming mail.
2.  **As an Admin**, I want to see a list of TLS failure reports sent by Google/Microsoft so I know if my certificates are misconfigured.
3.  **As a System**, I want to verify DANE records when delivering to security-conscious domains (like ProtonMail) to ensure I am talking to the real server.

## Non-Functional Requirements

-   **Performance**: MTA-STS file serving must be fast and cached (it is hit frequently by major providers).
-   **Security**: The MTA-STS endpoint MUST be served over HTTPS with a valid certificate (Critical).
-   **Reliability**: DANE verification must fallback gracefully or fail hard depending on standard RFC behavior (RFC 7672).

## Technical Considerations

-   **Mox Reference**: Look at `mox/mtasts` and `mox/dane` for implementation details.
-   **Dependencies**: Requires robust DNS resolution (Go `net` package might need helper for explicit DNSSEC checks if the system resolver is insufficient, or just rely on `miekg/dns` for TLSA lookups).
-   **HTTP**: Expanding the existing generic HTTP adapter to serve these specific `.well-known` paths.

## Out of Scope
-   Generating/Sending TLS-RPT reports to *other* domains (we are only receiving reports about ourselves for now).
-   DMARC Aggregate Report generation (XML).



