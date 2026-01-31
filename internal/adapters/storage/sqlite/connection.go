package sqlite

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite" // CGO-free SQLite driver
)

// Connection represents a SQLite database connection
type Connection struct {
	DB *sql.DB
}

// NewConnection creates a new SQLite connection with proper configuration
func NewConnection(dbPath string) (*Connection, error) {
	// Ensure parent directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Connection string with WAL mode and full durability
	// Note: PRAGMA statements in migration file will also be applied
	connectionString := fmt.Sprintf("file:%s?_journal_mode=WAL&_synchronous=FULL&_foreign_keys=ON&_busy_timeout=5000", dbPath)

	db, err := sql.Open("sqlite", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)   // Limit concurrent connections
	db.SetMaxIdleConns(10)   // Keep idle connections for reuse
	db.SetConnMaxLifetime(0) // No maximum lifetime
	db.SetConnMaxIdleTime(0) // No idle timeout

	// Verify connection works
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	return &Connection{DB: db}, nil
}

// CheckIntegrity performs PRAGMA integrity_check on startup
func (c *Connection) CheckIntegrity() error {
	var result string
	err := c.DB.QueryRow("PRAGMA integrity_check").Scan(&result)
	if err != nil {
		return fmt.Errorf("integrity check query failed: %w", err)
	}

	if result != "ok" {
		return fmt.Errorf("database integrity check failed: %s", result)
	}

	return nil
}

// RunMigrations executes SQL migration files
func (c *Connection) RunMigrations(migrationPath string) error {
	// Read migration file
	migrationSQL, err := os.ReadFile(migrationPath)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	// Execute migration (SQLite driver handles multiple statements)
	_, err = c.DB.Exec(string(migrationSQL))
	if err != nil {
		return fmt.Errorf("migration execution failed: %w", err)
	}

	return nil
}

// Close closes the database connection
func (c *Connection) Close() error {
	return c.DB.Close()
}
