package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Kartikey2011yadav/mailraven-server/internal/config"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/ports"
	"github.com/Kartikey2011yadav/mailraven-server/internal/observability"
)

type SystemHandler struct {
	updater ports.UpdateManager
	logger  *observability.Logger
}

func NewSystemHandler(updater ports.UpdateManager, logger *observability.Logger) *SystemHandler {
	return &SystemHandler{
		updater: updater,
		logger:  logger,
	}
}

// CheckUpdate checks for available updates
func (h *SystemHandler) CheckUpdate(w http.ResponseWriter, r *http.Request) {
	info, err := h.updater.CheckForUpdate(r.Context(), config.Version)
	if err != nil {
		h.logger.Error("failed to check for updates", "error", err)
		http.Error(w, "Failed to check for updates", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if info == nil {
		// No update available
		// Return 204 No Content or a JSON saying up to date?
		// Mox returns 200 with empty fields or specific status.
		// Let's return 200 with an object indicating status.
		//nolint:errcheck // Ignore encode error
		json.NewEncoder(w).Encode(map[string]interface{}{
			"available": false,
			"current":   config.Version,
		})
		return
	}

	//nolint:errcheck // Ignore encode error
	json.NewEncoder(w).Encode(map[string]interface{}{
		"available":     true,
		"current":       config.Version,
		"latest":        info.Version,
		"release_notes": info.ReleaseNotes,
		"published_at":  info.PublishedAt,
	})
}

// PerformUpdate applies the update
func (h *SystemHandler) PerformUpdate(w http.ResponseWriter, r *http.Request) {
	// First check again to get the info (stateless)
	// Or we could accept info in body, but that's insecure/trusting client.
	// Best to fetch again.
	info, err := h.updater.CheckForUpdate(r.Context(), config.Version)
	if err != nil {
		http.Error(w, "Failed to resolve update info", http.StatusInternalServerError)
		return
	}
	if info == nil {
		http.Error(w, "No update available", http.StatusBadRequest)
		return
	}

	// Apply
	if err := h.updater.ApplyUpdate(r.Context(), info); err != nil {
		h.logger.Error("update application failed", "error", err)
		http.Error(w, "Failed to apply update: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Info("Update applied successfully", "version", info.Version)
	w.WriteHeader(http.StatusOK)
	//nolint:errcheck // Ignore encode error
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Update applied. Please restart the server.",
	})

	// Option: We could exit the process here to force restart by supervisor?
	// But responding first is polite.
}
