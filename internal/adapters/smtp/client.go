package smtp

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"

	"github.com/Kartikey2011yadav/mailraven-server/internal/observability"
	"golang.org/x/net/idna"
)

// Client handles sending emails to external SMTP servers
type Client struct {
	logger  *observability.Logger
	dialer  *net.Dialer
	tlsConf *tls.Config
	Port    string // SMTP port to connect to (default "25")

	// LookupMX is the function used to look up MX records. Defaults to net.LookupMX.
	LookupMX func(name string) ([]*net.MX, error)
}

// NewClient creates a new SMTP client
func NewClient(logger *observability.Logger) *Client {
	return &Client{
		logger: logger,
		dialer: &net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		},
		tlsConf: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
		Port:     "25",
		LookupMX: net.LookupMX,
	}
}

// Send delivers a message to the recipient's mail server
// recipient is a full email address (e.g. "user@example.com")
// from is the sender email address
// data is the signed MIME message
func (c *Client) Send(ctx context.Context, from string, recipient string, data []byte) error {
	// 1. Extract domain from recipient
	parts := strings.Split(recipient, "@")
	if len(parts) != 2 {
		return fmt.Errorf("invalid recipient address: %s", recipient)
	}
	domain := parts[1]

	// Handle IDN (Internationalized Domain Names)
	asciiDomain, err := idna.ToASCII(domain)
	if err != nil {
		return fmt.Errorf("invalid domain %s: %w", domain, err)
	}

	// 2. Lookup MX records
	mxs, err := c.LookupMX(asciiDomain)
	if err != nil {
		// Fallback to A record if no MX (RFC 5321)
		// Go's net/smtp/LookupMX doesn't do A record fallback automatically for us in terms of returning dummy MX
		// But in practice, we should try MX first.
		// If LookupMX fails, it might be DNS temporary error or NXDOMAIN.
		// If it's NXDOMAIN for MX, we check A record.
		// Simplifying: Just treat lookup error as failure for now, unless it's strictly "no such host" AND we want to support A-record fallback.
		// Modern email delivery relies heavily on MX.
		return fmt.Errorf("MX lookup failed for %s: %w", asciiDomain, err)
	}

	if len(mxs) == 0 {
		// No MX records found. Try A record implicit MX (RFC 5321 Section 5)
		// We'll create a synthetic MX record pointing to the domain itself
		mxs = []*net.MX{{Host: asciiDomain, Pref: 0}}
	}

	// 3. Try each MX in order of preference
	// net.LookupMX returns sorted by preference usually, but let's be safe?
	// Actually strict RFC says we must sort. Go docs say "sorted by preference".

	var lastErr error
	for _, mx := range mxs {
		// MX host might be IDN too
		mxHost, err := idna.ToASCII(mx.Host)
		if err != nil {
			c.logger.Warn("invalid MX host", "host", mx.Host, "error", err)
			continue
		}

		c.logger.Info("attempting delivery", "mx", mxHost, "recipient", recipient)

		err = c.deliverToHost(ctx, mxHost, from, recipient, data)
		if err == nil {
			c.logger.Info("delivery successful", "mx", mxHost, "recipient", recipient)
			return nil
		}

		c.logger.Warn("delivery failed to MX", "mx", mxHost, "error", err)
		lastErr = err
	}

	return fmt.Errorf("delivery failed to all MX records for %s. Last error: %v", domain, lastErr)
}

func (c *Client) deliverToHost(ctx context.Context, host string, from string, to string, data []byte) error {
	// Add port
	port := c.Port
	if port == "" {
		port = "25"
	}
	addr := net.JoinHostPort(host, port)

	// Check context before dialing
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Connect
	conn, err := c.dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("connect failed: %w", err)
	}
	defer conn.Close()

	// Create SMTP client
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("smtp handshake failed: %w", err)
	}
	defer func() { _ = client.Quit() }() // Try to quit nicely

	// STARTTLS if supported
	if ok, _ := client.Extension("STARTTLS"); ok {
		// Clone config and set ServerName to prevent MITM
		tlsConf := c.tlsConf.Clone()
		tlsConf.ServerName = host

		if err := client.StartTLS(tlsConf); err != nil {
			return fmt.Errorf("starttls failed: %w", err)
		}
	}

	// Mail Command
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("MAIL FROM failed: %w", err)
	}

	// Rcpt Command
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("RCPT TO failed: %w", err)
	}

	// Data Command
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("DATA command failed: %w", err)
	}

	// Write Body
	if _, err := w.Write(data); err != nil {
		_ = w.Close() // Close writer to free resources, ignore error as we return original error
		return fmt.Errorf("write data failed: %w", err)
	}

	// Close Data writer to signal end of message (".")
	if err := w.Close(); err != nil {
		return fmt.Errorf("close data failed: %w", err)
	}

	return nil
}
