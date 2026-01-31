# Implementation Checklist: Sieve Filtering

- [ ] **Infrastructure Setup**
  - [ ] Add `github.com/emersion/go-sieve` dependency
  - [ ] Add `github.com/emersion/go-sasl` dependency
  - [ ] Run `go mod tidy`

- [ ] **Data Access Layer**
  - [ ] Create migration for `sieve_scripts` table
  - [ ] Create migration for `vacation_trackers` table
  - [ ] Implement `SieveRepository` (Save, GetActive, List, Delete)
  - [ ] Implement `VacationRepository` (CheckRateLimit, RecordReply)

- [ ] **Core Logic (Sieve Engine)**
  - [ ] Create `internal/core/sieve/engine.go`
  - [ ] Implement `Run` function taking (email, script)
  - [ ] Hook up `fileinto` extension to Delivery Agent
  - [ ] Hook up `vacation` extension (sending email logic)

- [ ] **ManageSieve Protocol Server**
  - [ ] Create `internal/adapters/managesieve/server.go`
  - [ ] Implement TCP Listener on port 4190
  - [ ] Wire up SASL Auth using `internal/core/auth` provider
  - [ ] Implement Commands: capability, putscript, setactive, etc.

- [ ] **Integration**
  - [ ] Edit `internal/core/smtp/delivery.go` (or equivalent)
  - [ ] Call Sieve Engine before saving message
  - [ ] Handle `Implicit Keep` behavior

- [ ] **Public API (Web Admin)**
  - [ ] Implement HTTP handlers matching `contracts/sieve_api.yaml`
  - [ ] Register routes in `webapisrv`

- [ ] **Testing**
  - [ ] Unit tests for Engine with mock emails
  - [ ] Integration test: Connect via `openssl` to ManageSieve
  - [ ] Integration test: Send email -> Check if sorted into folder
