package ports

import (
	"context"
)

// BlobStore defines storage operations for large binary objects (email bodies)
type BlobStore interface {
	// Write stores compressed message body and returns storage path
	// Compression is transparent to caller
	Write(ctx context.Context, messageID string, content []byte) (path string, err error)

	// Read retrieves and decompresses message body
	// Returns ErrNotFound if path doesn't exist
	Read(ctx context.Context, path string) ([]byte, error)

	// Delete removes a blob from storage
	Delete(ctx context.Context, path string) error

	// Verify checks if blob exists and is readable
	Verify(ctx context.Context, path string) error
}
