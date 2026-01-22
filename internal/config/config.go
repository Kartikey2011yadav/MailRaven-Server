package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the MailRaven server configuration
type Config struct {
	Domain string      `yaml:"domain"` // Primary mail domain (e.g., mail.example.com)
	SMTP   SMTPConfig  `yaml:"smtp"`
	API    APIConfig   `yaml:"api"`
	Storage StorageConfig `yaml:"storage"`
	DKIM   DKIMConfig  `yaml:"dkim"`
	Logging LogConfig   `yaml:"logging"`
}

// SMTPConfig contains SMTP server settings
type SMTPConfig struct {
	Port     int    `yaml:"port"`     // SMTP listen port (default: 25)
	Hostname string `yaml:"hostname"` // SMTP HELO hostname
	MaxSize  int64  `yaml:"max_size"` // Maximum message size in bytes (default: 10MB)
}

// APIConfig contains REST API settings
type APIConfig struct {
	Port      int    `yaml:"port"`       // HTTP listen port (default: 8080)
	JWTSecret string `yaml:"jwt_secret"` // JWT signing secret
	TLSCert   string `yaml:"tls_cert"`   // Optional TLS certificate path
	TLSKey    string `yaml:"tls_key"`    // Optional TLS key path
}

// StorageConfig contains database and blob storage settings
type StorageConfig struct {
	DBPath   string `yaml:"db_path"`   // SQLite database file path
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

// LoadFromFile loads configuration from a YAML file
func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
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
	if cfg.API.Port == 0 {
		cfg.API.Port = 8080
	}
	if cfg.DKIM.Selector == "" {
		cfg.DKIM.Selector = "default"
	}
	if cfg.Logging.Level == "" {
		cfg.Logging.Level = "info"
	}
	if cfg.Logging.Format == "" {
		cfg.Logging.Format = "json"
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

	if err := os.WriteFile(path, data, 0640); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
