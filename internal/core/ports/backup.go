package ports

import "context"

// BackupService defines the interface for performing system backups.
type BackupService interface {
	// PerformBackup triggers a hot backup to the specified location.
	// Returns the job ID or result path.
	PerformBackup(ctx context.Context, location string) (string, error)
}

// DatabaseBackup defines the interface for backing up the database
type DatabaseBackup interface {
	PerformBackup(ctx context.Context, targetPath string) error
}

// BlobStorageBackup defines the interface for backing up blob storage
type BlobStorageBackup interface {
	PerformBackup(ctx context.Context, targetDir string) error
}
