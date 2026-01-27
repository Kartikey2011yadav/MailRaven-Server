# Specification: Admin Management API & CLI

**Status**: Draft
**Owner**: DevOps Team
**Feature**: Phase 8 - Admin Management

## 1. Overview
Currently, MailRaven lacks a runtime interface for managing users and system state. User creation is manual or hardcoded. This feature introduces a robust **Admin API** and a **CLI Tool** to manage the server without direct database access.

## 2. Goals
- **User Management**: Create, List, Delete, and Reset Password for users via API.
- **Role-Based Access Control (RBAC)**: Support `ADMIN` vs `USER` roles in the database.
- **CLI Tool**: A dedicated `mailraven-cli` binary for sysadmins to interact with the API.
- **System Stats**: Endpoint to view storage usage and message counts.

## 3. Data Model Changes

### 3.1 User Entity
Update `User` struct to include a role.

```go
type Role string

const (
    RoleUser  Role = "user"
    RoleAdmin Role = "admin"
)

type User struct {
    Email        string
    PasswordHash string
    Role         Role      // New field
    CreatedAt    time.Time
    LastLoginAt  time.Time
}
```

### 3.2 Repository Inteface
Update `UserRepository` to support storage management.

```go
type UserRepository interface {
    // Existing methods...
    
    // New methods
    List(ctx context.Context, limit, offset int) ([]*domain.User, error)
    Delete(ctx context.Context, email string) error
    UpdatePassword(ctx context.Context, email, passwordHash string) error
    UpdateRole(ctx context.Context, email string, role domain.Role) error
}
```

## 4. API Design

### 4.1 Authentication
- Admin endpoints require a JWT token with `role: "admin"` claim.
- Base Path: `/api/v1/admin`

### 4.2 Endpoints

| Method | Path | Description | Access |
|--------|------|-------------|--------|
| GET | `/users` | List users (paginated) | Admin |
| POST | `/users` | Create new user | Admin |
| DELETE | `/users/{email}` | Delete user | Admin |
| PUT | `/users/{email}/password` | Reset user password | Admin |
| PUT | `/users/{email}/role` | Promote/Demote user | Admin |
| GET | `/system/stats` | Storage and DB stats | Admin |

## 5. CLI Design (`mailraven-cli`)

A separate binary that authenticates against the running server API.

Usage:
```bash
# Configuration
export MAILRAVEN_API="http://localhost:8443"
export MAILRAVEN_ADMIN_TOKEN="<jwt>"

# Commands
mailraven-cli users list
mailraven-cli users create <email> <password> [--admin]
mailraven-cli users delete <email>
mailraven-cli users passwd <email> <new-password>
mailraven-cli system stats
```

## 6. Implementation Plan
1.  **Migration**: Add `role` column to `users` table.
2.  **Core**: Update `User` domain and `UserRepository`.
3.  **Adapter**: Update `SQLiteUserRepository`.
4.  **Service**: Allow `AuthService` to issue tokens with roles.
5.  **HTTP**: Create `AdminUserHandler` and `AdminMiddleware`.
6.  **CLI**: Build the CLI client.
