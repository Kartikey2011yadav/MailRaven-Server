package handlers

import (
	"net/http"

	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
)

// MTASTSConfigProvider defines how to get policy config (usually from global config or DB)
type MTASTSConfigProvider interface {
	GetMTASTSPolicy(hostname string) (*domain.MTASTSPolicy, error)
}

type MTASTSHandler struct {
	// In a real multi-tenant system, we might look up policy by domain.
	// For MailRaven simple server, we use a static config derived from env vars.
	Policy *domain.MTASTSPolicy
}

func NewMTASTSHandler(policy *domain.MTASTSPolicy) *MTASTSHandler {
	return &MTASTSHandler{
		Policy: policy,
	}
}

// ServePolicy serves the .well-known/mta-sts.txt file
func (h *MTASTSHandler) ServePolicy(w http.ResponseWriter, r *http.Request) {
	// RFC 8461 Section 3.2: Content-Type must be text/plain
	w.Header().Set("Content-Type", "text/plain")

	// We trust the routing layer to only route mta-sts.* requests here.

	body := h.Policy.BuildPolicyString()
	//nolint:errcheck // Ignore write errors
	w.Write([]byte(body))
}
