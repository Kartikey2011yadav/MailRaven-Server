# Plan: Client Bundling & Auto-Updates

## Architecture

### 1. Static Asset Serving
We will use Go 1.16+ `embed` package.
*   **Location**: `internal/adapters/http/static/`
*   **Mechanism**: A `fs.FS` wrapper that handles the `index.html` fallback for SPA routing.
*   **Integration**: wired into `internal/adapters/http/server.go` as a fallback handler (after API routes).

### 2. Update Mechanism
We will use a library like `github.com/minio/selfupdate` or implement a custom robust swapper (safe for Windows).
*   **Source**: GitHub Releases (via GitHub API).
*   **Verification**: Minisign (`.minisig` files).
*   **Flow**:
    1.  GET releases (filter by Tag > current Version).
    2.  Download binary + signature.
    3.  Verify signature against embedded public key.
    4.  `selfupdate.Apply` (moves old binary to .old, writes new).
    5.  Restart process (or instruct user to restart).

## Components

1.  **Build Pipeline** (`Makefile`):
    *   Target `build-client`: `cd client && npm install && npm run build`
    *   Target `build`: `build-client` -> `go build`.

2.  **HTTP Adapter**:
    *   Modify `server.go` to accept an `fs.FS` for the frontend.
    *   Add `SPAHandler` middleware.

3.  **Core Domain**:
    *   `UpdateService`: Business logic for checking/verifying.

4.  **Admin API**:
    *   `POST /admin/system/update`

## Phase 1: Client Embedding (Blocking)
*   Ensure the React app can be built and embedded.
*   Serve it correctly.
*   Verify API still works.

## Phase 2: Update Logic
*   Implement the self-update logic.
*   Add the API endpoints.
*   Test on Windows/Linux.

## Dependencies
*   `github.com/minio/selfupdate` (or similar)
*   `jedisct1/go-minisign` (for verification)
