package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/ports"
	"github.com/google/uuid"
)

// TLSRptRepository implements ports.TLSRptRepository using SQLite
type TLSRptRepository struct {
	db *sql.DB
}

// NewTLSRptRepository creates a new SQLite report repository
func NewTLSRptRepository(db *sql.DB) *TLSRptRepository {
	return &TLSRptRepository{db: db}
}

// Save stores an incoming TLS report
func (r *TLSRptRepository) Save(ctx context.Context, report *domain.TLSReport) error {
	id := report.ID
	if id == uuid.Nil {
		id = uuid.New()
		report.ID = id
	}
	if report.IngestedAt.IsZero() {
		report.IngestedAt = time.Now()
	}

	query := `
		INSERT INTO tls_reports (
			id, report_id, provider, start_date, end_date, 
			total_count, success_count, failure_count, raw_json, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	// Convert RawJSON if it's nil/empty, though unlikely from handler
	raw := report.RawJSON
	if len(raw) == 0 {
		raw = []byte("{}")
	}

	_, err := r.db.ExecContext(ctx, query,
		id.String(),
		report.ReportID,
		report.Provider,
		report.DateRange.StartDatetime,
		report.DateRange.EndDatetime,
		report.TotalCount,
		report.SuccessCount,
		report.FailureCount,
		string(raw),
		report.IngestedAt,
	)

	if err != nil {
		return ports.ErrStorageFailure
	}

	return nil
}

// FindLatest retrieves the most recent reports
func (r *TLSRptRepository) FindLatest(ctx context.Context, limit int) ([]*domain.TLSReport, error) {
	query := `
		SELECT 
			id, report_id, provider, start_date, end_date, 
			total_count, success_count, failure_count, raw_json, created_at
		FROM tls_reports 
		ORDER BY created_at DESC 
		LIMIT ?
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, ports.ErrStorageFailure
	}
	defer rows.Close()

	var reports []*domain.TLSReport
	for rows.Next() {
		var rpt domain.TLSReport
		var idStr, rawJSON string

		err := rows.Scan(
			&idStr,
			&rpt.ReportID,
			&rpt.Provider,
			&rpt.DateRange.StartDatetime,
			&rpt.DateRange.EndDatetime,
			&rpt.TotalCount,
			&rpt.SuccessCount,
			&rpt.FailureCount,
			&rawJSON,
			&rpt.IngestedAt,
		)
		if err != nil {
			return nil, ports.ErrStorageFailure
		}

		rpt.ID = uuid.MustParse(idStr)
		rpt.RawJSON = json.RawMessage(rawJSON)

		// Re-inflate the policies from RawJSON for the struct usage
		// Note: The struct definition has policies separate from RawJSON,
		// usually we want to parse the JSON back into the struct fields if needed.
		// However, for pure storage retrieval, sometimes just the summary is enough.
		// Let's unmarshal carefully.
		var fullStruct domain.TLSReport
		if err := json.Unmarshal([]byte(rawJSON), &fullStruct); err == nil {
			rpt.Policies = fullStruct.Policies
			rpt.ContactInfo = fullStruct.ContactInfo
		}

		reports = append(reports, &rpt)
	}

	return reports, nil
}
