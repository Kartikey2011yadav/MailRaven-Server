# MailRaven Client Developer Handover

**Version**: 1.0.0
**API Version**: v1
**Target Platform**: Kotlin Multiplatform (KMP)

## Overview

This document is intended for the **Client Developer** (or AI Agent) building the `MailRaven` KMP mobile client. It details how to connect to the backend, the authentication flow, and the expected API behavior.

## Infrastructure

### Base URLs
- **Local Dev**: `http://localhost:8080/api/v1` (Backend)
- **Local HTTPS**: `https://localhost:8443/api/v1` (If TLS enabled)
- **Live Demo**: `https://api.mailraven.net/v1` (Example)

### Dependencies
The client should support:
- **HTTP Client**: Ktor recommended.
- **Serialization**: Kotlinx.serialization (JSON).
- **Storage**: SQLDelight or secure collection storage for JWT.

## Authentication Flow

MailRaven uses **JWT** (JSON Web Tokens).

1.  **Login**: `POST /auth/login`
    ```json
    { "email": "user@domain.com", "password": "secure_password" }
    ```
    **Response**:
    ```json
    { "token": "eyJhbG...", "expires_at": "2026-01-30T12:00:00Z" }
    ```
    *Action*: Store `token` securely (Keychain/EncryptedSharedPreferences).

2.  **Authenticated Requests**:
    Add header: `Authorization: Bearer <token>`

3.  **Refresh**: `POST /auth/refresh`
    *Action*: Call when 401 Unauthorized is received.

## Key Sync Strategy (Mobile-First)

MailRaven supports **Delta Sync** to minimize bandwidth.

1.  **Initial Fetch**: `GET /messages`
    *Returns*: List of messages + `next_cursor` (if paginated).

2.  **Sync**:
    *The generic `/messages` endpoint currently supports cursor-based pagination.*
    *Future Protocol*: Client should request `GET /sync?since=<timestamp>` (To be implemented in Feature 007).
    *Current Strategy*: Fetch page 1. Compare IDs with local DB. Stop when ID exists locally.

## API Contracts (OpenAPI)

Refer to: `specs/004-web-admin/contracts/openapi.yaml` (most up-to-date spec).

### Key Entities

**Message**:
```kotlin
@Serializable
data class Message(
    val id: String,
    val subject: String,
    val from: String, // "Name <email@addr.com>"
    val to: List<String>,
    val snippet: String,
    val date: String, // ISO8601
    val hasAttachment: Boolean,
    val read: Boolean,
    val folder: String // "inbox", "sent", "trash"
)
```

## Setup for Client Dev

1.  **Start Backend**:
    ```bash
    git clone https://github.com/Kartikey2011yadav/MailRaven-Server.git
    cd MailRaven-Server
    # Windows
    .\scripts\setup.ps1
    # Run
    .\bin\mailraven.exe serve
    ```

2.  **Seed Data**:
    Send yourself emails using `swaks` or another SMTP tool to populate `http://localhost:8080`.

## Roadmap Alignment

- **Phase 1 (Current)**:
    - Login / Logout.
    - List Inbox.
    - View Message Details.
    - Compose / Send (Simple Text).

- **Phase 2 (Upcoming)**:
    - Push Notifications (FCM).
    - Offline Support (Local DB sync).
    - Attachments.
