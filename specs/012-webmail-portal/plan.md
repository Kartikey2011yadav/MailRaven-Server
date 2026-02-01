# Plan: User Portal & Webmail

## Architecture

### Frontend (React)
We will expand the existing `client/` application.
*   **Routing**:
    *   `/login`: Existing (Unified).
    *   `/admin/*`: Protected Route (Role: Admin).
    *   `/mail/*`: Protected Route (Role: User | Admin).
        *   `/mail/inbox`: Message List.
        *   `/mail/message/:id`: Message Reader.
        *   `/mail/compose`: Composer.
    *   `/settings`: User Settings (Password, Sieve).

*   **State Management**:
    *   Extend `AuthContext` to handle non-admin users.
    *   Use `TanStack Query` (React Query) for fetching messages/threads.

### Backend (Go)
Existing APIs cover most needs, but some refinements are required.

*   **API Updates**:
    *   `POST /auth/login`: Ensure it returns Role (Admin vs User).
    *   `GET /api/v1/messages`: Ensure it filters by the *authenticated user* (already does, but verify).
    *   `POST /api/v1/messages/send`: Ensure it uses the *authenticated user's* email as From address (security).
    *   `PUT /api/v1/users/self/password`: New endpoint for self-service password change.

## Phases

### Phase 1: Foundation & Settings (The "Account" Gap)
**Goal**: Users can manage their account.
1.  Frontend: Update Router to handle "User" role.
2.  Backend: Add `update-self` endpoints.
3.  Frontend: Implement "Settings" page (Password, Vacation Form).

### Phase 2: Webmail - Read (Viewer)
**Goal**: Users can see their email.
1.  Frontend: Implement `MailLayout` (Sidebar with folders).
2.  Frontend: Implement `MessageList` component.
3.  Frontend: Implement `MessageReader` with HTML sanitization (`isomorphic-dompurify`).

### Phase 3: Webmail - Write (Sender)
**Goal**: Users can send email.
1.  Frontend: Implement `Composer` modal/page.
2.  Backend: Verify `SendHandler` enforces sender identity.

## Technical Decisions
*   **HTML Sanitization**: Essential for Webmail. We will use `dompurify` on the client side.
*   **Rich Text**: Use a lightweight editor like `tiptap` or just a `textarea` for MVP. Let's stick to `textarea` (Plain Text) or very basic HTML for MVP.
*   **Sieve Config**: The backend exposes raw script management. The Frontend needs a "Form" that generates the Sieve script for Vacation.

## Risks
*   **XSS**: Rendering email HTML is dangerous. CSP (Content Security Policy) must be strict.
