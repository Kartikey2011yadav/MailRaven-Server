---

description: "Task list template for feature implementation"
---

# Tasks: Web Admin UI & Domain Management

**Input**: Design documents from `/specs/004-web-admin/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/openapi.yaml
**Tests**: OPTIONAL - only included where strict contract verification is needed.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: [US1] Stats, [US2] Users, [US3] Domains
- Includes exact file paths.

## Phase 1: Setup (Frontend Initialization & Backend Prep)

**Purpose**: Initialize the React project and prepare backend for CORS without breaking existing logic.

- [ ] T001 Initialize Vite React project in `web/` directory
- [ ] T002 [P] Configure Tailwind CSS and PostCSS in `web/tailwind.config.js` and `web/postcss.config.js`
- [ ] T003 [P] Add Shadcn/UI CLI and initialize `web/components.json`
- [ ] T004 Install frontend dependencies (axios, react-router-dom, lucide-react, clsx, tailwind-merge) in `web/package.json`
- [ ] T005 [P] Implement `web/src/lib/utils.ts` (cn function) for shadcn
- [ ] T006 [P] Create CORS middleware in `internal/adapters/http/middleware/cors.go` allowing localhost frontend
- [ ] T007 Apply CORS middleware to global router in `internal/adapters/http/server.go`

## Phase 2: Foundational (Auth & Base Components)

**Purpose**: Enable Login flow so the dashboard can be accessed.

- [ ] T008 [US1] Create Login Page component in `web/src/pages/Login.tsx`
- [ ] T009 [US1] Create Auth Context/Provider in `web/src/contexts/AuthContext.tsx` handling JWT storage
- [ ] T010 [US1] Configure Protected Routes wrapper in `web/src/components/ProtectedRoute.tsx`
- [ ] T011 [US1] Setup React Router in `web/src/App.tsx` with Login and Dashboard routes
- [ ] T012 [P] [US1] Create Main Layout component (Sidebar + Header) in `web/src/layout/MainLayout.tsx`

## Phase 3: User Story 1 (Admin Dashboard & System Stats)

**Purpose**: Display system vitals.

- [ ] T013 [US1] Implement stats handler `GetSystemStats` in `internal/adapters/http/handlers/admin_stats.go`
- [ ] T014 [US1] Register `GET /api/v1/admin/stats` route in `internal/adapters/http/server.go`
- [ ] T015 [US1] Create API client function `fetchStats` in `web/src/lib/api.ts`
- [ ] T016 [US1] Build Dashboard Page in `web/src/pages/Dashboard.tsx` displaying widgets (Uptime, Users, Memory)

## Phase 4: User Story 2 (User Management UI)

**Purpose**: Manage users without CLI.

- [ ] T017 [US2] Create User API client functions (List, Create, Delete, Update Role) in `web/src/lib/api.ts`
- [ ] T018 [US2] Build Users List Table component in `web/src/pages/Users.tsx` (using shadcn Table)
- [ ] T019 [US2] Create "Add User" Dialog component in `web/src/components/users/CreateUserDialog.tsx`
- [ ] T020 [P] [US2] Create "Delete User" Confirmation Dialog in `web/src/components/users/DeleteUserDialog.tsx`

## Phase 5: User Story 3 (Domain Management)

**Purpose**: Support multiple domains.

### Backend Implementation
- [ ] T021 [US3] Create migration `internal/adapters/storage/sqlite/migrations/003_add_domains.sql`
- [ ] T022 [US3] Define `Domain` entity in `internal/core/domain/domains.go`
- [ ] T023 [US3] Define `DomainRepository` interface in `internal/core/ports/repositories.go`
- [ ] T024 [US3] Implement SQLite Domain Repository in `internal/adapters/storage/sqlite/domain_repo.go`
- [ ] T025 [US3] Implement Domain Admin Handlers in `internal/adapters/http/handlers/admin_domains.go`
- [ ] T026 [US3] Update `CreateUser` logic in `internal/adapters/http/handlers/admin_users.go` to validate domain against DB or Config
- [ ] T027 [US3] Register accessible domain routes in `internal/adapters/http/server.go`

### Frontend Implementation
- [ ] T028 [P] [US3] Create Domain API client functions in `web/src/lib/api.ts`
- [ ] T029 [P] [US3] Build Domains Page in `web/src/pages/Domains.tsx`
- [ ] T030 [P] [US3] Create "Add Domain" Dialog in `web/src/components/domains/CreateDomainDialog.tsx`

## Phase 6: Polish & Cross-Cutting

- [ ] T031 Configure error boundaries or toast notifications (shadcn/sonner) for API errors in `web/src/App.tsx`
- [ ] T032 Verify mobile responsiveness of Sidebar and Tables
- [ ] T033 Create `web/vercel.json` for Rewrite configuration (proxying API)

## Dependencies

- **US2** depends on **User API** (already implemented) + **Phase 2** (Auth)
- **US3 Frontend** depends on **US3 Backend** (Tasks T021-T027)

## Parallel Execution Opportunities

- Frontend Components (T012, T018, T019, T029) can be built in parallel with Backend Handlers (T013, T025) using mock data.
- Backend Database work (T021-T024) is independent of Frontend setup (T001-T005).

## Implementation Strategy

1.  **MVP**: Setup Frontend + Login + Dashboard (Stats).
2.  **Increment 1**: Users Page (uses existing API).
3.  **Increment 2**: Domains (New DB tables + Full Stack).
