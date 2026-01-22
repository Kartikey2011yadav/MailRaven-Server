package observability

import (
	"sync"
	"time"
)

// Metrics holds Prometheus-style metrics for MailRaven
// For MVP, using in-memory counters. In production, integrate with prometheus/client_golang
type Metrics struct {
	mu sync.RWMutex

	// SMTP metrics
	MessagesReceived  int64
	MessagesRejected  int64
	SMTPConnections   int64
	SMTPErrors        int64

	// API metrics
	APIRequests       int64
	APIErrors         int64
	
	// Storage metrics
	StorageWrites     int64
	StorageReads      int64
	StorageErrors     int64

	// Request duration histogram (simplified for MVP)
	APIRequestDurations []time.Duration
}

// NewMetrics creates a new metrics collector
func NewMetrics() *Metrics {
	return &Metrics{
		APIRequestDurations: make([]time.Duration, 0, 1000),
	}
}

// IncrementMessagesReceived increments the messages received counter
func (m *Metrics) IncrementMessagesReceived() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.MessagesReceived++
}

// IncrementMessagesRejected increments the messages rejected counter
func (m *Metrics) IncrementMessagesRejected() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.MessagesRejected++
}

// IncrementSMTPConnections increments the SMTP connections counter
func (m *Metrics) IncrementSMTPConnections() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.SMTPConnections++
}

// IncrementSMTPErrors increments the SMTP errors counter
func (m *Metrics) IncrementSMTPErrors() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.SMTPErrors++
}

// IncrementAPIRequests increments the API requests counter
func (m *Metrics) IncrementAPIRequests() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.APIRequests++
}

// IncrementAPIErrors increments the API errors counter
func (m *Metrics) IncrementAPIErrors() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.APIErrors++
}

// IncrementStorageWrites increments the storage writes counter
func (m *Metrics) IncrementStorageWrites() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.StorageWrites++
}

// IncrementStorageReads increments the storage reads counter
func (m *Metrics) IncrementStorageReads() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.StorageReads++
}

// IncrementStorageErrors increments the storage errors counter
func (m *Metrics) IncrementStorageErrors() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.StorageErrors++
}

// RecordAPIRequestDuration records an API request duration
func (m *Metrics) RecordAPIRequestDuration(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Keep last 1000 durations for simple P95 calculation
	if len(m.APIRequestDurations) >= 1000 {
		m.APIRequestDurations = m.APIRequestDurations[1:]
	}
	m.APIRequestDurations = append(m.APIRequestDurations, duration)
}

// GetSnapshot returns a read-only snapshot of current metrics
func (m *Metrics) GetSnapshot() MetricsSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return MetricsSnapshot{
		MessagesReceived:    m.MessagesReceived,
		MessagesRejected:    m.MessagesRejected,
		SMTPConnections:     m.SMTPConnections,
		SMTPErrors:          m.SMTPErrors,
		APIRequests:         m.APIRequests,
		APIErrors:           m.APIErrors,
		StorageWrites:       m.StorageWrites,
		StorageReads:        m.StorageReads,
		StorageErrors:       m.StorageErrors,
		RequestDurationCount: len(m.APIRequestDurations),
	}
}

// MetricsSnapshot is a read-only view of metrics at a point in time
type MetricsSnapshot struct {
	MessagesReceived     int64
	MessagesRejected     int64
	SMTPConnections      int64
	SMTPErrors           int64
	APIRequests          int64
	APIErrors            int64
	StorageWrites        int64
	StorageReads         int64
	StorageErrors        int64
	RequestDurationCount int
}
