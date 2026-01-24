package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/http/dto"
	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/http/middleware"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/ports"
	"github.com/Kartikey2011yadav/mailraven-server/internal/observability"
	"github.com/go-chi/chi/v5"
)

// MessageHandler handles message-related HTTP requests
type MessageHandler struct {
	emailRepo ports.EmailRepository
	blobStore ports.BlobStore
	searchIdx ports.SearchIndex
	logger    *observability.Logger
	metrics   *observability.Metrics
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(
	emailRepo ports.EmailRepository,
	blobStore ports.BlobStore,
	searchIdx ports.SearchIndex,
	logger *observability.Logger,
	metrics *observability.Metrics,
) *MessageHandler {
	return &MessageHandler{
		emailRepo: emailRepo,
		blobStore: blobStore,
		searchIdx: searchIdx,
		logger:    logger,
		metrics:   metrics,
	}
}

// ListMessages handles GET /v1/messages
func (h *MessageHandler) ListMessages(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract authenticated user email from context
	email, ok := middleware.GetUserEmail(r)
	if !ok {
		h.sendError(w, http.StatusUnauthorized, "Missing user email in context")
		return
	}

	// Parse query parameters
	limit := h.parseIntParam(r, "limit", 20)
	offset := h.parseIntParam(r, "offset", 0)
	unreadOnly := r.URL.Query().Get("unread_only") == "true"

	// Validate parameters
	if limit < 1 || limit > 1000 {
		h.sendError(w, http.StatusBadRequest, "limit must be between 1 and 1000")
		return
	}
	if offset < 0 {
		h.sendError(w, http.StatusBadRequest, "offset must be non-negative")
		return
	}

	h.logger.Info("Listing messages",
		"method", "GET", "path", "/v1/messages", "user", email, "limit", limit, "offset", offset, "unread_only", unreadOnly)

	// Get messages from repository (ignoring unread_only filter for MVP - filter client-side)
	messages, err := h.emailRepo.FindByUser(ctx, email, limit, offset)
	if err != nil {
		h.logger.Error("Failed to retrieve messages", "error", err)
		h.metrics.IncrementAPIErrors()
		h.sendError(w, http.StatusInternalServerError, "Failed to retrieve messages")
		return
	}

	// Get total count
	total, err := h.emailRepo.CountByUser(ctx, email)
	if err != nil {
		h.logger.Error("Failed to count messages", "error", err)
		total = 0 // Non-fatal, continue with zero count
	}

	// Convert to DTOs
	dtoMessages := make([]dto.MessageSummary, len(messages))
	for i, msg := range messages {
		dtoMessages[i] = dto.ToMessageSummary(msg)
	}

	// Build response
	response := dto.MessageListResponse{
		Messages: dtoMessages,
		Total:    total,
		Limit:    limit,
		Offset:   offset,
		HasMore:  offset+len(messages) < total,
	}

	h.sendJSON(w, http.StatusOK, response)
}

// GetMessage handles GET /v1/messages/{id}
func (h *MessageHandler) GetMessage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract authenticated user email
	email, ok := middleware.GetUserEmail(r)
	if !ok {
		h.sendError(w, http.StatusUnauthorized, "Missing user email in context")
		return
	}

	// Extract message ID from URL
	messageID := chi.URLParam(r, "id")
	if messageID == "" {
		h.sendError(w, http.StatusBadRequest, "Missing message ID")
		return
	}

	h.logger.Info("Getting message",
		"method", "GET", "path", "/v1/messages/{id}", "user", email, "message_id", messageID)

	// Get message metadata
	message, err := h.emailRepo.FindByID(ctx, messageID)
	if err != nil {
		if err == ports.ErrNotFound {
			h.sendError(w, http.StatusNotFound, "Message not found")
			return
		}
		h.logger.Error("Failed to retrieve message", "error", err)
		h.metrics.IncrementAPIErrors()
		h.sendError(w, http.StatusInternalServerError, "Failed to retrieve message")
		return
	}

	// Verify message belongs to authenticated user
	if message.Recipient != email {
		h.sendError(w, http.StatusNotFound, "Message not found")
		return
	}

	// Read message body from blob store
	bodyBytes, err := h.blobStore.Read(ctx, message.BodyPath)
	if err != nil {
		h.logger.Error("Failed to read message body", "error", err, "path", message.BodyPath)
		h.metrics.IncrementStorageErrors()
		h.sendError(w, http.StatusInternalServerError, "Failed to read message body")
		return
	}

	// Convert bytes to string
	body := string(bodyBytes)

	// Get body size (from compressed file)
	bodySize := int64(len(bodyBytes))

	// Convert to DTO
	response := dto.ToMessageFull(message, body, bodySize)

	h.metrics.IncrementStorageReads()
	h.sendJSON(w, http.StatusOK, response)
}

// UpdateMessage handles PATCH /v1/messages/{id}
func (h *MessageHandler) UpdateMessage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract authenticated user email
	email, ok := middleware.GetUserEmail(r)
	if !ok {
		h.sendError(w, http.StatusUnauthorized, "Missing user email in context")
		return
	}

	// Extract message ID
	messageID := chi.URLParam(r, "id")
	if messageID == "" {
		h.sendError(w, http.StatusBadRequest, "Missing message ID")
		return
	}

	// Parse request body
	var req dto.UpdateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	if req.ReadState == nil {
		h.sendError(w, http.StatusBadRequest, "Missing read_state field")
		return
	}

	h.logger.Info("Updating message",
		"method", "PATCH", "path", "/v1/messages/{id}", "user", email, "message_id", messageID, "read_state", *req.ReadState)

	// Verify message exists and belongs to user
	message, err := h.emailRepo.FindByID(ctx, messageID)
	if err != nil {
		if err == ports.ErrNotFound {
			h.sendError(w, http.StatusNotFound, "Message not found")
			return
		}
		h.logger.Error("Failed to retrieve message", "error", err)
		h.metrics.IncrementAPIErrors()
		h.sendError(w, http.StatusInternalServerError, "Failed to retrieve message")
		return
	}

	if message.Recipient != email {
		h.sendError(w, http.StatusNotFound, "Message not found")
		return
	}

	// Update read state
	if err := h.emailRepo.UpdateReadState(ctx, messageID, *req.ReadState); err != nil {
		h.logger.Error("Failed to update message", "error", err)
		h.metrics.IncrementAPIErrors()
		h.sendError(w, http.StatusInternalServerError, "Failed to update message")
		return
	}

	// Fetch updated message
	message.ReadState = *req.ReadState
	response := dto.ToMessageSummary(message)

	h.sendJSON(w, http.StatusOK, response)
}

// GetMessagesSince handles GET /v1/messages/since
func (h *MessageHandler) GetMessagesSince(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract authenticated user email
	email, ok := middleware.GetUserEmail(r)
	if !ok {
		h.sendError(w, http.StatusUnauthorized, "Missing user email in context")
		return
	}

	// Parse since timestamp
	sinceStr := r.URL.Query().Get("since")
	if sinceStr == "" {
		h.sendError(w, http.StatusBadRequest, "Missing 'since' query parameter")
		return
	}

	since, err := time.Parse(time.RFC3339, sinceStr)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid timestamp format (expected ISO 8601)")
		return
	}

	// Parse limit
	limit := h.parseIntParam(r, "limit", 100)
	if limit < 1 || limit > 1000 {
		h.sendError(w, http.StatusBadRequest, "limit must be between 1 and 1000")
		return
	}

	h.logger.Info("Getting messages since timestamp",
		"method", "GET", "path", "/v1/messages/since", "user", email, "since", since, "limit", limit)

	// Get messages from repository
	messages, err := h.emailRepo.FindSince(ctx, email, since, limit)
	if err != nil {
		h.logger.Error("Failed to retrieve messages", "error", err)
		h.metrics.IncrementAPIErrors()
		h.sendError(w, http.StatusInternalServerError, "Failed to retrieve messages")
		return
	}

	// Convert to DTOs
	dtoMessages := make([]dto.MessageSummary, len(messages))
	for i, msg := range messages {
		dtoMessages[i] = dto.ToMessageSummary(msg)
	}

	// Build response
	response := dto.MessagesSinceResponse{
		Messages: dtoMessages,
		Count:    len(messages),
		Since:    since,
	}

	h.sendJSON(w, http.StatusOK, response)
}

// parseIntParam parses an integer query parameter with a default value
func (h *MessageHandler) parseIntParam(r *http.Request, name string, defaultVal int) int {
	str := r.URL.Query().Get(name)
	if str == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(str)
	if err != nil {
		return defaultVal
	}
	return val
}

// sendJSON sends a JSON response
func (h *MessageHandler) sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to encode JSON response", "error", err)
	}
}

// sendError sends an error response
func (h *MessageHandler) sendError(w http.ResponseWriter, status int, message string) {
	resp := dto.ErrorResponse{
		Error:   http.StatusText(status),
		Message: message,
	}
	h.sendJSON(w, status, resp)
}
