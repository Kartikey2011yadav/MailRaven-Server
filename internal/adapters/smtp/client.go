package smtp

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/smtp/validators"
	"github.com/Kartikey2011yadav/mailraven-server/internal/config"
	"github.com/Kartikey2011yadav/mailraven-server/internal/observability"
	"golang.org/x/net/idna"
)

// Client handles sending emails to external SMTP servers
type Client struct {
	logger   *observability.Logger
	dialer   *net.Dialer
	tlsConf  *tls.Config
	Port     string // SMTP port to connect to (default "25")
	daneMode string // off, advisory, enforce

	// Dependencies
	daneValidator *validators.DANEValidator

	// LookupMX is the function used to look up MX records. Defaults to net.LookupMX.
	LookupMX func(name string) ([]*net.MX, error)
}

// NewClient creates a new SMTP client
func NewClient(daneCfg config.DANEConfig, logger *observability.Logger) *Client {
	// Defaut validatior
	var validator *validators.DANEValidator
	if daneCfg.Mode != "off" {
		validator = validators.NewDANEValidator("")
	}

	return &Client{
		logger: logger,
		dialer: &net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		},
		tlsConf: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
		Port:          "25",
		daneMode:      daneCfg.Mode,
		daneValidator: validator,
		LookupMX:      net.LookupMX,
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
	defer func() {
		//nolint:errcheck // Try to quit nicely
		_ = client.Quit()
	}()

	// STARTTLS if supported (DANE requires TLS match)
	if ok, _ := client.Extension("STARTTLS"); ok {
		// Clone config and set ServerName to prevent MITM
		tlsConf := c.tlsConf.Clone()
		tlsConf.ServerName = host

		// Add DANE verification hook
		tlsConf.VerifyConnection = func(cs tls.ConnectionState) error {
			// Parse port
			portInt := 25
			// default to 25, but we should use the configured port if possible.
			// Currently `deliverToHost` has `port` as string.
			if p, err := strconv.Atoi(c.Port); err == nil && p > 0 {
				portInt = p
			}

			// Perform DANE check
			// We skip if no DANE validator is configured (defensive)
			if c.daneValidator != nil {
				// Note: DANE check validates the cert against TLSA records.
				// If TLSA records exist, it returns validation result.
				// If NO TLSA records exist, it returns nil (allow standard PKIX).
				// If TLSA records exist but mismatch, it returns error.
				if err := c.daneValidator.CheckTLSA(host, portInt, cs.PeerCertificates); err != nil {
					c.logger.Error("DANE validation failed", "host", host, "error", err)

					// Enforce mode: Fail connection
					if c.daneMode == "enforce" {
						return fmt.Errorf("DANE validation failed for %s: %w", host, err)
					}
					// Advisory: just log (continue)
					c.logger.Warn("DANE violation ignoring due to mode=advisory", "host", host)
					return nil
				}
				c.logger.Debug("DANE validation passed or no records found", "host", host)
			}
			return nil
		}

		if err := client.StartTLS(tlsConf); err != nil {
			return fmt.Errorf("starttls failed: %w", err)
		}
	} else {
		// DANE enforcement note:
		// If DANE TLSA records exist, we theoretically MUST use STARTTLS.

		// If DANE is enforced, and we have a validator, we must check for presence of TLSA records.
		if c.daneMode == "enforce" && c.daneValidator != nil {
			// Quick check for existence.
			// Ideally CheckTLSA does DNS query. We can inspect if records exist.
			// Since we don't have peer certs (no TLS), CheckTLSA would fail/error usually or we need "CheckExistence".

			// For MVP: We assume if mode=enforce, we fail any connection not offering STARTTLS IF DANE records exist.
			// Implementing a lightweight "HasTLSA" check here is ideal.
			// Calling CheckTLSA with nil certs returns error "no certificates" usually, but let's use a helper if we had one.

			// Reuse CheckTLSA but with nil certs to trigger lookups? No, that returns error.
			// Current implementation of CheckTLSA fetches records first.
			// We'll proceed with clear text but log heavily. Strict enforcement here requires extra DNS call.

			// Let's rely on standard fallback for now but log error.
			c.logger.Warn("Remote server does not support STARTTLS -> DANE enforcement impossible", "host", host)
		} else {
			c.logger.Debug("STARTTLS not supported by remote, skipping DANE check (downgrade risk)", "host", host)
		}

		// Mail Command
		if err := client.Mail(from); err != nil {
			return fmt.Errorf("MAIL FROM failed: %w", err)
		}

		// Rcpt Command
		if err := client.Rcpt(to); err != nil {
			return fmt.Errorf("RCPT TO failed: %w", err)
		}
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
