# Implementation Plan: Web Admin UI & Domain Management

**Branch**: `004-web-admin` | **Date**: 2026-01-27 | **Spec**: [specs/004-web-admin/spec.md](spec.md)
**Input**: Feature specification from `/specs/004-web-admin/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Build a React-based Single Page Application (SPA) for system administration, deployable on Vercel. Implement backend APIs for System Stats and Domain Management to support the UI and multi-tenancy.

## Technical Context

**Language/Version**: Go 1.22+ (Backend), TypeScript/React 18+ (Frontend)
**Primary Dependencies**: 
- Backend: `chi` (Router), `sqlite` (Storage), `jwt` (Auth)
- Frontend: `Vite` (Build tool), `React`, `Tailwind CSS`, `shadcn/ui` (Components), `lucide-react` (Icons), `axios`/`fetch` (API Client)
**Storage**: SQLite (adding `domains` table)
**Testing**: `testify` (Go), `vitest`/`playwright` (Frontend - TBD)
**Target Platform**: 
- Backend: Linux/Windows Server (Go binary)
- Frontend: Vercel (Static/Edge)
**Project Type**: Web Application (Backend API + Decoupled Frontend)
**Performance Goals**: Dashboard load < 500ms
**Constraints**: 
- Frontend must consume API over HTTP (requires CORS config on Backend).
- Multi-domain support must strictly validate sender identity.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Verify compliance with [MailRaven Constitution](../../.specify/memory/constitution.md):

- [x] **Reliability**: Stats API is read-only; Domain management uses atomic transactions (SQLite).
- [x] **Protocol Parity**: Domain management doesn't break RFC compliance (just restricts who can send/receive).
- [x] **Mobile-First Architecture**: API responses are JSON, minimal payload size.
- [x] **Dependency Minimalism**: Frontend uses standard modern stack (React/Vite); Backend adds no new heavy libs.
- [x] **Observability**: Admin actions (add/remove domain) log structured events.
- [x] **Interface-Driven Design**: `DomainRepository` interface will be defined.
- [x] **Extensibility**: Domain logic decoupled from core SMTP via Repository.
- [x] **Protocol Isolation**: Admin API separated from SMTP/IMAP logic.
- [x] **Testing Standards**: E2E tests for Admin API; Component tests for React (optional but recommended).

**Violations Requiring Justification**: "None"

## Project Structure

### Documentation (this feature)

```text
specs/004-web-admin/
├── plan.md              # This file
├── research.md          # [Phase 0 output](research.md)
├── data-model.md        # [Phase 1 output](data-model.md)
├── quickstart.md        # [Phase 1 output](quickstart.md)
├── contracts/           # [Phase 1 output](contracts/openapi.yaml)
└── tasks.md             # Phase 2 output
```

### Source Code (repository root)

```text
# Backend
internal/
├── core/
│   ├── domain/           # Add domain.go
│   └── ports/            # Update repositories.go
├── adapters/
│   ├── http/
│   │   ├── handlers/     # Add admin_stats.go, admin_domains.go
│   │   └── middleware/   # Update cors.go (if needed)
│   └── storage/
│       └── sqlite/       # Add domain_repo.go, updates/003_domains.sql

# Frontend (New)
web/                      # Root for Vercel deployment
├── public/
├── src/
│   ├── components/       # ui (shadcn), layout, features
│   ├── lib/              # utils, api client
│   ├── pages/            # Dashboard, Users, Domains, Login
│   ├── App.tsx
│   └── main.tsx
├── package.json
├── tailwind.config.js
├── tsconfig.json
└── vite.config.ts
```
