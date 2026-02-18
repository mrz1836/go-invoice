package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrz1836/go-invoice/internal/cli"
	versionpkg "github.com/mrz1836/go-invoice/internal/version"
)

// newTestApp creates a minimal App for testing
func newTestApp() *App {
	return &App{
		logger: cli.NewLogger(false),
	}
}

// --- buildUpgradeCommand tests ---

func TestBuildUpgradeCommand(t *testing.T) {
	app := newTestApp()
	cmd := app.buildUpgradeCommand()

	assert.Equal(t, "upgrade", cmd.Use)
	assert.Equal(t, "Upgrade go-invoice to the latest version", cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.NotEmpty(t, cmd.Example)
	assert.NotNil(t, cmd.RunE)

	// Verify all expected flags exist
	assert.NotNil(t, cmd.Flags().Lookup("force"))
	assert.NotNil(t, cmd.Flags().Lookup("check"))
	assert.NotNil(t, cmd.Flags().Lookup("verbose"))
	assert.NotNil(t, cmd.Flags().Lookup("use-binary"))
}

func TestBuildUpgradeCommandFlagDefaults(t *testing.T) {
	app := newTestApp()
	cmd := app.buildUpgradeCommand()

	force, err := cmd.Flags().GetBool("force")
	require.NoError(t, err)
	assert.False(t, force)

	check, err := cmd.Flags().GetBool("check")
	require.NoError(t, err)
	assert.False(t, check)

	verbose, err := cmd.Flags().GetBool("verbose")
	require.NoError(t, err)
	assert.False(t, verbose)

	useBinary, err := cmd.Flags().GetBool("use-binary")
	require.NoError(t, err)
	assert.False(t, useBinary)
}

// --- formatVersion tests ---

func TestFormatVersion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "dev version", input: "dev", expected: "dev"},
		{name: "empty version", input: "", expected: "dev"},
		{name: "version without prefix", input: "1.2.3", expected: "v1.2.3"},
		{name: "version with v prefix", input: "v1.2.3", expected: "v1.2.3"},
		{name: "patch version", input: "0.9.0", expected: "v0.9.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatVersion(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// --- isLikelyCommitHash tests ---

func TestIsLikelyCommitHash(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{name: "short commit hash", input: "abc1234", expected: true},
		{name: "full commit hash", input: "abc123def456abc123def456abc123def456abc1", expected: true},
		{name: "dirty commit hash", input: "abc1234-dirty", expected: true},
		{name: "semantic version", input: "1.2.3", expected: false},
		{name: "dev version", input: "dev", expected: false},
		{name: "too short", input: "abc", expected: false},
		{name: "non-hex chars", input: "xyz12345", expected: false},
		{name: "pure numeric", input: "1234567", expected: false}, // no letters = not a hash
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isLikelyCommitHash(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// --- getCurrentVersion tests ---

func TestGetCurrentVersion(t *testing.T) {
	original := version
	defer func() { version = original }()

	version = "1.2.3"
	assert.Equal(t, "1.2.3", getCurrentVersion())

	version = "dev"
	assert.Equal(t, "dev", getCurrentVersion())
}

// --- runUpgrade tests with mock server ---

// setupMockGitHubServer creates a test HTTP server that returns a mock GitHub release response
// returning tag "v2.0.0".
func setupMockGitHubServer(t *testing.T) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"tag_name":"v2.0.0","name":"Release v2.0.0","draft":false,"prerelease":false,"body":"Test release notes"}`))
	}))
	return server
}

// makeUpgradeCmd builds a cobra.Command with a context for upgrade testing
func makeUpgradeCmd(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().BoolP("force", "f", false, "")
	cmd.Flags().Bool("check", false, "")
	cmd.Flags().BoolP("verbose", "v", false, "")
	cmd.Flags().Bool("use-binary", false, "")
	cmd.SetContext(ctx)
	return cmd
}

func TestRunUpgrade_DevVersionNoForce(t *testing.T) {
	original := version
	defer func() { version = original }()
	version = "dev"

	app := newTestApp()
	cmd := makeUpgradeCmd(context.Background())
	cfg := upgradeConfig{Force: false, CheckOnly: false}

	err := app.runUpgrade(cmd, cfg)
	assert.ErrorIs(t, err, ErrDevVersionNoForce)
}

func TestRunUpgrade_DevVersionCheckOnlyDoesNotError(t *testing.T) {
	original := version
	defer func() { version = original }()
	version = "dev"

	server := setupMockGitHubServer(t)
	defer server.Close()

	versionpkg.SetDefaultClient(versionpkg.NewClient(
		versionpkg.WithBaseURL(server.URL),
	))
	defer versionpkg.ResetDefaultClient()

	app := newTestApp()
	cmd := makeUpgradeCmd(context.Background())
	cfg := upgradeConfig{Force: false, CheckOnly: true}

	// dev version + check-only should NOT return error
	err := app.runUpgrade(cmd, cfg)
	assert.NoError(t, err)
}

func TestRunUpgrade_CheckOnlyNewerAvailable(t *testing.T) {
	original := version
	defer func() { version = original }()
	version = "1.0.0"

	server := setupMockGitHubServer(t)
	defer server.Close()

	versionpkg.SetDefaultClient(versionpkg.NewClient(
		versionpkg.WithBaseURL(server.URL),
	))
	defer versionpkg.ResetDefaultClient()

	app := newTestApp()
	cmd := makeUpgradeCmd(context.Background())
	cfg := upgradeConfig{Force: false, CheckOnly: true}

	err := app.runUpgrade(cmd, cfg)
	assert.NoError(t, err)
}

func TestRunUpgrade_CheckOnlyAlreadyLatest(t *testing.T) {
	original := version
	defer func() { version = original }()
	version = "2.0.0"

	server := setupMockGitHubServer(t)
	defer server.Close()

	versionpkg.SetDefaultClient(versionpkg.NewClient(
		versionpkg.WithBaseURL(server.URL),
	))
	defer versionpkg.ResetDefaultClient()

	app := newTestApp()
	cmd := makeUpgradeCmd(context.Background())
	cfg := upgradeConfig{Force: false, CheckOnly: true}

	err := app.runUpgrade(cmd, cfg)
	assert.NoError(t, err)
}

func TestRunUpgrade_AlreadyLatestNoForce(t *testing.T) {
	original := version
	defer func() { version = original }()
	version = "2.0.0"

	server := setupMockGitHubServer(t)
	defer server.Close()

	versionpkg.SetDefaultClient(versionpkg.NewClient(
		versionpkg.WithBaseURL(server.URL),
	))
	defer versionpkg.ResetDefaultClient()

	app := newTestApp()
	cmd := makeUpgradeCmd(context.Background())
	cfg := upgradeConfig{Force: false, CheckOnly: false}

	// Should return nil (already on latest, no force)
	err := app.runUpgrade(cmd, cfg)
	assert.NoError(t, err)
}

func TestRunUpgrade_DevVersionWithForce_CheckOnly(t *testing.T) {
	// Verify that Force=true bypasses the dev version guard even in check-only mode.
	// We use CheckOnly to avoid triggering real go install in tests.
	original := version
	defer func() { version = original }()
	version = "dev"

	server := setupMockGitHubServer(t)
	defer server.Close()

	versionpkg.SetDefaultClient(versionpkg.NewClient(
		versionpkg.WithBaseURL(server.URL),
	))
	defer versionpkg.ResetDefaultClient()

	app := newTestApp()
	cmd := makeUpgradeCmd(context.Background())
	cfg := upgradeConfig{Force: true, CheckOnly: true}

	// With Force=true, dev version guard is bypassed; check-only skips actual install
	err := app.runUpgrade(cmd, cfg)
	assert.NoError(t, err)
}

// --- extractInvoiceBinaryFromArchive tests ---

func TestExtractInvoiceBinaryFromArchive_Empty(t *testing.T) {
	// Empty reader should fail gracefully
	_, err := extractInvoiceBinaryFromArchive(bytes.NewReader([]byte{}), t.TempDir())
	require.Error(t, err)
}

func TestExtractInvoiceBinaryFromArchive_InvalidData(t *testing.T) {
	// Non-gzip data should fail
	_, err := extractInvoiceBinaryFromArchive(bytes.NewReader([]byte("not a tar.gz")), t.TempDir())
	require.Error(t, err)
}

// --- getInvoiceBinaryLocation tests ---

func TestGetInvoiceBinaryLocation(_ *testing.T) {
	// This test validates the function can be called without panicking.
	// The binary may or may not be in PATH; either result is acceptable.
	_, _ = getInvoiceBinaryLocation()
}

// --- verifyFileChecksum tests ---

func TestVerifyFileChecksum_Match(t *testing.T) {
	// Create a temp file with known content
	content := []byte("hello world")
	tmpFile, err := os.CreateTemp(t.TempDir(), "checksum-test-*")
	require.NoError(t, err)
	_, err = tmpFile.Write(content)
	require.NoError(t, err)
	require.NoError(t, tmpFile.Close())

	// Compute expected SHA256
	sum := sha256.Sum256(content)
	expectedHex := hex.EncodeToString(sum[:])

	err = verifyFileChecksum(tmpFile.Name(), expectedHex)
	assert.NoError(t, err)
}

func TestVerifyFileChecksum_Mismatch(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "checksum-test-*")
	require.NoError(t, err)
	_, err = tmpFile.WriteString("hello world")
	require.NoError(t, err)
	require.NoError(t, tmpFile.Close())

	err = verifyFileChecksum(tmpFile.Name(), "deadbeef")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrChecksumMismatch)
}

func TestVerifyFileChecksum_FileNotFound(t *testing.T) {
	err := verifyFileChecksum("/nonexistent/path/file.txt", "deadbeef")
	require.Error(t, err)
}

// --- fetchExpectedChecksum tests ---

func TestFetchExpectedChecksum_Found(t *testing.T) {
	// Serve a mock checksums file
	checksumsContent := "abc123def456  go-invoice_1.0.0_darwin_arm64.tar.gz\n" +
		"deadbeef1234  go-invoice_1.0.0_linux_amd64.tar.gz\n"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(checksumsContent))
	}))
	defer server.Close()

	app := newTestApp()
	client := &http.Client{Timeout: 5 * time.Second}

	hash, err := app.fetchExpectedChecksum(client, server.URL+"/checksums.txt", "go-invoice_1.0.0_darwin_arm64.tar.gz")
	require.NoError(t, err)
	assert.Equal(t, "abc123def456", hash)
}

func TestFetchExpectedChecksum_NotFound(t *testing.T) {
	checksumsContent := "abc123def456  go-invoice_1.0.0_linux_amd64.tar.gz\n"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(checksumsContent))
	}))
	defer server.Close()

	app := newTestApp()
	client := &http.Client{Timeout: 5 * time.Second}

	_, err := app.fetchExpectedChecksum(client, server.URL+"/checksums.txt", "go-invoice_1.0.0_darwin_arm64.tar.gz")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "checksum not found")
}

func TestFetchExpectedChecksum_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	app := newTestApp()
	client := &http.Client{Timeout: 5 * time.Second}

	_, err := app.fetchExpectedChecksum(client, server.URL+"/checksums.txt", "go-invoice_1.0.0_darwin_arm64.tar.gz")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrChecksumsDownloadFailed)
}

// --- downloadFile tests ---

func TestDownloadFile_Success(t *testing.T) {
	content := []byte("binary content here")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(content)
	}))
	defer server.Close()

	app := newTestApp()
	client := &http.Client{Timeout: 5 * time.Second}
	destPath := filepath.Join(t.TempDir(), "downloaded.tar.gz")

	err := app.downloadFile(client, server.URL+"/file.tar.gz", destPath)
	require.NoError(t, err)

	data, err := os.ReadFile(destPath) //nolint:gosec // destPath is constructed safely via t.TempDir()
	require.NoError(t, err)
	assert.Equal(t, content, data)
}

func TestDownloadFile_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	app := newTestApp()
	client := &http.Client{Timeout: 5 * time.Second}
	destPath := filepath.Join(t.TempDir(), "downloaded.tar.gz")

	err := app.downloadFile(client, server.URL+"/file.tar.gz", destPath)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrArchiveDownloadFailed)
}

// --- Error sentinel tests ---

func TestErrSentinels(t *testing.T) {
	require.Error(t, ErrDevVersionNoForce)
	require.Error(t, ErrVersionParseFailed)
	require.Error(t, ErrDownloadFailed)
	require.Error(t, ErrBinaryNotFoundInArchive)
	require.Error(t, ErrChecksumMismatch)
	require.Error(t, ErrChecksumNotFound)
	require.Error(t, ErrChecksumsDownloadFailed)
	require.Error(t, ErrArchiveDownloadFailed)
}
