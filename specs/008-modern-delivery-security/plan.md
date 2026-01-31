# Implementation Plan: Modern Delivery Security

**Branch**: `008-modern-delivery-security` | **Date**: 2026-01-30 | **Spec**: [specs/008-modern-delivery-security/spec.md](spec.md)

## Summary

Implement advanced security protocols (MTA-STS, TLS-RPT, DANE) to harden email delivery. This involves adding HTTP endpoints for policy serving/reporting and enhancing the SMTP client to perform DANE verification.

## Technical Context

**Language/Version**: Go 1.24+
**Dependencies**:
- `lib/dns` (Standard `net` or `github.com/miekg/dns` for TLSA)
- Database: `internal/adapters/storage` (SQLite/Postgres)
**Mox Reference**:
- `mox/mtasts`: Policy parsing/serving.
- `mox/tlsrpt`: Report structure.
- `mox/dane`: TLSA verification logic.

**Unknowns**:
- **DANE Verification**: specific logic for validating the AD (Authenticated Data) bit from DNS responses using pure Go. May need `github.com/miekg/dns`. *Resolution: Add as Research Task in Phase 0.*
- **MTA-STS Host Handling**: How to route `mta-sts.` subdomain requests to the existing HTTP adapter? *Resolution: Host matching middleware.*

## Constitution Check

- [x] **Reliability**: DANE failure must not silently drop mail; it should defer or fail based on policy.
- [x] **Protocol Parity**: Aligning with Mox's high-security defaults.
- [x] **Mobile-First**: N/A (Backend feature), but ensures reliability for mobile users.
- [x] **Dependency Minimalism**: Evaluate if `miekg/dns` is strictly necessary or if `net` suffices.

## Project Structure

### Documentation

```text
specs/008-modern-delivery-security/
├── plan.md
├── research.md          # Output of Phase 0
├── data-model.md        # Output of Phase 1
└── tasks.md             # Output of Phase 2
```

### Source Code

```text
internal/
├── adapters/
│   ├── http/
│   │   ├── handlers/
│   │   │   ├── mtasts.go       # [NEW] Policy handler
│   │   │   └── tlsrpt.go       # [NEW] Report collector
│   │   └── server.go           # Update router for mta-sts host
│   └── smtp/
│       └── validators/
│           └── dane.go         # [NEW] DANE Logic
└── core/
    ├── domain/
    │   └── tlsrpt.go           # [NEW] Report Entity
    └── ports/
        └── repositories.go     # Add TLSRptRepository
```

## Phases

### Phase 0: Research & Prototyping
1.  **DANE Research**: Determine how to reliably query TLSA records and check DNSSEC status in Go. Check `mox` implementation.
2.  **MTA-STS Hosting**: Verify how `chi` router handles subdomain routing or if we need a refined entry point.

### Phase 1: Core Domain & Data
1.  Define `TLSReport` struct (JSON mapping).
2.  Create `TLSRptRepository` interface and SQLite/Postgres implementations.
3.  Define `MTASTSPolicy` struct.

### Phase 2: HTTP Implementation (Receive Side)
1.  Implement `GET /.well-known/mta-sts.txt` handler (Policy Builder).
2.  Implement `POST /.well-known/tlsrpt` (or configured path) handler.
3.  Wire up subdomain routing for `mta-sts.*`.

### Phase 3: SMTP Implementation (Send Side)
1.  Implement `dane.Verify(dialer, destination)` function.
2.  Integrate DANE check into `internal/adapters/smtp/sender.go` (Outbound Delivery).
3.  Add "Strict/Audit" modes to configuration.

### Phase 4: Integration & Testing
1.  Test MTA-STS endpoint with `curl`.
2.  Test TLS-RPT ingestion with sample JSON.
3.  Unit test DANE logic against mocked DNS records.
