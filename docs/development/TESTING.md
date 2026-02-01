# Testing Guide

This document outlines the testing strategy, setup instructions, and execution implementation for the MailRaven project.

## Overview

The project uses a comprehensive testing approach divided into two main categories:

1.  **End-to-End (E2E) Testing**: Uses **Playwright** to verify the full user journey via the React Frontend, interacting with a real running Backend.
2.  **Backend Testing**: Uses Go's standard `testing` package for unit and integration tests of the API and Core logic.

---

## 1. End-to-End (E2E) Testing

E2E tests ensure that the Frontend and Backend communicate correctly and that key user flows (Login, Navigation, etc.) function as expected.

### 📍 Location
*   **Tests**: `client/tests/`
*   **Config**: `client/playwright.config.ts`

### ⚙️ Environment Setup

Before running E2E tests, the "Complete System" (Frontend + Backend + Database) must be running locally.

#### Prerequisites
1.  **Ports**: Ensure ports `8081` (API) and `2526` (SMTP) are free. Modify `config.yaml` if needed.
2.  **Database**: Start with a known state using the seeding script.

#### Step 1: Seed the Database
Initialize the database with the default Admin user.
```powershell
# From project root
go run scripts/seed_admin.go
```
*   **User**: `admin@example.com`
*   **Password**: `admin123`

#### Step 2: Start the Backend Server
```powershell
# From project root
go run ./cmd/mailraven serve --config config.yaml
```

#### Step 3: Start the Frontend Server
We must point the frontend to the correct API port (8081) if changed in config.
```powershell
# From client/ directory
$env:VITE_API_URL="http://localhost:8081/api/v1"
npm run dev
```

### 🚀 Running the Tests
Once both servers are running, execute Playwright from the `client/` directory.

```powershell
cd client

# Run all tests (headless mode)
npx playwright test

# Run a specific test file
npx playwright test login.spec.ts

# Run in UI mode (interactive debugger)
npx playwright test --ui
```

### 📊 Viewing Reports
If tests fail, or to see a detailed breakdown:
```powershell
npx playwright show-report
```

### 🧪 Current Test Suites
*   `login.spec.ts`: 
    *   Verifies successful login with valid credentials.
    *   Verifies error messages for invalid credentials.
*   `navigation.spec.ts`:
    *   Verifies sidebar navigation to "Domains" and "Users" pages.
    *   Checks for correct routing and page headers.

---

## 2. Backend Testing

Backend tests cover the Go business logic, API handlers, and database interactions.

### 📍 Location
*   **Unit Tests**: Co-located with code (e.g., `internal/core/services/*_test.go`).
*   **Integration Tests**: `tests/` folder (if applicable) or specifically marked tests.

### 🚀 Running Tests
```powershell
# Run all tests recursively
go test ./...

# Run tests with verbose output
go test -v ./...
```

### 📈 Coverage
Generate a coverage report to see which parts of the code are untested.
```powershell
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

## ⚠️ Troubleshooting Common Issues

### "Connection Refused" / Test Timeouts
*   **Cause**: The Frontend cannot talk to the Backend, or Playwright cannot reach the Frontend.
*   **Fix**: 
    1. Check if `go run cmd/mailraven` is running and healthy (no panic logs).
    2. Verify `VITE_API_URL` is set correctly in the Frontend terminal.
    3. Ensure `config.yaml` ports match what you expect.

### "Table users already exists" (Migration Error)
*   **Cause**: The server tries to run migrations on start, but the DB is already initialized.
*   **Fix**: This is usually a warning and can be ignored. To reset completely, delete `data/mailraven.db` and re-run the `seed_admin.go` script.

### "Unique constraint failed"
*   **Cause**: Running `seed_admin.go` twice.
*   **Fix**: The script handles this gracefully, but you can ignore it if the user exists.

### Port Conflicts (8080/2525 in use)
*   **Fix**: Use PowerShell to find and kill the process:
    ```powershell
    netstat -ano | findstr ":8080"
    Stop-Process -Id <PID> -Force
    ```
