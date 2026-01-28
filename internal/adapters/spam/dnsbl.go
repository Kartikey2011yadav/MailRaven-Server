package spam

import (
	"fmt"
	"net"
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

// Check checks if an IP is listed
func (c *DNSBLChecker) Check(ip string) error {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return fmt.Errorf("invalid IP: %s", ip)
	}

	if parsedIP.IsPrivate() || parsedIP.IsLoopback() {
		return nil
	}

	v4 := parsedIP.To4()
	if v4 == nil {
		return nil // Only IPv4
	}

	reverseIP := fmt.Sprintf("%d.%d.%d.%d", v4[3], v4[2], v4[1], v4[0])

	for _, list := range c.providers {
		query := fmt.Sprintf("%s.%s", reverseIP, list)
		ips, err := net.LookupHost(query)
		if err == nil && len(ips) > 0 {
			return fmt.Errorf("blocked by %s", list)
		}
	}
	return nil
}
