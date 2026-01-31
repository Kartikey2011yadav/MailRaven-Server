package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/http/dto"
	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/http/middleware"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain/sieve"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/ports"
	"github.com/Kartikey2011yadav/mailraven-server/internal/observability"
	"github.com/go-chi/chi/v5"
)

type SieveHandler struct {
	repo   ports.ScriptRepository
	logger *observability.Logger
}

func NewSieveHandler(repo ports.ScriptRepository, logger *observability.Logger) *SieveHandler {
	return &SieveHandler{
		repo:   repo,
		logger: logger,
	}
}

// ListScripts returns all scripts for the authenticated user
func (h *SieveHandler) ListScripts(w http.ResponseWriter, r *http.Request) {
	email, ok := middleware.GetUserEmail(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	scripts, err := h.repo.List(r.Context(), email)
	if err != nil {
		h.logger.Error("failed to list scripts", "user", email, "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	resp := make([]dto.SieveScriptResponse, len(scripts))
	for i, s := range scripts {
		resp[i] = dto.SieveScriptResponse{
			Name:      s.Name,
			Content:   s.Content,
			IsActive:  s.IsActive,
			CreatedAt: s.CreatedAt,
			UpdatedAt: s.UpdatedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

// CreateScript creates or updates a script
func (h *SieveHandler) CreateScript(w http.ResponseWriter, r *http.Request) {
	email, ok := middleware.GetUserEmail(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req dto.CreateSieveScriptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Content == "" {
		http.Error(w, "Name and Content are required", http.StatusBadRequest)
		return
	}

	script := &sieve.SieveScript{
		Name:      req.Name,
		Content:   req.Content,
		UserID:    email,
		CreatedAt: time.Now(), // Repo should handle timestamps really
		UpdatedAt: time.Now(),
		// IsActive defaults to false
	}
	// Note: Create endpoint doesn't usually set active state unless specified?
	// RFC usually separates management from activation.

	if err := h.repo.Save(r.Context(), script); err != nil {
		h.logger.Error("failed to save script", "user", email, "name", req.Name, "error", err)
		http.Error(w, "Failed to save script", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// GetScript returns a single script content
func (h *SieveHandler) GetScript(w http.ResponseWriter, r *http.Request) {
	email, ok := middleware.GetUserEmail(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	name := chi.URLParam(r, "name")

	if name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	script, err := h.repo.Get(r.Context(), email, name)
	if err != nil {
		// Detect not found (needs proper error type from repo)
		// Assuming generic error for now, logging it
		h.logger.Warn("failed to get script", "user", email, "name", name, "error", err)
		http.Error(w, "Script not found", http.StatusNotFound)
		return
	}
	if script == nil {
		http.Error(w, "Script not found", http.StatusNotFound)
		return
	}

	resp := dto.SieveScriptResponse{
		Name:      script.Name,
		Content:   script.Content,
		IsActive:  script.IsActive,
		CreatedAt: script.CreatedAt,
		UpdatedAt: script.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

// DeleteScript removes a script
func (h *SieveHandler) DeleteScript(w http.ResponseWriter, r *http.Request) {
	email, ok := middleware.GetUserEmail(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	name := chi.URLParam(r, "name")

	if name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	if err := h.repo.Delete(r.Context(), email, name); err != nil {
		h.logger.Error("failed to delete script", "user", email, "name", name, "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ActivateScript sets the active script
func (h *SieveHandler) ActivateScript(w http.ResponseWriter, r *http.Request) {
	email, ok := middleware.GetUserEmail(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	name := chi.URLParam(r, "name")

	// If name is "active" keyword or similar? No, route is /scripts/{name}/active
	// If user wants to deactivate all, maybe an empty body to a general /activate endpoint?
	// T020 says "Activate Script".
	// Repo SetActive(name) - if name empty, deactivate all.

	if err := h.repo.SetActive(r.Context(), email, name); err != nil {
		h.logger.Error("failed to activate script", "user", email, "name", name, "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
