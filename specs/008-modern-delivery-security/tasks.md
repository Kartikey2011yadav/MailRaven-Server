# Tasks: Modern Delivery Security

## 1. Core Domain & Data
- [ ] **Define Entities**: Create `internal/core/domain/security.go` with `TLSReport` and `MTASTSPolicy` structs. <!-- id: 1 -->
- [ ] **Create Repository Interface**: Add `TLSRptRepository` to `internal/core/ports/repositories.go`. <!-- id: 2 -->
- [ ] **Implement SQLite Repo**: Create `internal/adapters/storage/sqlite/tlsrpt_repo.go`. <!-- id: 3 -->
- [ ] **Migration**: Add SQL migration for `tls_reports` table. <!-- id: 4 -->

## 2. HTTP Adapters (Receiver)
- [ ] **MTA-STS Handler**: Create `internal/adapters/http/handlers/mtasts.go` to serve policy text. <!-- id: 5 -->
- [ ] **TLS-RPT Handler**: Create `internal/adapters/http/handlers/tlsrpt.go` to ingest JSON. <!-- id: 6 -->
- [ ] **Host Routing**: Update `internal/adapters/http/server.go` to intercept `mta-sts.*` host header. <!-- id: 7 -->
- [ ] **DTOs**: Create `internal/adapters/http/dto/tlsrpt.go`. <!-- id: 8 -->

## 3. SMTP Adapter (Sender)
- [ ] **Add Dependency**: Add `github.com/miekg/dns` to `go.mod`. <!-- id: 9 -->
- [ ] **DANE Logic**: Implement `internal/adapters/smtp/validators/dane.go` (TLSA fetch & verify). <!-- id: 10 -->
- [ ] **Integrate Sender**: Update `internal/adapters/smtp/sender.go` to call DANE validator before dialing/during handshake. <!-- id: 11 -->

## 4. Testing & Verification
- [ ] **Unit Tests**: Test DANE logic with mock DNS responses. <!-- id: 12 -->
- [ ] **Integration Test**: Verify MTA-STS endpoint returns correct Content-Type and body. <!-- id: 13 -->
- [ ] **E2E Test**: Simulate full reporting flow. <!-- id: 14 -->
