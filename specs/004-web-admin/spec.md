# Feature Specification: Web Admin UI & Domain Management

**Feature Branch**: `004-web-admin`  
**Created**: 2026-01-27  
**Status**: Draft  
**Input**: User description: "Build a Web Admin UI (React), consuming these APIs. Implement System Stats endpoint to show uptime, user count, etc. Add Domain Management for multi-domain support."

## User Scenarios & Testing

<!--
  Prioritized user journeys.
-->

### User Story 1 - Admin Dashboard & System Stats (Priority: P1)

As an Administrator, I want to see an overview of the system's health and usage statistics so that I can monitor server performance and growth.

**Why this priority**: Core visibility into the system is essential for administration.

**Independent Test**: Can be tested by accessing the `/admin` dashboard and verifying that stats (uptime, user counts) are displayed and match actual database values.

**Acceptance Scenarios**:

1.  **Given** the server is running, **When** I navigate to the Admin Dashboard, **Then** I see the current uptime, total registered users, and system memory usage.
2.  **Given** I am not logged in, **When** I access `/admin`, **Then** I am redirected to a login page.
3.  **Given** I am logged in as a standard user, **When** I access `/admin`, **Then** I am denied access (403 or redirect).

---

### User Story 2 - User Management via UI (Priority: P2)

As an Administrator, I want to manage users through a web interface so that I don't need to use the CLI for common tasks.

**Why this priority**: Improves usability and efficiency for admins.

**Independent Test**: Create a user via the UI and verify they can log in.

**Acceptance Scenarios**:

1.  **Given** I am on the Users page, **When** I click "Create User" and fill the form, **Then** the user is created in the database.
2.  **Given** a list of users, **When** I click "Delete" on a user, **Then** the user is removed from the system.
3.  **Given** a user exists, **When** I change their role to "Admin" via the UI, **Then** they gain admin privileges.

---

### User Story 3 - Domain Management (Priority: P3)

As an Administrator, I want to manage multiple domains supported by the mail server so that I can host email for different organizations.

**Why this priority**: Enables multi-tenancy and expansion.

**Independent Test**: Add a domain via UI, then try to create a user with that domain email.

**Acceptance Scenarios**:

1.  **Given** I am on the Domains page, **When** I add `example.org`, **Then** it appears in the list of active domains.
2.  **Given** `example.org` is added, **When** I create a user `bob@example.org`, **Then** the creation succeeds (whereas it would fail if the domain were not allowed).
3.  **Given** `example.org` is removed, **When** I try to create an email address for that domain, **Then** it fails.

### Edge Cases

-   **Invalid Domain Format**: What happens if admin tries to add "invalid-domain"? System MUST reject it.
-   **Duplicate Domain**: What happens if admin adds a domain that already exists? System MUST reject it gracefully.
-   **Removing Primary Domain**: What happens if admin tries to remove the config-defined primary domain? System MUST prevent this.
-   **Removing Domain with Users**: What happens if admin tries to remove a domain that has active users? System SHOULD warn or block (or cascade delete, but blocking is safer default).
-   **Stats Timeout**: What happens if DB is slow when fetching stats? Endpoint SHOULD return partial stats or specific error, not hang indefinitely.

## Requirements

### Functional Requirements

#### Web Admin UI
-   **FR-001**: System MUST serve a Single Page Application (SPA) at the `/admin` path.
-   **FR-002**: The UI MUST require authentication for all pages.
-   **FR-003**: The UI MUST provide a Login form exchanging credentials for a JWT.
-   **FR-004**: The UI MUST store the JWT securely (e.g., localStorage or cookie) and attach it to API requests.
-   **FR-005**: The UI MUST handle 401/403 errors by redirecting to login.

#### System Stats API
-   **FR-006**: System MUST expose an endpoint `GET /api/v1/admin/stats` protected by Admin role.
-   **FR-007**: The Stats API MUST return:
    -   Uptime (seconds)
    -   Memory Usage (MB)
    -   Total User Count
    -   Total Domain Count
    -   Active Database Connections (if available)

#### Domain Management
-   **FR-008**: System MUST persist a list of authorized Domains in the database.
-   **FR-009**: System MUST expose endpoints to List, Add, and Remove domains (`/api/v1/admin/domains`).
-   **FR-010**: System MUST enforce that new User emails belong to an authorized Domain.
-   **FR-011**: Configuration MUST allow defining an initial "primary domain" (from config.yaml) which is automatically treated as authorized.

### Key Entities

-   **Domain**:
    -   Name (Response: "example.com")
    -   CreatedAt (Timestamp)
    -   IsPrimary (Boolean, optional - derived from config)

## Assumptions

-   The Frontend will be built with React.
-   The Go server will define a "catch-all" handler for `/admin/*` to serve the React `index.html`.
-   Deployment process includes building the frontend assets and placing them in a specific directory expected by the server.
-   Password management (reset/change) is out of scope for this specific feature (User Management implies CRUD, specialized operations like Password Reset might be basic or deferred).
