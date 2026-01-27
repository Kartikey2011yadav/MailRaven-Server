package handlers

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"time"

	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/ports"
	"github.com/Kartikey2011yadav/mailraven-server/internal/observability"
	"github.com/go-chi/chi/v5"
)

type AdminDomainHandler struct {
	repo   ports.DomainRepository
	logger *observability.Logger
}

func NewAdminDomainHandler(repo ports.DomainRepository, logger *observability.Logger) *AdminDomainHandler {
	return &AdminDomainHandler{repo: repo, logger: logger}
}

type CreateDomainRequest struct {
	Name string `json:"name"`
}

// ListDomains GET /api/v1/admin/domains
func (h *AdminDomainHandler) ListDomains(w http.ResponseWriter, r *http.Request) {
	domains, err := h.repo.List(r.Context(), 100, 0)
	if err != nil {
		h.logger.Error("Failed to list domains", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(domains)
}

// CreateDomain POST /api/v1/admin/domains
func (h *AdminDomainHandler) CreateDomain(w http.ResponseWriter, r *http.Request) {
	var req CreateDomainRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Domain name is required", http.StatusBadRequest)
		return
	}

	// Generate DKIM keys
	privKey, pubKey, err := generateDKIMKeys()
	if err != nil {
		h.logger.Error("Failed to generate DKIM keys", "error", err)
		http.Error(w, "Failed to generate DKIM keys", http.StatusInternalServerError)
		return
	}

	d := &domain.Domain{
		Name:           req.Name,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Active:         true,
		DKIMSelector:   "default",
		DKIMPrivateKey: privKey,
		DKIMPublicKey:  pubKey,
	}

	if err := h.repo.Create(r.Context(), d); err != nil {
		if err == ports.ErrAlreadyExists {
			http.Error(w, "Domain already exists", http.StatusConflict)
			return
		}
		h.logger.Error("Failed to create domain", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(d)
}

// DeleteDomain DELETE /api/v1/admin/domains/{name}
func (h *AdminDomainHandler) DeleteDomain(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if name == "" {
		http.Error(w, "Domain name required", http.StatusBadRequest)
		return
	}

	if err := h.repo.Delete(r.Context(), name); err != nil {
		if err == ports.ErrNotFound {
			http.Error(w, "Domain not found", http.StatusNotFound)
			return
		}
		h.logger.Error("Failed to delete domain", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Helper to generate 2048-bit RSA keys for DKIM
func generateDKIMKeys() (string, string, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", err
	}

	// Private Key to PEM
	privBytes := x509.MarshalPKCS1PrivateKey(key)
	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privBytes,
	})

	// Public Key to PEM (PKIX)
	pubBytes, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return "", "", err
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	})

	return string(privPEM), string(pubPEM), nil
}
