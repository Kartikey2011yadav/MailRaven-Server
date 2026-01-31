# Tasks: Client Bundling & Auto-Updates

**Reviewer**: Copilot
**Feature**: Client Bundling & Auto-Updates (Feature 011)

## Phase 1: Client Embedding
**Goal**: Single binary distribution.

- [x] T001 Update `Makefile` to include `build-client` target (npm install & build)
- [x] T002 Update `client/vite.config.ts` to ensure build output is compatible (correct output dir)
- [x] T003 Create `internal/adapters/http/static/fs.go` with `//go:embed` directive
- [x] T004 Implement `SPAFileSystem` wrapper in `internal/adapters/http/static/spa.go` to handle `index.html` fallback
- [x] T005 Update `internal/adapters/http/server.go` to mount the static file handler
- [x] T006 Verify: Build binary, run, and access `http://localhost:8080/` to see the React app

## Phase 2: Updater Core
**Goal**: Safe self-update capability.

- [x] T007 Add dependencies `github.com/minio/selfupdate` and `github.com/jedisct1/go-minisign`
- [x] T008 Define `UpdateManager` interface in `internal/core/ports/system.go`
- [x] T009 Implement `GitHubUpdater` in `internal/adapters/updater/github.go` (Check, Download, Verify)
- [x] T010 Implement `ApplyUpdate` logic (Windows-safe binary replacement)
- [x] T011 Create `minisign` public key constant in `internal/config` (or embedded)

## Phase 3: Update API
**Goal**: Admin control via Web UI.

- [x] T012 Implement HTTP handlers `CheckUpdate` and `PerformUpdate` in `internal/adapters/http/handlers/system.go`
- [x] T013 Register new routes in `internal/adapters/http/server.go`
- [x] T014 Integration: Verify API triggers download and replacement (Mocked for test)

## Implementation Strategy
Start with embedding (Phase 1) as it simplifies the dev workflow and is low risk. Phase 2 requires careful testing on Windows due to file locking.
