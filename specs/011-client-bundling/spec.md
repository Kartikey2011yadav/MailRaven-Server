# Specification: Client Bundling & Auto-Updates

**Feature ID**: 011
**Status**: Draft
**Priority**: Medium

## Context
MailRaven currently requires separate processes for the Backend (Go) and Frontend (React/Node). Reference implementation `mox` distributes a single binary with embedded web assets and built-in auto-update capabilities. To simplify deployment and user experience, we aim to match this "single binary" philosophy.

## Goals
1.  **Single Binary Distribution**: Embed the compiled React frontend into the Go binary.
2.  **Zero-Config Web UI**: Serving the Admin/Client UI directly from the main HTTP port without a separate Node.js server.
3.  **Self-Updating**: Allow the server to check for updates, verify signatures, and upgrade itself in place.

## Requirements

### Functional Requirements

#### F1: Static Asset Embedding
*   The build process MUST compile the React client to static files (`dist/`).
*   The Go backend MUST embed these files using `//go:embed`.
*   The HTTP server MUST serve these assets on the root path `/`.
*   The server MUST handle SPA routing (returning `index.html` for unknown non-API routes).

#### F2: Auto-Update Mechanism
*   The server MUST provide an API endpoint to check for the latest release (GitHub Releases or designated update server).
*   The server MUST be able to download the new binary for the current OS/Arch.
*   The server MUST verify the cryptographic signature of the downloaded binary.
*   The server MUST apply the update (replace binary) and restart the service.
*   Admins MUST be able to trigger an update via the Web Admin API.

### Non-Functional Requirements
*   **Safety**: Updates MUST NOT corrupt the existing installation on failure.
*   **Security**: Updates MUST be signed (Minisign or GPG).
*   **Performance**: Embedded assets SHOULD be served with control headers (ETag/Cache-Control).

## User Stories
*   **As an Admin**, I want to download a single `.exe` or binary and have the full mail server and UI working immediately.
*   **As an Admin**, I want to click "Update" in the web dashboard to upgrade to the latest version without manual file replacement.

## Out of Scope
*   Automatic background updates (should be manual trigger for now).
*   Updating the configuration file structure automatically (migration is separate).
