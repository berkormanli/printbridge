package update

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	// GitHubAPIURL is the base URL for GitHub API
	GitHubAPIURL = "https://api.github.com"
	// DefaultOwner is the default repository owner
	DefaultOwner = "berkormanli"
	// DefaultRepo is the default repository name
	DefaultRepo = "printbridge"
)

// Release represents a GitHub release
type Release struct {
	TagName     string  `json:"tag_name"`
	Name        string  `json:"name"`
	Body        string  `json:"body"`
	PublishedAt string  `json:"published_at"`
	HTMLURL     string  `json:"html_url"`
	Assets      []Asset `json:"assets"`
}

// Asset represents a release asset (downloadable file)
type Asset struct {
	Name               string `json:"name"`
	Size               int64  `json:"size"`
	BrowserDownloadURL string `json:"browser_download_url"`
	ContentType        string `json:"content_type"`
}

// UpdateInfo contains information about an available update
type UpdateInfo struct {
	Available      bool
	CurrentVersion string
	LatestVersion  string
	DownloadURL    string
	ReleaseNotes   string
	ReleaseURL     string
}

// CheckForUpdates checks GitHub for newer releases
func CheckForUpdates(currentVersion string) (*UpdateInfo, error) {
	return CheckForUpdatesRepo(currentVersion, DefaultOwner, DefaultRepo)
}

// CheckForUpdatesRepo checks a specific GitHub repo for newer releases
func CheckForUpdatesRepo(currentVersion, owner, repo string) (*UpdateInfo, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", GitHubAPIURL, owner, repo)

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// GitHub API requires User-Agent header
	req.Header.Set("User-Agent", "PrintBridge-Updater")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return &UpdateInfo{Available: false, CurrentVersion: currentVersion}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse release: %w", err)
	}

	// Extract version from tag (remove 'v' prefix if present)
	latestVersion := strings.TrimPrefix(release.TagName, "v")
	currentVersion = strings.TrimPrefix(currentVersion, "v")

	// Find the Windows installer asset
	var downloadURL string
	for _, asset := range release.Assets {
		if strings.HasSuffix(strings.ToLower(asset.Name), "-setup.exe") ||
			strings.HasSuffix(strings.ToLower(asset.Name), "-setup-"+latestVersion+".exe") ||
			strings.Contains(strings.ToLower(asset.Name), "setup") && strings.HasSuffix(strings.ToLower(asset.Name), ".exe") {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	// Check if update is available
	updateAvailable := CompareVersions(currentVersion, latestVersion) < 0

	return &UpdateInfo{
		Available:      updateAvailable,
		CurrentVersion: currentVersion,
		LatestVersion:  latestVersion,
		DownloadURL:    downloadURL,
		ReleaseNotes:   release.Body,
		ReleaseURL:     release.HTMLURL,
	}, nil
}

// CompareVersions compares two semantic version strings
// Returns: -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
func CompareVersions(v1, v2 string) int {
	// Parse version parts
	parts1 := parseVersion(v1)
	parts2 := parseVersion(v2)

	// Compare each part
	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var p1, p2 int
		if i < len(parts1) {
			p1 = parts1[i]
		}
		if i < len(parts2) {
			p2 = parts2[i]
		}

		if p1 < p2 {
			return -1
		}
		if p1 > p2 {
			return 1
		}
	}

	return 0
}

// parseVersion extracts numeric parts from a version string
func parseVersion(v string) []int {
	// Remove common prefixes
	v = strings.TrimPrefix(v, "v")
	v = strings.TrimPrefix(v, "V")

	// Find all numeric parts
	re := regexp.MustCompile(`\d+`)
	matches := re.FindAllString(v, -1)

	parts := make([]int, len(matches))
	for i, m := range matches {
		parts[i], _ = strconv.Atoi(m)
	}

	return parts
}

// DownloadInstaller downloads the update installer to a temporary location
func DownloadInstaller(downloadURL string) (string, error) {
	if downloadURL == "" {
		return "", fmt.Errorf("no download URL provided")
	}

	// Create temp file for installer
	tempDir := os.TempDir()
	tempFile, err := os.CreateTemp(tempDir, "PrintBridge-Setup-*.exe")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tempFile.Close()

	// Download the file
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Get(downloadURL)
	if err != nil {
		os.Remove(tempFile.Name())
		return "", fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		os.Remove(tempFile.Name())
		return "", fmt.Errorf("download returned status %d", resp.StatusCode)
	}

	// Copy to temp file
	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		os.Remove(tempFile.Name())
		return "", fmt.Errorf("failed to save installer: %w", err)
	}

	return tempFile.Name(), nil
}

// DownloadProgress represents download progress
type DownloadProgress struct {
	TotalBytes      int64
	DownloadedBytes int64
	Percent         float64
}

// DownloadInstallerWithProgress downloads with progress reporting
func DownloadInstallerWithProgress(downloadURL string, progressCh chan<- DownloadProgress) (string, error) {
	if downloadURL == "" {
		return "", fmt.Errorf("no download URL provided")
	}

	// Create temp file for installer
	tempDir := os.TempDir()
	tempFile, err := os.CreateTemp(tempDir, "PrintBridge-Setup-*.exe")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tempFile.Close()

	// Download the file
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Get(downloadURL)
	if err != nil {
		os.Remove(tempFile.Name())
		return "", fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		os.Remove(tempFile.Name())
		return "", fmt.Errorf("download returned status %d", resp.StatusCode)
	}

	totalSize := resp.ContentLength
	var downloaded int64

	// Create a buffer for reading
	buf := make([]byte, 32*1024) // 32KB buffer

	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			_, writeErr := tempFile.Write(buf[:n])
			if writeErr != nil {
				os.Remove(tempFile.Name())
				return "", fmt.Errorf("failed to write: %w", writeErr)
			}
			downloaded += int64(n)

			// Report progress
			if progressCh != nil && totalSize > 0 {
				progressCh <- DownloadProgress{
					TotalBytes:      totalSize,
					DownloadedBytes: downloaded,
					Percent:         float64(downloaded) / float64(totalSize) * 100,
				}
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			os.Remove(tempFile.Name())
			return "", fmt.Errorf("failed to read: %w", err)
		}
	}

	if progressCh != nil {
		close(progressCh)
	}

	return tempFile.Name(), nil
}
