package minio

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/Kartikey2011yadav/mailraven-server/internal/config"
)

// BlobStore implements ports.BlobStore using MinIO (S3-compatible) object storage.
type BlobStore struct {
	client *minio.Client
	bucket string
}

// NewBlobStore creates a MinIO blob store and ensures the bucket exists.
func NewBlobStore(cfg config.ObjectStoreConfig) (*BlobStore, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("minio client creation failed: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	exists, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("minio bucket check failed: %w", err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("minio bucket creation failed: %w", err)
		}
	}

	return &BlobStore{client: client, bucket: cfg.Bucket}, nil
}

func (b *BlobStore) Write(ctx context.Context, messageID string, content []byte) (string, error) {
	now := time.Now()
	objectKey := fmt.Sprintf("%d/%02d/%02d/%s.eml.gz", now.Year(), now.Month(), now.Day(), messageID)

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(content); err != nil {
		return "", fmt.Errorf("gzip write failed: %w", err)
	}
	if err := gz.Close(); err != nil {
		return "", fmt.Errorf("gzip close failed: %w", err)
	}

	_, err := b.client.PutObject(ctx, b.bucket, objectKey, &buf, int64(buf.Len()), minio.PutObjectOptions{
		ContentType: "application/gzip",
	})
	if err != nil {
		return "", fmt.Errorf("minio put failed: %w", err)
	}

	return objectKey, nil
}

func (b *BlobStore) Read(ctx context.Context, path string) ([]byte, error) {
	obj, err := b.client.GetObject(ctx, b.bucket, path, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("minio get failed: %w", err)
	}
	defer obj.Close()

	gz, err := gzip.NewReader(obj)
	if err != nil {
		return nil, fmt.Errorf("gzip reader failed: %w", err)
	}
	defer gz.Close()

	data, err := io.ReadAll(gz)
	if err != nil {
		return nil, fmt.Errorf("read decompressed data failed: %w", err)
	}

	return data, nil
}

func (b *BlobStore) Delete(ctx context.Context, path string) error {
	return b.client.RemoveObject(ctx, b.bucket, path, minio.RemoveObjectOptions{})
}

func (b *BlobStore) Verify(ctx context.Context, path string) error {
	_, err := b.client.StatObject(ctx, b.bucket, path, minio.StatObjectOptions{})
	if err != nil {
		return fmt.Errorf("blob not found: %w", err)
	}
	return nil
}

// Ping checks if MinIO is reachable.
func (b *BlobStore) Ping(ctx context.Context) error {
	_, err := b.client.BucketExists(ctx, b.bucket)
	return err
}
