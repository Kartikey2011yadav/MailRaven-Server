# Tasks: Modern Delivery Security

**Feature**: 008-modern-delivery-security

## Phase 1: Setup & Foundational
*Blocking prerequisites for user stories*

- [x] T001 Add `github.com/miekg/dns` dependency to go.mod
- [x] T002 Create `internal/core/domain/security.go` with `TLSReport` and `MTASTSPolicy` structs
- [x] T003 Create `TLSRptRepository` interface in `internal/core/ports/repositories.go`
- [x] T004 Implement `TLSRptRepository` in `internal/adapters/storage/sqlite/tlsrpt_repo.go`
- [x] T005 Create migration `005_add_tls_reports.sql` in `internal/adapters/storage/sqlite/migrations/`

## Phase 2: User Story 1 (MTA-STS Policy)
*As an Admin, I want to enable MTA-STS "Enforce" mode so that attackers cannot strip TLS from incoming mail.*

- [x] T006 [US1] P Implement MTA-STS policy builder logic in `internal/core/domain/mtasts.go`
- [x] T007 [US1] Create `internal/adapters/http/handlers/mtasts.go` handler for `GET /.well-known/mta-sts.txt`
- [x] T008 [US1] Implement host-based routing wrapper in `internal/adapters/http/server.go` to intercept `mta-sts.*`
- [x] T009 [US1] Add integration test for MTA-STS serving in `tests/mtasts_test.go`

## Phase 3: User Story 2 (TLS Reporting)
*As an Admin, I want to see a list of TLS failure reports sent by Google/Microsoft so I know if my certificates are misconfigured.*

- [x] T010 [US2] P Create `internal/adapters/http/dto/tlsrpt.go` for JSON unmarshaling
- [x] T011 [US2] Create `internal/adapters/http/handlers/tlsrpt.go` handler for `POST /.well-known/tlsrpt`
- [x] T012 [US2] Register TLS-RPT route in `internal/adapters/http/server.go`
- [x] T013 [US2] Add integration test for TLS-RPT ingestion in `tests/tlsrpt_test.go`

## Phase 4: User Story 3 (DANE Verification)
*As a System, I want to verify DANE records when delivering to security-conscious domains (like ProtonMail) to ensure I am talking to the real server.*

- [x] T014 [US3] P Implement `internal/adapters/smtp/validators/dane.go` with TLSA fetching logic using `miekg/dns`
- [x] T015 [US3] Add unit tests for DANE verification in `internal/adapters/smtp/validators/dane_test.go`
- [x] T016 [US3] Integrate DANE check into `internal/adapters/smtp/client.go` (DialWithDANE logic in `deliverToHost`)
- [x] T017 [US3] Add configuration flags for DANE (Enforce/Advisory) in `internal/config/config.go`

## Phase 5: Polish & Integration
- [x] T018 Verify full flow with Quickstart guide scenarios
- [x] T019 Update `README.md` with new security features documentation

## Dependencies
1. T001 -> T014 (DANE requires DNS lib)
2. T002 -> T006, T010 (Domain structs needed for logic/DTOs)
3. T003/T004 -> T011 (Handler needs Repo)
4. T008 -> T007 (Routing needs Handler, but implemented together usually)

## Parallel Execution Opportunities
- US1 (MTA-STS) and US3 (DANE) are largely independent and can be developed in parallel after generic Setup.
- US2 (TLS-RPT) depends on the DB Setup (Phase 1) but is independent of US1/US3 logic.
