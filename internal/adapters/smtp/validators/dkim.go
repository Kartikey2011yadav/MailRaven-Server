package validators

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"net"
	"strings"
)

// DKIMResult represents the result of a DKIM verification
type DKIMResult string

const (
	DKIMPass      DKIMResult = "pass"
	DKIMFail      DKIMResult = "fail"
	DKIMNone      DKIMResult = "none"
	DKIMTempError DKIMResult = "temperror"
	DKIMPermError DKIMResult = "permerror"
)

// VerifyDKIM validates DKIM signature in email message
// Implements RFC 6376 - DomainKeys Identified Mail
func VerifyDKIM(ctx context.Context, rawMessage []byte) (DKIMResult, error) {
	// RFC 6376 Section 3.5: Parse DKIM-Signature header
	signature := extractDKIMSignature(rawMessage)
	if signature == "" {
		// No DKIM signature present
		return DKIMNone, nil
	}

	// Parse signature parameters
	// RFC 6376 Section 3.5: Tag-value list
	params := parseDKIMParams(signature)

	domain, ok := params["d"]
	if !ok {
		return DKIMPermError, fmt.Errorf("missing domain (d=) tag")
	}

	selector, ok := params["s"]
	if !ok {
		return DKIMPermError, fmt.Errorf("missing selector (s=) tag")
	}

	bodyHash, ok := params["bh"]
	if !ok {
		return DKIMPermError, fmt.Errorf("missing body hash (bh=) tag")
	}

	signatureB64, ok := params["b"]
	if !ok {
		return DKIMPermError, fmt.Errorf("missing signature (b=) tag")
	}

	// RFC 6376 Section 3.6: DNS query for public key
	dkimRecord := fmt.Sprintf("%s._domainkey.%s", selector, domain)
	txtRecords, err := net.LookupTXT(dkimRecord)
	if err != nil {
		return DKIMTempError, fmt.Errorf("DNS lookup failed: %w", err)
	}

	if len(txtRecords) == 0 {
		return DKIMFail, fmt.Errorf("no DKIM record found")
	}

	// Parse public key from DNS
	publicKey, err := parseDKIMPublicKey(txtRecords[0])
	if err != nil {
		return DKIMPermError, fmt.Errorf("failed to parse public key: %w", err)
	}

	// RFC 6376 Section 3.7: Compute body hash
	bodyHashComputed := computeBodyHash(rawMessage)
	bodyHashDecoded, _ := base64.StdEncoding.DecodeString(bodyHash)
	if string(bodyHashComputed) != string(bodyHashDecoded) {
		return DKIMFail, fmt.Errorf("body hash mismatch")
	}

	// RFC 6376 Section 3.8: Verify signature
	signatureBytes, err := base64.StdEncoding.DecodeString(signatureB64)
	if err != nil {
		return DKIMPermError, fmt.Errorf("failed to decode signature: %w", err)
	}

	// Compute hash of signed headers
	headerHash := computeHeaderHash(rawMessage, signature)

	// Verify RSA signature
	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, headerHash, signatureBytes)
	if err != nil {
		return DKIMFail, fmt.Errorf("signature verification failed: %w", err)
	}

	return DKIMPass, nil
}

// extractDKIMSignature extracts DKIM-Signature header from message
func extractDKIMSignature(rawMessage []byte) string {
	lines := strings.Split(string(rawMessage), "\r\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "DKIM-Signature:") {
			return strings.TrimPrefix(line, "DKIM-Signature:")
		}
	}
	return ""
}

// parseDKIMParams parses tag=value pairs from DKIM signature
func parseDKIMParams(signature string) map[string]string {
	params := make(map[string]string)
	signature = strings.ReplaceAll(signature, " ", "")
	signature = strings.ReplaceAll(signature, "\t", "")
	
	pairs := strings.Split(signature, ";")
	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			params[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	
	return params
}

// parseDKIMPublicKey extracts RSA public key from DKIM DNS record
func parseDKIMPublicKey(record string) (*rsa.PublicKey, error) {
	params := parseDKIMParams(record)
	
	publicKeyB64, ok := params["p"]
	if !ok {
		return nil, fmt.Errorf("no public key (p=) in DNS record")
	}

	publicKeyBytes, err := base64.StdEncoding.DecodeString(publicKeyB64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key: %w", err)
	}

	publicKeyInterface, err := x509.ParsePKIXPublicKey(publicKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	publicKey, ok := publicKeyInterface.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("public key is not RSA")
	}

	return publicKey, nil
}

// computeBodyHash computes SHA256 hash of message body (simplified for MVP)
func computeBodyHash(rawMessage []byte) []byte {
	// Find body (after \r\n\r\n)
	parts := strings.SplitN(string(rawMessage), "\r\n\r\n", 2)
	body := ""
	if len(parts) == 2 {
		body = parts[1]
	}

	hash := sha256.Sum256([]byte(body))
	return hash[:]
}

// computeHeaderHash computes SHA256 hash of signed headers (simplified for MVP)
func computeHeaderHash(rawMessage []byte, signature string) []byte {
	// Extract headers (before \r\n\r\n)
	parts := strings.SplitN(string(rawMessage), "\r\n\r\n", 2)
	headers := parts[0]

	hash := sha256.Sum256([]byte(headers))
	return hash[:]
}
