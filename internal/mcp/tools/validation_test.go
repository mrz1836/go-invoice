package tools

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// ValidationTestSuite provides comprehensive tests for the input validation system
type ValidationTestSuite struct {
	suite.Suite

	validator *DefaultInputValidator
	logger    *MockLogger
	// No context stored in struct - pass through method parameters instead
}

func (s *ValidationTestSuite) SetupTest() {
	s.logger = new(MockLogger)
	s.validator = NewDefaultInputValidator(s.logger)
	// Context created as needed in individual test methods
}

func (s *ValidationTestSuite) TearDownTest() {
	s.logger.AssertExpectations(s.T())
}

func (s *ValidationTestSuite) TestNewDefaultInputValidator() {
	s.Run("ValidCreation", func() {
		logger := new(MockLogger)
		validator := NewDefaultInputValidator(logger)
		s.NotNil(validator, "Validator should be created")
		s.Equal(logger, validator.logger, "Logger should be assigned")
		s.NotNil(validator.formatValidators, "Format validators should be initialized")
	})

	s.Run("NilLoggerPanic", func() {
		s.Panics(func() {
			NewDefaultInputValidator(nil)
		}, "Should panic with nil logger")
	})

	s.Run("BuiltinFormatValidators", func() {
		logger := new(MockLogger)
		validator := NewDefaultInputValidator(logger)

		expectedFormats := []string{"date", "date-time", "email", "uuid", "uri"}
		for _, format := range expectedFormats {
			s.Contains(validator.formatValidators, format, "Should have %s format validator", format)
		}
	})
}

func (s *ValidationTestSuite) TestValidateAgainstSchema_BasicValidation() {
	tests := []struct {
		name        string
		input       map[string]interface{}
		schema      map[string]interface{}
		expectError bool
		errorMsg    string
	}{
		{
			name: "ValidObjectWithRequiredFields",
			input: map[string]interface{}{
				"name":  "Test Name",
				"email": "test@example.com",
			},
			schema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type": "string",
					},
					"email": map[string]interface{}{
						"type": "string",
					},
				},
				"required": []interface{}{"name", "email"},
			},
			expectError: false,
		},
		{
			name: "MissingRequiredField",
			input: map[string]interface{}{
				"name": "Test Name",
			},
			schema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type": "string",
					},
					"email": map[string]interface{}{
						"type": "string",
					},
				},
				"required": []interface{}{"name", "email"},
			},
			expectError: true,
			errorMsg:    "missing required fields",
		},
		{
			name: "InvalidSchemaType",
			input: map[string]interface{}{
				"name": "Test",
			},
			schema: map[string]interface{}{
				"type": "array",
			},
			expectError: true,
			errorMsg:    "expected object type",
		},
		{
			name: "AdditionalPropertiesNotAllowed",
			input: map[string]interface{}{
				"name":       "Test Name",
				"unexpected": "value",
			},
			schema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type": "string",
					},
				},
				"additionalProperties": false,
			},
			expectError: true,
			errorMsg:    "unexpected property not allowed",
		},
		{
			name:  "EmptyInputValidSchema",
			input: map[string]interface{}{},
			schema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"optional": map[string]interface{}{
						"type": "string",
					},
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// Setup logger expectations
			s.logger.On("Debug", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()

			ctx := context.Background()
			err := s.validator.ValidateAgainstSchema(ctx, tt.input, tt.schema)

			if tt.expectError {
				s.Require().Error(err, "Should return validation error")
				if tt.errorMsg != "" {
					s.Contains(err.Error(), tt.errorMsg, "Error message should contain expected text")
				}
			} else {
				s.NoError(err, "Should pass validation")
			}
		})
	}
}

func (s *ValidationTestSuite) TestValidateRequired() {
	ctx := context.Background()
	tests := []struct {
		name           string
		input          map[string]interface{}
		requiredFields []string
		expectError    bool
		errorMsg       string
	}{
		{
			name: "AllRequiredFieldsPresent",
			input: map[string]interface{}{
				"name":   "John Doe",
				"email":  "john@example.com",
				"amount": 100.0,
			},
			requiredFields: []string{"name", "email", "amount"},
			expectError:    false,
		},
		{
			name: "MissingOneRequiredField",
			input: map[string]interface{}{
				"name":  "John Doe",
				"email": "john@example.com",
			},
			requiredFields: []string{"name", "email", "amount"},
			expectError:    true,
			errorMsg:       "missing required fields: amount",
		},
		{
			name: "EmptyRequiredField",
			input: map[string]interface{}{
				"name":  "",
				"email": "john@example.com",
			},
			requiredFields: []string{"name", "email"},
			expectError:    true,
			errorMsg:       "empty required fields: name",
		},
		{
			name: "NilRequiredField",
			input: map[string]interface{}{
				"name":  nil,
				"email": "john@example.com",
			},
			requiredFields: []string{"name", "email"},
			expectError:    true,
			errorMsg:       "empty required fields: name",
		},
		{
			name: "EmptyArrayField",
			input: map[string]interface{}{
				"items": []interface{}{},
				"name":  "test",
			},
			requiredFields: []string{"items", "name"},
			expectError:    true,
			errorMsg:       "empty required fields: items",
		},
		{
			name: "EmptyObjectField",
			input: map[string]interface{}{
				"config": map[string]interface{}{},
				"name":   "test",
			},
			requiredFields: []string{"config", "name"},
			expectError:    true,
			errorMsg:       "empty required fields: config",
		},
		{
			name: "WhitespaceOnlyString",
			input: map[string]interface{}{
				"name": "   \t\n  ",
			},
			requiredFields: []string{"name"},
			expectError:    true,
			errorMsg:       "empty required fields: name",
		},
		{
			name:           "NoRequiredFields",
			input:          map[string]interface{}{"optional": "value"},
			requiredFields: []string{},
			expectError:    false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// Setup logger expectations
			s.logger.On("Debug", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()

			err := s.validator.ValidateRequired(ctx, tt.input, tt.requiredFields)

			if tt.expectError {
				s.Require().Error(err, "Should return validation error")
				if tt.errorMsg != "" {
					s.Contains(err.Error(), tt.errorMsg, "Error message should contain expected text")
				}
			} else {
				s.NoError(err, "Should pass validation")
			}
		})
	}
}

func (s *ValidationTestSuite) TestValidateFormat() {
	ctx := context.Background()
	tests := []struct {
		name        string
		fieldName   string
		value       interface{}
		format      string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "ValidDate",
			fieldName:   "date_field",
			value:       "2025-08-03",
			format:      "date",
			expectError: false,
		},
		{
			name:        "InvalidDate",
			fieldName:   "date_field",
			value:       "2025-13-01",
			format:      "date",
			expectError: true,
			errorMsg:    "invalid date format",
		},
		{
			name:        "ValidDateTime",
			fieldName:   "timestamp",
			value:       "2025-08-03T10:30:00Z",
			format:      "date-time",
			expectError: false,
		},
		{
			name:        "InvalidDateTime",
			fieldName:   "timestamp",
			value:       "2025-08-03 10:30:00",
			format:      "date-time",
			expectError: true,
			errorMsg:    "invalid date-time format",
		},
		{
			name:        "ValidEmail",
			fieldName:   "email",
			value:       "user@example.com",
			format:      "email",
			expectError: false,
		},
		{
			name:        "InvalidEmail",
			fieldName:   "email",
			value:       "not-an-email",
			format:      "email",
			expectError: true,
			errorMsg:    "invalid email format",
		},
		{
			name:        "ValidUUID",
			fieldName:   "id",
			value:       "123e4567-e89b-12d3-a456-426614174000",
			format:      "uuid",
			expectError: false,
		},
		{
			name:        "InvalidUUID",
			fieldName:   "id",
			value:       "not-a-uuid",
			format:      "uuid",
			expectError: true,
			errorMsg:    "invalid UUID format",
		},
		{
			name:        "ValidURI",
			fieldName:   "url",
			value:       "https://example.com/path",
			format:      "uri",
			expectError: false,
		},
		{
			name:        "InvalidURI",
			fieldName:   "url",
			value:       "not-a-uri",
			format:      "uri",
			expectError: true,
			errorMsg:    "invalid URI format",
		},
		{
			name:        "UnknownFormat",
			fieldName:   "field",
			value:       "value",
			format:      "unknown",
			expectError: false, // Unknown formats are not validated (lenient approach)
		},
		{
			name:        "NonStringValueForStringFormat",
			fieldName:   "field",
			value:       123,
			format:      "email",
			expectError: true,
			errorMsg:    "expected string value",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// Setup logger expectations
			s.logger.On("Debug", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()
			s.logger.On("Warn", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()

			err := s.validator.ValidateFormat(ctx, tt.fieldName, tt.value, tt.format)

			if tt.expectError {
				s.Require().Error(err, "Should return validation error")
				if tt.errorMsg != "" {
					s.Contains(err.Error(), tt.errorMsg, "Error message should contain expected text")
				}
			} else {
				s.NoError(err, "Should pass validation")
			}
		})
	}
}

func (s *ValidationTestSuite) TestValidateField() {
	ctx := context.Background()
	tests := []struct {
		name        string
		fieldName   string
		value       interface{}
		fieldSchema interface{}
		expectError bool
		errorMsg    string
	}{
		{
			name:      "ValidStringField",
			fieldName: "name",
			value:     "John Doe",
			fieldSchema: map[string]interface{}{
				"type": "string",
			},
			expectError: false,
		},
		{
			name:      "InvalidStringFieldType",
			fieldName: "name",
			value:     123,
			fieldSchema: map[string]interface{}{
				"type": "string",
			},
			expectError: true,
			errorMsg:    "expected type string, got number",
		},
		{
			name:      "StringTooShort",
			fieldName: "password",
			value:     "123",
			fieldSchema: map[string]interface{}{
				"type":      "string",
				"minLength": 8.0,
			},
			expectError: true,
			errorMsg:    "string too short",
		},
		{
			name:      "StringTooLong",
			fieldName: "title",
			value:     "This is a very long title that exceeds the maximum length",
			fieldSchema: map[string]interface{}{
				"type":      "string",
				"maxLength": 20.0,
			},
			expectError: true,
			errorMsg:    "string too long",
		},
		{
			name:      "StringPatternMatch",
			fieldName: "code",
			value:     "ABC123",
			fieldSchema: map[string]interface{}{
				"type":    "string",
				"pattern": "^[A-Z]{3}[0-9]{3}$",
			},
			expectError: false,
		},
		{
			name:      "StringPatternNoMatch",
			fieldName: "code",
			value:     "abc123",
			fieldSchema: map[string]interface{}{
				"type":    "string",
				"pattern": "^[A-Z]{3}[0-9]{3}$",
			},
			expectError: true,
			errorMsg:    "does not match required pattern",
		},
		{
			name:      "ValidNumberField",
			fieldName: "amount",
			value:     100.5,
			fieldSchema: map[string]interface{}{
				"type": "number",
			},
			expectError: false,
		},
		{
			name:      "NumberTooSmall",
			fieldName: "amount",
			value:     -5.0,
			fieldSchema: map[string]interface{}{
				"type":    "number",
				"minimum": 0.0,
			},
			expectError: true,
			errorMsg:    "value too small",
		},
		{
			name:      "NumberTooLarge",
			fieldName: "amount",
			value:     1000.0,
			fieldSchema: map[string]interface{}{
				"type":    "number",
				"maximum": 500.0,
			},
			expectError: true,
			errorMsg:    "value too large",
		},
		{
			name:      "ValidBooleanField",
			fieldName: "active",
			value:     true,
			fieldSchema: map[string]interface{}{
				"type": "boolean",
			},
			expectError: false,
		},
		{
			name:        "InvalidFieldSchemaType",
			fieldName:   "field",
			value:       "value",
			fieldSchema: "not_a_map",
			expectError: true,
			errorMsg:    "invalid field schema definition",
		},
		{
			name:      "ValidEmailFormat",
			fieldName: "email",
			value:     "user@example.com",
			fieldSchema: map[string]interface{}{
				"type":   "string",
				"format": "email",
			},
			expectError: false,
		},
		{
			name:      "InvalidEmailFormat",
			fieldName: "email",
			value:     "not-an-email",
			fieldSchema: map[string]interface{}{
				"type":   "string",
				"format": "email",
			},
			expectError: true,
			errorMsg:    "invalid email format",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// Setup logger expectations
			s.logger.On("Debug", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()

			err := s.validator.validateField(ctx, tt.fieldName, tt.value, tt.fieldSchema)

			if tt.expectError {
				s.Require().Error(err, "Should return validation error")
				if tt.errorMsg != "" {
					s.Contains(err.Error(), tt.errorMsg, "Error message should contain expected text")
				}
			} else {
				s.NoError(err, "Should pass validation")
			}
		})
	}
}

func (s *ValidationTestSuite) TestBuildValidationError() {
	ctx := context.Background()
	tests := []struct {
		name        string
		fieldPath   string
		message     string
		suggestions []string
		expectedErr *ValidationError
	}{
		{
			name:        "CompleteError",
			fieldPath:   "client_name",
			message:     "field is required",
			suggestions: []string{"provide client name", "use client_id instead"},
			expectedErr: &ValidationError{
				Field:       "client_name",
				Message:     "field is required",
				Code:        "validation_failed",
				Suggestions: []string{"provide client name", "use client_id instead"},
			},
		},
		{
			name:        "ErrorWithoutSuggestions",
			fieldPath:   "amount",
			message:     "must be positive",
			suggestions: nil,
			expectedErr: &ValidationError{
				Field:       "amount",
				Message:     "must be positive",
				Code:        "validation_failed",
				Suggestions: nil,
			},
		},
		{
			name:        "ErrorWithoutField",
			fieldPath:   "",
			message:     "general validation error",
			suggestions: []string{"check input"},
			expectedErr: &ValidationError{
				Field:       "",
				Message:     "general validation error",
				Code:        "validation_failed",
				Suggestions: []string{"check input"},
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			err := s.validator.BuildValidationError(ctx, tt.fieldPath, tt.message, tt.suggestions)

			s.Require().Error(err, "Should return an error")

			var validationErr *ValidationError
			s.Require().ErrorAs(err, &validationErr, "Should return ValidationError type")
			s.Equal(tt.expectedErr.Field, validationErr.Field)
			s.Equal(tt.expectedErr.Message, validationErr.Message)
			s.Equal(tt.expectedErr.Code, validationErr.Code)
			s.Equal(tt.expectedErr.Suggestions, validationErr.Suggestions)
		})
	}
}

func (s *ValidationTestSuite) TestContextCancellation() {
	ctx := context.Background()
	tests := []struct {
		name     string
		testFunc func(context.Context) error
	}{
		{
			name: "ValidateAgainstSchemaCancellation",
			testFunc: func(ctx context.Context) error {
				return s.validator.ValidateAgainstSchema(ctx, map[string]interface{}{}, map[string]interface{}{})
			},
		},
		{
			name: "ValidateRequiredCancellation",
			testFunc: func(ctx context.Context) error {
				return s.validator.ValidateRequired(ctx, map[string]interface{}{}, []string{})
			},
		},
		{
			name: "ValidateFormatCancellation",
			testFunc: func(ctx context.Context) error {
				return s.validator.ValidateFormat(ctx, "field", "value", "unknown")
			},
		},
		{
			name: "BuildValidationErrorCancellation",
			testFunc: func(ctx context.Context) error {
				return s.validator.BuildValidationError(ctx, "field", "message", nil)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			cancelCtx, cancel := context.WithCancel(ctx)
			cancel() // Cancel immediately

			err := tt.testFunc(cancelCtx)
			s.Require().Error(err, "Should return context cancellation error")
			s.Equal(context.Canceled, err, "Should be context.Canceled error")
		})
	}
}

func (s *ValidationTestSuite) TestConcurrentValidation() {
	ctx := context.Background()
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type": "string",
			},
		},
		"required": []interface{}{"name"},
	}

	// Setup logger expectations for concurrent access
	s.logger.On("Debug", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()

	done := make(chan bool, 100)
	for i := 0; i < 100; i++ {
		go func(_ int) {
			defer func() { done <- true }()

			input := map[string]interface{}{
				"name": "concurrent_test",
			}

			err := s.validator.ValidateAgainstSchema(ctx, input, schema)
			s.NoError(err, "Concurrent validation should succeed")
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 100; i++ {
		<-done
	}
}

func (s *ValidationTestSuite) TestFormatValidatorEdgeCases() {
	ctx := context.Background()
	tests := []struct {
		name        string
		format      string
		value       interface{}
		expectError bool
	}{
		{
			name:        "DateWithTime",
			format:      "date",
			value:       "2025-08-03T10:30:00Z",
			expectError: true,
		},
		{
			name:        "DateTimeWithoutTimeZone",
			format:      "date-time",
			value:       "2025-08-03T10:30:00",
			expectError: true,
		},
		{
			name:        "EmailWithDisplayName",
			format:      "email",
			value:       "John Doe <john@example.com>",
			expectError: false,
		},
		{
			name:        "UUIDUppercase",
			format:      "uuid",
			value:       "123E4567-E89B-12D3-A456-426614174000",
			expectError: false,
		},
		{
			name:        "UUIDWithoutHyphens",
			format:      "uuid",
			value:       "123e4567e89b12d3a456426614174000",
			expectError: true,
		},
		{
			name:        "URIWithQuery",
			format:      "uri",
			value:       "https://example.com/path?param=value",
			expectError: false,
		},
		{
			name:        "URIRelative",
			format:      "uri",
			value:       "/relative/path",
			expectError: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.logger.On("Debug", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()

			err := s.validator.ValidateFormat(ctx, "test_field", tt.value, tt.format)

			if tt.expectError {
				s.Error(err, "Should return validation error for %s format with value %v", tt.format, tt.value)
			} else {
				s.NoError(err, "Should pass validation for %s format with value %v", tt.format, tt.value)
			}
		})
	}
}

func (s *ValidationTestSuite) TestComplexSchemaValidation() {
	ctx := context.Background()
	complexSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"invoice": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"number": map[string]interface{}{
						"type":    "string",
						"pattern": "^INV-[0-9]{3,6}$",
					},
					"amount": map[string]interface{}{
						"type":    "number",
						"minimum": 0.01,
						"maximum": 1000000.0,
					},
					"due_date": map[string]interface{}{
						"type":   "string",
						"format": "date",
					},
				},
				"required": []interface{}{"number", "amount"},
			},
			"client": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"email": map[string]interface{}{
						"type":   "string",
						"format": "email",
					},
					"name": map[string]interface{}{
						"type":      "string",
						"minLength": 1.0,
						"maxLength": 100.0,
					},
				},
				"required": []interface{}{"email", "name"},
			},
		},
		"required": []interface{}{"invoice", "client"},
	}

	tests := []struct {
		name        string
		input       map[string]interface{}
		expectError bool
		description string
	}{
		{
			name: "ValidComplexInput",
			input: map[string]interface{}{
				"invoice": map[string]interface{}{
					"number":   "INV-12345",
					"amount":   250.75,
					"due_date": "2025-09-01",
				},
				"client": map[string]interface{}{
					"email": "client@example.com",
					"name":  "Acme Corporation",
				},
			},
			expectError: false,
			description: "Valid complex nested object",
		},
		{
			name: "InvalidInvoiceNumber",
			input: map[string]interface{}{
				"invoice": map[string]interface{}{
					"number": "INVALID",
					"amount": 250.75,
				},
				"client": map[string]interface{}{
					"email": "client@example.com",
					"name":  "Acme Corporation",
				},
			},
			expectError: true,
			description: "Invalid invoice number pattern",
		},
		{
			name: "MissingNestedRequired",
			input: map[string]interface{}{
				"invoice": map[string]interface{}{
					"amount": 250.75,
				},
				"client": map[string]interface{}{
					"email": "client@example.com",
					"name":  "Acme Corporation",
				},
			},
			expectError: true,
			description: "Missing required nested field",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.logger.On("Debug", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()

			err := s.validator.ValidateAgainstSchema(ctx, tt.input, complexSchema)

			if tt.expectError {
				s.Error(err, tt.description)
			} else {
				s.NoError(err, tt.description)
			}
		})
	}
}

// TestValidationTestSuite runs the complete validation test suite
func TestValidationTestSuite(t *testing.T) {
	suite.Run(t, new(ValidationTestSuite))
}

// Benchmark tests for performance validation
func BenchmarkDefaultInputValidator_ValidateAgainstSchema(b *testing.B) {
	logger := new(MockLogger)
	logger.On("Debug", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()

	validator := NewDefaultInputValidator(logger)
	ctx := context.Background()

	input := map[string]interface{}{
		"name":   "Test Name",
		"email":  "test@example.com",
		"amount": 100.0,
	}

	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type": "string",
			},
			"email": map[string]interface{}{
				"type":   "string",
				"format": "email",
			},
			"amount": map[string]interface{}{
				"type":    "number",
				"minimum": 0.0,
			},
		},
		"required": []interface{}{"name", "email", "amount"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.ValidateAgainstSchema(ctx, input, schema)
	}
}

func BenchmarkDefaultInputValidator_ValidateFormat(b *testing.B) {
	logger := new(MockLogger)
	logger.On("Debug", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()

	validator := NewDefaultInputValidator(logger)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.ValidateFormat(ctx, "email", "test@example.com", "email")
	}
}

// Unit tests for specific edge cases
func TestDefaultInputValidator_EdgeCases(t *testing.T) {
	logger := new(MockLogger)
	logger.On("Debug", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()
	logger.On("Warn", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()

	validator := NewDefaultInputValidator(logger)
	ctx := context.Background()

	t.Run("EmptyInput", func(t *testing.T) {
		err := validator.ValidateAgainstSchema(ctx, map[string]interface{}{}, map[string]interface{}{
			"type": "object",
		})
		assert.NoError(t, err, "Empty input should be valid for object schema without required fields")
	})

	t.Run("NilSchema", func(t *testing.T) {
		err := validator.ValidateAgainstSchema(ctx, map[string]interface{}{}, nil)
		assert.NoError(t, err, "Should handle nil schema gracefully")
	})

	t.Run("SchemaWithoutType", func(t *testing.T) {
		err := validator.ValidateAgainstSchema(ctx, map[string]interface{}{}, map[string]interface{}{
			"properties": map[string]interface{}{},
		})
		assert.NoError(t, err, "Should handle schema without type field")
	})

	t.Run("InvalidPatternInSchema", func(t *testing.T) {
		schema := map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"field": map[string]interface{}{
					"type":    "string",
					"pattern": "[",
				},
			},
		}
		input := map[string]interface{}{
			"field": "test",
		}
		err := validator.ValidateAgainstSchema(ctx, input, schema)
		assert.Error(t, err, "Should return error for invalid regex pattern")
	})
}

func TestFormatValidators_EdgeCases(t *testing.T) {
	logger := new(MockLogger)
	logger.On("Debug", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()

	validator := NewDefaultInputValidator(logger)
	ctx := context.Background()

	t.Run("LeapYearDate", func(t *testing.T) {
		err := validator.ValidateFormat(ctx, "date", "2024-02-29", "date")
		assert.NoError(t, err, "Should handle leap year dates")
	})

	t.Run("InvalidLeapYearDate", func(t *testing.T) {
		err := validator.ValidateFormat(ctx, "date", "2023-02-29", "date")
		assert.Error(t, err, "Should reject invalid leap year dates")
	})

	t.Run("DateTimeWithNanoseconds", func(t *testing.T) {
		err := validator.ValidateFormat(ctx, "datetime", "2025-08-03T10:30:00.123456789Z", "date-time")
		assert.NoError(t, err, "Should handle datetime with nanoseconds")
	})

	t.Run("EmailWithSpecialCharacters", func(t *testing.T) {
		err := validator.ValidateFormat(ctx, "email", "test+tag@example.co.uk", "email")
		assert.NoError(t, err, "Should handle email with special characters")
	})

	t.Run("UUIDAllZeros", func(t *testing.T) {
		err := validator.ValidateFormat(ctx, "uuid", "00000000-0000-0000-0000-000000000000", "uuid")
		assert.NoError(t, err, "Should handle UUID with all zeros")
	})

	t.Run("URIWithFragment", func(t *testing.T) {
		err := validator.ValidateFormat(ctx, "uri", "https://example.com/path#fragment", "uri")
		assert.NoError(t, err, "Should handle URI with fragment")
	})
}

// Race condition tests
func TestValidation_RaceConditions(t *testing.T) {
	logger := new(MockLogger)
	logger.On("Debug", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()

	validator := NewDefaultInputValidator(logger)
	ctx := context.Background()

	t.Run("ConcurrentFormatValidation", func(t *testing.T) {
		done := make(chan bool, 100)
		for i := 0; i < 100; i++ {
			go func(_ int) {
				defer func() { done <- true }()
				err := validator.ValidateFormat(ctx, "email", "test@example.com", "email")
				assert.NoError(t, err)
			}(i)
		}

		for i := 0; i < 100; i++ {
			<-done
		}
	})

	t.Run("ConcurrentSchemaValidation", func(t *testing.T) {
		schema := map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type": "string",
				},
			},
		}

		done := make(chan bool, 50)
		for i := 0; i < 50; i++ {
			go func(_ int) {
				defer func() { done <- true }()
				input := map[string]interface{}{
					"name": "concurrent_test",
				}
				err := validator.ValidateAgainstSchema(ctx, input, schema)
				assert.NoError(t, err)
			}(i)
		}

		for i := 0; i < 50; i++ {
			<-done
		}
	})
}
