# Tasks: Admin Management API & CLI

**Feature**: Admin Management
**Status**: Pending
**Spec**: [specs/003-admin-management/spec.md](specs/003-admin-management/spec.md)

## Phase 1: Core & Storage
**Goal**: Update the domain model and database layer to support user roles and management operations.

- [x] T033 Update `internal/core/domain/user.go` to include `Role` type and field
- [x] T034 Update `internal/core/ports/repositories.go` with `List`, `Delete`, `UpdatePassword`, `UpdateRole` in `UserRepository`
- [x] T035 Create SQL migration `002_add_user_roles.sql` in `internal/adapters/storage/sqlite/migrations`
- [x] T036 Update `internal/adapters/storage/sqlite/user_repo.go` implementing new interface methods

## Phase 2: Authentication & Logic
**Goal**: Update authentication to support roles and implement admin business logic.

- [x] T037 Update `internal/core/services/auth_service.go` to support role-based login claims
- [x] T038 Create `internal/adapters/http/middleware/admin.go` for Role-Based Access Control (RBAC)
- [x] T039 Create `AdminUserHandler` in `internal/adapters/http/handlers/admin_users.go` combining `UserService` and `SystemStats` logic

## Phase 3: API & Wiring
**Goal**: Expose the new functionality via HTTP.

- [x] T040 Register new routes in `internal/adapters/http/server.go` under `/api/v1/admin/users` protected by Admin Middleware
- [x] T041 Verify Admin API with integration tests in `tests/E2E_Admin_test.go`

## Phase 4: CLI Tool
**Goal**: Build the command-line interface for ease of use.

- [x] T042 Create `cmd/mailraven-cli/main.go` structure with subcommands (using `flag` or `cobra`)
- [x] T043 Implement `users` subcommand in `cmd/mailraven-cli/users.go`
- [x] T044 Implement `system` subcommand in `cmd/mailraven-cli/system.go`
- [x] T045 Add `cli-build` target to `Makefile`

## Phase 5: Documentation
**Goal**: Document the new tools.

- [x] T046 Create `docs/CLI.md` and update `docs/API.md`
