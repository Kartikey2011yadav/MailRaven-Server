package config

import (
	"fmt"

	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
)

// GenerateDNSRecords creates DNS records for a mail domain
func GenerateDNSRecords(cfg *Config, dkimPublicKey string) []domain.DNSRecord {
	records := []domain.DNSRecord{
		// MX Record - points to the mail server
		{
			Type:     "MX",
			Name:     "@",
			Value:    fmt.Sprintf("%s.", cfg.SMTP.Hostname),
			TTL:      3600,
			Priority: 10,
		},

		// SPF Record - allows this server to send mail
		{
			Type:  "TXT",
			Name:  "@",
			Value: fmt.Sprintf("v=spf1 mx ip4:%s -all", "YOUR_SERVER_IP"),
			TTL:   3600,
		},

		// DKIM Record - public key for signature verification
		{
			Type:  "TXT",
			Name:  fmt.Sprintf("%s._domainkey", cfg.DKIM.Selector),
			Value: fmt.Sprintf("v=DKIM1; k=rsa; p=%s", dkimPublicKey),
			TTL:   3600,
		},

		// DMARC Record - email authentication policy
		{
			Type:  "TXT",
			Name:  "_dmarc",
			Value: "v=DMARC1; p=quarantine; rua=mailto:dmarc@" + cfg.Domain,
			TTL:   3600,
		},
	}

	return records
}

// FormatDNSRecordsForConsole formats DNS records for human-readable output
func FormatDNSRecordsForConsole(records []domain.DNSRecord) string {
	output := "\n=== DNS Configuration ===\n\n"
	output += "Add these records to your DNS provider:\n\n"

	for _, record := range records {
		output += fmt.Sprintf("Type: %s\n", record.Type)
		output += fmt.Sprintf("Name: %s\n", record.Name)

		if record.Type == "MX" {
			output += fmt.Sprintf("Priority: %d\n", record.Priority)
		}

		output += fmt.Sprintf("Value: %s\n", record.Value)
		output += fmt.Sprintf("TTL: %d seconds\n", record.TTL)
		output += "\n"
	}

	output += "Note: Replace 'YOUR_SERVER_IP' in the SPF record with your actual server IP address.\n"
	output += "DNS propagation can take up to 48 hours, but usually completes within 1-2 hours.\n"

	return output
}

// ValidateDNSConfiguration checks if DNS records appear correctly configured
// (For MVP, this is a placeholder - real validation would query DNS)
func ValidateDNSConfiguration(domain string) error {
	// TODO: Implement DNS lookup validation
	// For now, just return nil (assume DNS is configured)
	return nil
}
