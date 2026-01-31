package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the MailRaven server configuration
type Config struct {
	Domain      string            `yaml:"domain"` // Primary mail domain (e.g., mail.example.com)
	SMTP        SMTPConfig        `yaml:"smtp"`
	API         APIConfig         `yaml:"api"`
	Storage     StorageConfig     `yaml:"storage"`
	DKIM        DKIMConfig        `yaml:"dkim"`
	Logging     LogConfig         `yaml:"logging"`
	TLS         TLSConfig         `yaml:"tls"`
	Spam        SpamConfig        `yaml:"spam"`
	IMAP        IMAPConfig        `yaml:"imap"`
	Backup      BackupConfig      `yaml:"backup"`
	ManageSieve ManageSieveConfig `yaml:"managesieve"`
}

// SMTPConfig contains SMTP server settings
type SMTPConfig struct {
	Port     int        `yaml:"port"`     // SMTP listen port (default: 25)
	Hostname string     `yaml:"hostname"` // SMTP HELO hostname
	MaxSize  int64      `yaml:"max_size"` // Maximum message size in bytes (default: 10MB)
	DANE     DANEConfig `yaml:"dane"`     // DANE verification settings
}

// DANEConfig contains DANE verification settings
type DANEConfig struct {
	Mode string `yaml:"mode"` // Mode: "off", "advisory" (log only), "enforce" (fail delivery). Default: "advisory"
}

// APIConfig contains REST API settings
type APIConfig struct {
	Host      string `yaml:"host"`       // HTTP listen host (default: "0.0.0.0")
	Port      int    `yaml:"port"`       // HTTP listen port (default: 8443)
	TLS       bool   `yaml:"tls"`        // Enable TLS (default: false for dev)
	TLSCert   string `yaml:"tls_cert"`   // TLS certificate path (required if TLS=true)
	TLSKey    string `yaml:"tls_key"`    // TLS key path (required if TLS=true)
	JWTSecret string `yaml:"jwt_secret"` // JWT signing secret (required)
}

// StorageConfig contains database and blob storage settings
type StorageConfig struct {
	Driver   string `yaml:"driver"`    // "sqlite" or "postgres" (default: "sqlite")
	DBPath   string `yaml:"db_path"`   // SQLite: database file path
	DSN      string `yaml:"dsn"`       // Postgres: connection string (e.g. "postgres://user:pass@localhost:5432/mailraven")
	BlobPath string `yaml:"blob_path"` // Blob storage directory path
}

// DKIMConfig contains DKIM signing settings
type DKIMConfig struct {
	Selector       string `yaml:"selector"`         // DKIM selector (default: "default")
	PrivateKeyPath string `yaml:"private_key_path"` // Path to DKIM private key
}

// LogConfig contains logging settings
type LogConfig struct {
	Level  string `yaml:"level"`  // Log level: debug, info, warn, error (default: info)
	Format string `yaml:"format"` // Log format: json, text (default: json)
}

// TLSConfig contains global TLS settings (including ACME)
type TLSConfig struct {
	Enabled bool       `yaml:"enabled"` // Enable TLS globally
	ACME    ACMEConfig `yaml:"acme"`    // Let's Encrypt configuration
}

// ACMEConfig contains Let's Encrypt settings
type ACMEConfig struct {
	Enabled  bool     `yaml:"enabled"`   // Enable ACME
	Email    string   `yaml:"email"`     // Email for registration
	Domains  []string `yaml:"domains"`   // List of domains for certs
	CacheDir string   `yaml:"cache_dir"` // Directory to cache certs
}

// SpamConfig contains spam protection settings
type SpamConfig struct {
	Enabled       bool            `yaml:"enabled"`        // Enable spam protection
	RspamdURL     string          `yaml:"rspamd_url"`     // URL to Rspamd (e.g. http://localhost:11333)
	DNSBLs        []string        `yaml:"dnsbls"`         // List of DNSBL providers
	MaxRecipients int             `yaml:"max_recipients"` // Max recipients per message
	RateLimit     RateLimitConfig `yaml:"rate_limit"`     // Rate limiting settings
	RejectScore   float64         `yaml:"reject_score"`   // Score threshold to reject
	HeaderScore   float64         `yaml:"header_score"`   // Score threshold to add header
	Greylist      GreylistConfig  `yaml:"greylist"`       // Greylisting settings
}

// GreylistConfig contains greylisting settings
type GreylistConfig struct {
	Enabled    bool   `yaml:"enabled"`     // Enable greylisting
	RetryDelay string `yaml:"retry_delay"` // Time to wait before retry (e.g. "5m")
	Expiration string `yaml:"expiration"`  // Time before record expires (e.g. "24h")
}

// IMAPConfig contains IMAP server settings
type IMAPConfig struct {
	Enabled           bool   `yaml:"enabled"`             // Enable IMAP server
	Port              int    `yaml:"port"`                // IMAP listen port (default: 143)
	PortTLS           int    `yaml:"port_tls"`            // IMAP TLS listen port (default: 993)
	AllowInsecureAuth bool   `yaml:"allow_insecure_auth"` // Allow LOGIN on insecure connection
	TLSCert           string `yaml:"tls_cert"`            // TLS certificate path
	TLSKey            string `yaml:"tls_key"`             // TLS key path
}

// RateLimitConfig contains rate limiting settings
type RateLimitConfig struct {
	Window string `yaml:"window"` // Time window (e.g. "1h")
	Count  int    `yaml:"count"`  // Max requests per window
}

// BackupConfig contains backup settings
type BackupConfig struct {
	Enabled       bool   `yaml:"enabled"`        // Enable automatic backups
	Schedule      string `yaml:"schedule"`       // Cron schedule
	Location      string `yaml:"location"`       // Backup directory
	RetentionDays int    `yaml:"retention_days"` // Retention period
}

// ManageSieveConfig contains ManageSieve server settings
type ManageSieveConfig struct {
	Enabled bool `yaml:"enabled"` // Enable ManageSieve server (default: true)
	Port    int  `yaml:"port"`    // ManageSieve listen port (default: 4190)
}

// LoadFromFile loads configuration from a YAML file
func LoadFromFile(path string) (*Config, error) {
	// Sanitize path
	cleanPath := filepath.Clean(path)
	data, err := os.ReadFile(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config YAML: %w", err)
	}

	// Apply defaults
	if cfg.SMTP.Port == 0 {
		cfg.SMTP.Port = 25
	}
	if cfg.SMTP.MaxSize == 0 {
		cfg.SMTP.MaxSize = 10 * 1024 * 1024 // 10MB
	}
	if cfg.SMTP.DANE.Mode == "" {
		cfg.SMTP.DANE.Mode = "advisory"
	}
	if cfg.API.Host == "" {
		cfg.API.Host = "0.0.0.0"
	}
	if cfg.API.Port == 0 {
		cfg.API.Port = 8443
	}
	if cfg.DKIM.Selector == "" {
		cfg.DKIM.Selector = "default"
	}
	if cfg.Spam.Greylist.RetryDelay == "" {
		cfg.Spam.Greylist.RetryDelay = "5m"
	}
	if cfg.Spam.Greylist.Expiration == "" {
		cfg.Spam.Greylist.Expiration = "24h"
	}
	if cfg.Logging.Level == "" {
		cfg.Logging.Level = "info"
	}
	if cfg.Logging.Format == "" {
		cfg.Logging.Format = "json"
	}
	if cfg.Spam.MaxRecipients == 0 {
		cfg.Spam.MaxRecipients = 50
	}
	if cfg.Spam.RateLimit.Count == 0 {
		cfg.Spam.RateLimit.Count = 100
	}
	if cfg.Spam.RateLimit.Window == "" {
		cfg.Spam.RateLimit.Window = "1h"
	}
	if cfg.Backup.Schedule == "" {
		cfg.Backup.Schedule = "0 2 * * *" // Daily at 2am
	}
	if cfg.Backup.RetentionDays == 0 {
		cfg.Backup.RetentionDays = 7
	}
	if cfg.ManageSieve.Port == 0 {
		cfg.ManageSieve.Port = 4190
	}
	return &cfg, nil
}

// Validate checks if configuration is valid
func (c *Config) Validate() error {
	if c.Domain == "" {
		return fmt.Errorf("domain is required")
	}
	if c.SMTP.Hostname == "" {
		return fmt.Errorf("smtp.hostname is required")
	}
	if c.API.JWTSecret == "" {
		return fmt.Errorf("api.jwt_secret is required")
	}
	if c.Storage.DBPath == "" {
		return fmt.Errorf("storage.db_path is required")
	}
	if c.Storage.BlobPath == "" {
		return fmt.Errorf("storage.blob_path is required")
	}
	if c.DKIM.PrivateKeyPath == "" {
		return fmt.Errorf("dkim.private_key_path is required")
	}

	return nil
}

// SaveToFile writes configuration to a YAML file
func (c *Config) SaveToFile(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
