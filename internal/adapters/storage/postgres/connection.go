package postgres

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Connection represents a PostgreSQL database connection
type Connection struct {
	DB *sql.DB
}

// NewConnection creates a new PostgreSQL connection
func NewConnection(dsn string) (*Connection, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	return &Connection{DB: db}, nil
}

// RunMigrations applies SQL migration files from the embedded filesystem
func (c *Connection) RunMigrations() error {
	content, err := MigrationsFS.ReadFile("migrations/000001_init.up.sql")
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	query := string(content)

	// Create transaction
	tx, err := c.DB.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(query); err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	return tx.Commit()
}

// CheckIntegrity is a placeholder for interface satisfaction if needed
func (c *Connection) CheckIntegrity() error {
	return c.DB.Ping()
}

// Close closes the database connection
func (c *Connection) Close() error {
	return c.DB.Close()
}
