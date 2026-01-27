# Quickstart: Web Admin Development

## Prerequisites
- Node.js 18+
- Go 1.22+

## Running the Backend
1. **Migrations**: Ensure DB is up to date (auto-runs on start, usually).
   ```bash
   go run cmd/mailraven/main.go
   ```

## Running the Frontend
1. Navigate to web directory:
   ```bash
   cd web
   ```
2. Install Dependencies:
   ```bash
   npm install
   ```
3. Start Dev Server:
   ```bash
   npm run dev
   ```
   (Access at http://localhost:5173)

## Environment Variables (.env)
Create `web/.env`:
```
VITE_API_URL=http://localhost:8443
```

## Deployment (Vercel)
1. Install Vercel CLI: `npm i -g vercel`
2. Run `vercel` in `web/` directory.
