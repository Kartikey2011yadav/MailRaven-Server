package validators

import (
	"context"
	"fmt"
	"net"
	"strings"
)

// DMARCResult represents the result of a DMARC evaluation
type DMARCResult string

const (
	DMARCPass      DMARCResult = "pass"
	DMARCFail      DMARCResult = "fail"
	DMARCNone      DMARCResult = "none"
	DMARCTempError DMARCResult = "temperror"
)

// DMARCPolicy represents the DMARC policy from DNS
type DMARCPolicy string

const (
	DMARCPolicyNone       DMARCPolicy = "none"
	DMARCPolicyQuarantine DMARCPolicy = "quarantine"
	DMARCPolicyReject     DMARCPolicy = "reject"
)

// EvaluateDMARC checks DMARC policy and validates alignment per RFC 7489
func EvaluateDMARC(_ context.Context, sender string, spfResult SPFResult, spfDomain string, dkimResult DKIMResult, dkimDomain string) (DMARCResult, DMARCPolicy, error) {
	parts := strings.Split(sender, "@")
	if len(parts) != 2 {
		return DMARCFail, DMARCPolicyNone, fmt.Errorf("invalid sender format")
	}
	fromDomain := strings.ToLower(parts[1])

	dmarcRecord := fmt.Sprintf("_dmarc.%s", fromDomain)
	txtRecords, err := net.LookupTXT(dmarcRecord)
	if err != nil || len(txtRecords) == 0 {
		return DMARCNone, DMARCPolicyNone, nil
	}

	record := strings.Join(txtRecords, "")
	if !strings.HasPrefix(record, "v=DMARC1") {
		return DMARCNone, DMARCPolicyNone, nil
	}

	policy, aspf, adkim := parseDMARCRecord(record)

	// RFC 7489 Section 3.1: Check alignment
	spfAligned := spfResult == SPFPass && checkAlignment(fromDomain, spfDomain, aspf)
	dkimAligned := dkimResult == DKIMPass && checkAlignment(fromDomain, dkimDomain, adkim)

	if spfAligned || dkimAligned {
		return DMARCPass, policy, nil
	}

	return DMARCFail, policy, nil
}

// checkAlignment verifies domain alignment per RFC 7489 Section 3.1
func checkAlignment(fromDomain, authDomain, mode string) bool {
	if authDomain == "" {
		return false
	}
	fromDomain = strings.ToLower(fromDomain)
	authDomain = strings.ToLower(authDomain)

	if mode == "s" {
		return fromDomain == authDomain
	}
	// Relaxed (default): organizational domain must match
	return getOrgDomain(fromDomain) == getOrgDomain(authDomain)
}

// getOrgDomain extracts the organizational domain (base registerable domain).
// For "sub.mail.example.com" returns "example.com".
// Simple heuristic: take last two labels (handles .com, .org, .net, etc.)
// For ccTLDs like .co.uk, this is imperfect but acceptable for most cases.
func getOrgDomain(domain string) string {
	labels := strings.Split(domain, ".")
	if len(labels) <= 2 {
		return domain
	}
	return strings.Join(labels[len(labels)-2:], ".")
}

// parseDMARCRecord extracts policy and alignment modes from DMARC record
func parseDMARCRecord(record string) (DMARCPolicy, string, string) {
	policy := DMARCPolicyNone
	aspf := "r" // relaxed default
	adkim := "r"

	pairs := strings.Split(record, ";")
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if strings.HasPrefix(pair, "p=") {
			switch strings.TrimSpace(strings.TrimPrefix(pair, "p=")) {
			case "reject":
				policy = DMARCPolicyReject
			case "quarantine":
				policy = DMARCPolicyQuarantine
			case "none":
				policy = DMARCPolicyNone
			}
		} else if strings.HasPrefix(pair, "aspf=") {
			aspf = strings.TrimSpace(strings.TrimPrefix(pair, "aspf="))
		} else if strings.HasPrefix(pair, "adkim=") {
			adkim = strings.TrimSpace(strings.TrimPrefix(pair, "adkim="))
		}
	}

	return policy, aspf, adkim
}

// ShouldRejectMessage determines if message should be rejected based on DMARC
func ShouldRejectMessage(dmarcResult DMARCResult, dmarcPolicy DMARCPolicy) bool {
	return dmarcResult == DMARCFail && dmarcPolicy == DMARCPolicyReject
}

// ShouldQuarantineMessage determines if message should be quarantined
func ShouldQuarantineMessage(dmarcResult DMARCResult, dmarcPolicy DMARCPolicy) bool {
	return dmarcResult == DMARCFail && dmarcPolicy == DMARCPolicyQuarantine
}
