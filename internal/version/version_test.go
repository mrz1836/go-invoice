package version

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	t.Parallel()

	t.Run("DefaultValues", func(t *testing.T) {
		t.Parallel()
		client := NewClient()

		assert.Equal(t, DefaultBaseURL, client.baseURL)
		assert.NotNil(t, client.httpClient)
		assert.Equal(t, DefaultTimeout, client.httpClient.Timeout)
		assert.Contains(t, client.userAgent, "go-invoice")
	})

	t.Run("WithBaseURL", func(t *testing.T) {
		t.Parallel()
		client := NewClient(WithBaseURL("https://custom.api.github.com/"))

		// Should trim trailing slash
		assert.Equal(t, "https://custom.api.github.com", client.baseURL)
	})

	t.Run("WithHTTPClient", func(t *testing.T) {
		t.Parallel()
		customClient := &http.Client{Timeout: 30 * time.Second}
		client := NewClient(WithHTTPClient(customClient))

		assert.Equal(t, customClient, client.httpClient)
	})

	t.Run("WithTimeout", func(t *testing.T) {
		t.Parallel()
		client := NewClient(WithTimeout(5 * time.Second))

		assert.Equal(t, 5*time.Second, client.httpClient.Timeout)
	})

	t.Run("WithUserAgent", func(t *testing.T) {
		t.Parallel()
		client := NewClient(WithUserAgent("custom-agent/1.0"))

		assert.Equal(t, "custom-agent/1.0", client.userAgent)
	})

	t.Run("MultipleOptions", func(t *testing.T) {
		t.Parallel()
		client := NewClient(
			WithBaseURL("https://custom.example.com"),
			WithTimeout(20*time.Second),
			WithUserAgent("multi-option/1.0"),
		)

		assert.Equal(t, "https://custom.example.com", client.baseURL)
		assert.Equal(t, 20*time.Second, client.httpClient.Timeout)
		assert.Equal(t, "multi-option/1.0", client.userAgent)
	})
}

func TestValidateOwnerRepo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		owner       string
		repo        string
		expectedErr error
	}{
		{
			name:        "ValidOwnerRepo",
			owner:       "mrz1836",
			repo:        "go-invoice",
			expectedErr: nil,
		},
		{
			name:        "EmptyOwner",
			owner:       "",
			repo:        "go-invoice",
			expectedErr: ErrInvalidOwner,
		},
		{
			name:        "EmptyRepo",
			owner:       "mrz1836",
			repo:        "",
			expectedErr: ErrInvalidRepo,
		},
		{
			name:        "BothEmpty",
			owner:       "",
			repo:        "",
			expectedErr: ErrInvalidOwner,
		},
		{
			name:        "OwnerWithSlash",
			owner:       "../etc",
			repo:        "passwd",
			expectedErr: ErrInvalidOwnerRepo,
		},
		{
			name:        "RepoWithSlash",
			owner:       "valid",
			repo:        "../etc/passwd",
			expectedErr: ErrInvalidOwnerRepo,
		},
		{
			name:        "OwnerStartsWithDot",
			owner:       ".hidden",
			repo:        "repo",
			expectedErr: ErrInvalidOwnerRepo,
		},
		{
			name:        "OwnerStartsWithHyphen",
			owner:       "-invalid",
			repo:        "repo",
			expectedErr: ErrInvalidOwnerRepo,
		},
		{
			name:        "ValidWithHyphens",
			owner:       "my-org",
			repo:        "my-repo",
			expectedErr: nil,
		},
		{
			name:        "ValidWithUnderscores",
			owner:       "my_org",
			repo:        "my_repo",
			expectedErr: nil,
		},
		{
			name:        "ValidWithDots",
			owner:       "my.org",
			repo:        "my.repo",
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateOwnerRepo(tt.owner, tt.repo)
			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClientGetLatestRelease(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		mockResponse    string
		mockStatusCode  int
		expectedRelease *GitHubRelease
		expectError     bool
		errorContains   string
	}{
		{
			name:           "ValidRelease",
			mockStatusCode: http.StatusOK,
			mockResponse: `{
				"tag_name": "v1.2.3",
				"name": "Release v1.2.3",
				"draft": false,
				"prerelease": false,
				"published_at": "2023-01-01T12:00:00Z",
				"body": "Bug fixes and improvements"
			}`,
			expectedRelease: &GitHubRelease{
				TagName:     "v1.2.3",
				Name:        "Release v1.2.3",
				Draft:       false,
				Prerelease:  false,
				PublishedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
				Body:        "Bug fixes and improvements",
			},
			expectError: false,
		},
		{
			name:           "InvalidJSON",
			mockStatusCode: http.StatusOK,
			mockResponse:   `{invalid json`,
			expectError:    true,
			errorContains:  "decoding response",
		},
		{
			name:           "NotFound",
			mockStatusCode: http.StatusNotFound,
			mockResponse:   `{"message": "Not Found"}`,
			expectError:    true,
			errorContains:  "GitHub API request failed",
		},
		{
			name:           "RateLimited",
			mockStatusCode: http.StatusForbidden,
			mockResponse:   `{"message": "API rate limit exceeded"}`,
			expectError:    true,
			errorContains:  "GitHub API request failed",
		},
		{
			name:           "InternalServerError",
			mockStatusCode: http.StatusInternalServerError,
			mockResponse:   `{"message": "Internal Server Error"}`,
			expectError:    true,
			errorContains:  "GitHub API request failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/repos/owner/repo/releases/latest", r.URL.Path)
				assert.Contains(t, r.Header.Get("User-Agent"), "go-invoice")
				assert.Equal(t, "application/vnd.github.v3+json", r.Header.Get("Accept"))

				w.WriteHeader(tt.mockStatusCode)
				_, _ = w.Write([]byte(tt.mockResponse))
			}))
			defer server.Close()

			// Create client with mock server URL
			client := NewClient(WithBaseURL(server.URL))

			// Test
			release, err := client.GetLatestRelease(context.Background(), "owner", "repo")

			if tt.expectError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.Nil(t, release)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedRelease, release)
			}
		})
	}
}

func TestGetLatestReleaseInputValidation(t *testing.T) {
	t.Parallel()

	client := NewClient()
	ctx := context.Background()

	t.Run("EmptyOwner", func(t *testing.T) {
		t.Parallel()
		_, err := client.GetLatestRelease(ctx, "", "repo")
		assert.ErrorIs(t, err, ErrInvalidOwner)
	})

	t.Run("EmptyRepo", func(t *testing.T) {
		t.Parallel()
		_, err := client.GetLatestRelease(ctx, "owner", "")
		assert.ErrorIs(t, err, ErrInvalidRepo)
	})

	t.Run("InvalidOwner", func(t *testing.T) {
		t.Parallel()
		_, err := client.GetLatestRelease(ctx, "../malicious", "repo")
		assert.ErrorIs(t, err, ErrInvalidOwnerRepo)
	})
}

func TestGetLatestReleaseContextCancellation(t *testing.T) {
	t.Parallel()

	// Create a slow server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"tag_name": "v1.0.0"}`))
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))

	// Create a context that will be canceled quickly
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := client.GetLatestRelease(ctx, "owner", "repo")
	require.Error(t, err)
	assert.True(t, errors.Is(err, context.DeadlineExceeded) || strings.Contains(err.Error(), "context"))
}

func TestGetLatestReleaseErrorBodyLimit(t *testing.T) {
	t.Parallel()

	// Create a server that returns a huge error body
	largeBody := strings.Repeat("x", maxErrorBodySize*2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(largeBody))
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))

	_, err := client.GetLatestRelease(context.Background(), "owner", "repo")
	require.Error(t, err)

	// Error should be truncated to maxErrorBodySize
	errStr := err.Error()
	// The error should not contain the full large body
	assert.Less(t, len(errStr), len(largeBody))
}

func TestPackageLevelGetLatestRelease(t *testing.T) {
	// This test verifies the package-level function works with the default client
	// We can't easily mock it without modifying the default client,
	// so we just verify it returns appropriate errors for invalid input
	t.Parallel()

	_, err := GetLatestRelease(context.Background(), "", "repo")
	assert.ErrorIs(t, err, ErrInvalidOwner)
}

func TestCompareVersions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		v1       string
		v2       string
		expected int
	}{
		{
			name:     "V1Greater",
			v1:       "1.2.3",
			v2:       "1.2.2",
			expected: 1,
		},
		{
			name:     "V2Greater",
			v1:       "1.2.2",
			v2:       "1.2.3",
			expected: -1,
		},
		{
			name:     "Equal",
			v1:       "1.2.3",
			v2:       "1.2.3",
			expected: 0,
		},
		{
			name:     "MajorVersionDifference",
			v1:       "2.0.0",
			v2:       "1.9.9",
			expected: 1,
		},
		{
			name:     "MinorVersionDifference",
			v1:       "1.3.0",
			v2:       "1.2.9",
			expected: 1,
		},
		{
			name:     "WithVPrefix",
			v1:       "v1.2.3",
			v2:       "v1.2.2",
			expected: 1,
		},
		{
			name:     "MixedVPrefix",
			v1:       "v1.2.3",
			v2:       "1.2.3",
			expected: 0,
		},
		{
			name:     "DevVersionVsRelease",
			v1:       "dev",
			v2:       "1.2.3",
			expected: -1,
		},
		{
			name:     "ReleaseVsDevVersion",
			v1:       "1.2.3",
			v2:       "dev",
			expected: 1,
		},
		{
			name:     "BothDevVersions",
			v1:       "dev",
			v2:       "dev",
			expected: 0,
		},
		{
			name:     "CommitHashVsRelease",
			v1:       "abc123def456",
			v2:       "1.2.3",
			expected: -1,
		},
		{
			name:     "ReleaseVsCommitHash",
			v1:       "1.2.3",
			v2:       "abc123def456",
			expected: 1,
		},
		{
			name:     "EmptyVersionVsRelease",
			v1:       "",
			v2:       "1.2.3",
			expected: -1,
		},
		{
			name:     "VersionWithSuffix",
			v1:       "1.2.3-rc1",
			v2:       "1.2.3",
			expected: 0, // Suffixes are ignored in basic comparison
		},
		{
			name:     "TwoPartVersion",
			v1:       "1.2",
			v2:       "1.2.0",
			expected: 0,
		},
		{
			name:     "SinglePartVersion",
			v1:       "2",
			v2:       "1.9.9",
			expected: 1,
		},
		{
			name:     "SemanticVersionComparison",
			v1:       "1.5.5",
			v2:       "1.6.0",
			expected: -1,
		},
		{
			name:     "PureNumericSevenDigitIsVersion",
			v1:       "1234567",
			v2:       "1.0.0",
			expected: 1, // Treated as a large major version, not a commit
		},
		{
			name:     "PureNumericTenDigitIsVersion",
			v1:       "2024010100",
			v2:       "1.0.0",
			expected: 1, // Date-based version, not a commit
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := CompareVersions(tt.v1, tt.v2)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsNewerVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		currentVersion string
		latestVersion  string
		expected       bool
	}{
		{
			name:           "NewerAvailable",
			currentVersion: "1.2.2",
			latestVersion:  "1.2.3",
			expected:       true,
		},
		{
			name:           "SameVersion",
			currentVersion: "1.2.3",
			latestVersion:  "1.2.3",
			expected:       false,
		},
		{
			name:           "CurrentNewer",
			currentVersion: "1.2.4",
			latestVersion:  "1.2.3",
			expected:       false,
		},
		{
			name:           "DevVersionNeedsUpgrade",
			currentVersion: "dev",
			latestVersion:  "1.2.3",
			expected:       true,
		},
		{
			name:           "CommitHashNeedsUpgrade",
			currentVersion: "abc123def456",
			latestVersion:  "1.2.3",
			expected:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := IsNewerVersion(tt.currentVersion, tt.latestVersion)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNormalizeVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		version  string
		expected string
	}{
		{
			name:     "WithVPrefix",
			version:  "v1.2.3",
			expected: "1.2.3",
		},
		{
			name:     "WithoutVPrefix",
			version:  "1.2.3",
			expected: "1.2.3",
		},
		{
			name:     "WithDashSuffix",
			version:  "1.2.3-rc1",
			expected: "1.2.3",
		},
		{
			name:     "WithPlusSuffix",
			version:  "1.2.3+build123",
			expected: "1.2.3",
		},
		{
			name:     "WithSpaces",
			version:  "  1.2.3  ",
			expected: "1.2.3",
		},
		{
			name:     "WithVPrefixAndDirtySuffix",
			version:  "v1.2.3-dirty",
			expected: "1.2.3",
		},
		{
			name:     "WithVPrefixAndBuildSuffix",
			version:  "v1.2.3+build456",
			expected: "1.2.3",
		},
		{
			name:     "EmptyString",
			version:  "",
			expected: "",
		},
		{
			name:     "OnlyV",
			version:  "v",
			expected: "",
		},
		{
			name:     "ComplexSuffix",
			version:  "1.2.3-rc1+build.456",
			expected: "1.2.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := NormalizeVersion(tt.version)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsCommitHash(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		version  string
		expected bool
	}{
		{
			name:     "ValidShortCommitHash",
			version:  "abc123d",
			expected: true,
		},
		{
			name:     "ValidLongCommitHash",
			version:  "abc123def456789012345678901234567890abcd",
			expected: true,
		},
		{
			name:     "ValidHashWithDirtySuffix",
			version:  "abc123d-dirty",
			expected: true,
		},
		{
			name:     "ValidMixedCaseHash",
			version:  "AbC123DeF456",
			expected: true,
		},
		{
			name:     "TooShort",
			version:  "abc12",
			expected: false,
		},
		{
			name:     "TooLong",
			version:  "abc123def456789012345678901234567890abcdef",
			expected: false,
		},
		{
			name:     "ContainsInvalidCharacters",
			version:  "abc123xyz",
			expected: false,
		},
		{
			name:     "ContainsSpecialCharacters",
			version:  "abc123-def",
			expected: false,
		},
		{
			name:     "EmptyString",
			version:  "",
			expected: false,
		},
		{
			name:     "StandardVersion",
			version:  "1.2.3",
			expected: false,
		},
		{
			name:     "DevVersion",
			version:  "dev",
			expected: false,
		},
		{
			name:     "OnlyNumbersSevenDigit",
			version:  "1234567",
			expected: false, // Pure numbers are versions, not commits
		},
		{
			name:     "OnlyNumbersTenDigit",
			version:  "1234567890",
			expected: false, // Pure numbers are versions, not commits
		},
		{
			name:     "DateBasedVersion",
			version:  "2024010100",
			expected: false, // Date-based version, not commit
		},
		{
			name:     "OnlyValidHexLetters",
			version:  "abcdefabcdef",
			expected: true,
		},
		{
			name:     "OnlyInvalidLetters",
			version:  "abcdefghijk",
			expected: false,
		},
		{
			name:     "MixedHexWithNumbers",
			version:  "1a2b3c4d",
			expected: true, // Has letters, is valid hex
		},
		{
			name:     "AllZeros",
			version:  "0000000",
			expected: false, // No letters, just zeros
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := isCommitHash(tt.version)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		version  string
		expected []int
	}{
		{
			name:     "StandardVersion",
			version:  "1.2.3",
			expected: []int{1, 2, 3},
		},
		{
			name:     "TwoPartVersion",
			version:  "1.2",
			expected: []int{1, 2},
		},
		{
			name:     "SinglePartVersion",
			version:  "1",
			expected: []int{1},
		},
		{
			name:     "VersionWithDashSuffix",
			version:  "1.2.3-rc1",
			expected: []int{1, 2, 3},
		},
		{
			name:     "VersionWithPlusSuffix",
			version:  "1.2.3+build123",
			expected: []int{1, 2, 3},
		},
		{
			name:     "EmptyString",
			version:  "",
			expected: []int{},
		},
		{
			name:     "InvalidVersion",
			version:  "abc.def.ghi",
			expected: []int{},
		},
		{
			name:     "MixedValidInvalid",
			version:  "1.abc.3",
			expected: []int{1, 3},
		},
		{
			name:     "LargeNumbers",
			version:  "999.888.777",
			expected: []int{999, 888, 777},
		},
		{
			name:     "ZeroVersion",
			version:  "0.0.0",
			expected: []int{0, 0, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := parseVersion(tt.version)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestVersionInfo(t *testing.T) {
	t.Parallel()

	info := Info{
		Current: "1.2.2",
		Latest:  "1.2.3",
		IsNewer: true,
	}

	assert.Equal(t, "1.2.2", info.Current)
	assert.Equal(t, "1.2.3", info.Latest)
	assert.True(t, info.IsNewer)
}

func TestGitHubRelease(t *testing.T) {
	t.Parallel()

	publishedAt := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	release := GitHubRelease{
		TagName:     "v1.2.3",
		Name:        "Release v1.2.3",
		Draft:       false,
		Prerelease:  false,
		PublishedAt: publishedAt,
		Body:        "Bug fixes and improvements",
	}

	assert.Equal(t, "v1.2.3", release.TagName)
	assert.Equal(t, "Release v1.2.3", release.Name)
	assert.False(t, release.Draft)
	assert.False(t, release.Prerelease)
	assert.Equal(t, publishedAt, release.PublishedAt)
	assert.Equal(t, "Bug fixes and improvements", release.Body)
}

// TestClientConcurrentUse verifies that Client can be used concurrently without races
func TestClientConcurrentUse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"tag_name": "v1.0.0"}`))
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))

	var wg sync.WaitGroup
	errCh := make(chan error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := client.GetLatestRelease(context.Background(), "owner", "repo")
			if err != nil {
				errCh <- err
			}
		}()
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Errorf("unexpected error: %v", err)
	}
}

// Benchmarks
func BenchmarkCompareVersions(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CompareVersions("1.2.3", "1.2.4")
	}
}

func BenchmarkIsNewerVersion(b *testing.B) {
	for i := 0; i < b.N; i++ {
		IsNewerVersion("1.2.3", "1.2.4")
	}
}

func BenchmarkNormalizeVersion(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NormalizeVersion("v1.2.3-rc1")
	}
}

func BenchmarkIsCommitHash(b *testing.B) {
	for i := 0; i < b.N; i++ {
		isCommitHash("abc123def456")
	}
}

func BenchmarkNewClient(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewClient(
			WithBaseURL("https://api.example.com"),
			WithTimeout(5*time.Second),
		)
	}
}

// Fuzz tests
func FuzzCompareVersions(f *testing.F) {
	// Seed corpus
	f.Add("1.2.3", "1.2.4")
	f.Add("v1.0.0", "2.0.0")
	f.Add("dev", "1.0.0")
	f.Add("abc123def", "1.0.0")
	f.Add("", "")
	f.Add("1.2.3-rc1", "1.2.3")
	f.Add("1.2.3+build", "1.2.3")
	f.Add("999999999.0.0", "0.0.1")

	f.Fuzz(func(t *testing.T, v1, v2 string) {
		// Skip long inputs to avoid timeout in CI
		if len(v1)+len(v2) > 1000 {
			t.Skipf("Input too large: %d bytes (limit: 1000)", len(v1)+len(v2))
		}

		// Create context with timeout to prevent expensive operations from hanging
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		// Check context before expensive operations
		select {
		case <-ctx.Done():
			t.Skipf("Context timeout before operation")
		default:
		}

		// Should never panic
		result := CompareVersions(v1, v2)
		// Result must be -1, 0, or 1
		if result < -1 || result > 1 {
			t.Errorf("CompareVersions(%q, %q) = %d, want -1, 0, or 1", v1, v2, result)
		}
	})
}

func FuzzParseVersion(f *testing.F) {
	f.Add("1.2.3")
	f.Add("v1.0.0-rc1")
	f.Add("1.2.3+build")
	f.Add("")
	f.Add("abc.def.ghi")
	f.Add("999.888.777")

	f.Fuzz(func(t *testing.T, version string) {
		// Skip long inputs to avoid timeout in CI
		if len(version) > 500 {
			t.Skipf("Input too large: %d bytes (limit: 500)", len(version))
		}

		// Create context with timeout to prevent expensive operations from hanging
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		// Check context before expensive operations
		select {
		case <-ctx.Done():
			t.Skipf("Context timeout before operation")
		default:
		}

		// Should never panic
		result := parseVersion(version)
		// Result should be a valid slice (not nil after make)
		if result == nil {
			t.Errorf("parseVersion(%q) returned nil", version)
		}
		// All elements should be non-negative
		for i, v := range result {
			if v < 0 {
				t.Errorf("parseVersion(%q)[%d] = %d, want >= 0", version, i, v)
			}
		}
	})
}

func FuzzNormalizeVersion(f *testing.F) {
	f.Add("v1.2.3")
	f.Add("1.2.3-dirty")
	f.Add("1.2.3+build")
	f.Add("  v1.2.3  ")
	f.Add("")

	f.Fuzz(func(t *testing.T, version string) {
		// Skip long inputs to avoid timeout in CI
		if len(version) > 500 {
			t.Skipf("Input too large: %d bytes (limit: 500)", len(version))
		}

		// Create context with timeout to prevent expensive operations from hanging
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		// Check context before expensive operations
		select {
		case <-ctx.Done():
			t.Skipf("Context timeout before operation")
		default:
		}

		// Should never panic
		result := NormalizeVersion(version)

		// Result should not have 'v' prefix
		if strings.HasPrefix(result, "v") {
			t.Errorf("NormalizeVersion(%q) = %q, should not have 'v' prefix", version, result)
		}

		// Result should not have leading/trailing whitespace
		if result != strings.TrimSpace(result) {
			t.Errorf("NormalizeVersion(%q) = %q, has leading/trailing whitespace", version, result)
		}
	})
}

func FuzzIsCommitHash(f *testing.F) {
	f.Add("abc123d")
	f.Add("1234567")
	f.Add("abc123def456789012345678901234567890abcd")
	f.Add("abc123-dirty")
	f.Add("")
	f.Add("dev")

	f.Fuzz(func(t *testing.T, s string) {
		// Skip long inputs to avoid timeout in CI
		if len(s) > 100 {
			t.Skipf("Input too large: %d bytes (limit: 100)", len(s))
		}

		// Create context with timeout to prevent expensive operations from hanging
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		// Check context before expensive operations
		select {
		case <-ctx.Done():
			t.Skipf("Context timeout before operation")
		default:
		}

		// Should never panic
		_ = isCommitHash(s)
	})
}
