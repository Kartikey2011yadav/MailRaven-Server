# MailRaven Architecture Documentation

MailRaven is a modern, modular email server designed with a "Mobile-First" philosophy. Unlike traditional mail servers that primarily serve IMAP/POP3 clients, MailRaven exposes a rich RESTful JSON API optimized for mobile devices, enabling features like delta sync, push notifications (future), and server-side search.

## High-Level Overview

MailRaven follows the **Ports and Adapters (Hexagonal)** architecture. This ensures that the core business logic (email processing, storage rules) is decoupled from external interfaces (SMTP listeners, HTTP APIs, Database implementations).

```mermaid
graph TD
    Client[Mobile Client / Web App] -->|HTTP/JSON| API[API Layer]
    SMTP_In[External SMTP Servers] -->|SMTP| Listener[SMTP Listener]
    
    subgraph "Core Domain"
        API --> Service[Email Service]
        Listener --> Service
        Service --> Domain[Domain Logic]
    end
    
    subgraph "Infrastructure / Adapters"
        Service --> Repo[Repository Port]
        Repo --> SQLite[SQLite Adapter]
        Repo --> FS[Blob Storage Adapter]
        Service --> FTS[Search Port]
        FTS --> SQLiteFTS[SQLite FTS5]
    end
```

## Layers

### 1. Interface Layer (Primary Adapters)
- **SMTP Listener**: Listens on port 25/587. Handles the SMTP protocol state machine (RFC 5321).
  - Implementation: `internal/adapters/smtp`
- **HTTP API**: Exposes endpoints for email retrieval, sending, and management.
  - Implementation: `internal/adapters/http`
  - Authentication: JWT-based.

### 2. Core Layer (Business Logic)
- **Services**: Orchestrate the flow of data.
  - `EmailService`: Handles receiving emails, validating them (SPF/DKIM/DMARC), and storing them.
  - `OutboundService`: Manages the sending queue and delivery to remote servers.
- **Ports**: Interfaces defining how the core interacts with the outside world (e.g., `EmailRepository`, `BlobStore`).
- **Domain**: Pure Go structs representing `Email`, `User`, `Thread`.

### 3. Infrastructure Layer (Secondary Adapters)
- **Storage**:
  - **Metadata**: Stored in SQLite.
  - **Blobs (Email Bodies)**: Stored in the filesystem, gzip compressed, partitioned by date (`YYYY/MM/DD`).
- **Search**:
  - Implemented using SQLite's FTS5 extension.
  - Indexes headers, body snippets, and recipients.

## Key Decisions

- **Why JSON API and not IMAP?**
  IMAP is chatty and hard to optimize for battery-constrained mobile devices. A JSON API allows for batching, partial updates, and simplified parsing on the client side.
- **Why SQLite?**
  For a self-hosted single-tenant or small multi-tenant server, SQLite provides excellent performance, zero operational overhead, and easy backups (just copy the file).
- **Blob Storage Separation**:
  Keeping binary email content out of the database keeps the index small and fast. Atomic writes to the filesystem ensure data integrity.

## Directory Structure

- `cmd/`: Entry points (main application).
- `internal/`: Private application code.
  - `core/`: Domain logic and interfaces.
  - `adapters/`: Implementations (HTTP, SMTP, SQLite).
- `deployment/`: Configuration and systemd files.
- `specs/`: Planning and design documents.
- `mox/`: (Reference) The Mox codebase used for comparison and inspiration.
