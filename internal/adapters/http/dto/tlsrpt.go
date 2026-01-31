package dto

import (
	"encoding/json"
	"time"

	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
	"github.com/google/uuid"
)

// TLSReportRequest represents the incoming JSON body for RFC 8460 reports
type TLSReportRequest struct {
	OrganizationName string                 `json:"organization-name"`
	DateRange        DateRangeRequest       `json:"date-range"`
	ContactInfo      string                 `json:"contact-info"`
	ReportID         string                 `json:"report-id"`
	Policies         []domain.PolicySummary `json:"policies"`
}

type DateRangeRequest struct {
	StartDatetime time.Time `json:"start-datetime"`
	EndDatetime   time.Time `json:"end-datetime"`
}

// MapToDomain converts the DTO to the core Domain entity
// rawJSON is passed separately because DTO decoding consumes the stream
func (req *TLSReportRequest) MapToDomain(rawJSON []byte) *domain.TLSReport {
	return &domain.TLSReport{
		ID:       uuid.New(),
		ReportID: req.ReportID,
		Provider: req.OrganizationName,
		DateRange: domain.DateRange{
			StartDatetime: req.DateRange.StartDatetime,
			EndDatetime:   req.DateRange.EndDatetime,
		},
		ContactInfo:  req.ContactInfo,
		Policies:     req.Policies,
		RawJSON:      json.RawMessage(rawJSON),
		IngestedAt:   time.Now(),
		TotalCount:   calculateTotal(req.Policies),
		SuccessCount: calculateSuccess(req.Policies),
		FailureCount: calculateFailure(req.Policies),
	}
}

func calculateTotal(policies []domain.PolicySummary) int64 {
	var total int64
	for _, p := range policies {
		total += p.Summary.TotalSuccessfulSessionCount + p.Summary.TotalFailureSessionCount
	}
	return total
}

func calculateSuccess(policies []domain.PolicySummary) int64 {
	var total int64
	for _, p := range policies {
		total += p.Summary.TotalSuccessfulSessionCount
	}
	return total
}

func calculateFailure(policies []domain.PolicySummary) int64 {
	var total int64
	for _, p := range policies {
		total += p.Summary.TotalFailureSessionCount
	}
	return total
}
