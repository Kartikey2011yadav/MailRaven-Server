# Specification: User Portal & Webmail

**Feature ID**: 012
**Status**: Draft
**Priority**: High

## Context
Currently, MailRaven provides a Web Admin UI for administrators but lacks any interface for end-users. Users cannot change their passwords, configure auto-replies (Vacation), or read email without a third-party IMAP client. The reference implementation (`mox`) provides a full suite: Webmail, Web Account, and Web Admin. To be competitive and usable, we must provide at least basic user self-service and web access.

## Goals
1.  **User Self-Service**: Allow standard users to log in, change their password, and configure Vacation/Sieve rules.
2.  **Basic Webmail**: Provide a "Lite" email client for emergency access or simple usage (Read/Reply/Send).
3.  **Unified Capability**: Integrate these features into the existing React SPA (`client/`), routed by user role.

## Requirements

### Functional Requirements

#### F1: User Authentication & Routing
*   The login page MUST distinguish between Admin and Standard User sessions (or unified session with roles).
*   Standard Users MUST be routed to `/mail` (Webmail) upon login.
*   Admins MUST be routed to `/admin` (Dashboard) but OPTIONALLY be able to switch to Webmail view.

#### F2: User Settings (The "Account" Portal)
*   **Password Change**: User can update their credentials.
*   **Vacation Auto-Reply**: User can enable/disable vacation mode, set subject/body, and end date. (Interacts with Sieve API from Feature 010).

#### F3: Webmail - Read
*   **Inbox List**: Display paginated list of messages (Subject, Sender, Date, Unread Status).
*   **Message View**: Display email content. Prefer HTML, fallback to Text. Sanitize HTML to prevent XSS.
*   **Attachments**: List attachments and allow download.

#### F4: Webmail - Write
*   **Compose**: Simple editor for sending emails.
*   **Reply/Forward**: Pre-fill subject and body.

### Non-Functional Requirements
*   **Security**: All HTML rendering MUST be sanitized (DOMPurify).
*   **Performance**: Message lists MUST be paginated (cursor-based).
*   **Responsive**: UI MUST work on Mobile devices (Mobile-First philosophy).

## User Stories
*   **As a User**, I want to log in to change my password so I don't have to ask an admin.
*   **As a User**, I want to set an "Out of Office" reply before I go on vacation.
*   **As a Traveler**, I want to check my email from a hotel computer without installing an IMAP client.

## Out of Scope
*   Advanced Webmail features (Drag & Drop folders, multi-select mass actions, rich text editor image uploads).
*   Calendar/Contacts (JMAP/CardDAV/CalDAV) - Future feature.
*   Import/Export of Mbox files (Mox has this, we defer it).

## Dependencies
*   **Feature 010 (Sieve)**: Required for Vacation settings.
*   **Feature 011 (Client Bundling)**: Required to serve the updated UI.
