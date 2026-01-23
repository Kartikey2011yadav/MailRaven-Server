package disk

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/Kartikey2011yadav/mailraven-server/internal/core/ports"
)

// BlobStore implements ports.BlobStore using file system with gzip compression
type BlobStore struct {
	basePath string // Root directory for blob storage
}

// NewBlobStore creates a new file system blob store
func NewBlobStore(basePath string) (*BlobStore, error) {
	// Ensure base directory exists
	if err := os.MkdirAll(basePath, 0750); err != nil {
		return nil, fmt.Errorf("failed to create blob store directory: %w", err)
	}

	return &BlobStore{basePath: basePath}, nil
}

// Write stores compressed message body and returns storage path
func (bs *BlobStore) Write(ctx context.Context, messageID string, content []byte) (string, error) {
	// Create date-based directory structure: YYYY/MM/DD
	now := time.Now()
	dateDir := filepath.Join(
		bs.basePath,
		fmt.Sprintf("%04d", now.Year()),
		fmt.Sprintf("%02d", now.Month()),
		fmt.Sprintf("%02d", now.Day()),
	)

	if err := os.MkdirAll(dateDir, 0750); err != nil {
		return "", fmt.Errorf("failed to create date directory: %w", err)
	}

	// File path: <basePath>/YYYY/MM/DD/<messageID>.eml.gz
	filename := fmt.Sprintf("%s.eml.gz", messageID)
	fullPath := filepath.Join(dateDir, filename)

	// Compress content
	var compressedBuf bytes.Buffer
	gzWriter := gzip.NewWriter(&compressedBuf)
	if _, err := gzWriter.Write(content); err != nil {
		return "", fmt.Errorf("compression failed: %w", err)
	}
	if err := gzWriter.Close(); err != nil {
		return "", fmt.Errorf("compression finalization failed: %w", err)
	}

	// Write atomically: write to temp file, then rename
	tempPath := fullPath + ".tmp"
	if err := os.WriteFile(tempPath, compressedBuf.Bytes(), 0640); err != nil {
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}

	// Fsync to ensure durability before rename
	file, err := os.OpenFile(tempPath, os.O_WRONLY, 0)
	if err != nil {
		os.Remove(tempPath)
		return "", fmt.Errorf("failed to open temp file for fsync: %w", err)
	}
	if err := file.Sync(); err != nil {
		file.Close()
		os.Remove(tempPath)
		return "", fmt.Errorf("fsync failed: %w", err)
	}
	file.Close()

	// Atomic rename
	if err := os.Rename(tempPath, fullPath); err != nil {
		os.Remove(tempPath)
		return "", fmt.Errorf("failed to rename temp file: %w", err)
	}

	// Return relative path from basePath
	relPath, err := filepath.Rel(bs.basePath, fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to compute relative path: %w", err)
	}

	return relPath, nil
}

// Read retrieves and decompresses message body
func (bs *BlobStore) Read(ctx context.Context, path string) ([]byte, error) {
	fullPath := filepath.Join(bs.basePath, path)

	// Read compressed file
	compressedData, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ports.ErrNotFound
		}
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Decompress
	gzReader, err := gzip.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		return nil, fmt.Errorf("decompression failed: %w", err)
	}
	defer gzReader.Close()

	decompressed, err := io.ReadAll(gzReader)
	if err != nil {
		return nil, fmt.Errorf("decompression read failed: %w", err)
	}

	return decompressed, nil
}

// Delete removes a blob from storage
func (bs *BlobStore) Delete(ctx context.Context, path string) error {
	fullPath := filepath.Join(bs.basePath, path)

	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return ports.ErrNotFound
		}
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// Verify checks if blob exists and is readable
func (bs *BlobStore) Verify(ctx context.Context, path string) error {
	fullPath := filepath.Join(bs.basePath, path)

	// Check if file exists and is readable
	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return ports.ErrNotFound
		}
		return fmt.Errorf("failed to stat file: %w", err)
	}

	// Verify it's a regular file
	if !info.Mode().IsRegular() {
		return fmt.Errorf("not a regular file")
	}

	return nil
}
