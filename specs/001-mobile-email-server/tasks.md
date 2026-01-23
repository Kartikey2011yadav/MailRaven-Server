# Tasks: Modular Email Server

**Feature Branch**: `001-mobile-email-server`  
**Input**: Design documents from `/specs/001-mobile-email-server/`  
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/rest-api.yaml

**Organization**: Tasks grouped by user story to enable independent implementation and testing.

## Task Format: `- [ ] [ID] [P?] [Story?] Description`

- **Checkbox**: `- [ ]` (marks task as incomplete)
- **ID**: Sequential task number (T001, T002, ...)
- **[P]**: Parallelizable - can run concurrently with other [P] tasks
- **[Story]**: User story label (US1, US2, US3, US4, US5)
- **Description**: Includes exact file path

## Implementation Strategy

**MVP Scope**: User Stories 1, 2, 3 (P1 priorities) - Server setup, email reception, mobile API sync

**User Story Mapping**:
- **US1** (P1): Initial server setup and configuration (quickstart command)
- **US2** (P1): Receive and store emails (SMTP listener, SPF/DKIM/DMARC, storage)
- **US3** (P1): Mobile app syncs emails via API (REST endpoints, pagination)
- **US4** (P2): Authentication and authorization (JWT tokens, middleware)
- **US5** (P3): Send outbound email via API (SMTP client, delivery queue)

**Dependencies**:
- Phase 1 (Setup) must complete before Phase 2
- Phase 2 (Foundational) must complete before any user story work
- User Stories 1, 2, 3 can proceed in parallel after Phase 2
- User Story 4 (Auth) should complete before Story 3 in production, but can be stubbed for development
- User Story 5 (Send) depends on Story 1 (config) and Story 4 (auth)

---

## Phase 1: Setup (Project Initialization)

**Purpose**: Create repository structure and initialize Go module

- [X] T001 Create project directory structure per plan.md (cmd/, internal/core/, internal/adapters/, tests/)
- [X] T002 Initialize Go module with `go mod init github.com/Kartikey2011yadav/mailraven-server`
- [X] T003 [P] Create go.mod with dependencies: modernc.org/sqlite, go-chi/chi/v5, golang.org/x/crypto
- [X] T004 [P] Create Makefile with build, test, lint targets
- [X] T005 [P] Setup .gitignore for Go project (bin/, *.db, *.log, vendor/)
- [X] T006 [P] Create README.md with project overview and build instructions

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure required before ANY user story can begin

**⚠️ CRITICAL**: No user story implementation starts until this phase is 100% complete

### Core Domain & Ports

- [X] T007 [P] Create domain entities in internal/core/domain/message.go (Message, MessageBody structs)
- [X] T008 [P] Create domain entities in internal/core/domain/user.go (User, AuthToken structs)
- [X] T009 [P] Create domain entities in internal/core/domain/smtp.go (SMTPSession struct)
- [X] T010 [P] Define EmailRepository interface in internal/core/ports/repositories.go
- [X] T011 [P] Define UserRepository interface in internal/core/ports/repositories.go
- [X] T012 [P] Define BlobStore interface in internal/core/ports/storage.go
- [X] T013 [P] Define SearchIndex interface in internal/core/ports/search.go
- [X] T014 [P] Define common errors in internal/core/ports/errors.go (ErrNotFound, ErrAlreadyExists, etc.)

### Storage Layer (SQLite + File System)

- [X] T015 Create SQLite schema in internal/adapters/storage/sqlite/migrations/001_init.sql (users, messages, indexes)
- [X] T016 Add FTS5 virtual table to migrations/001_init.sql (messages_fts for full-text search)
- [X] T017 [P] Implement EmailRepository in internal/adapters/storage/sqlite/email_repo.go (Save, FindByID, FindByUser, UpdateReadState, CountByUser, FindSince methods)
- [X] T018 [P] Implement UserRepository in internal/adapters/storage/sqlite/user_repo.go (Create, FindByEmail, Authenticate, UpdateLastLogin methods)
- [X] T019 [P] Implement SearchIndex in internal/adapters/storage/sqlite/search_repo.go (Index, Search, Delete methods using FTS5)
- [X] T020 [P] Implement BlobStore in internal/adapters/storage/disk/blob_store.go (Write with gzip compression, Read with decompression, Delete, Verify methods)
- [X] T021 Add connection pooling and WAL mode setup in internal/adapters/storage/sqlite/connection.go
- [X] T022 Add database integrity check on startup in internal/adapters/storage/sqlite/connection.go (PRAGMA integrity_check)

### Configuration & Observability

- [X] T023 [P] Create Config struct in internal/config/config.go (domain, ports, storage paths, DKIM settings)
- [X] T024 [P] Implement YAML config loader in internal/config/config.go using gopkg.in/yaml.v3
- [X] T025 [P] Create structured logger in internal/observability/logger.go using log/slog
- [X] T026 [P] Create Prometheus metrics in internal/observability/metrics.go (messages_received, messages_rejected, api_request_duration, etc.)

**Checkpoint**: Foundation complete - user story implementation can begin

---

## Phase 3: User Story 1 - Initial Server Setup (Priority: P1) 🎯 MVP

**Goal**: Administrator can run `mailraven quickstart` to generate config and DNS records

**Independent Test**: Run quickstart, verify config files created, DNS records printed, server starts successfully

### Implementation for US1

- [X] T027 [P] [US1] Implement DKIM key generation in internal/config/dkim.go (RSA-2048 private/public key pair)
- [X] T028 [P] [US1] Create DNS record generator in internal/config/dns.go (MX, SPF, DKIM, DMARC record formatting)
- [X] T029 [US1] Implement quickstart command in cmd/mailraven/quickstart.go (calls DKIM gen, config gen, DNS gen)
- [X] T030 [US1] Add user creation logic to quickstart (prompt for admin password, call UserRepository.Create)
- [X] T031 [US1] Add config file writer to quickstart (generate /etc/mailraven/config.yaml with default values)
- [X] T032 [US1] Format and print DNS records to console in quickstart command
- [X] T033 [US1] Add server start validation in quickstart (check ports available, config valid)

**Checkpoint**: Administrator can run quickstart and get working config

---

## Phase 4: User Story 2 - Receive and Store Emails (Priority: P1) 🎯 MVP

**Goal**: External SMTP servers can send emails to MailRaven, which validates SPF/DMARC and stores messages durably

**Independent Test**: Send email via SMTP client, verify "250 OK", message in SQLite, body in file store, SPF/DMARC results recorded

### SPF/DKIM/DMARC Validation (Ported from mox)

- [X] T034 [P] [US2] Implement SPF validation in internal/adapters/smtp/validators/spf.go (DNS lookups, mechanism evaluation per RFC 7208)
- [ ] T035 [P] [US2] Add SPF test cases in internal/adapters/smtp/validators/spf_test.go (table-driven tests with RFC test vectors)
- [X] T036 [P] [US2] Implement DKIM verification in internal/adapters/smtp/validators/dkim.go (signature parsing, DNS key retrieval, validation per RFC 6376)
- [ ] T037 [P] [US2] Add DKIM test cases in internal/adapters/smtp/validators/dkim_test.go (valid/invalid signatures, missing keys)
- [X] T038 [P] [US2] Implement DMARC evaluation in internal/adapters/smtp/validators/dmarc.go (policy retrieval, alignment checks per RFC 7489)
- [ ] T039 [P] [US2] Add DMARC test cases in internal/adapters/smtp/validators/dmarc_test.go (policy enforcement scenarios)
- [X] T040 [P] [US2] Add RFC section comments to all validator code (e.g., `// RFC 7208 section 4.6.4`)

### MIME Parsing

- [X] T041 [P] [US2] Implement MIME parser in internal/adapters/smtp/mime/parser.go (multipart message parsing per RFC 2045)
- [X] T042 [P] [US2] Add body text extraction in MIME parser (plaintext and HTML parts, generate 200-char snippet)
- [ ] T043 [P] [US2] Add MIME parser tests in internal/adapters/smtp/mime/parser_test.go (multipart messages, attachments, edge cases)

### SMTP Server & Middleware

- [X] T044 [US2] Define Middleware interface in internal/adapters/smtp/middleware.go (MessageHandler func type, Chain function)
- [X] T045 [US2] Implement SPF validator middleware in internal/adapters/smtp/validators/spf.go (wraps SPF validation as middleware)
- [X] T046 [US2] Implement DKIM validator middleware in internal/adapters/smtp/validators/dkim.go (wraps DKIM validation as middleware)
- [X] T047 [US2] Implement DMARC validator middleware in internal/adapters/smtp/validators/dmarc.go (wraps DMARC evaluation as middleware)
- [X] T048 [US2] Create SMTP server in internal/adapters/smtp/server.go (net/smtp based, port 25 listener)
- [X] T049 [US2] Implement SMTP command handler in internal/adapters/smtp/handler.go (EHLO, MAIL FROM, RCPT TO, DATA commands per RFC 5321)
- [X] T050 [US2] Wire middleware pipeline in SMTP handler (SPF → DKIM → DMARC → storage)
- [X] T051 [US2] Add SMTP session logging in handler (session ID, remote IP, sender, recipients, SPF/DMARC results)
- [X] T052 [US2] Implement atomic storage in handler (begin transaction, save to EmailRepository, write to BlobStore, commit, fsync, then "250 OK")
- [X] T053 [US2] Add error handling in handler (4xx for temporary failures, 5xx for permanent failures, rollback on error)

### Integration Tests for US2

- [ ] T054 [US2] Create SMTP integration test in tests/smtp_test.go (send email via net/smtp client, verify storage)
- [ ] T055 [US2] Add durability test in tests/durability_test.go (kill -9 immediately after "250 OK", verify message recovered on restart)
- [ ] T056 [US2] Add SPF/DKIM/DMARC integration test (send emails with different auth results, verify correct validation)

**Checkpoint**: Server receives emails, validates senders, stores durably

---

## Phase 5: User Story 3 - Mobile App Syncs Emails (Priority: P1) 🎯 MVP

**Goal**: REST API allows mobile app to list messages, get message details, update read state

**Independent Test**: Use curl/Postman to call API endpoints, verify JSON responses match OpenAPI spec, pagination works

### HTTP Server & Routing

- [X] T057 [US3] Create HTTP server in internal/adapters/http/server.go (go-chi router, port 443, TLS setup)
- [X] T058 [US3] Define routes in internal/adapters/http/routes.go (/v1/messages, /v1/messages/{id}, /v1/messages/since, /v1/messages/search)
- [X] T059 [P] [US3] Create DTO structs in internal/adapters/http/dto/message.go (MessageSummary, MessageFull matching OpenAPI spec)

### Message Handlers

- [X] T060 [P] [US3] Implement GET /v1/messages handler in internal/adapters/http/handlers/messages.go (calls EmailRepository.FindByUser with pagination)
- [X] T061 [P] [US3] Implement GET /v1/messages/{id} handler in internal/adapters/http/handlers/messages.go (calls EmailRepository.FindByID and BlobStore.Read)
- [X] T062 [P] [US3] Implement PATCH /v1/messages/{id} handler in internal/adapters/http/handlers/messages.go (calls EmailRepository.UpdateReadState)
- [X] T063 [P] [US3] Implement GET /v1/messages/since handler in internal/adapters/http/handlers/messages.go (calls EmailRepository.FindSince for delta sync)
- [X] T064 [P] [US3] Implement GET /v1/messages/search handler in internal/adapters/http/handlers/search.go (calls SearchIndex.Search with FTS5 query)

### HTTP Middleware

- [X] T065 [P] [US3] Create logging middleware in internal/adapters/http/middleware/logging.go (log request method, path, duration, status)
- [X] T066 [P] [US3] Create compression middleware in internal/adapters/http/middleware/compression.go (gzip responses when Accept-Encoding: gzip)
- [X] T067 [P] [US3] Create rate limiting middleware in internal/adapters/http/middleware/ratelimit.go (100 req/min per IP)
- [X] T068 [US3] Apply middleware to routes in internal/adapters/http/routes.go (logging → compression → rate limit → auth)

### Integration Tests for US3

- [X] T069 [US3] Create API integration test in tests/api_test.go (list messages, get message, update read state)
- [X] T070 [US3] Add pagination test in tests/api_test.go (request messages with limit/offset, verify paging works)
- [X] T071 [US3] Add compression test in tests/api_test.go (verify gzip compression when requested)
- [X] T072 [US3] Add delta sync test in tests/api_test.go (call /messages/since, verify only new messages returned)

**Checkpoint**: Mobile app can sync email list, read messages, mark as read ✅

---
## Phase 6: User Story 4 - Authentication (Priority: P2)

**Goal**: API endpoints require valid JWT token, login endpoint issues tokens

**Independent Test**: Attempt API call without token (401), authenticate (/auth/login), use token (200)

### Implementation for US4

- [ ] T073 [P] [US4] Create JWT utilities in internal/adapters/http/auth/jwt.go (GenerateToken, ValidateToken functions)
- [ ] T074 [P] [US4] Implement POST /v1/auth/login handler in internal/adapters/http/handlers/auth.go (calls UserRepository.Authenticate, returns JWT)
- [ ] T075 [US4] Create auth middleware in internal/adapters/http/middleware/auth.go (extract JWT from Authorization header, validate, set user context)
- [ ] T076 [US4] Apply auth middleware to protected routes in internal/adapters/http/routes.go (all /v1/* except /auth/login)
- [ ] T077 [US4] Add error responses for auth failures in handlers (401 Unauthorized with JSON error message)

### Integration Tests for US4

- [ ] T078 [US4] Create auth integration test in tests/api_test.go (login with valid credentials, receive token)
- [ ] T079 [US4] Add unauthorized test in tests/api_test.go (call API without token, verify 401)
- [ ] T080 [US4] Add expired token test in tests/api_test.go (use token past expiration, verify 401)

**Checkpoint**: API is secured with JWT authentication

---

## Phase 7: User Story 5 - Send Outbound Email (Priority: P3)

**Goal**: Mobile app can compose and send emails via API, server delivers via SMTP

**Independent Test**: POST to /v1/messages/send, verify DKIM signature generated, SMTP delivery attempted, sent message stored

### Implementation for US5

- [ ] T081 [P] [US5] Create outbound message queue in internal/core/domain/queue.go (OutboundMessage struct, queue operations)
- [ ] T082 [P] [US5] Implement POST /v1/messages/send handler in internal/adapters/http/handlers/send.go (validate input, sign with DKIM, enqueue message)
- [ ] T083 [P] [US5] Create DKIM signer in internal/adapters/smtp/dkim/signer.go (generate DKIM-Signature header per RFC 6376)
- [ ] T084 [US5] Implement SMTP client in internal/adapters/smtp/client.go (MX lookup, connect to recipient's server, deliver message)
- [ ] T085 [US5] Create delivery worker in internal/adapters/smtp/delivery.go (process queue, attempt delivery, retry with backoff on failure)
- [ ] T086 [US5] Add delivery status tracking (update message record with delivery timestamp, status, recipient response)
- [ ] T087 [US5] Add exponential backoff retry logic in delivery worker (retry after 1min, 5min, 15min, 1hour)

### Integration Tests for US5

- [ ] T088 [US5] Create send integration test in tests/smtp_client_test.go (send email to test SMTP server, verify delivery)
- [ ] T089 [US5] Add DKIM signature test (verify signature validates with public key)
- [ ] T090 [US5] Add retry test (simulate delivery failure, verify retry attempts with backoff)

**Checkpoint**: Users can send emails from mobile app

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Final refinements for production readiness

### Observability

- [ ] T091 [P] Add Prometheus /metrics endpoint in cmd/mailraven/main.go
- [ ] T092 [P] Add structured logging to all SMTP sessions (session ID, remote IP, SPF/DMARC results)
- [ ] T093 [P] Add API request logging (endpoint, method, duration, status code)
- [ ] T094 [P] Add metrics collection (messages_received, messages_sent, api_requests, storage_operations)

### Application Entrypoint

- [ ] T095 Create main.go in cmd/mailraven/main.go (parse CLI args, load config, wire dependencies)
- [ ] T096 Add serve command in cmd/mailraven/serve.go (start SMTP server and HTTP server concurrently)
- [ ] T097 Add graceful shutdown in serve command (handle SIGTERM/SIGINT, close connections cleanly)
- [ ] T098 Add startup checks in serve command (verify config valid, ports available, database accessible)

### Documentation

- [ ] T099 [P] Create deployment/mailraven.service (systemd service file)
- [ ] T100 [P] Create deployment/config.example.yaml (example configuration with comments)
- [ ] T101 [P] Update README.md with build instructions, quickstart guide link, license info

### Testing Infrastructure

- [ ] T102 [P] Add test fixtures in tests/testdata/ (sample MIME messages, SPF records, DKIM keys)
- [ ] T103 [P] Create test helper functions in tests/helpers.go (setup test database, generate test users, seed messages)
- [ ] T104 [P] Add integration test suite runner in tests/integration_test.go (orchestrate SMTP + API + storage tests)

**Checkpoint**: Production-ready email server

---

## Dependencies & Execution Strategy

### Critical Path (Must Complete in Order)

1. **Phase 1** (Setup) → **Phase 2** (Foundational) → All user stories can begin
2. **US1** (Quickstart) → Generates config needed by US2 and US3
3. **US2** (Receive) → Provides messages for US3 to query
4. **US4** (Auth) → Should wrap US3 endpoints before production
5. **US5** (Send) → Requires US1 (config) and US4 (auth)

### Parallel Execution Opportunities

**After Phase 2 completes**, these can run in parallel:

- **Parallel Track A**: T027-T033 (US1: Quickstart)
- **Parallel Track B**: T034-T056 (US2: SMTP + Validation + Storage)
- **Parallel Track C**: T057-T072 (US3: REST API)

**After US1, US2, US3 complete**:

- **Parallel Track D**: T073-T080 (US4: Authentication)
- **Parallel Track E**: T081-T090 (US5: Send Email)
- **Parallel Track F**: T091-T104 (Polish & Observability)

### MVP Delivery (P1 Stories Only)

Minimum viable product includes:
- Phase 1: Setup (T001-T006)
- Phase 2: Foundational (T007-T026)
- Phase 3: US1 - Quickstart (T027-T033)
- Phase 4: US2 - Receive (T034-T056)
- Phase 5: US3 - API Sync (T057-T072)
- Phase 8: Partial (T095-T098 for entrypoint, T099-T100 for deployment)

**Estimated MVP Time**: 12-17 days with 1 developer

**Full Feature Time**: 20-28 days including US4 (Auth) and US5 (Send)

---

## Task Count Summary

- **Phase 1 (Setup)**: 6 tasks
- **Phase 2 (Foundational)**: 18 tasks (blocking)
- **Phase 3 (US1 - Quickstart)**: 7 tasks
- **Phase 4 (US2 - Receive)**: 23 tasks
- **Phase 5 (US3 - API Sync)**: 16 tasks
- **Phase 6 (US4 - Auth)**: 8 tasks
- **Phase 7 (US5 - Send)**: 10 tasks
- **Phase 8 (Polish)**: 14 tasks

**Total**: 104 tasks across 5 user stories

**Parallelizable**: 52 tasks marked with [P] can run concurrently

---

## Validation Checklist

All tasks follow required format:
- ✅ Every task has checkbox `- [ ]`
- ✅ Every task has sequential ID (T001-T104)
- ✅ Parallelizable tasks marked with [P]
- ✅ User story tasks marked with [Story] label (US1-US5)
- ✅ Every task includes specific file path
- ✅ Tasks organized by user story for independent testing
- ✅ Setup and Foundational phases complete before user stories
- ✅ Dependencies documented
- ✅ MVP scope clearly identified (P1 stories)
- ✅ Parallel execution opportunities identified

**Status**: Tasks ready for implementation ✨
