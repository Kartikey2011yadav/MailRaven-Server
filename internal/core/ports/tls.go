package ports

import "crypto/tls"

// CertificateManager defines the interface for managing TLS certificates.
type CertificateManager interface {
	// GetCertificate returns a Certificate based on the ClientHelloInfo.
	// This matches the signature of tls.Config.GetCertificate.
	GetCertificate(hello *tls.ClientHelloInfo) (*tls.Certificate, error)
}
