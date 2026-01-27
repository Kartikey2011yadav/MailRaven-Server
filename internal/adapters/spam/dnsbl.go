package spam

import (
	"fmt"

	"github.com/mrichman/godnsbl"
)

// DNSBLChecker checks IPs against DNS Blocklists
type DNSBLChecker struct {
	providers []string
}

// NewDNSBLChecker creates a new DNSBL checker
func NewDNSBLChecker(providers []string) *DNSBLChecker {
	return &DNSBLChecker{
		providers: providers,
	}
}

// Check checks if the IP is listed in any of the configured blocklists
func (c *DNSBLChecker) Check(ip string) error {
	if len(c.providers) == 0 {
		return nil
	}

	for _, provider := range c.providers {
		result := godnsbl.Lookup(provider, ip)

		for _, res := range result.Results {
			if res.Error {
				// Log error (res.ErrorType) but don't fail just because of DNS lookup error
				continue
			}
			if res.Listed {
				return fmt.Errorf("IP listed in %s: %s", provider, res.Text)
			}
		}
	}
	return nil
}
