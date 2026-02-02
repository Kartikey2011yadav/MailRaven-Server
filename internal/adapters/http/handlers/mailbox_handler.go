package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Kartikey2011yadav/mailraven-server/internal/core/services"
	"github.com/Kartikey2011yadav/mailraven-server/internal/observability"
	"github.com/go-chi/chi/v5"
)

type MailboxHandler struct {
	emailService *services.EmailService
	logger       *observability.Logger
}

func NewMailboxHandler(emailService *services.EmailService, logger *observability.Logger) *MailboxHandler {
	return &MailboxHandler{
		emailService: emailService,
		logger:       logger,
	}
}

type UpdateACLRequest struct {
	Identifier string `json:"identifier"`
	Rights     string `json:"rights"`
}

// UpdateACL handles PUT /users/{userID}/mailboxes/{mailboxName}/acl
func (h *MailboxHandler) UpdateACL(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	mailboxName := chi.URLParam(r, "mailboxName")

	var req UpdateACLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.emailService.UpdateACL(r.Context(), userID, mailboxName, req.Identifier, req.Rights); err != nil {
		h.logger.Error("failed to update ACL", "error", err)
		// TODO: Better error mapping (404, 400)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	//nolint:errcheck // Write is final action
	w.Write([]byte(`{"status":"ok"}`))
}
