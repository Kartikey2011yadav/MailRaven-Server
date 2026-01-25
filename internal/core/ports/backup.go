package ports

import "context"

// BackupService defines the interface for performing system backups.
type BackupService interface {
	// PerformBackup triggers a hot backup to the specified location.
	// Returns the job ID or result path.
	PerformBackup(ctx context.Context, location string) (string, error)
}
