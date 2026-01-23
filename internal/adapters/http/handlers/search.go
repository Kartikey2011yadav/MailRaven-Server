package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/http/dto"
	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/http/middleware"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/ports"
	"github.com/Kartikey2011yadav/mailraven-server/internal/observability"
)

// SearchHandler handles search-related HTTP requests
type SearchHandler struct {
	emailRepo ports.EmailRepository
	searchIdx ports.SearchIndex
	logger    *observability.Logger
	metrics   *observability.Metrics
}

// NewSearchHandler creates a new search handler
func NewSearchHandler(
	emailRepo ports.EmailRepository,
	searchIdx ports.SearchIndex,
	logger *observability.Logger,
	metrics *observability.Metrics,
) *SearchHandler {
	return &SearchHandler{
		emailRepo: emailRepo,
		searchIdx: searchIdx,
		logger:    logger,
		metrics:   metrics,
	}
}

// SearchMessages handles GET /v1/messages/search
func (h *SearchHandler) SearchMessages(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract authenticated user email
	email, ok := middleware.GetUserEmail(r)
	if !ok {
		h.sendError(w, http.StatusUnauthorized, "Missing user email in context")
		return
	}

	// Parse query parameters
	query := r.URL.Query().Get("q")
	if query == "" {
		h.sendError(w, http.StatusBadRequest, "Missing 'q' query parameter")
		return
	}

	if len(query) > 1000 {
		h.sendError(w, http.StatusBadRequest, "Query too long (max 1000 characters)")
		return
	}

	limit := h.parseIntParam(r, "limit", 20)
	offset := h.parseIntParam(r, "offset", 0)

	if limit < 1 || limit > 1000 {
		h.sendError(w, http.StatusBadRequest, "limit must be between 1 and 1000")
		return
	}
	if offset < 0 {
		h.sendError(w, http.StatusBadRequest, "offset must be non-negative")
		return
	}

	h.logger.Info("Searching messages",
		"method", "GET", "path", "/v1/messages/search", "user", email, "query", query, "limit", limit, "offset", offset)

	// Execute search
	results, err := h.searchIdx.Search(ctx, email, query, limit, offset)
	if err != nil {
		h.logger.Error("Search failed", "error", err, "query", query)
		h.metrics.IncrementAPIErrors()
		h.sendError(w, http.StatusBadRequest, "Invalid search query syntax")
		return
	}

	// Convert to DTOs with relevance scores
	dtoResults := make([]dto.SearchResult, len(results))
	for i, result := range results {
		// Get full message details
		message, err := h.emailRepo.FindByID(ctx, result.MessageID)
		if err != nil {
			h.logger.Error("Failed to retrieve message for search result", "error", err, "id", result.MessageID)
			continue // Skip this result
		}

		dtoResults[i] = dto.ToSearchResult(message, result.Relevance)
	}

	// Build response
	response := dto.SearchResponse{
		Results:      dtoResults,
		Query:        query,
		Count:        len(dtoResults),
		TotalMatches: len(dtoResults), // TODO: Get actual total from search index
	}

	h.sendJSON(w, http.StatusOK, response)
}

// parseIntParam parses an integer query parameter with a default value
func (h *SearchHandler) parseIntParam(r *http.Request, name string, defaultVal int) int {
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
func (h *SearchHandler) sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to encode JSON response", "error", err)
	}
}

// sendError sends an error response
func (h *SearchHandler) sendError(w http.ResponseWriter, status int, message string) {
	resp := dto.ErrorResponse{
		Error:   http.StatusText(status),
		Message: message,
	}
	h.sendJSON(w, status, resp)
}
