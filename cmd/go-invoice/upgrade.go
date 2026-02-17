package main

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"

	versionpkg "github.com/mrz1836/go-invoice/internal/version"
)

const (
	// devVersionString is the string used for development versions
	devVersionString = "dev"
)

var (
	// ErrDevVersionNoForce is returned when trying to upgrade a dev version without --force
	ErrDevVersionNoForce = errors.New("cannot upgrade development build without --force")
	// ErrVersionParseFailed is returned when version cannot be parsed from output
	ErrVersionParseFailed = errors.New("could not parse version from output")
	// ErrDownloadFailed is returned when binary download fails
	ErrDownloadFailed = errors.New("failed to download binary")
	// ErrBinaryNotFoundInArchive is returned when the binary is not found in the archive
	ErrBinaryNotFoundInArchive = errors.New("go-invoice binary not found in archive")
	// ErrChecksumMismatch is returned when the downloaded file's checksum doesn't match
	ErrChecksumMismatch = errors.New("checksum mismatch: downloaded file may be corrupted or tampered")
	// ErrChecksumNotFound is returned when the checksum for the target file is not in checksums.txt
	ErrChecksumNotFound = errors.New("checksum not found in checksums file")
	// ErrChecksumsDownloadFailed is returned when the checksums file cannot be downloaded
	ErrChecksumsDownloadFailed = errors.New("checksums download failed")
	// ErrArchiveDownloadFailed is returned when the release archive cannot be downloaded
	ErrArchiveDownloadFailed = errors.New("archive download failed")
)

// upgradeConfig holds configuration for the upgrade command
type upgradeConfig struct {
	Force     bool
	CheckOnly bool
	UseBinary bool
}

// buildUpgradeCommand creates the upgrade command
func (a *App) buildUpgradeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade go-invoice to the latest version",
		Long: `Upgrade the go-invoice CLI to the latest version available.

This command will:
  - Check the latest version available on GitHub
  - Compare with the currently installed version
  - Upgrade if a newer version is available`,
		Example: `  # Check for available updates
  go-invoice upgrade --check

  # Upgrade to latest version
  go-invoice upgrade

  # Force upgrade even if already on latest
  go-invoice upgrade --force`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg := upgradeConfig{}
			var err error

			cfg.Force, err = cmd.Flags().GetBool("force")
			if err != nil {
				return err
			}

			cfg.CheckOnly, err = cmd.Flags().GetBool("check")
			if err != nil {
				return err
			}

			cfg.UseBinary, err = cmd.Flags().GetBool("use-binary")
			if err != nil {
				return err
			}

			return a.runUpgrade(cmd, cfg)
		},
	}

	// Add flags
	cmd.Flags().BoolP("force", "f", false, "Force upgrade even if already on latest version")
	cmd.Flags().Bool("check", false, "Check for updates without upgrading")
	cmd.Flags().BoolP("verbose", "v", false, "Show release notes after upgrade")
	cmd.Flags().Bool("use-binary", false, "Download and install pre-built binary instead of using go install")

	return cmd
}

// runUpgrade executes the upgrade logic
func (a *App) runUpgrade(cmd *cobra.Command, cfg upgradeConfig) error {
	currentVersion := getCurrentVersion()

	// Handle development version or commit hash
	if currentVersion == devVersionString || currentVersion == "" || isLikelyCommitHash(currentVersion) {
		if !cfg.Force && !cfg.CheckOnly {
			a.logger.Printf("⚠️  Current version appears to be a development build (%s)\n", currentVersion)
			a.logger.Println("   Use --force to upgrade anyway")
			return ErrDevVersionNoForce
		}
	}

	a.logger.Printf("ℹ️  Current version: %s\n", formatVersion(currentVersion))

	// Fetch latest release
	a.logger.Println("ℹ️  Checking for updates...")
	release, err := versionpkg.GetLatestRelease(cmd.Context(), "mrz1836", "go-invoice")
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	latestVersion := strings.TrimPrefix(release.TagName, "v")
	a.logger.Printf("ℹ️  Latest version: %s\n", formatVersion(latestVersion))

	// Compare versions
	isNewer := versionpkg.IsNewerVersion(currentVersion, latestVersion)

	if !isNewer && !cfg.Force {
		a.logger.Printf("✅ You are already on the latest version (%s)\n", formatVersion(currentVersion))
		return nil
	}

	if cfg.CheckOnly {
		if isNewer {
			a.logger.Printf("⚠️  A newer version is available: %s → %s\n", formatVersion(currentVersion), formatVersion(latestVersion))
			a.logger.Println("   Run 'go-invoice upgrade' to upgrade")
		} else {
			a.logger.Println("✅ You are on the latest version")
		}
		return nil
	}

	// Perform upgrade
	if isNewer {
		a.logger.Printf("ℹ️  Upgrading from %s to %s...\n", formatVersion(currentVersion), formatVersion(latestVersion))
	} else if cfg.Force {
		a.logger.Printf("ℹ️  Force reinstalling version %s...\n", formatVersion(latestVersion))
	}

	// Perform upgrade using selected method
	if cfg.UseBinary {
		if err := a.upgradeBinary(latestVersion); err != nil {
			a.logger.Println("⚠️  Binary upgrade failed, falling back to go install...")
			if err := a.upgradeGoInstall(latestVersion); err != nil {
				return fmt.Errorf("both binary and go install upgrade methods failed: %w", err)
			}
		}
	} else {
		if err := a.upgradeGoInstall(latestVersion); err != nil {
			a.logger.Println("⚠️  go install failed, falling back to binary download...")
			if err := a.upgradeBinary(latestVersion); err != nil {
				return fmt.Errorf("both go install and binary upgrade methods failed: %w", err)
			}
		}
	}

	a.logger.Printf("✅ Successfully upgraded to version %s\n", formatVersion(latestVersion))

	// Show release notes if available and verbose
	verbose, _ := cmd.Flags().GetBool("verbose")
	if release.Body != "" && verbose {
		a.logger.Printf("\nRelease notes for v%s:\n", latestVersion)
		lines := strings.Split(release.Body, "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				a.logger.Printf("  %s\n", line)
			}
		}
	}

	return nil
}

// formatVersion ensures version strings have a consistent "vX.Y.Z" format for display.
func formatVersion(v string) string {
	if v == devVersionString || v == "" {
		return devVersionString
	}
	if !strings.HasPrefix(v, "v") {
		return "v" + v
	}
	return v
}

// getCurrentVersion returns the current version of go-invoice
func getCurrentVersion() string {
	return Version
}

// upgradeGoInstall upgrades using go install command (primary method)
func (a *App) upgradeGoInstall(latestVersion string) error {
	installCmd := fmt.Sprintf("github.com/mrz1836/go-invoice/cmd/go-invoice@v%s", latestVersion)

	a.logger.Printf("ℹ️  Running: go install %s\n", installCmd)

	execCmd := exec.CommandContext(context.Background(), "go", "install", installCmd) //nolint:gosec // Command is constructed safely
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	if err := execCmd.Run(); err != nil {
		return fmt.Errorf("go install failed: %w", err)
	}

	return nil
}

// upgradeBinary downloads and installs pre-built binary (fallback method)
func (a *App) upgradeBinary(latestVersion string) error {
	// Get current binary location
	currentBinary, err := getInvoiceBinaryLocation()
	if err != nil {
		return fmt.Errorf("could not determine current binary location: %w", err)
	}

	archiveName := fmt.Sprintf("go-invoice_%s_%s_%s.tar.gz", latestVersion, runtime.GOOS, runtime.GOARCH)

	// Construct download URLs
	downloadURL := fmt.Sprintf("https://github.com/mrz1836/go-invoice/releases/download/v%s/%s",
		latestVersion, archiveName)
	checksumsURL := fmt.Sprintf("https://github.com/mrz1836/go-invoice/releases/download/v%s/go-invoice_%s_checksums.txt",
		latestVersion, latestVersion)

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "go-invoice-upgrade-*")
	if err != nil {
		return fmt.Errorf("could not create temporary directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	client := &http.Client{Timeout: 30 * time.Second}

	// Download checksums file for verification
	a.logger.Printf("ℹ️  Downloading checksums from: %s\n", checksumsURL)
	expectedChecksum, err := a.fetchExpectedChecksum(client, checksumsURL, archiveName)
	if err != nil {
		return fmt.Errorf("could not fetch checksums: %w", err)
	}

	// Download the archive
	a.logger.Printf("ℹ️  Downloading binary from: %s\n", downloadURL)
	archivePath := filepath.Join(tempDir, archiveName)
	if err = a.downloadFile(client, downloadURL, archivePath); err != nil {
		return fmt.Errorf("%w: %w", ErrDownloadFailed, err)
	}

	// Verify checksum
	a.logger.Println("ℹ️  Verifying checksum...")
	if err = verifyFileChecksum(archivePath, expectedChecksum); err != nil {
		return err
	}
	a.logger.Println("✅ Checksum verified")

	// Extract binary from tar.gz archive
	archiveFile, err := os.Open(archivePath) //nolint:gosec // Path is constructed safely in temp dir
	if err != nil {
		return fmt.Errorf("could not open archive: %w", err)
	}
	defer func() { _ = archiveFile.Close() }()

	extractedBinary, err := extractInvoiceBinaryFromArchive(archiveFile, tempDir)
	if err != nil {
		return fmt.Errorf("could not extract binary: %w", err)
	}

	// Backup current binary
	backupFile := currentBinary + ".backup"
	if err := os.Rename(currentBinary, backupFile); err != nil {
		return fmt.Errorf("could not backup current binary: %w", err)
	}

	// Replace with new binary
	if err := os.Rename(extractedBinary, currentBinary); err != nil {
		// Restore backup on failure
		_ = os.Rename(backupFile, currentBinary)
		return fmt.Errorf("could not replace binary: %w", err)
	}

	// Remove backup on success
	_ = os.Remove(backupFile)

	a.logger.Println("ℹ️  Binary upgrade completed successfully")
	return nil
}

// fetchExpectedChecksum downloads the checksums file and returns the SHA256 for the given filename.
func (a *App) fetchExpectedChecksum(client *http.Client, checksumsURL, filename string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, checksumsURL, nil)
	if err != nil {
		return "", fmt.Errorf("creating checksums request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("downloading checksums: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: HTTP %d", ErrChecksumsDownloadFailed, resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		return "", fmt.Errorf("reading checksums: %w", err)
	}

	// Parse checksums file (format: "<sha256>  <filename>")
	for _, line := range strings.Split(string(body), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[1] == filename {
			return fields[0], nil
		}
	}

	return "", fmt.Errorf("%w: %s", ErrChecksumNotFound, filename)
}

// downloadFile downloads a URL to a local file path.
func (a *App) downloadFile(client *http.Client, url, destPath string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("creating download request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("downloading file: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: HTTP %d", ErrArchiveDownloadFailed, resp.StatusCode)
	}

	file, err := os.Create(destPath) //nolint:gosec // Path is constructed safely in temp dir
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}

	_, copyErr := io.Copy(file, resp.Body)
	closeErr := file.Close()

	if copyErr != nil {
		return fmt.Errorf("writing file: %w", copyErr)
	}
	if closeErr != nil {
		return fmt.Errorf("closing file: %w", closeErr)
	}

	return nil
}

// verifyFileChecksum calculates the SHA256 of a file and compares it to the expected hash.
func verifyFileChecksum(filePath, expectedHex string) error {
	file, err := os.Open(filePath) //nolint:gosec // Path is constructed safely in temp dir
	if err != nil {
		return fmt.Errorf("opening file for checksum: %w", err)
	}
	defer func() { _ = file.Close() }()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return fmt.Errorf("computing checksum: %w", err)
	}

	actual := hex.EncodeToString(hasher.Sum(nil))
	if actual != expectedHex {
		return fmt.Errorf("%w: expected %s, got %s", ErrChecksumMismatch, expectedHex, actual)
	}

	return nil
}

// extractInvoiceBinaryFromArchive extracts the go-invoice binary from a tar.gz archive
func extractInvoiceBinaryFromArchive(reader io.Reader, destDir string) (string, error) {
	// Create gzip reader
	gzipReader, err := gzip.NewReader(reader)
	if err != nil {
		return "", fmt.Errorf("could not create gzip reader: %w", err)
	}
	defer func() { _ = gzipReader.Close() }()

	// Create tar reader
	tarReader := tar.NewReader(gzipReader)

	// Extract files from tar
	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", fmt.Errorf("could not read tar entry: %w", err)
		}

		// Look for the go-invoice binary
		if filepath.Base(header.Name) == "go-invoice" && header.Typeflag == tar.TypeReg {
			// Create destination file
			destPath := filepath.Join(destDir, "go-invoice")
			file, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755) //nolint:gosec // Need executable permissions
			if err != nil {
				return "", fmt.Errorf("could not create binary file: %w", err)
			}

			// Copy binary content with size limit for security
			limitedReader := io.LimitReader(tarReader, 100*1024*1024) // 100MB limit
			_, copyErr := io.Copy(file, limitedReader)
			closeErr := file.Close()

			// Check copy error first (more likely to indicate actual failure)
			if copyErr != nil {
				return "", fmt.Errorf("could not write binary: %w", copyErr)
			}
			// Check close error (can indicate disk full, I/O errors, etc.)
			if closeErr != nil {
				return "", fmt.Errorf("could not close binary file: %w", closeErr)
			}

			return destPath, nil
		}
	}

	return "", ErrBinaryNotFoundInArchive
}

// isLikelyCommitHash checks if a version string looks like a commit hash.
// It requires the string to:
// - Be 7-40 characters long (short to full SHA-1)
// - Contain only hex characters (0-9, a-f, A-F)
// - Contain at least one letter (to distinguish from pure numeric versions)
func isLikelyCommitHash(version string) bool {
	// Remove any -dirty suffix
	version = strings.TrimSuffix(version, "-dirty")

	// Commit hashes are typically 7-40 hex characters
	if len(version) < 7 || len(version) > 40 {
		return false
	}

	hasLetter := false
	for _, c := range version {
		isDigit := c >= '0' && c <= '9'
		isLowerHex := c >= 'a' && c <= 'f'
		isUpperHex := c >= 'A' && c <= 'F'

		if !isDigit && !isLowerHex && !isUpperHex {
			return false
		}
		if isLowerHex || isUpperHex {
			hasLetter = true
		}
	}

	// Require at least one letter to distinguish from pure numeric versions
	// like "1234567" or "2024010100"
	return hasLetter
}

// getInvoiceBinaryLocation returns the location of the go-invoice binary
func getInvoiceBinaryLocation() (string, error) {
	if runtime.GOOS == "windows" {
		return exec.LookPath("go-invoice.exe")
	}
	return exec.LookPath("go-invoice")
}
