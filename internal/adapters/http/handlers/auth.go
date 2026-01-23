package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/http/dto"
	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/http/middleware"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/ports"
	"github.com/Kartikey2011yadav/mailraven-server/internal/observability"
	"github.com/golang-jwt/jwt/v5"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	userRepo  ports.UserRepository
	jwtSecret string
	logger    *observability.Logger
	metrics   *observability.Metrics
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(
	userRepo ports.UserRepository,
	jwtSecret string,
	logger *observability.Logger,
	metrics *observability.Metrics,
) *AuthHandler {
	return &AuthHandler{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
		logger:    logger,
		metrics:   metrics,
	}
}

// Login handles POST /auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request body
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	// Validate input
	if req.Email == "" {
		h.sendError(w, http.StatusBadRequest, "Missing email field")
		return
	}
	if req.Password == "" {
		h.sendError(w, http.StatusBadRequest, "Missing password field")
		return
	}
	if len(req.Password) < 8 {
		h.sendError(w, http.StatusBadRequest, "Password must be at least 8 characters")
		return
	}

	h.logger.Info("Login attempt", "method", "POST", "path", "/auth/login", "email", req.Email)

	// Authenticate user
	user, err := h.userRepo.Authenticate(ctx, req.Email, req.Password)
	if err != nil {
		if err == ports.ErrInvalidCredentials {
			h.logger.Info("Login failed: invalid credentials", "email", req.Email)
			h.sendError(w, http.StatusUnauthorized, "Invalid email or password")
			return
		}
		h.logger.Error("Login failed", "error", err, "email", req.Email)
		h.metrics.IncrementAPIErrors()
		h.sendError(w, http.StatusInternalServerError, "Authentication failed")
		return
	}

	// Generate JWT token (valid for 7 days)
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	claims := &middleware.Claims{
		Email: user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(h.jwtSecret))
	if err != nil {
		h.logger.Error("Failed to sign JWT token", "error", err)
		h.metrics.IncrementAPIErrors()
		h.sendError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	// Update last login time
	if err := h.userRepo.UpdateLastLogin(ctx, user.Email); err != nil {
		// Non-fatal, log and continue
		h.logger.Error("Failed to update last login", "error", err)
	}

	h.logger.Info("Login successful", "email", user.Email)

	// Send response
	response := dto.LoginResponse{
		Token:     tokenString,
		Email:     user.Email,
		ExpiresAt: expiresAt,
	}
	h.sendJSON(w, http.StatusOK, response)
}

// sendJSON sends a JSON response
func (h *AuthHandler) sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to encode JSON response", "error", err)
	}
}

// sendError sends an error response
func (h *AuthHandler) sendError(w http.ResponseWriter, status int, message string) {
	resp := dto.ErrorResponse{
		Error:   http.StatusText(status),
		Message: message,
	}
	h.sendJSON(w, status, resp)
}
