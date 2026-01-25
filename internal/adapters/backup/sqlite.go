package backup

import (
	"context"
	"database/sql"
	"fmt"
)

// SQLiteBackup handles database backup
type SQLiteBackup struct {
	db *sql.DB
}

// NewSQLiteBackup creates a new SQLite backup handler
func NewSQLiteBackup(db *sql.DB) *SQLiteBackup {
	return &SQLiteBackup{db: db}
}

// PerformBackup executes VACUUM INTO to create a safe snapshot
func (b *SQLiteBackup) PerformBackup(ctx context.Context, targetPath string) error {
	// VACUUM INTO requires a file path argument using single quotes in the SQL
	// or bound parameter if supported. modernc.org/sqlite supports binding.
	_, err := b.db.ExecContext(ctx, "VACUUM INTO ?", targetPath)
	if err != nil {
		return fmt.Errorf("VACUUM INTO failed: %w", err)
	}
	return nil
}
