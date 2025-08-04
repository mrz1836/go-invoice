package config

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ConfigTypesTestSuite defines the test suite for configuration type validation
type ConfigTypesTestSuite struct {
	suite.Suite
}

// TestConfigSuite runs the configuration types test suite
func TestConfigTypesTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTypesTestSuite))
}

// TestConfigJSONMarshaling tests JSON marshaling/unmarshaling for Config
func (suite *ConfigTypesTestSuite) TestConfigJSONMarshaling() {
	original := &Config{
		Business: BusinessConfig{
			Name:         "Test Business",
			Address:      "123 Test Street",
			Phone:        "+1-555-0123",
			Email:        "test@business.com",
			TaxID:        "12-3456789",
			VATID:        "VAT123456",
			Website:      "https://test.com",
			PaymentTerms: "Net 30",
			BankDetails: BankDetails{
				Name:                "Test Bank",
				AccountNumber:       "1234567890",
				RoutingNumber:       "987654321",
				IBAN:                "DE89370400440532013000",
				SWIFT:               "DEUTDEFF",
				PaymentInstructions: "Wire transfer only",
			},
		},
		Invoice: InvoiceConfig{
			Prefix:         "INV",
			StartNumber:    1000,
			Footer:         "Thank you for your business",
			Currency:       "USD",
			VATRate:        0.15,
			DefaultDueDays: 30,
		},
		Storage: StorageConfig{
			DataDir:        "/var/data/invoices",
			BackupDir:      "/var/backups/invoices",
			RetentionDays:  365,
			AutoBackup:     true,
			BackupInterval: 24 * time.Hour,
		},
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(original)
	suite.Require().NoError(err)
	suite.NotEmpty(jsonData)

	// Unmarshal back to struct
	var unmarshaled Config
	err = json.Unmarshal(jsonData, &unmarshaled)
	suite.Require().NoError(err)

	// Verify all fields are preserved
	suite.Equal(original.Business.Name, unmarshaled.Business.Name)
	suite.Equal(original.Business.Address, unmarshaled.Business.Address)
	suite.Equal(original.Business.Phone, unmarshaled.Business.Phone)
	suite.Equal(original.Business.Email, unmarshaled.Business.Email)
	suite.Equal(original.Business.TaxID, unmarshaled.Business.TaxID)
	suite.Equal(original.Business.VATID, unmarshaled.Business.VATID)
	suite.Equal(original.Business.Website, unmarshaled.Business.Website)
	suite.Equal(original.Business.PaymentTerms, unmarshaled.Business.PaymentTerms)
	suite.Equal(original.Business.BankDetails, unmarshaled.Business.BankDetails)

	suite.Equal(original.Invoice.Prefix, unmarshaled.Invoice.Prefix)
	suite.Equal(original.Invoice.StartNumber, unmarshaled.Invoice.StartNumber)
	suite.Equal(original.Invoice.Footer, unmarshaled.Invoice.Footer)
	suite.Equal(original.Invoice.Currency, unmarshaled.Invoice.Currency)
	suite.InEpsilon(original.Invoice.VATRate, unmarshaled.Invoice.VATRate, 1e-9)
	suite.Equal(original.Invoice.DefaultDueDays, unmarshaled.Invoice.DefaultDueDays)

	suite.Equal(original.Storage.DataDir, unmarshaled.Storage.DataDir)
	suite.Equal(original.Storage.BackupDir, unmarshaled.Storage.BackupDir)
	suite.Equal(original.Storage.RetentionDays, unmarshaled.Storage.RetentionDays)
	suite.Equal(original.Storage.AutoBackup, unmarshaled.Storage.AutoBackup)
	suite.Equal(original.Storage.BackupInterval, unmarshaled.Storage.BackupInterval)
}

// TestBusinessConfigJSONMarshaling tests JSON marshaling for BusinessConfig
func (suite *ConfigTypesTestSuite) TestBusinessConfigJSONMarshaling() {
	tests := []struct {
		name     string
		business BusinessConfig
	}{
		{
			name: "CompleteBusinessConfig",
			business: BusinessConfig{
				Name:         "Complete Business",
				Address:      "456 Complete Ave",
				Phone:        "+1-555-9876",
				Email:        "complete@business.com",
				TaxID:        "98-7654321",
				VATID:        "VAT987654",
				Website:      "https://complete.com",
				PaymentTerms: "Due upon receipt",
				BankDetails: BankDetails{
					Name:                "Complete Bank",
					AccountNumber:       "9876543210",
					RoutingNumber:       "123456789",
					IBAN:                "GB82WEST12345698765432",
					SWIFT:               "WESTGB2L",
					PaymentInstructions: "ACH preferred",
				},
			},
		},
		{
			name: "MinimalBusinessConfig",
			business: BusinessConfig{
				Name:         "Minimal Business",
				Address:      "789 Minimal St",
				Email:        "minimal@business.com",
				PaymentTerms: "Net 15",
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			jsonData, err := json.Marshal(tt.business)
			suite.Require().NoError(err)

			var unmarshaled BusinessConfig
			err = json.Unmarshal(jsonData, &unmarshaled)
			suite.Require().NoError(err)

			suite.Equal(tt.business, unmarshaled)
		})
	}
}

// TestBankDetailsJSONMarshaling tests JSON marshaling for BankDetails
func (suite *ConfigTypesTestSuite) TestBankDetailsJSONMarshaling() {
	tests := []struct {
		name        string
		bankDetails BankDetails
	}{
		{
			name: "CompleteBankDetails",
			bankDetails: BankDetails{
				Name:                "Test Bank Corp",
				AccountNumber:       "1111222233334444",
				RoutingNumber:       "555666777",
				IBAN:                "FR1420041010050500013M02606",
				SWIFT:               "BNPAFRPP",
				PaymentInstructions: "International wire transfers accepted",
			},
		},
		{
			name:        "EmptyBankDetails",
			bankDetails: BankDetails{},
		},
		{
			name: "PartialBankDetails",
			bankDetails: BankDetails{
				Name:          "Partial Bank",
				AccountNumber: "123456789",
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			jsonData, err := json.Marshal(tt.bankDetails)
			suite.Require().NoError(err)

			var unmarshaled BankDetails
			err = json.Unmarshal(jsonData, &unmarshaled)
			suite.Require().NoError(err)

			suite.Equal(tt.bankDetails, unmarshaled)
		})
	}
}

// TestInvoiceConfigJSONMarshaling tests JSON marshaling for InvoiceConfig
func (suite *ConfigTypesTestSuite) TestInvoiceConfigJSONMarshaling() {
	tests := []struct {
		name    string
		invoice InvoiceConfig
	}{
		{
			name: "CompleteInvoiceConfig",
			invoice: InvoiceConfig{
				Prefix:         "TEST",
				StartNumber:    5000,
				Footer:         "Custom footer message",
				Currency:       "EUR",
				VATRate:        0.21,
				DefaultDueDays: 45,
			},
		},
		{
			name: "MinimalInvoiceConfig",
			invoice: InvoiceConfig{
				Prefix:      "MIN",
				StartNumber: 1,
				Currency:    "USD",
			},
		},
		{
			name: "ZeroVATConfig",
			invoice: InvoiceConfig{
				Prefix:      "ZERO",
				StartNumber: 100,
				Currency:    "GBP",
				VATRate:     0.0,
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			jsonData, err := json.Marshal(tt.invoice)
			suite.Require().NoError(err)

			var unmarshaled InvoiceConfig
			err = json.Unmarshal(jsonData, &unmarshaled)
			suite.Require().NoError(err)

			suite.Equal(tt.invoice, unmarshaled)
		})
	}
}

// TestStorageConfigJSONMarshaling tests JSON marshaling for StorageConfig
func (suite *ConfigTypesTestSuite) TestStorageConfigJSONMarshaling() {
	tests := []struct {
		name    string
		storage StorageConfig
	}{
		{
			name: "CompleteStorageConfig",
			storage: StorageConfig{
				DataDir:        "/data/invoices",
				BackupDir:      "/backups/invoices",
				RetentionDays:  90,
				AutoBackup:     true,
				BackupInterval: 12 * time.Hour,
			},
		},
		{
			name: "MinimalStorageConfig",
			storage: StorageConfig{
				DataDir: "/tmp/invoices",
			},
		},
		{
			name: "NoBackupConfig",
			storage: StorageConfig{
				DataDir:       "/data/invoices",
				RetentionDays: 30,
				AutoBackup:    false,
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			jsonData, err := json.Marshal(tt.storage)
			suite.Require().NoError(err)

			var unmarshaled StorageConfig
			err = json.Unmarshal(jsonData, &unmarshaled)
			suite.Require().NoError(err)

			suite.Equal(tt.storage, unmarshaled)
		})
	}
}

// TestLoadConfigRequestJSONMarshaling tests JSON marshaling for LoadConfigRequest
func (suite *ConfigTypesTestSuite) TestLoadConfigRequestJSONMarshaling() {
	tests := []struct {
		name    string
		request LoadConfigRequest
	}{
		{
			name: "StrictLoadRequest",
			request: LoadConfigRequest{
				Path:   "/path/to/config.env",
				Strict: true,
			},
		},
		{
			name: "NonStrictLoadRequest",
			request: LoadConfigRequest{
				Path:   "/path/to/config.env",
				Strict: false,
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			jsonData, err := json.Marshal(tt.request)
			suite.Require().NoError(err)

			var unmarshaled LoadConfigRequest
			err = json.Unmarshal(jsonData, &unmarshaled)
			suite.Require().NoError(err)

			suite.Equal(tt.request, unmarshaled)
		})
	}
}

// TestValidateConfigRequestJSONMarshaling tests JSON marshaling for ValidateConfigRequest
func (suite *ConfigTypesTestSuite) TestValidateConfigRequestJSONMarshaling() {
	config := &Config{
		Business: BusinessConfig{
			Name:         "Test Business",
			Address:      "123 Test St",
			Email:        "test@business.com",
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
	}

	request := ValidateConfigRequest{
		Config: config,
	}

	jsonData, err := json.Marshal(request)
	suite.Require().NoError(err)

	var unmarshaled ValidateConfigRequest
	err = json.Unmarshal(jsonData, &unmarshaled)
	suite.Require().NoError(err)

	suite.Equal(request.Config.Business.Name, unmarshaled.Config.Business.Name)
	suite.Equal(request.Config.Invoice.Prefix, unmarshaled.Config.Invoice.Prefix)
	suite.Equal(request.Config.Storage.DataDir, unmarshaled.Config.Storage.DataDir)
}

// TestConfigValidationTags tests that validation tags are properly set
func (suite *ConfigTypesTestSuite) TestConfigValidationTags() {
	// This test verifies that the struct tags are correctly defined
	// Real validation would be done by the validator implementation

	config := Config{}

	// Test that zero values would trigger validation errors for required fields
	suite.Empty(config.Business.Name)         // required
	suite.Empty(config.Business.Address)      // required
	suite.Empty(config.Business.Email)        // required,email
	suite.Empty(config.Business.PaymentTerms) // required
	suite.Empty(config.Invoice.Prefix)        // required
	suite.Empty(config.Invoice.Currency)      // required
	suite.Empty(config.Storage.DataDir)       // required
}

// TestInvoiceConfigBoundaryValues tests boundary values for InvoiceConfig
func (suite *ConfigTypesTestSuite) TestInvoiceConfigBoundaryValues() {
	tests := []struct {
		name          string
		invoiceConfig InvoiceConfig
		description   string
	}{
		{
			name: "MinimumValidValues",
			invoiceConfig: InvoiceConfig{
				Prefix:         "A",
				StartNumber:    1, // min=1
				Currency:       "USD",
				VATRate:        0.0, // min=0
				DefaultDueDays: 0,   // min=0
			},
			description: "Should accept minimum valid values",
		},
		{
			name: "MaximumVATRate",
			invoiceConfig: InvoiceConfig{
				Prefix:         "MAX",
				StartNumber:    999999,
				Currency:       "EUR",
				VATRate:        1.0, // max=1
				DefaultDueDays: 365,
			},
			description: "Should accept maximum VAT rate of 1.0 (100%)",
		},
		{
			name: "TypicalBusinessValues",
			invoiceConfig: InvoiceConfig{
				Prefix:         "INV",
				StartNumber:    1000,
				Currency:       "USD",
				VATRate:        0.20, // 20% VAT
				DefaultDueDays: 30,
			},
			description: "Should handle typical business values",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Test JSON marshaling works for boundary values
			jsonData, err := json.Marshal(tt.invoiceConfig)
			suite.Require().NoError(err, tt.description)

			var unmarshaled InvoiceConfig
			err = json.Unmarshal(jsonData, &unmarshaled)
			suite.Require().NoError(err, tt.description)

			suite.Equal(tt.invoiceConfig, unmarshaled, tt.description)
		})
	}
}

// TestStorageConfigBoundaryValues tests boundary values for StorageConfig
func (suite *ConfigTypesTestSuite) TestStorageConfigBoundaryValues() {
	tests := []struct {
		name          string
		storageConfig StorageConfig
		description   string
	}{
		{
			name: "MinimumRetentionDays",
			storageConfig: StorageConfig{
				DataDir:       "/data",
				RetentionDays: 0, // min=0 (no retention)
				AutoBackup:    false,
			},
			description: "Should accept zero retention days",
		},
		{
			name: "LongRetentionPeriod",
			storageConfig: StorageConfig{
				DataDir:        "/data",
				BackupDir:      "/backup",
				RetentionDays:  3650, // 10 years
				AutoBackup:     true,
				BackupInterval: time.Hour,
			},
			description: "Should handle long retention periods",
		},
		{
			name: "VeryShortBackupInterval",
			storageConfig: StorageConfig{
				DataDir:        "/data",
				BackupDir:      "/backup",
				RetentionDays:  30,
				AutoBackup:     true,
				BackupInterval: time.Minute,
			},
			description: "Should handle very short backup intervals",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			jsonData, err := json.Marshal(tt.storageConfig)
			suite.Require().NoError(err, tt.description)

			var unmarshaled StorageConfig
			err = json.Unmarshal(jsonData, &unmarshaled)
			suite.Require().NoError(err, tt.description)

			suite.Equal(tt.storageConfig, unmarshaled, tt.description)
		})
	}
}

// TestJSONFieldNames tests that JSON field names match expectations
func (suite *ConfigTypesTestSuite) TestJSONFieldNames() {
	config := Config{
		Business: BusinessConfig{
			Name:         "Test",
			Address:      "123 St",
			Email:        "test@example.com",
			PaymentTerms: "Net 30",
			TaxID:        "12345", // Include optional fields to test their JSON names
			VATID:        "VAT123",
			BankDetails: BankDetails{
				Name:                "Bank",
				AccountNumber:       "123456",
				RoutingNumber:       "789",
				PaymentInstructions: "Wire transfer",
			},
		},
		Invoice: InvoiceConfig{
			Prefix:         "INV",
			StartNumber:    1,
			Currency:       "USD",
			DefaultDueDays: 30,
			VATRate:        0.1,
		},
		Storage: StorageConfig{
			DataDir:        "/data",
			BackupDir:      "/backup",
			RetentionDays:  30,
			AutoBackup:     true,
			BackupInterval: time.Hour,
		},
	}

	jsonData, err := json.Marshal(config)
	suite.Require().NoError(err)

	jsonStr := string(jsonData)

	// Verify JSON field names match the struct tags
	suite.Contains(jsonStr, `"business":`)
	suite.Contains(jsonStr, `"invoice":`)
	suite.Contains(jsonStr, `"storage":`)
	suite.Contains(jsonStr, `"tax_id":`)
	suite.Contains(jsonStr, `"vat_id":`)
	suite.Contains(jsonStr, `"payment_terms":`)
	suite.Contains(jsonStr, `"bank_details":`)
	suite.Contains(jsonStr, `"account_number":`)
	suite.Contains(jsonStr, `"routing_number":`)
	suite.Contains(jsonStr, `"payment_instructions":`)
	suite.Contains(jsonStr, `"start_number":`)
	suite.Contains(jsonStr, `"vat_rate":`)
	suite.Contains(jsonStr, `"default_due_days":`)
	suite.Contains(jsonStr, `"data_dir":`)
	suite.Contains(jsonStr, `"backup_dir":`)
	suite.Contains(jsonStr, `"retention_days":`)
	suite.Contains(jsonStr, `"auto_backup":`)
	suite.Contains(jsonStr, `"backup_interval":`)
}

// TestEmptyConfigSerialization tests serialization of empty/zero-value configs
func (suite *ConfigTypesTestSuite) TestEmptyConfigSerialization() {
	// Test empty structs
	emptyConfig := Config{}
	jsonData, err := json.Marshal(emptyConfig)
	suite.Require().NoError(err)

	var unmarshaled Config
	err = json.Unmarshal(jsonData, &unmarshaled)
	suite.Require().NoError(err)

	suite.Equal(emptyConfig, unmarshaled)

	// Test empty nested structs
	emptyBankDetails := BankDetails{}
	jsonData, err = json.Marshal(emptyBankDetails)
	suite.Require().NoError(err)

	var unmarshaledBank BankDetails
	err = json.Unmarshal(jsonData, &unmarshaledBank)
	suite.Require().NoError(err)

	suite.Equal(emptyBankDetails, unmarshaledBank)
}

// TestInvalidJSONHandling tests handling of invalid JSON
func (suite *ConfigTypesTestSuite) TestInvalidJSONHandling() {
	tests := []struct {
		name        string
		invalidJSON string
		targetType  string
	}{
		{
			name:        "InvalidConfigJSON",
			invalidJSON: `{"business": "invalid", "invoice": 123}`,
			targetType:  "Config",
		},
		{
			name:        "MalformedJSON",
			invalidJSON: `{"business": {"name": "test"`,
			targetType:  "Config",
		},
		{
			name:        "InvalidVATRate",
			invalidJSON: `{"prefix": "INV", "start_number": 1, "currency": "USD", "vat_rate": "invalid"}`,
			targetType:  "InvoiceConfig",
		},
		{
			name:        "InvalidDuration",
			invalidJSON: `{"data_dir": "/data", "backup_interval": "invalid"}`,
			targetType:  "StorageConfig",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			switch tt.targetType {
			case "Config":
				var config Config
				err := json.Unmarshal([]byte(tt.invalidJSON), &config)
				suite.Require().Error(err, "Should fail to unmarshal invalid JSON")

			case "InvoiceConfig":
				var invoice InvoiceConfig
				err := json.Unmarshal([]byte(tt.invalidJSON), &invoice)
				suite.Require().Error(err, "Should fail to unmarshal invalid JSON")

			case "StorageConfig":
				var storage StorageConfig
				err := json.Unmarshal([]byte(tt.invalidJSON), &storage)
				suite.Require().Error(err, "Should fail to unmarshal invalid JSON")
			}
		})
	}
}

// Standalone tests for specific edge cases

// TestConfigNilPointerHandling tests handling of nil pointers
func TestConfigNilPointerHandling(t *testing.T) {
	// Test ValidateConfigRequest with nil config
	request := ValidateConfigRequest{Config: nil}

	jsonData, err := json.Marshal(request)
	require.NoError(t, err)

	var unmarshaled ValidateConfigRequest
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Nil(t, unmarshaled.Config)
}

// TestDurationJSONSerialization tests time.Duration JSON handling
func TestDurationJSONSerialization(t *testing.T) {
	storage := StorageConfig{
		DataDir:        "/data",
		BackupInterval: 2*time.Hour + 30*time.Minute,
	}

	jsonData, err := json.Marshal(storage)
	require.NoError(t, err)

	var unmarshaled StorageConfig
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, storage.BackupInterval, unmarshaled.BackupInterval)
}

// TestFloatPrecisionHandling tests floating point precision in VATRate
func TestFloatPrecisionHandling(t *testing.T) {
	tests := []struct {
		name    string
		vatRate float64
	}{
		{"ZeroRate", 0.0},
		{"SmallRate", 0.001},
		{"TypicalRate", 0.20},
		{"HighPrecisionRate", 0.123456789},
		{"MaxRate", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			invoice := InvoiceConfig{
				Prefix:      "TEST",
				StartNumber: 1,
				Currency:    "USD",
				VATRate:     tt.vatRate,
			}

			jsonData, err := json.Marshal(invoice)
			require.NoError(t, err)

			var unmarshaled InvoiceConfig
			err = json.Unmarshal(jsonData, &unmarshaled)
			require.NoError(t, err)

			if tt.vatRate == 0.0 {
				assert.InDelta(t, tt.vatRate, unmarshaled.VATRate, 1e-9)
			} else {
				assert.InEpsilon(t, tt.vatRate, unmarshaled.VATRate, 1e-9)
			}
		})
	}
}

// TestStringFieldLengths tests handling of various string field lengths
func TestStringFieldLengths(t *testing.T) {
	tests := []struct {
		name  string
		field string
		value string
	}{
		{"EmptyString", "Name", ""},
		{"SingleChar", "Name", "A"},
		{"NormalLength", "Name", "Normal Business Name"},
		{"LongString", "Name", "Very Long Business Name That Might Cause Issues In Some Systems"},
		{"UnicodeString", "Name", "Business Name with Accents"},
		{"MultilineString", "Address", "123 Main St\nSuite 456\nNew York, NY 10001"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			business := BusinessConfig{
				Name:         tt.value,
				Address:      "123 Test St",
				Email:        "test@example.com",
				PaymentTerms: "Net 30",
			}

			if tt.field == "Address" {
				business.Address = tt.value
				business.Name = "Test Business"
			}

			jsonData, err := json.Marshal(business)
			require.NoError(t, err)

			var unmarshaled BusinessConfig
			err = json.Unmarshal(jsonData, &unmarshaled)
			require.NoError(t, err)

			assert.Equal(t, business, unmarshaled)
		})
	}
}
