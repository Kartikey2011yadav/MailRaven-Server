package backup

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// BlobBackup handles blob storage backup
type BlobBackup struct {
	sourceDir string
}

// NewBlobBackup creates a new blob backup handler
func NewBlobBackup(sourceDir string) *BlobBackup {
	return &BlobBackup{sourceDir: sourceDir}
}

// PerformBackup copies all blobs to the target directory
func (b *BlobBackup) PerformBackup(ctx context.Context, targetDir string) error {
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target dir: %w", err)
	}

	return filepath.Walk(b.sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if info.IsDir() {
			// Create subdirectories in target
			relPath, err := filepath.Rel(b.sourceDir, path)
			if err != nil {
				return err
			}
			return os.MkdirAll(filepath.Join(targetDir, relPath), 0755)
		}

		// Copy file
		relPath, err := filepath.Rel(b.sourceDir, path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(targetDir, relPath)
		return copyFile(path, destPath)
	})
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
