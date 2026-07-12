package handlers

import (
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/ports"
	"github.com/Kartikey2011yadav/mailraven-server/internal/observability"
	"golang.org/x/crypto/bcrypt"
)

type SetupHandler struct {
	mu         sync.Mutex
	userRepo   ports.UserRepository
	domainRepo ports.DomainRepository
	logger     *observability.Logger
}

func NewSetupHandler(userRepo ports.UserRepository, domainRepo ports.DomainRepository, logger *observability.Logger) *SetupHandler {
	return &SetupHandler{userRepo: userRepo, domainRepo: domainRepo, logger: logger}
}

type SetupStatusResponse struct {
	SetupRequired bool `json:"setup_required"`
}

type SetupCompleteRequest struct {
	Domain       string `json:"domain"`
	AdminEmail   string `json:"admin_email"`
	AdminPassword string `json:"admin_password"`
	SMTPHostname string `json:"smtp_hostname"`
}

type DNSRecord struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

type SetupCompleteResponse struct {
	Success    bool        `json:"success"`
	DNSRecords []DNSRecord `json:"dns_records"`
}

// GetStatus returns whether initial setup is required (no users exist)
func (h *SetupHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	counts, err := h.userRepo.Count(r.Context())
	if err != nil {
		h.logger.Error("Failed to count users for setup check", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	total := counts["total"]

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(SetupStatusResponse{
		SetupRequired: total == 0,
	})
}

// Complete performs initial server setup: creates domain and admin user
func (h *SetupHandler) Complete(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Guard: only works if no users exist
	counts, err := h.userRepo.Count(r.Context())
	if err != nil {
		h.logger.Error("Setup: failed to count users", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if counts["total"] > 0 {
		http.Error(w, "Setup already completed", http.StatusConflict)
		return
	}

	var req SetupCompleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Domain == "" || req.AdminEmail == "" || req.AdminPassword == "" {
		http.Error(w, "domain, admin_email, and admin_password are required", http.StatusBadRequest)
		return
	}

	if len(req.AdminPassword) < 8 {
		http.Error(w, "Password must be at least 8 characters", http.StatusBadRequest)
		return
	}

	// Validate email belongs to the domain
	parts := strings.SplitN(req.AdminEmail, "@", 2)
	if len(parts) != 2 || parts[1] != req.Domain {
		http.Error(w, "Admin email must belong to the configured domain", http.StatusBadRequest)
		return
	}

	// Generate DKIM keys
	privKeyPEM, pubKeyPEM, err := generateDKIMKeys()
	if err != nil {
		h.logger.Error("Setup: failed to generate DKIM keys", "error", err)
		http.Error(w, "Failed to generate DKIM keys", http.StatusInternalServerError)
		return
	}

	// Create domain
	d := &domain.Domain{
		Name:           req.Domain,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Active:         true,
		DKIMSelector:   "default",
		DKIMPrivateKey: privKeyPEM,
		DKIMPublicKey:  pubKeyPEM,
	}
	if err := h.domainRepo.Create(r.Context(), d); err != nil {
		h.logger.Error("Setup: failed to create domain", "error", err)
		http.Error(w, "Failed to create domain", http.StatusInternalServerError)
		return
	}

	// Create admin user
	hash, err := bcrypt.GenerateFromPassword([]byte(req.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		h.logger.Error("Setup: failed to hash password", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	user := &domain.User{
		Email:        req.AdminEmail,
		PasswordHash: string(hash),
		Role:         domain.RoleAdmin,
		CreatedAt:    time.Now(),
	}
	if err := h.userRepo.Create(r.Context(), user); err != nil {
		h.logger.Error("Setup: failed to create admin user", "error", err)
		http.Error(w, "Failed to create admin user", http.StatusInternalServerError)
		return
	}

	// Build DNS records
	hostname := req.SMTPHostname
	if hostname == "" {
		hostname = "mail." + req.Domain
	}

	dkimValue := formatDKIMDNSValue(pubKeyPEM)

	dnsRecords := []DNSRecord{
		{Type: "MX", Name: req.Domain, Value: fmt.Sprintf("10 %s.", hostname)},
		{Type: "A", Name: hostname, Value: "<YOUR_SERVER_IP>"},
		{Type: "TXT", Name: req.Domain, Value: "v=spf1 mx -all"},
		{Type: "TXT", Name: fmt.Sprintf("default._domainkey.%s", req.Domain), Value: dkimValue},
		{Type: "TXT", Name: fmt.Sprintf("_dmarc.%s", req.Domain), Value: fmt.Sprintf("v=DMARC1; p=quarantine; rua=mailto:postmaster@%s", req.Domain)},
	}

	h.logger.Info("Setup completed", "domain", req.Domain, "admin", req.AdminEmail)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(SetupCompleteResponse{
		Success:    true,
		DNSRecords: dnsRecords,
	})
}

func formatDKIMDNSValue(pubKeyPEM string) string {
	block, _ := pem.Decode([]byte(pubKeyPEM))
	if block == nil {
		return ""
	}
	b64 := base64.StdEncoding.EncodeToString(block.Bytes)
	return fmt.Sprintf("v=DKIM1; k=rsa; p=%s", b64)
}
