# Quickstart: Archive and Spam Features

## 1. List Messages

You can now filter messages by mailbox, read status, starred status, and date.

**Get Inbox (default):**
```http
GET /api/v1/messages
```

**Get Archived Messages:**
```http
GET /api/v1/messages?mailbox=Archive
```

**Get Starred Messages (Important):**
```http
GET /api/v1/messages?is_starred=true
```

**Get Recent Unread Messages:**
```http
GET /api/v1/messages?is_read=false&start_date=2024-01-01T00:00:00Z
```

## 2. Managing Messages

**Archive a Message:**
```http
PATCH /api/v1/messages/{id}
Content-Type: application/json

{
  "mailbox": "Archive"
}
```

**Star a Message:**
```http
PATCH /api/v1/messages/{id}
Content-Type: application/json

{
  "is_starred": true
}
```

## 3. Spam Management

**Report as Spam:**
Moves message to "Junk" and trains the filter.
```http
POST /api/v1/messages/{id}/spam
```

**Report Not Spam (Ham):**
Moves message to "Inbox" and trains the filter.
```http
POST /api/v1/messages/{id}/ham
```
