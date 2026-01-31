package validators

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
)

// Helper to generate a dummy certificate
func generateSelfSignedCert() (*x509.Certificate, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"MailRaven Test"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour),
		KeyUsage:  x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
		},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, err
	}

	return x509.ParseCertificate(derBytes)
}

// Since CheckTLSA makes real network calls, we can't easily unit test it without mocking the DNS client
// or spinning up a test DNS server.
// miekg/dns provides a way to run a test server easily.

func TestDANE_VerifyRecord(t *testing.T) {
	cert, err := generateSelfSignedCert()
	assert.NoError(t, err)

	// Calculate SHA256 of the FULL CERTIFICATE (Selector 0)
	hash0 := sha256.Sum256(cert.Raw)
	hex0 := hex.EncodeToString(hash0[:])

	// Record Matching: Selector=0, Matching=1(SHA256)
	rec0 := &dns.TLSA{
		Usage:        3, // DANE-EE
		Selector:     0, // Cert
		MatchingType: 1, // SHA256
		Certificate:  hex0,
	}

	err = verifyRecord(rec0, cert)
	assert.NoError(t, err, "Should match correctly")

	// Mismatch test
	recBad := &dns.TLSA{
		Usage:        3,
		Selector:     0,
		MatchingType: 1,
		Certificate:  "deadbeef",
	}
	err = verifyRecord(recBad, cert)
	assert.Error(t, err, "Should fail on mismatch")
}

func TestDANE_VerifySPKI(t *testing.T) {
	cert, err := generateSelfSignedCert()
	assert.NoError(t, err)

	// Calculate SHA256 of SPKI (Selector 1)
	hash1 := sha256.Sum256(cert.RawSubjectPublicKeyInfo)
	hex1 := hex.EncodeToString(hash1[:])

	rec1 := &dns.TLSA{
		Usage:        3, // DANE-EE
		Selector:     1, // SPKI
		MatchingType: 1, // SHA256
		Certificate:  hex1,
	}

	err = verifyRecord(rec1, cert)
	assert.NoError(t, err, "Should match SPKI correctly")
}
