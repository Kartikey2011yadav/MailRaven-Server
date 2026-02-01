package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/http/dto"
	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/http/middleware"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/ports"
	"github.com/Kartikey2011yadav/mailraven-server/internal/observability"
	"golang.org/x/crypto/bcrypt"
)

// UserSelfHandler handles requests for user self-management
type UserSelfHandler struct {
	userRepo ports.UserRepository
	logger   *observability.Logger
}

// NewUserSelfHandler creates a new user self handler
func NewUserSelfHandler(userRepo ports.UserRepository, logger *observability.Logger) *UserSelfHandler {
	return &UserSelfHandler{
		userRepo: userRepo,
		logger:   logger,
	}
}

// ChangePassword handles PUT /api/v1/users/self/password
func (h *UserSelfHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. Get authenticated user email
	email, ok := middleware.GetUserEmail(r)
	if !ok {
		h.logger.Error("ChangePassword called without authenticated user")
		h.sendError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// 2. Parse request body
	var req dto.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	// 3. Validate input
	if req.CurrentPassword == "" {
		h.sendError(w, http.StatusBadRequest, "Current password is required")
		return
	}
	if req.NewPassword == "" {
		h.sendError(w, http.StatusBadRequest, "New password is required")
		return
	}
	if len(req.NewPassword) < 8 {
		h.sendError(w, http.StatusBadRequest, "New password must be at least 8 characters")
		return
	}

	h.logger.Info("Password change attempt", "email", email)

	// 4. Authenticate current password
	// We use userRepo.Authenticate to verify the current credentials
	_, err := h.userRepo.Authenticate(ctx, email, req.CurrentPassword)
	if err != nil {
		if err == ports.ErrInvalidCredentials {
			h.logger.Info("Password change failed: invalid current password", "email", email)
			h.sendError(w, http.StatusUnauthorized, "Invalid current password")
			return
		}
		h.logger.Error("Password change failed during authentication", "error", err, "email", email)
		h.sendError(w, http.StatusInternalServerError, "Failed to verify current password")
		return
	}

	// 5. Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		h.logger.Error("Failed to hash new password", "error", err)
		h.sendError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	// 6. Update password
	err = h.userRepo.UpdatePassword(ctx, email, string(hashedPassword))
	if err != nil {
		h.logger.Error("Failed to update password", "error", err, "email", email)
		h.sendError(w, http.StatusInternalServerError, "Failed to update password")
		return
	}

	h.logger.Info("Password changed successfully", "email", email)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Password updated successfully"})
}

func (h *UserSelfHandler) sendError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
