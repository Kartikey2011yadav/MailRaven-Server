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

// EvaluateDMARC checks DMARC policy and validates alignment
// Implements RFC 7489 - Domain-based Message Authentication, Reporting, and Conformance
func EvaluateDMARC(ctx context.Context, sender string, spfResult SPFResult, dkimResult DKIMResult) (DMARCResult, DMARCPolicy, error) {
	// Extract domain from sender
	parts := strings.Split(sender, "@")
	if len(parts) != 2 {
		return DMARCFail, DMARCPolicyNone, fmt.Errorf("invalid sender format")
	}
	domain := parts[1]

	// RFC 7489 Section 6.6.3: DNS query for DMARC policy
	dmarcRecord := fmt.Sprintf("_dmarc.%s", domain)
	txtRecords, err := net.LookupTXT(dmarcRecord)
	if err != nil {
		// No DMARC record found
		return DMARCNone, DMARCPolicyNone, nil
	}

	if len(txtRecords) == 0 {
		return DMARCNone, DMARCPolicyNone, nil
	}

	// Parse DMARC record
	// RFC 7489 Section 6.3: DMARC record format
	policy := parseDMARCPolicy(txtRecords[0])

	// RFC 7489 Section 3.1: DMARC evaluation
	// Check SPF and DKIM alignment
	spfAligned := spfResult == SPFPass
	dkimAligned := dkimResult == DKIMPass

	// RFC 7489 Section 3.1.1: At least one must pass
	if spfAligned || dkimAligned {
		return DMARCPass, policy, nil
	}

	return DMARCFail, policy, nil
}

// parseDMARCPolicy extracts policy from DMARC DNS record
func parseDMARCPolicy(record string) DMARCPolicy {
	// DMARC record format: v=DMARC1; p=quarantine; ...
	if !strings.HasPrefix(record, "v=DMARC1") {
		return DMARCPolicyNone
	}

	pairs := strings.Split(record, ";")
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if strings.HasPrefix(pair, "p=") {
			policyValue := strings.TrimPrefix(pair, "p=")
			policyValue = strings.TrimSpace(policyValue)

			switch policyValue {
			case "reject":
				return DMARCPolicyReject
			case "quarantine":
				return DMARCPolicyQuarantine
			case "none":
				return DMARCPolicyNone
			default:
				return DMARCPolicyNone
			}
		}
	}

	// RFC 7489 Section 6.3: Default policy is "none"
	return DMARCPolicyNone
}

// ShouldRejectMessage determines if message should be rejected based on DMARC
func ShouldRejectMessage(dmarcResult DMARCResult, dmarcPolicy DMARCPolicy) bool {
	// RFC 7489 Section 6.3: Policy application
	if dmarcResult == DMARCFail {
		return dmarcPolicy == DMARCPolicyReject
	}
	return false
}

// ShouldQuarantineMessage determines if message should be quarantined
func ShouldQuarantineMessage(dmarcResult DMARCResult, dmarcPolicy DMARCPolicy) bool {
	// RFC 7489 Section 6.3: Policy application
	if dmarcResult == DMARCFail {
		return dmarcPolicy == DMARCPolicyQuarantine
	}
	return false
}
