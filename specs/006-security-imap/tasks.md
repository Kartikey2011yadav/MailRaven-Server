# Tasks: Security & IMAP Groundwork

**Feature Branch**: `006-security-imap`
**Spec**: [specs/006-security-imap/spec.md](specs/006-security-imap/spec.md)
**Status**: In Progress

## Phase 1: Setup

Goal: Initialize project structure and configuration for the new feature.
Independent Test: Application compiles and loads new configuration without errors.

- [x] T001 Create directory structure for spam and imap adapters in `internal/adapters/spam` and `internal/adapters/imap`.
- [x] T002 Update `internal/config/config.go` to include `SpamConfig` (RspamdUrl, Dnsbls) and `ImapConfig` structs.

## Phase 2: Foundational

Goal: Establish external dependencies and shared infrastructure.
Independent Test: `docker-compose up` starts Rspamd and MailRaven successfully.

- [x] T003 Update `deployment/docker-compose.yml` to include the Rspamd service definition.

## Phase 3: User Story 1 (Spam Filtering)

Goal: Incoming emails are scanned by Rspamd and rejected/flagged based on score.
Priority: P1
Independent Test: Send an email with GTUBE string; verify 550 rejection or X-Spam header.

- [x] T004 [P] [US1] Define Rspamd client struct and check result types in `internal/adapters/spam/rspamd.go`.
- [x] T005 [P] [US1] Implement Rspamd HTTP client `Check` method in `internal/adapters/spam/rspamd.go`.
- [x] T006 [US1] Integrate Rspamd check into the SMTP DATA phase in `internal/adapters/smtp/server.go`.

## Phase 4: User Story 2 (DNSBL Protection)

Goal: Reject connections from known spam IPs.
Priority: P2
Independent Test: Unit test `CheckIP` with a known test IP (or mocked resolver) returns error.

- [x] T007 [P] [US2] Implement DNSBL lookup logic `CheckIP` in `internal/adapters/spam/dnsbl.go`.
- [x] T008 [US2] Integrate DNSBL check into the SMTP Connect or HELO phase in `internal/adapters/smtp/server.go`.

## Phase 5: User Story 3 (IMAP Connectivity)

Goal: Standard IMAP client can connect, TLS handshake, and see capabilities.
Priority: P2
Independent Test: `telnet localhost 143` shows greeting; `A01 CAPABILITY` lists supported features.

- [x] T009 [P] [US3] Create IMAP Server struct and main listener loop in `internal/adapters/imap/server.go`.
- [x] T010 [US3] Define `Session` struct and `State` constants (NotAuth, Auth) in `internal/adapters/imap/session.go`.
- [x] T011 [US3] Implement basic command parser (Tag/Command/Args) in `internal/adapters/imap/parser.go`.
- [x] T012 [P] [US3] Implement `CAPABILITY`, `NOOP`, and `LOGOUT` command handlers in `internal/adapters/imap/commands.go`.
- [x] T013 [US3] Integrate IMAP server startup into `cmd/main.go`.

## Phase 6: User Story 4 (IMAP Authentication)

Goal: Users can authenticate via LOGIN command and establish a session.
Priority: P2
Independent Test: `A01 LOGIN user pass` returns OK; `A02 LOGIN bad pass` returns NO.

- [ ] T014 [US4] Implement `STARTTLS` command handling in `internal/adapters/imap/commands.go`.
- [ ] T015 [US4] Implement `LOGIN` command handler using `UserRepository` in `internal/adapters/imap/commands.go`.
- [ ] T016 [US4] Update Session logic to transition state upon successful login in `internal/adapters/imap/session.go`.

## Phase 7: Polish & Cleanup

Goal: Ensure code quality and robust testing.

- [ ] T017 [P] Create unit tests for Rspamd response parsing in `internal/adapters/spam/rspamd_test.go`.
- [ ] T018 [P] Create unit tests for IMAP command parsing in `internal/adapters/imap/parser_test.go`.
- [ ] T019 Update `README.md` with new Rspamd and IMAP configuration details.

## Dependencies

- **US1 & US2** are independent of IMAP stories (US3, US4).
- **US3** (Connectivity) is a prerequisite for **US4** (Auth).
- **Setup** is prerequisite for all.

## Implementation Strategy

We will tackle the Spam Protection features first (US1, US2) as they directly impact the existing SMTP pipeline and offer immediate security value. The IMAP implementation (US3, US4) will follow, building a new listener from scratch.

## Parallel Execution Opportunities

- Rspamd Client (T005) and DNSBL (T007) can be implemented in parallel.
- IMAP Command Handlers (T012, T015) can be implemented by different developers once the Session struct is defined.
