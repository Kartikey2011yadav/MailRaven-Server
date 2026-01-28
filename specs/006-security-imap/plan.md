# Implementation Plan: Security & IMAP Groundwork

**Branch**: `006-security-imap` | **Date**: 2026-01-28 | **Spec**: [specs/006-security-imap/spec.md](specs/006-security-imap/spec.md)
**Input**: Feature specification from `/specs/006-security-imap/spec.md`

## Summary

This feature hardens the server by adding Spam Protection (via **Rspamd** and **DNSBL**) and lays the groundwork for standard **IMAP4** support. It focuses on security middleware for SMTP and establishing a basic RFC-compliant IMAP listener and authentication state machine.

## Technical Context

**Language/Version**: Go 1.25+
**Primary Dependencies**:
- `Rspamd` (External Docker service)
- `net` (Stdlib for DNSBL/TCP)
- `crypto/tls` (Stdlib for IMAP TLS)
**Storage**: None for this feature (Logic only).
**Testing**: Integration tests (GTUBE), Telnet simulation.
**Target Platform**: Linux (primary), Windows (dev).
**Project Type**: Backend Service.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Verify compliance with [MailRaven Constitution](../../.specify/memory/constitution.md):

- [x] **Reliability**: Spam checks must fail-safe (or fail-open based on config) to avoid losing mail during outages.
- [x] **Protocol Parity**: IMAP implementation aims for strict RFC 3501 compliance (connection/auth phases).
- [x] **Mobile-First Architecture**: N/A for this backend feature (though IMAP enables standard mobile clients).
- [x] **Dependency Minimalism**: Implementing a basic IMAP listener from scratch (or minimal stdlib usage) rather than a heavy library to keep control. Rspamd is an external optional dependency.
- [x] **Observability**: Rspamd results will be logged with scores and symbols.
- [x] **Interface-Driven Design**: `SpamScorer` interface allows swapping Rspamd for something else later.
- [x] **Extensibility**: Middleware pattern used for SMTP Data phase.
- [x] **Protocol Isolation**: IMAP Protocol handling (`adapters/imap`) is separate from core logic.
- [x] **Testing Standards**: Integration tests defined for GTUBE and Login.

**Violations Requiring Justification**: None. Rspamd is a standard "sidecar" pattern.

## Project Structure

### Source Code (repository root)

```text
f:/MailRaven-Server/
├── internal/
│   ├── adapters/
│   │   ├── spam/                # NEW: Spam adapters
│   │   │   ├── dnsbl.go         # DNSBL implementation
│   │   │   └── rspamd.go        # Rspamd client
│   │   ├── imap/                # NEW: IMAP Protocol
│   │   │   ├── server.go        # Listener & Loop
│   │   │   ├── session.go       # State Machine
│   │   │   └── commands.go      # Parsers
│   │   └── smtp/
│   │       └── server.go        # UPDATE: Add Middleware hook
│   └── config/
│       └── config.go            # UPDATE: Add Spam/IMAP config
├── cmd/
│   └── mailraven/
│       └── serve.go             # UPDATE: Start IMAP listener
└── deployment/
    └── docker-compose.yml       # UPDATE: Add Rspamd service (optional)
```

## Phases

### Phase 1: Spam Protection

1. **DNSBL**: Implement simple IP lookup against `zen.spamhaus.org`.
2. **Rspamd**: Implement HTTP client to `/checkv2`.
3. **Integration**: Hook into `smtp/server.go`.

### Phase 2: IMAP4 Groundwork

1. **Listener**: TCP server on 143/993 with TLS support.
2. **Session**: State machine (NotAuth -> Auth).
3. **Auth**: Validate credentials against `UserRepository`.

