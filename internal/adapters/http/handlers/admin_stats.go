package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/http/middleware"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/ports"
	"github.com/Kartikey2011yadav/mailraven-server/internal/observability"
)

type AdminStatsHandler struct {
	userRepo  ports.UserRepository
	emailRepo ports.EmailRepository
	queueRepo ports.QueueRepository
	logger    *observability.Logger
}

func NewAdminStatsHandler(
	userRepo ports.UserRepository,
	emailRepo ports.EmailRepository,
	queueRepo ports.QueueRepository,
	logger *observability.Logger,
) *AdminStatsHandler {
	return &AdminStatsHandler{
		userRepo:  userRepo,
		emailRepo: emailRepo,
		queueRepo: queueRepo,
		logger:    logger,
	}
}

type SystemStatsResponse struct {
	Users struct {
		Total  int64 `json:"total"`
		Active int64 `json:"active"`
		Admin  int64 `json:"admin"`
	} `json:"users"`
	Emails struct {
		Total int64 `json:"total"`
	} `json:"emails"`
	Queue struct {
		Pending    int64 `json:"pending"`
		Processing int64 `json:"processing"`
		Failed     int64 `json:"failed"`
		Completed  int64 `json:"completed"`
	} `json:"queue"`
}

func (h *AdminStatsHandler) GetSystemStats(w http.ResponseWriter, r *http.Request) {
	// Check if user is admin
	role, ok := middleware.GetUserRole(r)
	if !ok || role != "ADMIN" {
		h.logger.Warn("Unauthorized stats access attempt", "role", role)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	ctx := r.Context()

	// 1. Get User Stats
	userStats, err := h.userRepo.Count(ctx)
	if err != nil {
		h.logger.Error("failed to get user stats", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 2. Get Email Stats
	totalEmails, err := h.emailRepo.CountTotal(ctx)
	if err != nil {
		h.logger.Error("failed to get email stats", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 3. Get Queue Stats
	pending, processing, failed, completed, err := h.queueRepo.Stats(ctx)
	if err != nil {
		h.logger.Error("failed to get queue stats", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	resp := SystemStatsResponse{}
	resp.Users.Total = userStats["total"]
	resp.Users.Active = userStats["active"]
	resp.Users.Admin = userStats["admin"]

	resp.Emails.Total = totalEmails

	resp.Queue.Pending = pending
	resp.Queue.Processing = processing
	resp.Queue.Failed = failed
	resp.Queue.Completed = completed

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}
