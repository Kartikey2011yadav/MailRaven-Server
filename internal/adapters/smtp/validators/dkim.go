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

// VerifyDKIM validates DKIM signature in email message per RFC 6376
func VerifyDKIM(_ context.Context, rawMessage []byte) (DKIMResult, string, error) {
	sig := extractDKIMSignature(rawMessage)
	if sig == "" {
		return DKIMNone, "", nil
	}

	params := parseDKIMParams(sig)

	domain, ok := params["d"]
	if !ok {
		return DKIMPermError, "", fmt.Errorf("missing domain (d=) tag")
	}

	selector, ok := params["s"]
	if !ok {
		return DKIMPermError, "", fmt.Errorf("missing selector (s=) tag")
	}

	bodyHashB64, ok := params["bh"]
	if !ok {
		return DKIMPermError, "", fmt.Errorf("missing body hash (bh=) tag")
	}

	signatureB64, ok := params["b"]
	if !ok {
		return DKIMPermError, "", fmt.Errorf("missing signature (b=) tag")
	}

	headersToVerify := strings.Split(params["h"], ":")
	for i := range headersToVerify {
		headersToVerify[i] = strings.TrimSpace(headersToVerify[i])
	}
	if len(headersToVerify) == 0 || headersToVerify[0] == "" {
		return DKIMPermError, "", fmt.Errorf("missing headers (h=) tag")
	}

	canon := params["c"]
	if canon == "" {
		canon = "simple/simple"
	}
	canonParts := strings.SplitN(canon, "/", 2)
	headerCanon := canonParts[0]
	bodyCanon := headerCanon
	if len(canonParts) == 2 {
		bodyCanon = canonParts[1]
	}

	// DNS lookup for public key
	dkimRecord := fmt.Sprintf("%s._domainkey.%s", selector, domain)
	txtRecords, err := net.LookupTXT(dkimRecord)
	if err != nil {
		return DKIMTempError, domain, fmt.Errorf("DNS lookup failed: %w", err)
	}
	if len(txtRecords) == 0 {
		return DKIMFail, domain, fmt.Errorf("no DKIM record found")
	}

	publicKey, err := parseDKIMPublicKey(strings.Join(txtRecords, ""))
	if err != nil {
		return DKIMPermError, domain, fmt.Errorf("failed to parse public key: %w", err)
	}

	// Verify body hash
	msgStr := string(rawMessage)
	bodyStart := strings.Index(msgStr, "\r\n\r\n")
	body := ""
	if bodyStart >= 0 {
		body = msgStr[bodyStart+4:]
	}

	computedBodyHash := canonicalizeAndHashBody(body, bodyCanon)
	expectedBodyHash, err := base64.StdEncoding.DecodeString(bodyHashB64)
	if err != nil {
		return DKIMFail, domain, fmt.Errorf("invalid body hash encoding")
	}
	if string(computedBodyHash) != string(expectedBodyHash) {
		return DKIMFail, domain, fmt.Errorf("body hash mismatch")
	}

	// Verify header signature
	signatureBytes, err := base64.StdEncoding.DecodeString(signatureB64)
	if err != nil {
		return DKIMPermError, domain, fmt.Errorf("failed to decode signature: %w", err)
	}

	headerHash := canonicalizeAndHashHeaders(rawMessage, headersToVerify, sig, headerCanon)

	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, headerHash, signatureBytes)
	if err != nil {
		return DKIMFail, domain, fmt.Errorf("signature verification failed: %w", err)
	}

	return DKIMPass, domain, nil
}

// canonicalizeAndHashBody applies RFC 6376 body canonicalization then SHA256
func canonicalizeAndHashBody(body string, mode string) []byte {
	h := sha256.New()

	lines := strings.Split(body, "\r\n")

	if mode == "relaxed" {
		for i, line := range lines {
			line = strings.TrimRight(line, " \t")
			var b strings.Builder
			lastSpace := false
			for _, r := range line {
				if r == ' ' || r == '\t' {
					if !lastSpace {
						b.WriteRune(' ')
						lastSpace = true
					}
				} else {
					b.WriteRune(r)
					lastSpace = false
				}
			}
			lines[i] = b.String()
		}
		// Remove trailing empty lines
		for len(lines) > 0 && lines[len(lines)-1] == "" {
			lines = lines[:len(lines)-1]
		}
	} else {
		// Simple: remove trailing empty lines only
		for len(lines) > 0 && lines[len(lines)-1] == "" {
			lines = lines[:len(lines)-1]
		}
	}

	if len(lines) == 0 {
		h.Write([]byte("\r\n"))
	} else {
		for _, line := range lines {
			h.Write([]byte(line + "\r\n"))
		}
	}

	return h.Sum(nil)
}

// canonicalizeAndHashHeaders builds the header hash input per RFC 6376 Section 3.7
func canonicalizeAndHashHeaders(rawMessage []byte, headersToVerify []string, dkimSigValue string, mode string) []byte {
	h := sha256.New()

	headerSection, _ := splitHeaderBody(rawMessage)
	headerMap := parseHeaders(headerSection)

	for _, name := range headersToVerify {
		lowerName := strings.ToLower(strings.TrimSpace(name))
		if values, ok := headerMap[lowerName]; ok && len(values) > 0 {
			val := values[0]
			values = values[1:] // consume in order
			headerMap[lowerName] = values

			if mode == "relaxed" {
				h.Write([]byte(relaxedHeaderLine(lowerName, val) + "\r\n"))
			} else {
				h.Write([]byte(name + ":" + val + "\r\n"))
			}
		}
	}

	// Append DKIM-Signature header with b= value empty
	dkimSigClean := removeBValue(dkimSigValue)
	if mode == "relaxed" {
		h.Write([]byte(relaxedHeaderLine("dkim-signature", dkimSigClean)))
	} else {
		h.Write([]byte("DKIM-Signature:" + dkimSigClean))
	}

	return h.Sum(nil)
}

func relaxedHeaderLine(name, value string) string {
	lowerKey := strings.ToLower(strings.TrimSpace(name))
	cleanValue := strings.ReplaceAll(value, "\r\n", "")
	cleanValue = strings.ReplaceAll(cleanValue, "\n", "")

	var b strings.Builder
	lastSpace := false
	for _, r := range cleanValue {
		if r == ' ' || r == '\t' {
			if !lastSpace {
				b.WriteRune(' ')
				lastSpace = true
			}
		} else {
			b.WriteRune(r)
			lastSpace = false
		}
	}
	return lowerKey + ":" + strings.TrimSpace(b.String())
}

func removeBValue(sigValue string) string {
	parts := strings.Split(sigValue, ";")
	for i, part := range parts {
		trimmed := strings.TrimSpace(part)
		if strings.HasPrefix(trimmed, "b=") {
			parts[i] = " b="
		}
	}
	return strings.Join(parts, ";")
}

func splitHeaderBody(raw []byte) (string, string) {
	s := string(raw)
	idx := strings.Index(s, "\r\n\r\n")
	if idx < 0 {
		return s, ""
	}
	return s[:idx], s[idx+4:]
}

func parseHeaders(headerSection string) map[string][]string {
	result := make(map[string][]string)
	lines := strings.Split(headerSection, "\r\n")

	var currentName string
	var currentValue string

	for _, line := range lines {
		if len(line) > 0 && (line[0] == ' ' || line[0] == '\t') {
			currentValue += line
			continue
		}
		if currentName != "" {
			result[strings.ToLower(currentName)] = append(result[strings.ToLower(currentName)], currentValue)
		}
		colonIdx := strings.IndexByte(line, ':')
		if colonIdx > 0 {
			currentName = line[:colonIdx]
			currentValue = line[colonIdx+1:]
		} else {
			currentName = ""
			currentValue = ""
		}
	}
	if currentName != "" {
		result[strings.ToLower(currentName)] = append(result[strings.ToLower(currentName)], currentValue)
	}
	return result
}

func extractDKIMSignature(rawMessage []byte) string {
	headerSection, _ := splitHeaderBody(rawMessage)
	lines := strings.Split(headerSection, "\r\n")

	var sigValue strings.Builder
	inSig := false

	for _, line := range lines {
		if strings.HasPrefix(strings.ToLower(line), "dkim-signature:") {
			inSig = true
			sigValue.WriteString(strings.TrimPrefix(line, line[:strings.IndexByte(line, ':')+1]))
			continue
		}
		if inSig {
			if len(line) > 0 && (line[0] == ' ' || line[0] == '\t') {
				sigValue.WriteString(line)
			} else {
				break
			}
		}
	}
	return sigValue.String()
}

func parseDKIMParams(signature string) map[string]string {
	params := make(map[string]string)
	clean := strings.ReplaceAll(signature, "\r\n", "")
	clean = strings.ReplaceAll(clean, "\n", "")

	pairs := strings.Split(clean, ";")
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			params[key] = val
		}
	}
	return params
}

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
