package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Static errors for err113 compliance
var (
	ErrInvalidServerPort    = errors.New("invalid server port")
	ErrInvalidServerTimeout = errors.New("invalid server timeout")
	ErrEmptyCLIPath         = errors.New("CLI path cannot be empty")
	ErrInvalidCLIMaxTimeout = errors.New("invalid CLI max timeout")
	ErrEmptyAllowedCommands = errors.New("allowed commands list cannot be empty")
	ErrInvalidLogLevel      = errors.New("invalid log level")
)

// Config represents the complete MCP server configuration
type Config struct {
	Server   ServerConfig   `json:"server"`
	CLI      CLIConfig      `json:"cli"`
	Security SecurityConfig `json:"security"`
	LogLevel string         `json:"logLevel"`
}

// ServerConfig represents server-specific configuration
type ServerConfig struct {
	Host        string        `json:"host"`
	Port        int           `json:"port"`
	Timeout     time.Duration `json:"timeout"`
	ReadTimeout time.Duration `json:"readTimeout"`
}

// CLIConfig represents CLI bridge configuration
type CLIConfig struct {
	Path       string        `json:"path"`
	WorkingDir string        `json:"workingDir"`
	MaxTimeout time.Duration `json:"maxTimeout"`
}

// SecurityConfig represents security-related configuration
type SecurityConfig struct {
	AllowedCommands       []string `json:"allowedCommands"`
	WorkingDir            string   `json:"workingDir"`
	SandboxEnabled        bool     `json:"sandboxEnabled"`
	FileAccessRestricted  bool     `json:"fileAccessRestricted"`
	MaxCommandTimeout     string   `json:"maxCommandTimeout"`
	EnableInputValidation bool     `json:"enableInputValidation"`
}

// LoadConfig loads the MCP server configuration from file with validation
func LoadConfig(ctx context.Context) (*Config, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	configPath := getConfigPath()

	// If config file doesn't exist, create default config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		defaultConfig := getDefaultConfig()
		if err := saveConfig(configPath, defaultConfig); err != nil {
			return nil, fmt.Errorf("failed to create default config at %s: %w", configPath, err)
		}
		return defaultConfig, nil
	}

	// #nosec G304 -- configPath is constructed from known sources (env vars, CLI args, defaults)
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", configPath, err)
	}

	// Validate configuration
	if err := validateConfig(ctx, &config); err != nil {
		return nil, fmt.Errorf("config validation failed for %s: %w", configPath, err)
	}

	// Apply environment variable overrides
	applyEnvironmentOverrides(&config)

	return &config, nil
}

// getConfigPath determines the configuration file path
func getConfigPath() string {
	// Check command line argument
	for i, arg := range os.Args {
		if arg == "--config" && i+1 < len(os.Args) {
			return os.Args[i+1]
		}
	}

	// Check environment variable
	if configPath := os.Getenv("MCP_CONFIG_PATH"); configPath != "" {
		return configPath
	}

	// Default to user home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "mcp-config.json"
	}

	return filepath.Join(homeDir, ".go-invoice", "mcp-config.json")
}

// getDefaultConfig returns the default configuration
func getDefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	invoiceDir := filepath.Join(homeDir, ".go-invoice")

	return &Config{
		Server: ServerConfig{
			Host:        "localhost",
			Port:        0, // Auto-assign port
			Timeout:     30 * time.Second,
			ReadTimeout: 10 * time.Second,
		},
		CLI: CLIConfig{
			Path:       "go-invoice",
			WorkingDir: invoiceDir,
			MaxTimeout: 60 * time.Second,
		},
		Security: SecurityConfig{
			AllowedCommands: []string{
				"go-invoice",
			},
			WorkingDir:            invoiceDir,
			SandboxEnabled:        true,
			FileAccessRestricted:  true,
			MaxCommandTimeout:     "60s",
			EnableInputValidation: true,
		},
		LogLevel: "info",
	}
}

// validateConfig validates the configuration for correctness and security
func validateConfig(ctx context.Context, config *Config) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Validate server configuration
	if config.Server.Port < 0 || config.Server.Port > 65535 {
		return fmt.Errorf("%w: %d (must be 0-65535)", ErrInvalidServerPort, config.Server.Port)
	}

	if config.Server.Timeout <= 0 {
		return fmt.Errorf("%w: %v (must be positive)", ErrInvalidServerTimeout, config.Server.Timeout)
	}

	// Validate CLI configuration
	if config.CLI.Path == "" {
		return ErrEmptyCLIPath
	}

	if config.CLI.MaxTimeout <= 0 {
		return fmt.Errorf("%w: %v (must be positive)", ErrInvalidCLIMaxTimeout, config.CLI.MaxTimeout)
	}

	// Validate security configuration
	if len(config.Security.AllowedCommands) == 0 {
		return ErrEmptyAllowedCommands
	}

	// Validate working directory exists or can be created
	if err := ensureDirectoryExists(config.CLI.WorkingDir); err != nil {
		return fmt.Errorf("failed to ensure CLI working directory %s exists: %w", config.CLI.WorkingDir, err)
	}

	if err := ensureDirectoryExists(config.Security.WorkingDir); err != nil {
		return fmt.Errorf("failed to ensure security working directory %s exists: %w", config.Security.WorkingDir, err)
	}

	// Validate log level
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}

	if !validLogLevels[config.LogLevel] {
		return fmt.Errorf("%w: %s (must be debug, info, warn, or error)", ErrInvalidLogLevel, config.LogLevel)
	}

	return nil
}

// applyEnvironmentOverrides applies environment variable overrides to configuration
func applyEnvironmentOverrides(config *Config) {
	if logLevel := os.Getenv("MCP_LOG_LEVEL"); logLevel != "" {
		config.LogLevel = logLevel
	}

	if cliPath := os.Getenv("GO_INVOICE_CLI_PATH"); cliPath != "" {
		config.CLI.Path = cliPath
	}

	if workingDir := os.Getenv("GO_INVOICE_HOME"); workingDir != "" {
		config.CLI.WorkingDir = workingDir
		config.Security.WorkingDir = workingDir
	}
}

// saveConfig saves configuration to file
func saveConfig(path string, config *Config) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := ensureDirectoryExists(dir); err != nil {
		return fmt.Errorf("failed to create config directory %s: %w", dir, err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("failed to write config file %s: %w", path, err)
	}

	return nil
}

// ensureDirectoryExists creates directory if it doesn't exist
func ensureDirectoryExists(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	return nil
}
