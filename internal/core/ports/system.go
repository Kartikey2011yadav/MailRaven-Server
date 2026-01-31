package ports

import "context"

// UpdateInfo contains information about an available update
type UpdateInfo struct {
	Version      string
	ReleaseNotes string
	DownloadURL  string
	AssetID      int64
	PublishedAt  string
}

// UpdateManager manages self-updates
type UpdateManager interface {
	// CheckForUpdate checks if a newer version is available compared to currentVersion
	CheckForUpdate(ctx context.Context, currentVersion string) (*UpdateInfo, error)

	// ApplyUpdate downloads the update from info, verifies it using the public key,
	// and replaces the current binary. It does NOT restart the process (caller must do that).
	ApplyUpdate(ctx context.Context, info *UpdateInfo) error
}
