# Feature Specification: Archive and Spam Mechanism

**Feature Branch**: `014-archive-spam-handling`  
**Created**: 2026-02-10  
**Status**: Draft  
**Input**: User description: "implement archive and spam mechanism"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Archive Message (Priority: P1)

Users encounter messages they want to keep but remove from their Inbox to declutter.

**Why this priority**: Core email workflow functionality. Essential for Inbox Zero workflows.

**Independent Test**: Can be tested by archiving a message and verifying it moves out of Inbox listing but remains searching/retrievable in Archive folder.

**Acceptance Scenarios**:

1. **Given** a message in "Inbox", **When** user chooses to "Archive" it, **Then** message moves to "Archive" location.
2. **Given** a message in "Archive", **When** user views the "Inbox", **Then** message is NOT visible.
3. **Given** a message in "Archive", **When** user views the "Archive" folder, **Then** message IS visible.

---

### User Story 2 - Report Spam (Mark as Junk) (Priority: P1)

Users identifying unsolicited messages want to move them out of their way and prevent future occurrences.

**Why this priority**: Critical for security and user experience. Helps improve the global spam filter.

**Independent Test**: Mark a message as spam; verify it moves to Junk and system logs indicate spam training occurred.

**Acceptance Scenarios**:

1. **Given** a message in "Inbox", **When** user reports it as Spam, **Then** message moves to "Junk" folder.
2. **Given** a message in "Inbox", **When** user reports it as Spam, **Then** system trains the spam filter with this message.

---

### User Story 3 - Report Not Spam (Mark as Ham) (Priority: P2)

Users find a legitimate message incorrectly filed in Junk/Spam folder.

**Why this priority**: Essential for correcting false positives (which cause missed important emails).

**Independent Test**: Mark a message in Junk as "Not Spam"; verify it moves to Inbox and system logs indicate ham training.

**Acceptance Scenarios**:

1. **Given** a message in "Junk", **When** user marks it as Not Spam, **Then** message moves to "Inbox" folder.
2. **Given** a message in "Junk", **When** user marks it as Not Spam, **Then** system trains the filter with this message as safe (ham).

---

### User Story 4 - Mark as Important (Star) (Priority: P2)

Users want to highlight specific messages to find them easily later.

**Why this priority**: Standard email feature for organization. 

**Independent Test**: Star a message, filter by "Starred", verify message appears.

**Acceptance Scenarios**:

1. **Given** a message, **When** user clicks "Star", **Then** message is marked as important.
2. **Given** a list of messages, **When** user filters by "Starred", **Then** only starred messages appear.

### Edge Cases

- **Already trained**: What if user toggles spam/ham multiple times? (System should ideally handle Re-training or Un-training, but for MVP re-training is acceptable).
- **Missing content**: If message body isn't available (deleted from blob store), training should fail gracefully but move should succeed.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Users MUST be able to move messages to standard folders (Archive, Trash, Junk, Inbox).
- **FR-002**: Users MUST be able to "Report Spam", which automatically moves the message to Junk and marks it for spam filter training.
- **FR-003**: Users MUST be able to "Report Not Spam" (Ham), which moves the message to Inbox and marks it for ham filter training.
- **FR-004**: Users MUST be able to view messages filtered by their folder/mailbox location.
- **FR-005**: Moving a message to "Archive" MUST simply change its folder location without triggering spam training.
- **FR-006**: Users MUST be able to filter messages by: Read/Unread status, Starred status, and Date Range.
- **FR-007**: Users MUST be able to toggle a "Starred" (Important) flag on any message.

### Success Criteria

- **Efficiency**: Archive/Spam actions must complete within 500ms for the user.
- **Reliability**: 100% of "Report Spam" actions must result in the message being added to the spam training queue.
- **Usability**: Users can reverse a spam decision (move back to Inbox) with a single action.

### Key Entities

- **Message**: Represents the email, now with `Mailbox` (Folder) and `IsStarred` properties.
- **Folder/Mailbox**: Represents a container for messages (e.g., "Inbox", "Archive", "Junk").

## Assumptions

- We assume standard mailbox names: "Inbox", "Archive", "Junk", "Trash", "Sent", "Drafts".
- The client application is responsible for presenting the folder names to the user.
- Spam training happens successfully if the message content is available.
