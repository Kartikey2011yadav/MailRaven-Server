package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Kartikey2011yadav/mailraven-server/internal/config"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/ports"
	"github.com/Kartikey2011yadav/mailraven-server/internal/observability"
)

// BackupService implements ports.BackupService
type BackupService struct {
	dbBackup   ports.DatabaseBackup
	blobBackup ports.BlobStorageBackup
	config     config.BackupConfig
	logger     *observability.Logger
}

// NewBackupService creates a new backup service
func NewBackupService(
	cfg config.BackupConfig,
	dbBackup ports.DatabaseBackup,
	blobBackup ports.BlobStorageBackup,
	logger *observability.Logger,
) *BackupService {
	return &BackupService{
		dbBackup:   dbBackup,
		blobBackup: blobBackup,
		config:     cfg,
		logger:     logger,
	}
}

// PerformBackup triggers a backup
func (s *BackupService) PerformBackup(ctx context.Context, location string) (string, error) {
	// Use default location if not provided
	if location == "" {
		location = s.config.Location
	}
	if location == "" {
		return "", fmt.Errorf("no backup location specified")
	}

	jobID := fmt.Sprintf("backup-%s", time.Now().Format("20060102-150405"))
	targetDir := filepath.Join(location, jobID)

	// Create target directory
	if err := os.MkdirAll(targetDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	s.logger.Info("starting backup", "job_id", jobID, "target", targetDir)

	// 1. Backup Database
	dbTarget := filepath.Join(targetDir, "meta.db")
	if err := s.dbBackup.PerformBackup(ctx, dbTarget); err != nil {
		s.logger.Error("database backup failed", "error", err)
		return "", fmt.Errorf("database backup failed: %w", err)
	}

	// 2. Backup Blobs
	blobTarget := filepath.Join(targetDir, "blobs")
	if err := s.blobBackup.PerformBackup(ctx, blobTarget); err != nil {
		s.logger.Error("blob backup failed", "error", err)
		return "", fmt.Errorf("blob backup failed: %w", err)
	}

	s.logger.Info("backup completed successfully", "job_id", jobID)
	return targetDir, nil
}
