# Tasks: User Portal & Webmail

**Reviewer**: Copilot
**Feature**: webmail-portal (Feature 012)

## Phase 1: Foundation & Settings
**Goal**: User login and self-management.

- [x] T001 Update `internal/adapters/http/dto` to include `UserRole` in LoginResponse
- [x] T002 Update `internal/adapters/http/handlers/auth.go` to return role
- [x] T003 Implement `PUT /api/v1/users/self/password` handler
- [x] T004 Frontend: Update `AuthProvider` to store/check Role
- [x] T005 Frontend: Create `UserLayout` and `/mail` route guard
- [x] T006 Frontend: Create `Settings` page with Password Change form
- [x] T007 Frontend: Create `VacationSettings` form (Generates Sieve script and calls Feature 010 API)

## Phase 2: Webmail Reader
**Goal**: Read-only access.

- [x] T008 Frontend: Install `dompurify` and types
- [x] T009 Frontend: Create `MailSidebar` (Inbox, Sent, Trash etc.)
- [x] T010 Frontend: Create `MessageList` component (using existing GET /messages API)
- [x] T011 Frontend: Create `MessageReader` component (Handle HTML/Text rendering)
- [x] T012 Backend: Verify `GET /messages` correctly filters by auth user (Integration Test)

## Phase 3: Webmail Sender
**Goal**: Sending capability.

- [x] T013 Frontend: Create `Composer` component
- [x] T014 Backend: Ensure `POST /messages/send` validates `From` matches Auth User
- [x] T015 Integration: End-to-end test (Login -> Compose -> Send -> Receive)

## Phase 4: Polish
**Goal**: UX Parity with Mox.

- [x] T016 Enable "Dark Mode" for Webmail (inherited from system)
- [x] T017 Add "Logout" button to User layout
