package domain

import (
	"time"
)

// OutboundStatus represents the state of a message in the queue
type OutboundStatus string

const (
	QueueStatusPending    OutboundStatus = "PENDING"
	QueueStatusProcessing OutboundStatus = "PROCESSING"
	QueueStatusSent       OutboundStatus = "SENT"
	QueueStatusFailed     OutboundStatus = "FAILED" // Permanent failure
	QueueStatusRetrying   OutboundStatus = "RETRYING"
)

// OutboundMessage represents an email waiting to be delivered
type OutboundMessage struct {
	ID          string // UUID
	Sender      string
	Recipient   string
	BlobKey     string // Key in blob store for the full message content
	Status      OutboundStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
	NextRetryAt time.Time
	RetryCount  int
	LastError   string
}
