package domain

import "time"

// DateRangeFilter defines a time window for filtering
type DateRangeFilter struct {
	Start *time.Time
	End   *time.Time
}

// MessageFilter defines criteria for listing messages
type MessageFilter struct {
	Limit     int
	Offset    int
	Mailbox   string // "INBOX", "Archive", "Junk", "Trash", "Sent", "Drafts"
	IsRead    *bool  // nil = all, true = read only, false = unread only
	IsStarred *bool  // nil = all, true = starred only
	DateRange DateRangeFilter
}
