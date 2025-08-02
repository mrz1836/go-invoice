package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Logger interface defined at point of use (consumer-driven design)
type Logger interface {
	Info(msg string, fields ...any)
	Error(msg string, fields ...any)
	Debug(msg string, fields ...any)
}

// Validator interface defined at point of use
type Validator interface {
	ValidateConfig(ctx context.Context, config *Config) error
}

// ConfigService provides configuration management with dependency injection
type ConfigService struct {
	logger    Logger
	validator Validator
}

// NewConfigService creates a new ConfigService with injected dependencies
func NewConfigService(logger Logger, validator Validator) *ConfigService {
	return &ConfigService{
		logger:    logger,
		validator: validator,
	}
}

// LoadConfig loads configuration from the specified path with context support
func (s *ConfigService) LoadConfig(ctx context.Context, path string) (*Config, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	s.logger.Info("loading configuration", "path", path)

	// Load .env file if it exists
	if err := s.loadEnvFile(ctx, path); err != nil {
		return nil, fmt.Errorf("failed to load env file from %s: %w", path, err)
	}

	// Build configuration from environment variables
	config, err := s.buildConfigFromEnv(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build config from environment: %w", err)
	}

	// Set defaults
	s.setDefaults(config)

	// Validate configuration
	if s.validator != nil {
		if err := s.validator.ValidateConfig(ctx, config); err != nil {
			return nil, fmt.Errorf("config validation failed: %w", err)
		}
	}

	s.logger.Info("configuration loaded successfully", "business", config.Business.Name)
	return config, nil
}

// ValidateConfig validates a configuration object
func (s *ConfigService) ValidateConfig(ctx context.Context, config *Config) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	if s.validator != nil {
		if err := s.validator.ValidateConfig(ctx, config); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}
	}

	return nil
}

// loadEnvFile loads environment variables from the specified file
func (s *ConfigService) loadEnvFile(ctx context.Context, path string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if path == "" {
		path = ".env.config"
	}

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		s.logger.Debug("env file not found, using system environment", "path", path)
		return nil
	}

	if err := godotenv.Load(path); err != nil {
		return fmt.Errorf("failed to load environment file: %w", err)
	}

	s.logger.Debug("loaded environment file", "path", path)
	return nil
}

// buildConfigFromEnv constructs a Config object from environment variables
func (s *ConfigService) buildConfigFromEnv(ctx context.Context) (*Config, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	config := &Config{
		Business: BusinessConfig{
			Name:         getEnv("BUSINESS_NAME", ""),
			Address:      getEnv("BUSINESS_ADDRESS", ""),
			Phone:        getEnv("BUSINESS_PHONE", ""),
			Email:        getEnv("BUSINESS_EMAIL", ""),
			TaxID:        getEnv("BUSINESS_TAX_ID", ""),
			VATID:        getEnv("BUSINESS_VAT_ID", ""),
			Website:      getEnv("BUSINESS_WEBSITE", ""),
			PaymentTerms: getEnv("PAYMENT_TERMS", "Net 30"),
			BankDetails: BankDetails{
				Name:                getEnv("BANK_NAME", ""),
				AccountNumber:       getEnv("BANK_ACCOUNT", ""),
				RoutingNumber:       getEnv("BANK_ROUTING", ""),
				IBAN:                getEnv("BANK_IBAN", ""),
				SWIFT:               getEnv("BANK_SWIFT", ""),
				PaymentInstructions: getEnv("PAYMENT_INSTRUCTIONS", ""),
			},
		},
		Invoice: InvoiceConfig{
			Prefix:         getEnv("INVOICE_PREFIX", "INV"),
			StartNumber:    getEnvInt("INVOICE_START_NUMBER", 1000),
			Footer:         getEnv("INVOICE_FOOTER", ""),
			Currency:       getEnv("CURRENCY", "USD"),
			VATRate:        getEnvFloat("VAT_RATE", 0.0),
			DefaultDueDays: getEnvInt("INVOICE_DUE_DAYS", 30),
		},
		Storage: StorageConfig{
			DataDir:        getEnv("DATA_DIR", getDefaultDataDir()),
			BackupDir:      getEnv("BACKUP_DIR", ""),
			RetentionDays:  getEnvInt("RETENTION_DAYS", 365),
			AutoBackup:     getEnvBool("AUTO_BACKUP", false),
			BackupInterval: getEnvDuration("BACKUP_INTERVAL", 24*time.Hour),
		},
	}

	return config, nil
}

// setDefaults sets default values for configuration
func (s *ConfigService) setDefaults(config *Config) {
	if config.Storage.BackupDir == "" {
		config.Storage.BackupDir = filepath.Join(config.Storage.DataDir, "backups")
	}
}

// getDefaultDataDir returns the default data directory
func getDefaultDataDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ".go-invoice"
	}
	return filepath.Join(homeDir, ".go-invoice")
}

// Helper functions for environment variable parsing

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseFloat(value, 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// Simple validator implementation
type SimpleValidator struct {
	logger Logger
}

// NewSimpleValidator creates a new simple validator
func NewSimpleValidator(logger Logger) *SimpleValidator {
	return &SimpleValidator{logger: logger}
}

// ValidateConfig performs basic validation on the configuration
func (v *SimpleValidator) ValidateConfig(ctx context.Context, config *Config) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	var errors []string

	// Validate business config
	if config.Business.Name == "" {
		errors = append(errors, "business name is required")
	}
	if config.Business.Address == "" {
		errors = append(errors, "business address is required")
	}
	if config.Business.Email == "" {
		errors = append(errors, "business email is required")
	}
	if config.Business.PaymentTerms == "" {
		errors = append(errors, "payment terms are required")
	}

	// Validate invoice config
	if config.Invoice.Prefix == "" {
		errors = append(errors, "invoice prefix is required")
	}
	if config.Invoice.StartNumber < 1 {
		errors = append(errors, "invoice start number must be greater than 0")
	}
	if config.Invoice.Currency == "" {
		errors = append(errors, "currency is required")
	}
	if config.Invoice.VATRate < 0 || config.Invoice.VATRate > 1 {
		errors = append(errors, "VAT rate must be between 0 and 1")
	}

	// Validate storage config
	if config.Storage.DataDir == "" {
		errors = append(errors, "data directory is required")
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed: %s", strings.Join(errors, "; "))
	}

	v.logger.Debug("configuration validation passed")
	return nil
}
