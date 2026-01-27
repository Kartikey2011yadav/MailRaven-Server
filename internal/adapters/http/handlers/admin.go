package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Kartikey2011yadav/mailraven-server/internal/core/ports"
	"github.com/Kartikey2011yadav/mailraven-server/internal/observability"
)

type AdminHandler struct {
	backupService ports.BackupService
	logger        *observability.Logger
	metrics       *observability.Metrics
}

func NewAdminHandler(backupService ports.BackupService, logger *observability.Logger, metrics *observability.Metrics) *AdminHandler {
	return &AdminHandler{
		backupService: backupService,
		logger:        logger,
		metrics:       metrics,
	}
}

type BackupRequest struct {
	Location string `json:"location"`
}

type BackupResponse struct {
	JobID  string `json:"job_id"`
	Status string `json:"status"`
	Path   string `json:"path"`
}

func (h *AdminHandler) TriggerBackup(w http.ResponseWriter, r *http.Request) {
	var req BackupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && r.ContentLength > 0 {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Trigger backup
	path, err := h.backupService.PerformBackup(r.Context(), req.Location)
	if err != nil {
		h.logger.Error("admin backup failed", "error", err)
		http.Error(w, "backup failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(BackupResponse{
		JobID:  "done", // Service should return ID, but sticking to simple return for now
		Status: "completed",
		Path:   path,
	})
}
