// Package version provides version comparison and GitHub release fetching utilities
package version

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"runtime"
	"strings"
	"time"
)

// Default configuration constants
const (
	DefaultBaseURL      = "https://api.github.com"
	DefaultTimeout      = 10 * time.Second
	maxErrorBodySize    = 1024      // 1KB limit for error response bodies
	maxResponseBodySize = 64 * 1024 // 64KB limit for success response bodies
)

// Errors returned by this package
var (
	ErrGitHubAPIFailed  = errors.New("GitHub API request failed")
	ErrInvalidOwner     = errors.New("owner cannot be empty")
	ErrInvalidRepo      = errors.New("repo cannot be empty")
	ErrInvalidOwnerRepo = errors.New("owner/repo contains invalid characters")
)

// validOwnerRepoPattern matches valid GitHub owner/repo names
// GitHub allows alphanumeric, hyphens, underscores, and dots
var validOwnerRepoPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]*$`)

// GitHubRelease represents a GitHub release
type GitHubRelease struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Draft       bool      `json:"draft"`
	Prerelease  bool      `json:"prerelease"`
	PublishedAt time.Time `json:"published_at"`
	Body        string    `json:"body"`
}

// Info contains version information
type Info struct {
	Current string
	Latest  string
	IsNewer bool
}

// Client provides methods for fetching GitHub releases with configurable settings.
// Use NewClient to create a properly initialized client.
type Client struct {
	baseURL    string
	httpClient *http.Client
	userAgent  string
}

// Option configures a Client
type Option func(*Client)

// WithBaseURL sets a custom base URL for the GitHub API
func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.baseURL = strings.TrimSuffix(url, "/")
	}
}

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) {
		c.httpClient = client
	}
}

// WithTimeout sets the HTTP client timeout
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// WithUserAgent sets a custom user agent string
func WithUserAgent(userAgent string) Option {
	return func(c *Client) {
		c.userAgent = userAgent
	}
}

// NewClient creates a new Client with the given options
func NewClient(opts ...Option) *Client {
	c := &Client{
		baseURL: DefaultBaseURL,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
		userAgent: fmt.Sprintf("go-invoice/dev (%s/%s)", runtime.GOOS, runtime.GOARCH),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// defaultClient is the package-level default client
var defaultClient = NewClient() //nolint:gochecknoglobals // Intentional package-level convenience client

// GetLatestRelease fetches the latest release from GitHub using the default client.
// This is a convenience wrapper around Client.GetLatestRelease.
func GetLatestRelease(ctx context.Context, owner, repo string) (*GitHubRelease, error) {
	return defaultClient.GetLatestRelease(ctx, owner, repo)
}

// SetDefaultClient replaces the package-level default client.
// Intended for use in tests to inject a custom client (e.g. pointing at a test server).
func SetDefaultClient(c *Client) {
	defaultClient = c
}

// ResetDefaultClient restores the package-level default client to its original value.
// Call this in test cleanup (e.g. defer versionpkg.ResetDefaultClient()).
func ResetDefaultClient() {
	defaultClient = NewClient()
}

// validateOwnerRepo validates the owner and repo parameters
func validateOwnerRepo(owner, repo string) error {
	if owner == "" {
		return ErrInvalidOwner
	}
	if repo == "" {
		return ErrInvalidRepo
	}
	if !validOwnerRepoPattern.MatchString(owner) || !validOwnerRepoPattern.MatchString(repo) {
		return ErrInvalidOwnerRepo
	}
	return nil
}

// GetLatestRelease fetches the latest release from GitHub
func (c *Client) GetLatestRelease(ctx context.Context, owner, repo string) (*GitHubRelease, error) {
	if err := validateOwnerRepo(owner, repo); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", c.baseURL, owner, repo)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := c.httpClient.Do(req) //nolint:gosec // G704: URL is a constructed GitHub API URL, not user-controlled
	if err != nil {
		return nil, fmt.Errorf("fetching release: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		// Limit error body read to prevent memory exhaustion
		limitedReader := io.LimitReader(resp.Body, maxErrorBodySize)
		body, _ := io.ReadAll(limitedReader)
		return nil, fmt.Errorf("%w: status %d: %s", ErrGitHubAPIFailed, resp.StatusCode, string(body))
	}

	// Limit response body to prevent memory exhaustion from malicious servers
	limitedReader := io.LimitReader(resp.Body, maxResponseBodySize)
	var release GitHubRelease
	if err := json.NewDecoder(limitedReader).Decode(&release); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &release, nil
}

// CompareVersions compares two version strings
// Returns:
//   - 1 if v1 > v2
//   - 0 if v1 == v2
//   - -1 if v1 < v2
func CompareVersions(v1, v2 string) int {
	// Clean versions (remove 'v' prefix if present)
	v1 = strings.TrimPrefix(v1, "v")
	v2 = strings.TrimPrefix(v2, "v")

	// Handle development versions and commit hashes
	isV1Dev := v1 == "dev" || v1 == "" || isCommitHash(v1)
	isV2Dev := v2 == "dev" || v2 == "" || isCommitHash(v2)

	if isV1Dev && isV2Dev {
		return 0
	}
	if isV1Dev {
		return -1 // dev/commit is always considered older than a release
	}
	if isV2Dev {
		return 1
	}

	// Split versions into parts
	parts1 := parseVersion(v1)
	parts2 := parseVersion(v2)

	// Compare major, minor, patch
	for i := 0; i < 3; i++ {
		if i >= len(parts1) && i >= len(parts2) {
			break
		}
		val1 := 0
		val2 := 0
		if i < len(parts1) {
			val1 = parts1[i]
		}
		if i < len(parts2) {
			val2 = parts2[i]
		}

		if val1 > val2 {
			return 1
		}
		if val1 < val2 {
			return -1
		}
	}

	return 0
}

// parseVersion parses a version string into major, minor, patch integers
func parseVersion(version string) []int {
	// Remove any suffixes like -dirty, -rc1, +build, etc.
	if idx := strings.IndexAny(version, "-+"); idx != -1 {
		version = version[:idx]
	}

	parts := strings.Split(version, ".")
	result := make([]int, 0, len(parts))

	for _, part := range parts {
		var num int
		if _, err := fmt.Sscanf(part, "%d", &num); err == nil {
			result = append(result, num)
		}
	}

	return result
}

// IsNewerVersion checks if latestVersion is newer than currentVersion
func IsNewerVersion(currentVersion, latestVersion string) bool {
	return CompareVersions(latestVersion, currentVersion) > 0
}

// NormalizeVersion ensures version strings are in a consistent format.
// It removes the 'v' prefix, trims whitespace, and removes any pre-release
// or build metadata suffixes (e.g., -rc1, -dirty, +build).
func NormalizeVersion(version string) string {
	// Remove any pre-release or build metadata suffixes
	if idx := strings.IndexAny(version, "-+"); idx != -1 {
		version = version[:idx]
	}

	// Remove leading/trailing whitespace and 'v' prefixes until stable
	for {
		trimmed := strings.TrimSpace(version)
		trimmed = strings.TrimLeft(trimmed, "v")
		if trimmed == version {
			break
		}
		version = trimmed
	}

	return version
}

// isCommitHash checks if a string looks like a git commit hash.
// It requires the string to:
// - Be 7-40 characters long (short to full SHA-1)
// - Contain only hex characters (0-9, a-f, A-F)
// - Contain at least one letter (to distinguish from pure numeric versions)
func isCommitHash(s string) bool {
	// Remove any -dirty suffix
	s = strings.TrimSuffix(s, "-dirty")

	// Commit hashes are typically 7-40 hex characters
	if len(s) < 7 || len(s) > 40 {
		return false
	}

	hasLetter := false
	for _, c := range s {
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
