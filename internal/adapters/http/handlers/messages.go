package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/http/dto"
	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/http/middleware"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/ports"
	"github.com/Kartikey2011yadav/mailraven-server/internal/observability"
	"github.com/go-chi/chi/v5"
)

// MessageHandler handles message-related HTTP requests
type MessageHandler struct {
	emailRepo  ports.EmailRepository
	blobStore  ports.BlobStore
	searchIdx  ports.SearchIndex
	spamFilter ports.SpamFilter
	logger     *observability.Logger
	metrics    *observability.Metrics
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(
	emailRepo ports.EmailRepository,
	blobStore ports.BlobStore,
	searchIdx ports.SearchIndex,
	spamFilter ports.SpamFilter,
	logger *observability.Logger,
	metrics *observability.Metrics,
) *MessageHandler {
	return &MessageHandler{
		emailRepo:  emailRepo,
		blobStore:  blobStore,
		searchIdx:  searchIdx,
		spamFilter: spamFilter,
		logger:     logger,
		metrics:    metrics,
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

	// Build filter
	filter := domain.MessageFilter{
		Limit:   limit,
		Offset:  offset,
		Mailbox: r.URL.Query().Get("mailbox"),
	}

	// Handle read status filter (support both is_read and legacy unread_only)
	if s := r.URL.Query().Get("is_read"); s != "" {
		val := s == "true"
		filter.IsRead = &val
	} else if r.URL.Query().Get("unread_only") == "true" {
		val := false
		filter.IsRead = &val
	}

	// Handle starred status filter
	if s := r.URL.Query().Get("is_starred"); s != "" {
		val := s == "true"
		filter.IsStarred = &val
	}

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
		"method", "GET", "path", "/v1/messages",
		"user", email,
		"limit", limit,
		"offset", offset,
		"mailbox", filter.Mailbox,
		"is_read", filter.IsRead,
		"is_starred", filter.IsStarred)

	// Get messages from repository using new List method
	messages, err := h.emailRepo.List(ctx, email, filter)
	if err != nil {
		h.logger.Error("Failed to retrieve messages", "error", err)
		h.metrics.IncrementAPIErrors()
		h.sendError(w, http.StatusInternalServerError, "Failed to retrieve messages")
		return
	}

	// Get total count
	// Note: Currently CountByUser returns total messages regardless of filter
	// TODO: Implement Count(filter) for accurate pagination in filtered views
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
		HasMore:  offset+len(messages) < total, // This check is approximate if we are filtering
	}

	if filter.Mailbox != "" || filter.IsRead != nil || filter.IsStarred != nil {
		// If filtering, HasMore logic using global total is flawed.
		// Fallback to checking if we got a full page?
		// If we got 'limit' messages, assume there might be more.
		response.HasMore = len(messages) == limit
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

	if req.ReadState == nil && req.IsStarred == nil && req.Mailbox == nil {
		h.sendError(w, http.StatusBadRequest, "No update fields provided")
		return
	}

	h.logger.Info("Updating message",
		"method", "PATCH", "path", "/v1/messages/{id}",
		"user", email,
		"message_id", messageID,
		"read_state", req.ReadState,
		"is_starred", req.IsStarred,
		"mailbox", req.Mailbox)

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
	if req.ReadState != nil {
		if err := h.emailRepo.UpdateReadState(ctx, messageID, *req.ReadState); err != nil {
			h.logger.Error("Failed to update read state", "error", err)
			h.metrics.IncrementAPIErrors()
			h.sendError(w, http.StatusInternalServerError, "Failed to update message")
			return
		}
	}

	// Update starred status
	if req.IsStarred != nil {
		if err := h.emailRepo.UpdateStarred(ctx, messageID, *req.IsStarred); err != nil {
			h.logger.Error("Failed to update starred status", "error", err)
			h.metrics.IncrementAPIErrors()
			h.sendError(w, http.StatusInternalServerError, "Failed to update message")
			return
		}
	}

	// Update mailbox
	if req.Mailbox != nil {
		if err := h.emailRepo.UpdateMailbox(ctx, messageID, *req.Mailbox); err != nil {
			h.logger.Error("Failed to update mailbox", "error", err)
			h.metrics.IncrementAPIErrors()
			h.sendError(w, http.StatusInternalServerError, "Failed to update message")
			return
		}
	}

	// Fetch updated message
	updatedMessage, err := h.emailRepo.FindByID(ctx, messageID)
	if err != nil {
		h.logger.Error("Failed to retrieve updated message", "error", err)
		h.sendError(w, http.StatusInternalServerError, "Failed to retrieve updated message")
		return
	}

	response := dto.ToMessageSummary(updatedMessage)
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

// ReportSpam handles POST /v1/messages/{id}/spam
func (h *MessageHandler) ReportSpam(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	email, ok := middleware.GetUserEmail(r)
	if !ok {
		h.sendError(w, http.StatusUnauthorized, "Missing user email in context")
		return
	}

	messageID := chi.URLParam(r, "id")
	if messageID == "" {
		h.sendError(w, http.StatusBadRequest, "Missing message ID")
		return
	}

	h.logger.Info("Reporting message as spam", "user", email, "message_id", messageID)

	message, err := h.emailRepo.FindByID(ctx, messageID)
	if err != nil {
		if err == ports.ErrNotFound {
			h.sendError(w, http.StatusNotFound, "Message not found")
			return
		}
		h.logger.Error("Failed to retrieve message", "error", err)
		h.sendError(w, http.StatusInternalServerError, "Failed to retrieve message")
		return
	}

	if message.Recipient != email {
		h.sendError(w, http.StatusNotFound, "Message not found")
		return
	}

	// Train spam filter
	if h.spamFilter != nil {
		bodyBytes, err := h.blobStore.Read(ctx, message.BodyPath)
		if err == nil {
			if err := h.spamFilter.TrainSpam(ctx, bytes.NewReader(bodyBytes)); err != nil {
				h.logger.Warn("Failed to train spam filter", "error", err)
			}
		} else {
			h.logger.Warn("Failed to read message body for spam training", "error", err)
		}
	}

	// Move to Junk
	if err := h.emailRepo.UpdateMailbox(ctx, messageID, "Junk"); err != nil {
		h.logger.Error("Failed to move message to Junk", "error", err)
		h.sendError(w, http.StatusInternalServerError, "Failed to update message")
		return
	}

	// Return updated message
	message.Mailbox = "Junk"
	h.sendJSON(w, http.StatusOK, dto.ToMessageSummary(message))
}

// ReportHam handles POST /v1/messages/{id}/ham
func (h *MessageHandler) ReportHam(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	email, ok := middleware.GetUserEmail(r)
	if !ok {
		h.sendError(w, http.StatusUnauthorized, "Missing user email in context")
		return
	}

	messageID := chi.URLParam(r, "id")
	if messageID == "" {
		h.sendError(w, http.StatusBadRequest, "Missing message ID")
		return
	}

	h.logger.Info("Reporting message as ham", "user", email, "message_id", messageID)

	message, err := h.emailRepo.FindByID(ctx, messageID)
	if err != nil {
		if err == ports.ErrNotFound {
			h.sendError(w, http.StatusNotFound, "Message not found")
			return
		}
		h.logger.Error("Failed to retrieve message", "error", err)
		h.sendError(w, http.StatusInternalServerError, "Failed to retrieve message")
		return
	}

	if message.Recipient != email {
		h.sendError(w, http.StatusNotFound, "Message not found")
		return
	}

	// Train ham filter
	if h.spamFilter != nil {
		bodyBytes, err := h.blobStore.Read(ctx, message.BodyPath)
		if err == nil {
			if err := h.spamFilter.TrainHam(ctx, bytes.NewReader(bodyBytes)); err != nil {
				h.logger.Warn("Failed to train ham filter", "error", err)
			}
		} else {
			h.logger.Warn("Failed to read message body for ham training", "error", err)
		}
	}

	// Move to INBOX
	if err := h.emailRepo.UpdateMailbox(ctx, messageID, "INBOX"); err != nil {
		h.logger.Error("Failed to move message to Inbox", "error", err)
		h.sendError(w, http.StatusInternalServerError, "Failed to update message")
		return
	}

	// Return updated message
	message.Mailbox = "INBOX"
	h.sendJSON(w, http.StatusOK, dto.ToMessageSummary(message))
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
