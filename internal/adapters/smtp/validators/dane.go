package validators

import (
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/miekg/dns"
)

// DANEValidator handles logic for fetching and verifying TLSA records
type DANEValidator struct {
	// DNS Client to make queries
	Client *dns.Client
	// Configurable DNS Server to query (e.g. "8.8.8.8:53" or local)
	DNSServer string
}

func NewDANEValidator(resolver string) *DANEValidator {
	if resolver == "" {
		// Use Google Public DNS by default if none provided,
		// though reliable AD bit usually requires a local recursive resolver (127.0.0.1:53)
		// or a trusted upstream that preserves DNSSEC data.
		resolver = "8.8.8.8:53"
	}
	return &DANEValidator{
		Client:    &dns.Client{},
		DNSServer: resolver,
	}
}

// CheckTLSA queries for TLSA records and verifies if any match the cert
// Returns nil if verification succeeds (or no records found/not applicable).
// Returns error if verification explicitly fails.
func (v *DANEValidator) CheckTLSA(domain string, port int, certs []*x509.Certificate) error {
	if len(certs) == 0 {
		return fmt.Errorf("no certificates provided")
	}
	leaf := certs[0]

	// Construct TLSA query name: _port._tcp.domain
	qName := fmt.Sprintf("_%d._tcp.%s.", port, dns.Fqdn(domain))

	msg := new(dns.Msg)
	msg.SetQuestion(qName, dns.TypeTLSA)
	msg.SetEdns0(4096, true) // Enable DNSSEC

	in, _, err := v.Client.Exchange(msg, v.DNSServer)
	if err != nil {
		// Network error querying DNS used to be soft-fail, but for high security might be an issue.
		// For now, treat as "can't verify" -> soft fail / allow connection unless policy enforcement is strict elsewhere.
		return fmt.Errorf("dns lookup failed: %w", err)
	}

	// If the AD (Authenticated Data) bit is not set, we cannot trust the denial of existence or the records.
	// RFC 7672: If DNSSEC is not validated (no AD bit), DANE is "unused".
	if !in.AuthenticatedData {
		// No DNSSEC -> No DANE. Return success (neutral).
		return nil
	}

	// Filter for TLSA records
	var tlsaRecords []*dns.TLSA
	for _, ans := range in.Answer {
		if t, ok := ans.(*dns.TLSA); ok {
			tlsaRecords = append(tlsaRecords, t)
		}
	}

	if len(tlsaRecords) == 0 {
		// Authenticated denial of existence -> No DANE records. neutral.
		return nil
	}

	// If we have records and AD bit is set, we MUST verify at least one matches.
	for _, record := range tlsaRecords {
		if err := verifyRecord(record, leaf); err == nil {
			return nil // Found a match! Secure.
		}
	}

	return fmt.Errorf("dane verification failed: certificate did not match any TLSA records")
}

func verifyRecord(r *dns.TLSA, cert *x509.Certificate) error {
	// Select data to verify based on Selector (0=Cert, 1=SPKI)
	var data []byte
	switch r.Selector {
	case 0: // Full Certificate
		data = cert.Raw
	case 1: // SubjectPublicKeyInfo
		data = cert.RawSubjectPublicKeyInfo
	default:
		return fmt.Errorf("unsupported selector %d", r.Selector)
	}

	// Match based on MatchingType (0=Exact, 1=SHA256, 2=SHA512)
	switch r.MatchingType {
	case 0: // Exact match
		return checkMatch(data, r.Certificate)
	case 1: // SHA2-256
		hash := sha256.Sum256(data)
		return checkMatch(hash[:], r.Certificate)
	case 2: // SHA2-512
		hash := sha512.Sum512(data)
		return checkMatch(hash[:], r.Certificate)
	default:
		return fmt.Errorf("unsupported matching type %d", r.MatchingType)
	}
	return nil
}

func checkMatch(calculated []byte, recordHex string) error {
	// r.Certificate in miekg/dns is a string of hex characters
	// We need to compare our calculated bytes against that hex string.
	// Easiest is to encode calculated to hex.
	calculatedHex := strings.ToUpper(hex.EncodeToString(calculated))
	recordHex = strings.ToUpper(recordHex)

	if calculatedHex == recordHex {
		return nil
	}
	return fmt.Errorf("mismatch: calculated %s != record %s", calculatedHex, recordHex)
}

func bytesEqual(a []byte, bHex string) bool {
	// Helper for case 0
	return checkMatch(a, bHex) == nil
}
