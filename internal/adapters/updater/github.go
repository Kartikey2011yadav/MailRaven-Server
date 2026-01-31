package updater

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/Kartikey2011yadav/mailraven-server/internal/config"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/ports"
	"github.com/jedisct1/go-minisign"
	"github.com/minio/selfupdate"
	"golang.org/x/mod/semver"
)

type GitHubUpdater struct {
	owner  string
	repo   string
	client *http.Client
}

func NewGitHubUpdater(owner, repo string) *GitHubUpdater {
	return &GitHubUpdater{
		owner:  owner,
		repo:   repo,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

type githubRelease struct {
	TagName     string        `json:"tag_name"`
	Name        string        `json:"name"`
	Body        string        `json:"body"`
	PublishedAt string        `json:"published_at"`
	Assets      []githubAsset `json:"assets"`
}

type githubAsset struct {
	ID                 int64  `json:"id"`
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

func (u *GitHubUpdater) CheckForUpdate(ctx context.Context, currentVersion string) (*ports.UpdateInfo, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", u.owner, u.repo)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := u.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to check for updates: status %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	// Compare versions
	// Ensure versions start with 'v' for semver comparison
	vCurrent := currentVersion
	if !strings.HasPrefix(vCurrent, "v") {
		vCurrent = "v" + vCurrent
	}
	vRelease := release.TagName
	if !strings.HasPrefix(vRelease, "v") {
		vRelease = "v" + vRelease
	}

	if semver.Compare(vRelease, vCurrent) <= 0 {
		return nil, nil // No update available
	}

	// Find matching asset
	assetURL := ""
	assetID := int64(0)
	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}

	// Relaxed matching: verify it contains goos/goarch

	for _, asset := range release.Assets {
		// Example: mailraven_v1.2.0_windows_amd64.exe
		normalizedName := strings.ToLower(asset.Name)
		if strings.Contains(normalizedName, runtime.GOOS) &&
			strings.Contains(normalizedName, runtime.GOARCH) &&
			strings.HasSuffix(normalizedName, ext) &&
			!strings.HasSuffix(normalizedName, ".minisig") { // Skip signature files
			assetURL = asset.BrowserDownloadURL
			assetID = asset.ID
			break
		}
	}

	if assetURL == "" {
		return nil, fmt.Errorf("no compatible asset found for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	return &ports.UpdateInfo{
		Version:      release.TagName,
		ReleaseNotes: release.Body,
		DownloadURL:  assetURL,
		AssetID:      assetID,
		PublishedAt:  release.PublishedAt,
	}, nil
}

func (u *GitHubUpdater) ApplyUpdate(ctx context.Context, info *ports.UpdateInfo) error {
	// 1. Download Binary
	binResp, err := u.download(ctx, info.DownloadURL)
	if err != nil {
		return fmt.Errorf("failed to download binary: %w", err)
	}
	defer binResp.Close()

	// Read full binary into memory to verify signature (unless we stream verify?)
	// SelfUpdate requires a reader, Minisign requires byte slice.
	binData, err := io.ReadAll(binResp)
	if err != nil {
		return err
	}

	// 2. Download Signature (.minisig)
	// Assuming it has same name + .minisig
	sigURL := info.DownloadURL + ".minisig"
	sigResp, err := u.download(ctx, sigURL)
	if err != nil {
		return fmt.Errorf("failed to download signature: %w", err)
	}
	defer sigResp.Close()

	sigData, err := io.ReadAll(sigResp)
	if err != nil {
		return err
	}

	// 3. Verify Signature
	if err := u.verifySignature(binData, sigData); err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}

	// 4. Apply Update
	reader := bytes.NewReader(binData)
	if err := selfupdate.Apply(reader, selfupdate.Options{}); err != nil {
		// Rollback is automatic by selfupdate? No, selfupdate replaces the file.
		// If it fails, the old file should be intact.
		return err
	}

	return nil
}

func (u *GitHubUpdater) download(ctx context.Context, url string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := u.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("download failed: %s", resp.Status)
	}
	return resp.Body, nil
}

func (u *GitHubUpdater) verifySignature(bin []byte, sigData []byte) error {
	pk, err := minisign.NewPublicKey(config.UpdatePublicKey)
	if err != nil {
		return fmt.Errorf("invalid public key configuration: %w", err)
	}

	sig, err := minisign.DecodeSignature(string(sigData))
	if err != nil {
		return fmt.Errorf("invalid signature format: %w", err)
	}

	valid, err := pk.Verify(bin, sig)
	if err != nil {
		return err
	}
	if !valid {
		return fmt.Errorf("signature verification failed (invalid)")
	}
	return nil
}
