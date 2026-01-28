# Research: Spam Filtering & IMAP Groundwork

**Feature**: 006-security-imap
**Status**: In Progress

## 1. Rspamd Integration

**Decision**: Use `Rspamd` via HTTP API (`/checkv2`).
**Rationale**:
- **Accuracy**: Industry standard, far superior to simple home-grown Bayesian.
- **Protocol**: HTTP is easier to implement than the Milter protocol in Go.
- **Maintenance**: Updates to rules are handled by the Rspamd container, not our Go code.

**Implementation Details**:
- **Endpoint**: `POST /checkv2`
- **Request**: Headers + Body.
- **Headers**: `Deliver-To`, `IP`, `Helo`, `User`.
- **Response**: JSON with `action` (reject, soft_reject, add_header, greylist, no_action).
- **Go Library**: Use standard `net/http` client. No external Go dependency needed.

## 2. IMAP4 Groundwork

**Decision**: Implement minimal TCP listener for ports 143/993 with a custom State Machine.
**Rationale**:
- **Minimalism**: We only need LOGIN/LOGOUT/CAPABILITY for this phase. Using a full IMAP library (like `go-imap`) might be overkill if we want strict control or mobile optimizations later, AND we want to minimize dependencies (Constitution). However, `emersion/go-imap` is very standard.
- **Constitution Check**: "Dependency Minimalism". Implementing IMAP from scratch is complex and error-prone.
- **Refinement**: We will stick to "Dependency Minimalism" but evaluating strict RFC parsing. Writing a reliable IMAP parser is non-trivial.
- **Comparison**: Mox implements its own `imapserver`. We should likely start with a basic listener that follows RFC 3501 structure but perhaps reuses `go-imap` structures *if* we decide to not reinvent the wheel.
- **Decision Update**: For "Groundwork", we will implement the **socket handling and command loop** ourselves to understand the protocol (Parity), but rely on standard libs for crypto.

**Missing Packages**:
- `rfc`: We need strict RFC 5322 parsing. (Currently we might rely on `net/mail`).
- `iprev`: Required for Rspamd (Reverse DNS lookup). `net.LookupAddr` is standard but proper "Forward-Confirmed reverse DNS" (FCrDNS) logic is distinct.

## 3. Libraries

- **DNSBL**: Use `net` package to query `reversed_ip.dnsbl.domain`.
- **RFC**: Continue using `net/mail` and `net/textproto` where possible.
- **SPF/DKIM**: Existing `internal/adapters/smtp/validators` are sufficient for now, but Rspamd handles these better if widely used. We will keep our internal validators for "Defense in Depth" (or if Rspamd is disabled).

## 4. Unknowns & Risks

- **Rspamd Dependency**: Requires user to run a Docker container. We must ensure the server runs *without* it (Fail open or explicit config).
- **IMAP Complexity**: Even basic Auth requires parsing.
