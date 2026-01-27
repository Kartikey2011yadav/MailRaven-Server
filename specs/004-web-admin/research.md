# Research: Web Admin UI & Domain Management

**Feature**: Web Admin UI (`004-web-admin`)
**Date**: 2026-01-27

## Decision Log

### 1. Frontend Architecture
- **Decision**: Decoupled React SPA (Vite) hosted on Vercel.
- **Rationale**:
  - **Scale**: Allows independent scaling of UI and API.
  - **Performance**: Vercel Edge Network for static assets.
  - **Security**: Separation of concerns; API credentials not baked into static build (env vars used).
  - **User Request**: Explicitly requested "deploying it on vercel".
- **Alternatives**:
  - *embedded*: Serving React static files from Go embed. (Rejected: user requested Vercel).
  - *SSR (Next.js)*: Too heavy for a simple admin dashboard; SPA + Client-side fetching is sufficient and cheaper for this use case.

### 2. UI Component Library
- **Decision**: Shadcn/UI + Tailwind CSS.
- **Rationale**:
  - **Customization**: Code-backed components (copy-paste) allow full control unlike installed libraries (MUI).
  - **Aesthetics**: "21st.dev" style implies modern, Vercel-like aesthetic which Shadcn provides out-of-box.
  - **Performance**: Lightweight, tree-shakeable.

### 3. API Communication & Security
- **Decision**: JWT in HTTP-Only Cookies (or LocalStorage if subdomain constraint prevents cookies).
  - *Refinement*: Since Vercel domain (`app.vercel.app`) != API domain (`api.mailraven.com`), HTTP-Only cookies require `SameSite=None; Secure`.
  - **Fallback**: Use LocalStorage for JWT in MVP to avoid complex cross-site cookie issues, or proxy via Vercel Rewrites.
  - **Selected Approach**: **Vercel Rewrites**. Configure `vercel.json` to proxy `/api/*` to the Go backend. This avoids CORS issues and allows cookie usage if needed, though LocalStorage is easiest for "Quickstart".
  - *Correction*: Codebase currently uses `Authorization: Bearer <token>`. We will stick to **Header-based Auth** (LocalStorage) for the React App to match existing CLI/API patterns without changing Backend Auth logic significantly (other than CORS).
- **CORS**: Backend MUST allow the Vercel app domain.

### 4. Domain Management Logic
- **Decision**: "Primary Domain" vs "Secondary Domains".
  - **Primary**: Defined in `config.yaml`. immutable via API (safeguard).
  - **Secondary**: Stored in SQLite `domains` table. mutable.
- **Validation**: All email checks must query both Config and DB.

## Open Questions Resolved

- **Q: How to proxy requests?**
  - A: Use `VITE_API_URL` env var. In dev: `vite.config.ts` proxy. In prod: Direct CORS or Vercel Rewrite. (Direct CORS is simpler for first iteration).

- **Q: "System Stats" Source?**
  - A:
    - Uptime: `time.Since(serverStartTime)`
    - Memory: `runtime.ReadMemStats`
    - Users: `COUNT(*)` from DB.
    - Domains: `COUNT(*) + 1`.
