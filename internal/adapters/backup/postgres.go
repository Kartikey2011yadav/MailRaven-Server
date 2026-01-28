package backup

import (
	"context"
	"fmt"
	"os/exec"
)

// PostgresBackup handles PostgreSQL database backup using pg_dump
type PostgresBackup struct {
	dsn string
}

// NewPostgresBackup creates a new PostgreSQL backup handler
func NewPostgresBackup(dsn string) *PostgresBackup {
	return &PostgresBackup{dsn: dsn}
}

// PerformBackup executes pg_dump to create a backup
func (b *PostgresBackup) PerformBackup(ctx context.Context, targetPath string) error {
	// Use pg_dump to export to file
	// Note: targetPath should include filename, or we assume it's a directory?
	// SQLite VACUUM INTO takes a filename.
	// BackupService logic probably creates a filename like "backup_timestamp/db.sqlite".
	// So targetPath is likely full path.

	cmd := exec.CommandContext(ctx, "pg_dump", b.dsn, "-f", targetPath, "-F", "c") // Custom format (compressed)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("pg_dump failed: %s: %w", string(out), err)
	}
	return nil
}
