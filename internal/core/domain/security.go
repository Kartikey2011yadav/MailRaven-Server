package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// MTASTSPolicy represents the policy served at .well-known/mta-sts.txt
type MTASTSPolicy struct {
	Version string     `json:"version"` // Must be STSv1
	Mode    MTASTSMode `json:"mode"`    // testing, enforce, none
	MX      []string   `json:"mx"`      // List of allowed MX patterns
	MaxAge  int        `json:"max_age"` // Cache duration in seconds
}

type MTASTSMode string

const (
	MTASTSModeTesting MTASTSMode = "testing"
	MTASTSModeEnforce MTASTSMode = "enforce"
	MTASTSModeNone    MTASTSMode = "none"
)

// TLSReport represents an aggregate report from RFC 8460
type TLSReport struct {
	ID           uuid.UUID       `json:"id"`
	ReportID     string          `json:"report_id"`
	Provider     string          `json:"provider"` // Organization name
	DateRange    DateRange       `json:"date_range"`
	ContactInfo  string          `json:"contact_info,omitempty"`
	Policies     []PolicySummary `json:"policies"`
	RawJSON      json.RawMessage `json:"-"` // Stored separately, not part of regular JSON marshaling
	IngestedAt   time.Time       `json:"ingested_at"`
	TotalCount   int64           `json:"total_count"`
	SuccessCount int64           `json:"success_count"`
	FailureCount int64           `json:"failure_count"`
}

type DateRange struct {
	StartDatetime time.Time `json:"start-datetime"`
	EndDatetime   time.Time `json:"end-datetime"`
}

// PolicySummary corresponds to the policy-level aggregation in the report
type PolicySummary struct {
	Policy         PolicyDetails    `json:"policy"`
	Summary        SummaryCounts    `json:"summary"`
	FailureDetails []FailureDetails `json:"failure-details,omitempty"`
}

type PolicyDetails struct {
	PolicyType   string   `json:"policy-type"`
	PolicyString []string `json:"policy-string,omitempty"`
	PolicyDomain string   `json:"policy-domain"`
	MXHost       []string `json:"mx-host,omitempty"`
}

type SummaryCounts struct {
	TotalSuccessfulSessionCount int64 `json:"total-successful-session-count"`
	TotalFailureSessionCount    int64 `json:"total-failure-session-count"`
}

type FailureDetails struct {
	ResultType            string `json:"result-type"`
	SendingMTAIP          string `json:"sending-mta-ip"`
	ReceivingMXHostname   string `json:"receiving-mx-hostname,omitempty"`
	ReceivingMXHelo       string `json:"receiving-mx-helo,omitempty"`
	ReceivingIP           string `json:"receiving-ip,omitempty"`
	FailedSessionCount    int64  `json:"failed-session-count"`
	AdditionalInformation string `json:"additional-information,omitempty"`
	FailureReasonCode     string `json:"failure-reason-code,omitempty"`
}
