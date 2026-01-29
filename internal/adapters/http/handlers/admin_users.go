package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/ports"
	"github.com/Kartikey2011yadav/mailraven-server/internal/observability"
	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"
)

type AdminUserHandler struct {
	userRepo   ports.UserRepository
	domainRepo ports.DomainRepository
	logger     *observability.Logger
}

func NewAdminUserHandler(userRepo ports.UserRepository, domainRepo ports.DomainRepository, logger *observability.Logger) *AdminUserHandler {
	return &AdminUserHandler{userRepo: userRepo, domainRepo: domainRepo, logger: logger}
}

type CreateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"` // optional, default "user"
}

// ListUsers GET /api/v1/admin/users
func (h *AdminUserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	limit := 50
	offset := 0

	if l := r.URL.Query().Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil {
			limit = val
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if val, err := strconv.Atoi(o); err == nil {
			offset = val
		}
	}

	users, err := h.userRepo.List(r.Context(), limit, offset)
	if err != nil {
		h.logger.Error("Failed to list users", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Sanitize output (hide password hashes)
	for _, u := range users {
		u.PasswordHash = ""
	}

	w.Header().Set("Content-Type", "application/json")
	//nolint:errcheck // Response write error is non-critical
	_ = json.NewEncoder(w).Encode(users)
}

// CreateUser POST /api/v1/admin/users
func (h *AdminUserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email and Password required", http.StatusBadRequest)
		return
	}

	// Validate domain ownership
	parts := strings.Split(req.Email, "@")
	if len(parts) != 2 {
		http.Error(w, "Invalid email format", http.StatusBadRequest)
		return
	}
	domainName := parts[1]

	// Only check domain exists if user is not admin?
	// Actually, even admins should probably be on a managed domain, OR we allow admins to be created on any domain (e.g. localhost) but that might be risky for email routing.
	// For "mail server", we usually only accept emails for domains we host.
	// So validating domain is correct.
	exists, err := h.domainRepo.Exists(r.Context(), domainName)
	if err != nil {
		h.logger.Error("Failed to check domain existence", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if !exists {
		http.Error(w, "Domain not managed by this server. Add domain first.", http.StatusBadRequest)
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Hashing failed", http.StatusInternalServerError)
		return
	}

	role := domain.RoleUser
	if req.Role == "admin" {
		role = domain.RoleAdmin
	}

	user := &domain.User{
		Email:        req.Email,
		PasswordHash: string(hashed),
		Role:         role,
		CreatedAt:    time.Now(),
		LastLoginAt:  time.Unix(0, 0),
	}

	if err := h.userRepo.Create(r.Context(), user); err != nil {
		if err == ports.ErrAlreadyExists {
			http.Error(w, "User already exists", http.StatusConflict)
			return
		}
		h.logger.Error("Failed to create user", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// DeleteUser DELETE /api/v1/admin/users/{email}
func (h *AdminUserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	email := chi.URLParam(r, "email")
	if email == "" {
		http.Error(w, "Email required", http.StatusBadRequest)
		return
	}

	if err := h.userRepo.Delete(r.Context(), email); err != nil {
		if err == ports.ErrNotFound {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		h.logger.Error("Failed to delete user", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// UpdateRole PUT /api/v1/admin/users/{email}/role
func (h *AdminUserHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	email := chi.URLParam(r, "email")
	var req struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	role := domain.Role(req.Role)
	if role != domain.RoleUser && role != domain.RoleAdmin {
		http.Error(w, "Invalid role", http.StatusBadRequest)
		return
	}

	if err := h.userRepo.UpdateRole(r.Context(), email, role); err != nil {
		if err == ports.ErrNotFound {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
