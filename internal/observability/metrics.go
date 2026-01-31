package observability

import (
	"fmt"
	"io"
	"sync"
	"time"
)

// Metrics holds Prometheus-style metrics for MailRaven
// For MVP, using in-memory counters. In production, integrate with prometheus/client_golang
type Metrics struct {
	mu sync.RWMutex

	// SMTP metrics
	MessagesReceived int64
	MessagesRejected int64
	SMTPConnections  int64
	SMTPErrors       int64

	// API metrics
	APIRequests int64
	APIErrors   int64

	// Storage metrics
	StorageWrites int64
	StorageReads  int64
	StorageErrors int64

	// Outbound metrics
	OutboundEnqueued        int64
	OutboundSent            int64
	OutboundFailedTransient int64
	OutboundFailedPermanent int64

	// Spam metrics
	SpamDetected    int64
	GreylistBlocked int64

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

// IncrementOutboundEnqueued increments the outbound enqueued counter
func (m *Metrics) IncrementOutboundEnqueued() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.OutboundEnqueued++
}

// IncrementOutboundSent increments the outbound sent counter
func (m *Metrics) IncrementOutboundSent() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.OutboundSent++
}

// IncrementOutboundFailedTransient increments the outbound failed transient counter
func (m *Metrics) IncrementOutboundFailedTransient() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.OutboundFailedTransient++
}

// IncrementOutboundFailedPermanent increments the outbound failed permanent counter
func (m *Metrics) IncrementOutboundFailedPermanent() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.OutboundFailedPermanent++
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
		MessagesReceived:        m.MessagesReceived,
		MessagesRejected:        m.MessagesRejected,
		SMTPConnections:         m.SMTPConnections,
		SMTPErrors:              m.SMTPErrors,
		APIRequests:             m.APIRequests,
		APIErrors:               m.APIErrors,
		StorageWrites:           m.StorageWrites,
		StorageReads:            m.StorageReads,
		StorageErrors:           m.StorageErrors,
		OutboundEnqueued:        m.OutboundEnqueued,
		OutboundSent:            m.OutboundSent,
		OutboundFailedTransient: m.OutboundFailedTransient,
		OutboundFailedPermanent: m.OutboundFailedPermanent,
		SpamDetected:            m.SpamDetected,
		GreylistBlocked:         m.GreylistBlocked,
		RequestDurationCount:    len(m.APIRequestDurations),
	}
}

// WritePrometheus writes metrics in Prometheus text format to w
func (m *Metrics) WritePrometheus(w io.Writer) {
	snap := m.GetSnapshot()

	writeMetric := func(name, help, typeStr string, value int64) {
		fmt.Fprintf(w, "# HELP %s %s\n", name, help)
		fmt.Fprintf(w, "# TYPE %s %s\n", name, typeStr)
		fmt.Fprintf(w, "%s %d\n", name, value)
	}

	writeMetric("mailraven_messages_received_total", "Total incoming messages received", "counter", snap.MessagesReceived)
	writeMetric("mailraven_messages_rejected_total", "Total incoming messages rejected", "counter", snap.MessagesRejected)
	writeMetric("mailraven_smtp_connections_total", "Total SMTP connections accepted", "counter", snap.SMTPConnections)
	writeMetric("mailraven_smtp_errors_total", "Total SMTP errors encountered", "counter", snap.SMTPErrors)

	writeMetric("mailraven_api_requests_total", "Total HTTP API requests", "counter", snap.APIRequests)
	writeMetric("mailraven_api_errors_total", "Total HTTP API errors", "counter", snap.APIErrors)

	writeMetric("mailraven_storage_writes_total", "Total storage write operations", "counter", snap.StorageWrites)
	writeMetric("mailraven_storage_reads_total", "Total storage read operations", "counter", snap.StorageReads)
	writeMetric("mailraven_storage_errors_total", "Total storage errors", "counter", snap.StorageErrors)

	writeMetric("mailraven_outbound_enqueued_total", "Total messages enqueued for outbound delivery", "counter", snap.OutboundEnqueued)
	writeMetric("mailraven_outbound_sent_total", "Total messages successfully delivered", "counter", snap.OutboundSent)
	writeMetric("mailraven_outbound_transient_failures_total", "Total transient outbound delivery failures", "counter", snap.OutboundFailedTransient)
	writeMetric("mailraven_outbound_permanent_failures_total", "Total permanent outbound delivery failures", "counter", snap.OutboundFailedPermanent)

	writeMetric("mailraven_spam_detected_total", "Total messages classified as spam", "counter", snap.SpamDetected)
	writeMetric("mailraven_greylist_blocked_total", "Total connections blocked by greylisting", "counter", snap.GreylistBlocked)
}

// IncrementSpamDetected increments the spam detected counter
func (m *Metrics) IncrementSpamDetected() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.SpamDetected++
}

// IncrementGreylistBlocked increments the greylist blocked counter
func (m *Metrics) IncrementGreylistBlocked() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.GreylistBlocked++
}

// MetricsSnapshot is a read-only view of metrics at a point in time
type MetricsSnapshot struct {
	MessagesReceived        int64
	MessagesRejected        int64
	SMTPConnections         int64
	SMTPErrors              int64
	APIRequests             int64
	APIErrors               int64
	StorageWrites           int64
	StorageReads            int64
	StorageErrors           int64
	OutboundEnqueued        int64
	OutboundSent            int64
	OutboundFailedTransient int64
	OutboundFailedPermanent int64
	SpamDetected            int64
	GreylistBlocked         int64
	RequestDurationCount    int
}
