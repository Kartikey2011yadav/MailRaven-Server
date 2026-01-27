package services

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/crypto/acme/autocert"

	"github.com/Kartikey2011yadav/mailraven-server/internal/config"
)

// ACMEService handles automatic TLS certificate management using Let's Encrypt
type ACMEService struct {
	manager *autocert.Manager
	config  config.ACMEConfig
}

// NewACMEService creates a new ACMEService instance
func NewACMEService(cfg config.ACMEConfig) (*ACMEService, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	// Ensure cache directory exists
	if cfg.CacheDir != "" {
		if err := os.MkdirAll(cfg.CacheDir, 0700); err != nil {
			return nil, fmt.Errorf("failed to create ACME cache directory: %w", err)
		}
	} else {
		return nil, fmt.Errorf("ACME cache directory is required")
	}

	m := &autocert.Manager{
		Cache:      autocert.DirCache(cfg.CacheDir),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(cfg.Domains...),
		Email:      cfg.Email,
	}

	return &ACMEService{
		manager: m,
		config:  cfg,
	}, nil
}

// GetCertificate implements ports.CertificateManager
func (s *ACMEService) GetCertificate(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	return s.manager.GetCertificate(hello)
}

// HTTPHandler returns a handler that resolves ACME challenges and redirects other traffic to HTTPS
func (s *ACMEService) HTTPHandler(fallback http.Handler) http.Handler {
	return s.manager.HTTPHandler(fallback)
}

// TLSConfig returns a TLS configuration that uses ACME
func (s *ACMEService) TLSConfig() *tls.Config {
	return s.manager.TLSConfig()
}
