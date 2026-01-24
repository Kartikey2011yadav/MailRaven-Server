package config

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
	"strings"
)

// DKIMKeyPair represents a DKIM RSA key pair
type DKIMKeyPair struct {
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
}

// GenerateDKIMKey generates a new RSA-2048 key pair for DKIM signing
func GenerateDKIMKey() (*DKIMKeyPair, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}

	return &DKIMKeyPair{
		PrivateKey: privateKey,
		PublicKey:  &privateKey.PublicKey,
	}, nil
}

// SavePrivateKey writes the private key to a PEM file
func (kp *DKIMKeyPair) SavePrivateKey(path string) error {
	// Encode private key to PKCS#1 ASN.1 DER
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(kp.PrivateKey)

	// Create PEM block
	pemBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	// Write to file with restricted permissions
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create private key file: %w", err)
	}
	defer file.Close()

	if err := pem.Encode(file, pemBlock); err != nil {
		return fmt.Errorf("failed to write PEM data: %w", err)
	}

	return nil
}

// GetPublicKeyDNS returns the public key in DNS TXT record format (base64-encoded)
func (kp *DKIMKeyPair) GetPublicKeyDNS() (string, error) {
	// Encode public key to PKIX format
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(kp.PublicKey)
	if err != nil {
		return "", fmt.Errorf("failed to marshal public key: %w", err)
	}

	// Base64 encode for DNS
	publicKeyB64 := base64.StdEncoding.EncodeToString(publicKeyBytes)

	// Split into 255-character chunks for DNS TXT record (RFC requirement)
	// Format: v=DKIM1; k=rsa; p=<base64>
	chunks := splitIntoChunks(publicKeyB64, 200) // Leave room for prefix
	return strings.Join(chunks, "\" \""), nil
}

// splitIntoChunks splits a string into chunks of specified size
func splitIntoChunks(s string, chunkSize int) []string {
	var chunks []string
	for len(s) > chunkSize {
		chunks = append(chunks, s[:chunkSize])
		s = s[chunkSize:]
	}
	if len(s) > 0 {
		chunks = append(chunks, s)
	}
	return chunks
}

// LoadPrivateKey loads a DKIM private key from a PEM file
func LoadPrivateKey(path string) (*DKIMKeyPair, error) {
	// Read PEM file
	pemData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	// Decode PEM block
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	// Parse RSA private key
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return &DKIMKeyPair{
		PrivateKey: privateKey,
		PublicKey:  &privateKey.PublicKey,
	}, nil
}
