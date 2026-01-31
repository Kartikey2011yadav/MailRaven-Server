package domain

import (
	"fmt"
	"strings"
)

// BuildPolicyString generates the text content of the mta-sts.txt file
// Format:
// version: STSv1
// mode: enforce
// mx: mail.example.com
// mx: *.example.net
// max_age: 86400
func (p *MTASTSPolicy) BuildPolicyString() string {
	var sb strings.Builder

	// Version is required
	if p.Version == "" {
		p.Version = "STSv1"
	}
	sb.WriteString(fmt.Sprintf("version: %s\n", p.Version))

	// Mode is required
	if p.Mode == "" {
		p.Mode = MTASTSModeTesting
	}
	sb.WriteString(fmt.Sprintf("mode: %s\n", p.Mode))

	// MX records are required
	for _, mx := range p.MX {
		sb.WriteString(fmt.Sprintf("mx: %s\n", mx))
	}

	// MaxAge is required
	if p.MaxAge == 0 {
		p.MaxAge = 86400 // Default 1 day
	}
	sb.WriteString(fmt.Sprintf("max_age: %d\n", p.MaxAge))

	return sb.String()
}
