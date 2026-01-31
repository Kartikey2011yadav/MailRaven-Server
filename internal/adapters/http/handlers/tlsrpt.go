package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/http/dto"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/ports"
	"github.com/Kartikey2011yadav/mailraven-server/internal/observability"
)

type TLSRptHandler struct {
	repo   ports.TLSRptRepository
	logger *observability.Logger
}

func NewTLSRptHandler(repo ports.TLSRptRepository, logger *observability.Logger) *TLSRptHandler {
	return &TLSRptHandler{
		repo:   repo,
		logger: logger,
	}
}

// HandleReport receives POST reports
func (h *TLSRptHandler) HandleReport(w http.ResponseWriter, r *http.Request) {
	// 1. Verify Content-Type (RFC 8460 Section 4)
	// Must be application/tlsrpt+json or application/json
	ct := r.Header.Get("Content-Type")
	if ct != "application/tlsrpt+json" && ct != "application/json" {
		http.Error(w, "Invalid Content-Type", http.StatusBadRequest)
		return
	}

	// 2. Read Body Limits (prevent DoS)
	r.Body = http.MaxBytesReader(w, r.Body, 1024*1024) // 1MB limit for reports should be plenty
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("failed to read tls-rpt body", "error", err)
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	// 3. Parse JSON
	var req dto.TLSReportRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		h.logger.Warn("invalid tls-rpt json", "error", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// 4. Map to Domain
	report := req.MapToDomain(bodyBytes)

	// 5. Save
	if err := h.repo.Save(r.Context(), report); err != nil {
		h.logger.Error("failed to save tls-rpt", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
