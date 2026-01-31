# Research: Modern Delivery Security

**Feature**: 008-modern-delivery-security
**Status**: Complete

## 1. DANE Verification Logic

**Decision**: Use `github.com/miekg/dns` for DNSSEC validation and TLSA record querying.

**Rationale**:
- The standard `net` package does not expose the DNSSEC "AD" (Authenticated Data) bit reliably across all platforms, nor does it parse TLSA records directly.
- `mox` uses a custom DNS library. Replicating that is too much effort.
- `miekg/dns` is the standard low-level DNS library for Go.

**Alternatives Considered**:
- **Parsing `/etc/resolv.conf` manually**: Too brittle.
- **Calling `dig` via logic**: Forbidden by "Pure Go" preference.

## 2. MTA-STS Host Routing

**Decision**: Use a Top-Level Host Mux in `internal/adapters/http/server.go`.

**Rationale**:
- MTA-STS requires serving content specifically on `mta-sts.<domain>`.
- Our current `chi` router is path-based.
- Adding a simple wrapper that checks `r.Host` before the main router is the simplest, least-intrusive way to isolate this traffic.

**Code Pattern**:
```go
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    if strings.HasPrefix(r.Host, "mta-sts.") {
        s.handleMTASTS(w, r)
        return
    }
    s.router.ServeHTTP(w, r)
}
```

## 3. TLS Reporting (TLS-RPT) Structure

**Decision**: Implement a dedicated `TLSReport` struct mapping to `application/tlsrpt+json`.

**Structure**:
Based on RFC 8460 and mox `report.go`:
- `organization-name`
- `date-range` (start/end)
- `contact-info`
- `report-id`
- `policies` (Array of policy summaries and failure details)

**Storage Strategy**:
- Store raw JSON in a `tls_reports` table with indexed fields for `Date`, `Domain`, and `Result` (Success/Fail).
- This allows easy querying by Admin API later (e.g., "Show me failures for the last 24h").
