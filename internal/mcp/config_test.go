package mcp

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ConfigTestSuite struct {
	suite.Suite

	tempDir string
}

func TestConfigSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}

func (s *ConfigTestSuite) SetupTest() {
	tempDir, err := os.MkdirTemp("", "mcp-config-test")
	s.Require().NoError(err)
	s.tempDir = tempDir
}

func (s *ConfigTestSuite) TearDownTest() {
	if s.tempDir != "" {
		err := os.RemoveAll(s.tempDir)
		s.Require().NoError(err)
	}
}

func (s *ConfigTestSuite) TestLoadConfigWithDefaultCreation() {
	ctx := context.Background()

	// Set config path to temp directory
	configPath := filepath.Join(s.tempDir, "test-config.json")
	err := os.Setenv("MCP_CONFIG_PATH", configPath)
	s.Require().NoError(err)
	defer func() {
		unsetErr := os.Unsetenv("MCP_CONFIG_PATH")
		s.Require().NoError(unsetErr)
	}()

	config, err := LoadConfig(ctx)
	s.Require().NoError(err)
	s.NotNil(config)

	// Verify default values
	s.Equal("localhost", config.Server.Host)
	s.Equal(0, config.Server.Port)
	s.Equal("go-invoice", config.CLI.Path)
	s.Equal("info", config.LogLevel)
	s.True(config.Security.SandboxEnabled)
	s.Contains(config.Security.AllowedCommands, "go-invoice")

	// Verify config file was created
	s.FileExists(configPath)
}

func (s *ConfigTestSuite) TestLoadConfigFromExistingFile() {
	ctx := context.Background()

	configPath := filepath.Join(s.tempDir, "existing-config.json")

	// Create test config
	testConfig := &Config{
		Server: ServerConfig{
			Host:        "test-host",
			Port:        8080,
			Timeout:     45 * time.Second,
			ReadTimeout: 15 * time.Second,
		},
		CLI: CLIConfig{
			Path:       "test-cli",
			WorkingDir: s.tempDir, // Use temp dir instead of /test/dir
			MaxTimeout: 30 * time.Second,
		},
		Security: SecurityConfig{
			AllowedCommands:       []string{"test-command"},
			WorkingDir:            s.tempDir, // Use temp dir instead of /test/security
			SandboxEnabled:        false,
			FileAccessRestricted:  false,
			MaxCommandTimeout:     "30s",
			EnableInputValidation: false,
		},
		LogLevel: "debug",
	}

	// Write test config
	data, err := json.MarshalIndent(testConfig, "", "  ")
	s.Require().NoError(err)
	s.Require().NoError(os.WriteFile(configPath, data, 0o600))

	// Set config path
	err = os.Setenv("MCP_CONFIG_PATH", configPath)
	s.Require().NoError(err)
	defer func() {
		unsetErr := os.Unsetenv("MCP_CONFIG_PATH")
		s.Require().NoError(unsetErr)
	}()

	config, err := LoadConfig(ctx)
	s.Require().NoError(err)

	// Verify loaded values
	s.Equal("test-host", config.Server.Host)
	s.Equal(8080, config.Server.Port)
	s.Equal("test-cli", config.CLI.Path)
	s.Equal("debug", config.LogLevel)
	s.False(config.Security.SandboxEnabled)
	s.Contains(config.Security.AllowedCommands, "test-command")
}

func (s *ConfigTestSuite) TestLoadConfigContextCancellation() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Immediately cancel

	_, err := LoadConfig(ctx)
	s.Equal(context.Canceled, err)
}

func (s *ConfigTestSuite) TestValidateConfigSuccess() {
	ctx := context.Background()

	config := getDefaultConfig()
	config.CLI.WorkingDir = s.tempDir
	config.Security.WorkingDir = s.tempDir

	err := validateConfig(ctx, config)
	s.NoError(err)
}

func (s *ConfigTestSuite) TestValidateConfigFailures() {
	ctx := context.Background()

	tests := []struct {
		name     string
		modifier func(*Config)
		contains string
	}{
		{
			name: "InvalidServerPort",
			modifier: func(c *Config) {
				c.Server.Port = -1
			},
			contains: "invalid server port",
		},
		{
			name: "InvalidServerTimeout",
			modifier: func(c *Config) {
				c.Server.Timeout = -1 * time.Second
			},
			contains: "invalid server timeout",
		},
		{
			name: "EmptyCLIPath",
			modifier: func(c *Config) {
				c.CLI.Path = ""
			},
			contains: "CLI path cannot be empty",
		},
		{
			name: "InvalidCLITimeout",
			modifier: func(c *Config) {
				c.CLI.MaxTimeout = -1 * time.Second
			},
			contains: "invalid CLI max timeout",
		},
		{
			name: "EmptyAllowedCommands",
			modifier: func(c *Config) {
				c.Security.AllowedCommands = []string{}
			},
			contains: "allowed commands list cannot be empty",
		},
		{
			name: "InvalidLogLevel",
			modifier: func(c *Config) {
				c.LogLevel = "invalid"
			},
			contains: "invalid log level",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			config := getDefaultConfig()
			config.CLI.WorkingDir = s.tempDir
			config.Security.WorkingDir = s.tempDir
			tt.modifier(config)

			err := validateConfig(ctx, config)
			s.Require().Error(err)
			s.Contains(err.Error(), tt.contains)
		})
	}
}

func (s *ConfigTestSuite) TestValidateConfigContextCancellation() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	config := getDefaultConfig()
	err := validateConfig(ctx, config)
	s.Equal(context.Canceled, err)
}

func (s *ConfigTestSuite) TestApplyEnvironmentOverrides() {
	config := getDefaultConfig()

	// Set environment variables
	err := os.Setenv("MCP_LOG_LEVEL", "debug")
	s.Require().NoError(err)
	err = os.Setenv("GO_INVOICE_CLI_PATH", "custom-cli")
	s.Require().NoError(err)
	err = os.Setenv("GO_INVOICE_HOME", "/custom/home")
	s.Require().NoError(err)
	defer func() {
		err := os.Unsetenv("MCP_LOG_LEVEL")
		s.Require().NoError(err)
		err = os.Unsetenv("GO_INVOICE_CLI_PATH")
		s.Require().NoError(err)
		err = os.Unsetenv("GO_INVOICE_HOME")
		s.Require().NoError(err)
	}()

	applyEnvironmentOverrides(config)

	s.Equal("debug", config.LogLevel)
	s.Equal("custom-cli", config.CLI.Path)
	s.Equal("/custom/home", config.CLI.WorkingDir)
	s.Equal("/custom/home", config.Security.WorkingDir)
}

func (s *ConfigTestSuite) TestGetConfigPath() {
	// Test command line argument
	originalArgs := os.Args
	os.Args = []string{"program", "--config", "/test/config.json"}
	defer func() { os.Args = originalArgs }()

	path := getConfigPath()
	s.Equal("/test/config.json", path)
}

func (s *ConfigTestSuite) TestGetConfigPathEnvironment() {
	// Clear args and test environment
	originalArgs := os.Args
	os.Args = []string{"program"}
	defer func() { os.Args = originalArgs }()

	err := os.Setenv("MCP_CONFIG_PATH", "/env/config.json")
	s.Require().NoError(err)
	defer func() {
		err := os.Unsetenv("MCP_CONFIG_PATH")
		s.Require().NoError(err)
	}()

	path := getConfigPath()
	s.Equal("/env/config.json", path)
}

func (s *ConfigTestSuite) TestSaveConfig() {
	configPath := filepath.Join(s.tempDir, "save-test.json")
	config := getDefaultConfig()

	err := saveConfig(configPath, config)
	s.Require().NoError(err)

	s.FileExists(configPath)

	// Verify file content
	// #nosec G304 -- configPath is constructed in the test from known temp directory
	data, err := os.ReadFile(configPath)
	s.Require().NoError(err)

	var loaded Config
	err = json.Unmarshal(data, &loaded)
	s.Require().NoError(err)

	s.Equal(config.LogLevel, loaded.LogLevel)
	s.Equal(config.CLI.Path, loaded.CLI.Path)
}

func (s *ConfigTestSuite) TestEnsureDirectoryExists() {
	nonExistentDir := filepath.Join(s.tempDir, "new", "nested", "dir")

	err := ensureDirectoryExists(nonExistentDir)
	s.Require().NoError(err)

	s.DirExists(nonExistentDir)
}

// Benchmark tests for performance validation
func BenchmarkLoadConfig(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "mcp-config-bench")
	require.NoError(b, err)
	defer func() {
		removeErr := os.RemoveAll(tempDir)
		require.NoError(b, removeErr)
	}()

	configPath := filepath.Join(tempDir, "bench-config.json")
	err = os.Setenv("MCP_CONFIG_PATH", configPath)
	require.NoError(b, err)
	defer func() {
		err := os.Unsetenv("MCP_CONFIG_PATH")
		require.NoError(b, err)
	}()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := LoadConfig(ctx)
		require.NoError(b, err)
	}
}

func BenchmarkValidateConfig(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "mcp-validate-bench")
	require.NoError(b, err)
	defer func() {
		err := os.RemoveAll(tempDir)
		require.NoError(b, err)
	}()

	config := getDefaultConfig()
	config.CLI.WorkingDir = tempDir
	config.Security.WorkingDir = tempDir
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := validateConfig(ctx, config)
		require.NoError(b, err)
	}
}
