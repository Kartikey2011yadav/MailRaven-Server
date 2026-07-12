package postgres

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Connection represents a PostgreSQL database connection
type Connection struct {
	DB *sql.DB
}

// NewConnection creates a new PostgreSQL connection with production-safe pool settings
func NewConnection(dsn string) (*Connection, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres connection: %w", err)
	}

	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(2 * time.Minute)

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

	tx, err := c.DB.Begin()
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if _, err := tx.Exec(query); err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	return tx.Commit()
}

// CheckIntegrity verifies the connection is healthy
func (c *Connection) CheckIntegrity() error {
	return c.DB.Ping()
}

// Close closes the database connection
func (c *Connection) Close() error {
	return c.DB.Close()
}
