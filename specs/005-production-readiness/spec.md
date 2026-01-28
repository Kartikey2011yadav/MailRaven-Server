# Feature Specification: Production Readiness & Polish

**Feature Branch**: `005-production-readiness`  
**Created**: 2026-01-28  
**Status**: Draft  
**Input**: User description: "Fix and polish project: update backend logs (descriptive, colors, error handling), analyze mox website capabilities vs ours, test UI/Production scenarios, generate production guide (domain, setup, deployment)."

## User Scenarios & Testing

### User Story 1 - Admin Troubleshoots Server Issues (Priority: P1)

As a System Admin, I want backend logs to be structured, colorful, and descriptive, so that I can quickly distinguish between standard operations, warnings, and critical errors during debugging.

**Why this priority**: Effective logging is critical for diagnosing issues in a "fully functional" mail server. Currently, logs might be hard to parse visually.

**Independent Test**: Can be tested by running the server, triggering various states (startup, login success, login fail, SMTP transaction), and verifying the console output.

**Acceptance Scenarios**:

1. **Given** the server is starting up, **When** configuration is loaded, **Then** an INFO log message appears in blue/green indicating success with details.
2. **Given** an invalid login attempt, **When** the error is logged, **Then** it appears in yellow (WARN) or red (ERROR) with specific details (IP, user) but without sensitive data (passwords).
3. **Given** a fatal error (e.g., port binding fail), **When** the server crashes, **Then** a clear red ERROR log explains the cause and potential fix.

---

### User Story 2 - User Deploys to Production (Priority: P1)

As a potential user, I want a comprehensive "Production Guide", so that I can successfully configure my domain (DNS, SPF, DKIM) and deploy the server on a public VPS.

**Why this priority**: The goal is a "user friendly easy to setup mail server". Without clear production guides, users cannot leave localhost.

**Independent Test**: Can be tested by a user reading the document and successfully provisioning a server (simulated or real).

**Acceptance Scenarios**:

1. **Given** a user has a VPS and Domain, **When** they follow the guide, **Then** they can configure the necessary DNS records (A, MX, TXT).
2. **Given** the remote server is ready, **When** they follow deployment steps, **Then** MailRaven starts and is accessible via the web guide.

---

### User Story 3 - Feature Exploration & Gap Analysis (Priority: P2)

As a product owner, I want to know if our WebAdmin allows users to "setup and play" with features as easily as the Mox demo site, so that we can identify gaps for future iterations.

**Why this priority**: To ensure feature parity and user engagement comparable to the inspiration project.

**Independent Test**: Delivery of an Analysis Report document.

**Acceptance Scenarios**:

1. **Given** the Mox demo site and MailRaven WebAdmin, **When** the features are compared, **Then** a document `docs/MOX_GAP_ANALYSIS.md` is created listing matching capabilities and missing "playground" features.

### User Story 4 - Cross-Platform Deployment (Priority: P1)

As a user, I want scripts to quickly setup, deploy, and verify the system on my OS (Windows, Linux, Docker, or Mac), so that I don't have to manually configure dependencies.

**Acceptance Scenarios**:
1. **Given** a clean Linux/Mac/Windows environment, **When** the setup script is run, **Then** dependencies are checked and the binary is built/installed.
2. **Given** a Docker environment, **When** `docker-compose up` is run, **Then** Backend, Frontend, and Database containers start and communicate successfully.

### User Story 5 - Database Flexibility (Priority: P2)

As an admin, I want to choose between SQLite (default) and PostgreSQL during setup, so that I can scale the database according to my load requirements.

**Acceptance Scenarios**:
1. **Given** the setup wizard/cli, **When** I select "PostgreSQL", **Then** the application connects to a Postgres instance instead of creating a SQLite file.
2. **Given** the system is running, **When** I trigger backup, **Then** the system provides a prompt/guide on how to backup the chosen database.

## Requirements

### Functional Requirements

#### Backend Logging
- **FR-001**: System MUST output logs to the console with color-coding based on severity (INFO=Green/Blue, WARN=Yellow, ERROR=Red).
- **FR-002**: Logs MUST include context fields: Timestamp, Component (HTTP/SMTP), RequestID (for HTTP), and Message.
- **FR-003**: System MUST catch error edge cases in:
    - Login (Wrong password, locked account, DB error)
    - SMTP Receiving (Max size exceeded, invalid recipient, DNSBL rejection)
    - Startup (Config errors, port conflicts)
- **FR-004**: Logs MUST NOT output sensitive information (raw passwords, API keys).

#### Deployment & Scripts
- **FR-005**: Provide Shell scripts (`setup.sh`, `check.sh`) for Linux/Mac and PowerShell scripts (`setup.ps1`, `check.ps1`) for Windows.
- **FR-006**: Update `docker-compose.yml` to orchestrate specific containers for Backend, Frontend, and optional Database.
- **FR-007**: System MUST support PostgreSQL configuration in `config.yaml` alongside SQLite.

#### Documentation & Analysis
- **FR-008**: A new file `docs/PRODUCTION.md` MUST be created covering DNS, SSL, and Deployment.
- **FR-009**: A Gap Analysis document `docs/GAP_ANALYSIS.md` MUST be created comparing MailRaven vs Mox usability.
- **FR-010**: Generate `docs/BACKUP_ANALYSIS_PROMPT.md` for analyzing the `DB-BACKUP` project capabilities.

## Success Criteria

1.  **Log Readability**: A developer can identify an error in a stream of logs within 5 seconds due to color/formatting.
2.  **Deployment Clarity**: The Production Guide allows a user to setup a server from scratch without needing to check source code.
3.  **Project Roadmap**: The Gap Analysis clearly lists at least 3 actionable items for the next iteration to improve UX.

## Assumptions

- We are using the existing `slog` or custom logger; we will wrap/enhance it for color support.
- The "Mox project" refers to the `mox` mail server (modern full-featured mail server) likely located at `github.com/mjl-/mox`.
- We have simulating capabilities for some production tests (e.g. using `localhost` with modified hosts file, or simple port checks).
