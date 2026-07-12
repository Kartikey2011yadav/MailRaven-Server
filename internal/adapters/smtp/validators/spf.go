package validators

import (
	"context"
	"fmt"
	"net"
	"strings"
)

// SPFResult represents the result of an SPF check
type SPFResult string

const (
	SPFPass      SPFResult = "pass"
	SPFFail      SPFResult = "fail"
	SPFSoftFail  SPFResult = "softfail"
	SPFNeutral   SPFResult = "neutral"
	SPFNone      SPFResult = "none"
	SPFTempError SPFResult = "temperror"
	SPFPermError SPFResult = "permerror"
)

const maxSPFLookups = 10

// ValidateSPF checks SPF record for the given sender per RFC 7208
func ValidateSPF(_ context.Context, remoteIP, sender, _ string) (SPFResult, error) {
	parts := strings.Split(sender, "@")
	if len(parts) != 2 {
		return SPFPermError, fmt.Errorf("invalid sender format")
	}
	domain := parts[1]

	lookups := 0
	visited := make(map[string]bool)
	return validateSPFInternal(remoteIP, domain, &lookups, visited)
}

func validateSPFInternal(remoteIP, domain string, lookups *int, visited map[string]bool) (SPFResult, error) {
	domain = strings.ToLower(domain)

	if visited[domain] {
		return SPFPermError, fmt.Errorf("SPF include loop detected for %s", domain)
	}
	visited[domain] = true

	*lookups++
	if *lookups > maxSPFLookups {
		return SPFPermError, fmt.Errorf("SPF DNS lookup limit exceeded (max %d)", maxSPFLookups)
	}

	txtRecords, err := net.LookupTXT(domain)
	if err != nil {
		return SPFTempError, fmt.Errorf("DNS lookup failed for %s: %w", domain, err)
	}

	var spfRecord string
	for _, record := range txtRecords {
		if strings.HasPrefix(record, "v=spf1") {
			spfRecord = record
			break
		}
	}

	if spfRecord == "" {
		return SPFNone, nil
	}

	mechanisms := strings.Fields(spfRecord)
	var redirectDomain string

	for _, mech := range mechanisms[1:] {
		qualifier := "+"
		mechanism := mech

		if len(mech) > 0 && strings.ContainsRune("+-~?", rune(mech[0])) {
			qualifier = string(mech[0])
			mechanism = mech[1:]
		}

		// redirect modifier (processed after all mechanisms)
		if strings.HasPrefix(mechanism, "redirect=") {
			redirectDomain = strings.TrimPrefix(mechanism, "redirect=")
			continue
		}

		if mechanism == "all" {
			return qualifierToResult(qualifier), nil
		}

		if strings.HasPrefix(mechanism, "ip4:") {
			cidr := strings.TrimPrefix(mechanism, "ip4:")
			if matchesIP(remoteIP, cidr) {
				return qualifierToResult(qualifier), nil
			}
		}

		if strings.HasPrefix(mechanism, "ip6:") {
			cidr := strings.TrimPrefix(mechanism, "ip6:")
			if matchesIP(remoteIP, cidr) {
				return qualifierToResult(qualifier), nil
			}
		}

		if mechanism == "mx" || strings.HasPrefix(mechanism, "mx:") {
			mxDomain := domain
			if strings.HasPrefix(mechanism, "mx:") {
				mxDomain = strings.TrimPrefix(mechanism, "mx:")
			}
			*lookups++
			if *lookups > maxSPFLookups {
				return SPFPermError, fmt.Errorf("SPF DNS lookup limit exceeded")
			}
			if matchesMX(remoteIP, mxDomain) {
				return qualifierToResult(qualifier), nil
			}
		}

		if mechanism == "a" || strings.HasPrefix(mechanism, "a:") {
			aDomain := domain
			if strings.HasPrefix(mechanism, "a:") {
				aDomain = strings.TrimPrefix(mechanism, "a:")
			}
			*lookups++
			if *lookups > maxSPFLookups {
				return SPFPermError, fmt.Errorf("SPF DNS lookup limit exceeded")
			}
			if matchesA(remoteIP, aDomain) {
				return qualifierToResult(qualifier), nil
			}
		}

		if strings.HasPrefix(mechanism, "include:") {
			includeDomain := strings.TrimPrefix(mechanism, "include:")
			result, err := validateSPFInternal(remoteIP, includeDomain, lookups, visited)
			if err != nil {
				return SPFTempError, fmt.Errorf("include %s: %w", includeDomain, err)
			}
			if result == SPFPass {
				return qualifierToResult(qualifier), nil
			}
		}
	}

	// RFC 7208 Section 6.1: redirect modifier processed if no match
	if redirectDomain != "" {
		return validateSPFInternal(remoteIP, redirectDomain, lookups, visited)
	}

	return SPFNeutral, nil
}

func qualifierToResult(qualifier string) SPFResult {
	switch qualifier {
	case "+":
		return SPFPass
	case "-":
		return SPFFail
	case "~":
		return SPFSoftFail
	case "?":
		return SPFNeutral
	default:
		return SPFNeutral
	}
}

func matchesIP(remoteIP, cidr string) bool {
	if !strings.Contains(cidr, "/") {
		if strings.Contains(cidr, ":") {
			cidr += "/128"
		} else {
			cidr += "/32"
		}
	}

	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}

	ip := net.ParseIP(remoteIP)
	if ip == nil {
		return false
	}

	return ipNet.Contains(ip)
}

func matchesMX(remoteIP, domain string) bool {
	mxRecords, err := net.LookupMX(domain)
	if err != nil {
		return false
	}

	for _, mx := range mxRecords {
		if matchesA(remoteIP, mx.Host) {
			return true
		}
	}

	return false
}

func matchesA(remoteIP, domain string) bool {
	ips, err := net.LookupIP(domain)
	if err != nil {
		return false
	}

	remoteIPParsed := net.ParseIP(remoteIP)
	if remoteIPParsed == nil {
		return false
	}

	for _, ip := range ips {
		if ip.Equal(remoteIPParsed) {
			return true
		}
	}

	return false
}
