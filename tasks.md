# Implementation Tasks - Standard Client Compliance

## Phase 1: Setup & Configuration
- [x] [TASK-101] Project Configuration & Dependencies
    - Verify `go.mod` (using Go 1.21+)
    - Verify `config` package structure
- [x] [TASK-102] Domain Entities Update
    - Update `domain/message.go` (UID, Flags, Directory)
    - Update `domain/mailbox.go` (Entities)

## Phase 2: Storage & Database
- [x] [TASK-201] Schema Migration
    - Create `migrations/004_imap_support.sql` (UIDs, Mailboxes)
- [x] [TASK-202] SQLite Repository Implementation
    - Update `email_repo.go` (UID Management, Mailbox CRUD)
    - Implement `AssignUID` and atomic `Save` with Notifications

## Phase 3: Autodiscover Support
- [x] [TASK-301] XML DTOs
    - Create `internal/adapters/http/dto/autodiscover.go`
- [x] [TASK-302] HTTP Handlers
    - Create `internal/adapters/http/handlers/autodiscover.go` (POST /autodiscover/autodiscover.xml, GET /mail/config-v1.1.xml)
- [x] [TASK-303] Route Registration
    - Update `internal/adapters/http/server.go`

## Phase 4: IMAP Core Implementation
- [x] [TASK-401] IMAP Server Infrastructure
    - Update `internal/adapters/imap/server.go` (Listeners)
    - Create `internal/adapters/imap/session.go` (State Machine)
- [x] [TASK-402] Command Parsers & Handlers
    - Create/Update `internal/adapters/imap/commands.go`
    - Implement CAPABILITY, LOGIN, LOGOUT, SELECT, LIST, CREATE, DELETE
- [x] [TASK-403] Message Retrieval & Management
    - Implement UID FETCH, FETCH (Headers, Body)
    - Implement STORE (Flags)

## Phase 5: IDLE & Real-time
- [x] [TASK-501] Notification Hub
    - Create `internal/core/notifications/hub.go`
- [x] [TASK-502] IDLE Command
    - Create `internal/adapters/imap/idle.go`
- [x] [TASK-503] Integration
    - Wire `Save` in `email_repo.go` to Notification Hub

## Phase 6: Polish & Verification
- [x] [TASK-601] Dependency Injection
    - Verify `cmd/mailraven/serve.go` wiring
- [x] [TASK-602] Integration Testing
    - Verify email flow (SMTP -> Repo -> IDLE -> Client)
