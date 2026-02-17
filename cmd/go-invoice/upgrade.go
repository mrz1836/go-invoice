package main

import (
	"archive/tar"
	"compress/gzip"
	"context"
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

	// Construct download URL for compressed archive
	downloadURL := fmt.Sprintf("https://github.com/mrz1836/go-invoice/releases/download/v%s/go-invoice_%s_%s_%s.tar.gz",
		latestVersion, latestVersion, runtime.GOOS, runtime.GOARCH)

	a.logger.Printf("ℹ️  Downloading binary from: %s\n", downloadURL)

	// Download the binary with context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrDownloadFailed, err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrDownloadFailed, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: HTTP %d", ErrDownloadFailed, resp.StatusCode)
	}

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "go-invoice-upgrade-*")
	if err != nil {
		return fmt.Errorf("could not create temporary directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Extract binary from tar.gz archive
	extractedBinary, err := extractInvoiceBinaryFromArchive(resp.Body, tempDir)
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
