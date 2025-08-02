package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// ConfigTestSuite defines the test suite for configuration functionality
type ConfigTestSuite struct {
	suite.Suite

	logger    *TestLogger
	validator *SimpleValidator
	service   *ConfigService
	tempDir   string
}

// TestLogger provides a test implementation of the Logger interface
type TestLogger struct {
	messages []string
}

func (l *TestLogger) Info(msg string, _ ...any)  { l.messages = append(l.messages, msg) }
func (l *TestLogger) Error(msg string, _ ...any) { l.messages = append(l.messages, msg) }
func (l *TestLogger) Debug(msg string, _ ...any) { l.messages = append(l.messages, msg) }

// SetupSuite runs once before all tests in the suite
func (suite *ConfigTestSuite) SetupSuite() {
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "go-invoice-test-*")
	suite.Require().NoError(err)
	suite.tempDir = tempDir
}

// TearDownSuite runs once after all tests in the suite
func (suite *ConfigTestSuite) TearDownSuite() {
	if err := os.RemoveAll(suite.tempDir); err != nil {
		// Log the error but don't fail the test since this is cleanup
		suite.T().Logf("Warning: failed to remove temp directory %s: %v", suite.tempDir, err)
	}
}

// SetupTest runs before each test
func (suite *ConfigTestSuite) SetupTest() {
	suite.logger = &TestLogger{}
	suite.validator = NewSimpleValidator(suite.logger)
	suite.service = NewConfigService(suite.logger, suite.validator)
}

// TestNewConfigService tests the constructor
func (suite *ConfigTestSuite) TestNewConfigService() {
	service := NewConfigService(suite.logger, suite.validator)
	suite.NotNil(service)
}

// TestLoadConfigFromEnv tests loading configuration from environment variables
func (suite *ConfigTestSuite) TestLoadConfigFromEnv() {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected func(*Config) bool
		wantErr  bool
	}{
		{
			name: "ValidMinimalConfig",
			envVars: map[string]string{
				"BUSINESS_NAME":    "Test Business",
				"BUSINESS_ADDRESS": "123 Test St",
				"BUSINESS_EMAIL":   "test@example.com",
				"PAYMENT_TERMS":    "Net 30",
			},
			expected: func(c *Config) bool {
				return c.Business.Name == "Test Business" &&
					c.Business.Email == "test@example.com" &&
					c.Invoice.Currency == "USD" &&
					c.Invoice.Prefix == "INV"
			},
			wantErr: false,
		},
		{
			name: "ValidCompleteConfig",
			envVars: map[string]string{
				"BUSINESS_NAME":        "Complete Business",
				"BUSINESS_ADDRESS":     "456 Complete Ave",
				"BUSINESS_EMAIL":       "complete@example.com",
				"BUSINESS_PHONE":       "+1-555-0123",
				"BUSINESS_TAX_ID":      "12-3456789",
				"PAYMENT_TERMS":        "Due upon receipt",
				"INVOICE_PREFIX":       "CB",
				"INVOICE_START_NUMBER": "5000",
				"CURRENCY":             "EUR",
				"VAT_RATE":             "0.20",
			},
			expected: func(c *Config) bool {
				return c.Business.Name == "Complete Business" &&
					c.Business.Phone == "+1-555-0123" &&
					c.Business.TaxID == "12-3456789" &&
					c.Invoice.Prefix == "CB" &&
					c.Invoice.StartNumber == 5000 &&
					c.Invoice.Currency == "EUR" &&
					c.Invoice.VATRate == 0.20
			},
			wantErr: false,
		},
		{
			name: "MissingRequiredFields",
			envVars: map[string]string{
				"BUSINESS_NAME": "Incomplete Business",
				// Missing required fields
			},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Clear environment
			suite.clearTestEnv()

			// Set test environment variables
			for key, value := range tt.envVars {
				suite.Require().NoError(os.Setenv(key, value))
			}

			ctx := context.Background()
			config, err := suite.service.LoadConfig(ctx, "nonexistent.env")

			if tt.wantErr {
				suite.Require().Error(err)
				suite.Nil(config)
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(config)
				if tt.expected != nil {
					suite.True(tt.expected(config), "config validation failed")
				}
			}

			// Clean up environment
			suite.clearTestEnv()
		})
	}
}

// TestLoadConfigFromFile tests loading configuration from .env file
func (suite *ConfigTestSuite) TestLoadConfigFromFile() {
	envContent := `# Test configuration
BUSINESS_NAME=File Business
BUSINESS_ADDRESS=789 File Road
BUSINESS_EMAIL=file@example.com
PAYMENT_TERMS=Net 15
INVOICE_PREFIX=FB
CURRENCY=GBP
`

	envFile := filepath.Join(suite.tempDir, "test.env")
	err := os.WriteFile(envFile, []byte(envContent), 0o600)
	suite.Require().NoError(err)

	ctx := context.Background()
	config, err := suite.service.LoadConfig(ctx, envFile)

	suite.Require().NoError(err)
	suite.Equal("File Business", config.Business.Name)
	suite.Equal("file@example.com", config.Business.Email)
	suite.Equal("FB", config.Invoice.Prefix)
	suite.Equal("GBP", config.Invoice.Currency)
}

// TestLoadConfigContextCancellation tests context cancellation
func (suite *ConfigTestSuite) TestLoadConfigContextCancellation() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	config, err := suite.service.LoadConfig(ctx, "test.env")

	suite.Require().Error(err)
	suite.Equal(context.Canceled, err)
	suite.Nil(config)
}

// TestValidateConfig tests configuration validation
func (suite *ConfigTestSuite) TestValidateConfig() {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "ValidConfig",
			config: &Config{
				Business: BusinessConfig{
					Name:         "Valid Business",
					Address:      "123 Valid St",
					Email:        "valid@example.com",
					PaymentTerms: "Net 30",
				},
				Invoice: InvoiceConfig{
					Prefix:      "VB",
					StartNumber: 1000,
					Currency:    "USD",
					VATRate:     0.10,
				},
				Storage: StorageConfig{
					DataDir: "/tmp/test",
				},
			},
			wantErr: false,
		},
		{
			name:    "NilConfig",
			config:  nil,
			wantErr: true,
		},
		{
			name: "MissingBusinessName",
			config: &Config{
				Business: BusinessConfig{
					Address:      "123 Valid St",
					Email:        "valid@example.com",
					PaymentTerms: "Net 30",
				},
				Invoice: InvoiceConfig{
					Prefix:      "VB",
					StartNumber: 1000,
					Currency:    "USD",
				},
				Storage: StorageConfig{
					DataDir: "/tmp/test",
				},
			},
			wantErr: true,
		},
		{
			name: "InvalidVATRate",
			config: &Config{
				Business: BusinessConfig{
					Name:         "Valid Business",
					Address:      "123 Valid St",
					Email:        "valid@example.com",
					PaymentTerms: "Net 30",
				},
				Invoice: InvoiceConfig{
					Prefix:      "VB",
					StartNumber: 1000,
					Currency:    "USD",
					VATRate:     1.5, // Invalid rate > 1
				},
				Storage: StorageConfig{
					DataDir: "/tmp/test",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			ctx := context.Background()
			err := suite.service.ValidateConfig(ctx, tt.config)

			if tt.wantErr {
				suite.Error(err)
			} else {
				suite.NoError(err)
			}
		})
	}
}

// TestEnvHelperFunctions tests the environment variable helper functions
func (suite *ConfigTestSuite) TestEnvHelperFunctions() {
	suite.Run("getEnvInt", func() {
		suite.Require().NoError(os.Setenv("TEST_INT", "42"))
		defer func() { suite.Require().NoError(os.Unsetenv("TEST_INT")) }()

		result := getEnvInt("TEST_INT", 10)
		suite.Equal(42, result)

		result = getEnvInt("NONEXISTENT_INT", 10)
		suite.Equal(10, result)
	})

	suite.Run("getEnvFloat", func() {
		suite.Require().NoError(os.Setenv("TEST_FLOAT", "3.14"))
		defer func() { suite.Require().NoError(os.Unsetenv("TEST_FLOAT")) }()

		result := getEnvFloat("TEST_FLOAT", 1.0)
		suite.InEpsilon(3.14, result, 1e-9)

		result = getEnvFloat("NONEXISTENT_FLOAT", 1.0)
		suite.InEpsilon(1.0, result, 1e-9)
	})

	suite.Run("getEnvBool", func() {
		suite.Require().NoError(os.Setenv("TEST_BOOL", "true"))
		defer func() { suite.Require().NoError(os.Unsetenv("TEST_BOOL")) }()

		result := getEnvBool("TEST_BOOL", false)
		suite.True(result)

		result = getEnvBool("NONEXISTENT_BOOL", false)
		suite.False(result)
	})

	suite.Run("getEnvDuration", func() {
		suite.Require().NoError(os.Setenv("TEST_DURATION", "5m"))
		defer func() {
			suite.Require().NoError(os.Unsetenv("TEST_DURATION"))
		}()

		result := getEnvDuration("TEST_DURATION", time.Hour)
		suite.Equal(5*time.Minute, result)

		result = getEnvDuration("NONEXISTENT_DURATION", time.Hour)
		suite.Equal(time.Hour, result)
	})
}

// TestDefaultDataDir tests the default data directory logic
func (suite *ConfigTestSuite) TestDefaultDataDir() {
	defaultDir := getDefaultDataDir()
	suite.NotEmpty(defaultDir)
	suite.Contains(defaultDir, ".go-invoice")
}

// clearTestEnv clears test environment variables
func (suite *ConfigTestSuite) clearTestEnv() {
	testEnvVars := []string{
		"BUSINESS_NAME", "BUSINESS_ADDRESS", "BUSINESS_EMAIL", "BUSINESS_PHONE",
		"BUSINESS_TAX_ID", "BUSINESS_VAT_ID", "BUSINESS_WEBSITE", "PAYMENT_TERMS",
		"BANK_NAME", "BANK_ACCOUNT", "BANK_ROUTING", "BANK_IBAN", "BANK_SWIFT",
		"PAYMENT_INSTRUCTIONS", "INVOICE_PREFIX", "INVOICE_START_NUMBER",
		"INVOICE_FOOTER", "CURRENCY", "VAT_RATE", "INVOICE_DUE_DAYS",
		"DATA_DIR", "BACKUP_DIR", "RETENTION_DAYS", "AUTO_BACKUP", "BACKUP_INTERVAL",
	}

	for _, envVar := range testEnvVars {
		_ = os.Unsetenv(envVar) // Cleanup, ignore error
	}
}

// TestConfigSuite runs the configuration test suite
func TestConfigSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}

// TestSimpleValidator tests the validator independently
func TestSimpleValidator(t *testing.T) {
	logger := &TestLogger{}
	validator := NewSimpleValidator(logger)

	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "ValidConfiguration",
			config: &Config{
				Business: BusinessConfig{
					Name:         "Test Business",
					Address:      "123 Test St",
					Email:        "test@example.com",
					PaymentTerms: "Net 30",
				},
				Invoice: InvoiceConfig{
					Prefix:      "TEST",
					StartNumber: 1,
					Currency:    "USD",
					VATRate:     0.0,
				},
				Storage: StorageConfig{
					DataDir: "/tmp/test",
				},
			},
			wantErr: false,
		},
		{
			name: "EmptyBusinessName",
			config: &Config{
				Business: BusinessConfig{
					Address:      "123 Test St",
					Email:        "test@example.com",
					PaymentTerms: "Net 30",
				},
				Invoice: InvoiceConfig{
					Prefix:      "TEST",
					StartNumber: 1,
					Currency:    "USD",
				},
				Storage: StorageConfig{
					DataDir: "/tmp/test",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := validator.ValidateConfig(ctx, tt.config)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
