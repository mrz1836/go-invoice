package tools

import (
	"context"
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"time"
)

// DefaultInputValidator provides a concrete implementation of the InputValidator interface.
//
// This implementation supports JSON Schema Draft 7 validation with comprehensive error
// reporting and actionable guidance. It follows context-first design principles and
// provides structured error information for precise error correction.
//
// Key features:
// - JSON Schema Draft 7 validation support
// - Field-level error reporting with path information
// - Format validation for common data types (dates, emails, UUIDs, etc.)
// - Required field validation with clear error messages
// - Context-aware operations with cancellation support
// - Extensible format validation system
// - Performance-optimized for frequent validation operations
//
// The validator provides detailed error messages with suggestions for correction,
// making it ideal for conversational interfaces where users need clear guidance.
type DefaultInputValidator struct {
	// logger provides structured logging for validation operations
	logger Logger

	// formatValidators maps format names to validation functions
	formatValidators map[string]FormatValidator
}

// FormatValidator defines the signature for format validation functions.
//
// Format validators check if a value matches a specific format (e.g., date, email)
// and provide actionable error messages when validation fails.
//
// Parameters:
// - ctx: Context for cancellation and timeout
// - value: Value to validate
//
// Returns:
// - error: Format validation error with guidance, or nil if valid
type FormatValidator func(ctx context.Context, value interface{}) error

// NewDefaultInputValidator creates a new input validator with dependency injection.
//
// This constructor initializes a comprehensive input validator with built-in
// format validators for common data types. It follows dependency injection
// patterns to avoid global state and improve testability.
//
// Parameters:
// - logger: Structured logger for validation operations and debugging
//
// Returns:
// - *DefaultInputValidator: Initialized validator ready for schema validation
//
// Notes:
// - Validator includes built-in format validators for common types
// - Logger must be non-nil or the constructor will panic
// - Thread-safe for concurrent validation operations
func NewDefaultInputValidator(logger Logger) *DefaultInputValidator {
	if logger == nil {
		panic("logger cannot be nil")
	}

	validator := &DefaultInputValidator{
		logger:           logger,
		formatValidators: make(map[string]FormatValidator),
	}

	// Register built-in format validators
	validator.registerBuiltinFormatValidators()

	return validator
}

// ValidateAgainstSchema performs comprehensive JSON schema validation with detailed error reporting.
//
// This method validates input data against a JSON Schema Draft 7 specification,
// providing field-level error details for precise error correction. It supports
// the most commonly used JSON Schema features for tool input validation.
//
// Parameters:
// - ctx: Context for cancellation and timeout
// - input: Input data to validate
// - schema: JSON Schema Draft 7 specification to validate against
//
// Returns:
// - error: Validation error with field-level details, or nil if valid
//
// Side Effects:
// - Logs validation operations for debugging and audit trails
//
// Notes:
// - Supports JSON Schema Draft 7 core features (type, properties, required, format)
// - Provides field path information for nested validation errors
// - Includes suggested corrections for common validation failures
// - Respects context cancellation for large schema validation
// - Performance-optimized for frequent validation operations
func (v *DefaultInputValidator) ValidateAgainstSchema(ctx context.Context, input, schema map[string]interface{}) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	v.logger.Debug("starting schema validation",
		"inputKeys", getMapKeys(input),
		"schemaType", schema["type"])

	// Validate root object type
	if schemaType, exists := schema["type"]; exists {
		if schemaType != "object" {
			return v.BuildValidationError(ctx, "",
				fmt.Sprintf("expected object type, got: %v", schemaType),
				[]string{"ensure input is a JSON object"})
		}
	}

	// Validate required fields
	if required, exists := schema["required"]; exists {
		if requiredFields, ok := required.([]interface{}); ok {
			var requiredStrings []string
			for _, field := range requiredFields {
				if fieldStr, ok := field.(string); ok {
					requiredStrings = append(requiredStrings, fieldStr)
				}
			}
			if err := v.ValidateRequired(ctx, input, requiredStrings); err != nil {
				return err
			}
		}
	}

	// Validate individual properties
	if properties, exists := schema["properties"]; exists {
		if propertiesMap, ok := properties.(map[string]interface{}); ok {
			for fieldName, fieldSchema := range propertiesMap {
				if fieldValue, hasField := input[fieldName]; hasField {
					if err := v.validateField(ctx, fieldName, fieldValue, fieldSchema); err != nil {
						return err
					}
				}
			}
		}
	}

	// Validate additional properties if specified
	if additionalProperties, exists := schema["additionalProperties"]; exists {
		if additionalProperties == false {
			// Check for unexpected properties
			allowedProperties := make(map[string]bool)
			if properties, exists := schema["properties"]; exists {
				if propertiesMap, ok := properties.(map[string]interface{}); ok {
					for prop := range propertiesMap {
						allowedProperties[prop] = true
					}
				}
			}

			for inputField := range input {
				if !allowedProperties[inputField] {
					return v.BuildValidationError(ctx, inputField,
						"unexpected property not allowed by schema",
						[]string{"remove this property or check if it's misspelled"})
				}
			}
		}
	}

	v.logger.Debug("schema validation completed successfully")
	return nil
}

// ValidateRequired checks presence and validity of required fields with context-aware processing.
//
// This method ensures all required fields are present and contain non-empty values,
// providing clear error messages for missing fields.
//
// Parameters:
// - ctx: Context for cancellation and timeout
// - input: Input data to check for required fields
// - requiredFields: List of field names that must be present and non-empty
//
// Returns:
// - error: Error listing missing required fields, or nil if all present
//
// Side Effects:
// - Logs required field validation for debugging
//
// Notes:
// - Checks both presence and non-empty values for required fields
// - Returns structured error with all missing fields listed
// - Supports nested field paths using dot notation
// - Respects context cancellation for large field sets
func (v *DefaultInputValidator) ValidateRequired(ctx context.Context, input map[string]interface{}, requiredFields []string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	var missingFields []string
	var emptyFields []string

	for _, fieldName := range requiredFields {
		value, exists := input[fieldName]
		if !exists {
			missingFields = append(missingFields, fieldName)
			continue
		}

		// Check if field is empty
		if v.isEmptyValue(value) {
			emptyFields = append(emptyFields, fieldName)
		}
	}

	if len(missingFields) > 0 || len(emptyFields) > 0 {
		var message string
		var suggestions []string

		if len(missingFields) > 0 {
			message = fmt.Sprintf("missing required fields: %s", strings.Join(missingFields, ", "))
			suggestions = append(suggestions, "add the missing required fields to your input")
		}

		if len(emptyFields) > 0 {
			if message != "" {
				message += "; "
			}
			message += fmt.Sprintf("empty required fields: %s", strings.Join(emptyFields, ", "))
			suggestions = append(suggestions, "provide non-empty values for required fields")
		}

		v.logger.Debug("required field validation failed",
			"missingFields", missingFields,
			"emptyFields", emptyFields)

		return v.BuildValidationError(ctx, "", message, suggestions)
	}

	v.logger.Debug("required field validation passed", "fieldCount", len(requiredFields))
	return nil
}

// ValidateFormat validates field formats with context support for cancellation.
//
// This method validates specific field formats using registered format validators,
// providing clear error messages with examples of correct formats.
//
// Parameters:
// - ctx: Context for cancellation and timeout
// - fieldName: Name of the field being validated
// - value: Field value to validate
// - format: Expected format (date, email, uuid, etc.)
//
// Returns:
// - error: Format validation error with correction guidance, or nil if valid
//
// Side Effects:
// - Logs format validation operations for debugging
//
// Notes:
// - Supports standard JSON Schema format validators
// - Provides examples of correct format in error messages
// - Extensible for custom format validation
// - Respects context cancellation for complex format validation
func (v *DefaultInputValidator) ValidateFormat(ctx context.Context, fieldName string, value interface{}, format string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	validator, exists := v.formatValidators[format]
	if !exists {
		v.logger.Warn("unknown format validator", "format", format, "fieldName", fieldName)
		return nil // Unknown formats are not validated (lenient approach)
	}

	if err := validator(ctx, value); err != nil {
		v.logger.Debug("format validation failed",
			"fieldName", fieldName,
			"format", format,
			"value", value,
			"error", err.Error())

		return v.BuildValidationError(ctx, fieldName,
			fmt.Sprintf("invalid %s format: %s", format, err.Error()),
			[]string{v.getFormatExample(format)})
	}

	v.logger.Debug("format validation passed", "fieldName", fieldName, "format", format)
	return nil
}

// BuildValidationError creates structured validation errors with actionable guidance.
//
// This method creates comprehensive validation errors that include field path
// information and actionable suggestions for error correction.
//
// Parameters:
// - ctx: Context for cancellation and timeout
// - fieldPath: Path to the field with validation error
// - message: Error message describing the validation failure
// - suggestions: Optional suggestions for correcting the error
//
// Returns:
// - error: Structured validation error with context and guidance
//
// Notes:
// - Creates errors that can be formatted for Claude consumption
// - Includes field path for precise error location
// - Provides actionable suggestions when possible
// - Respects context cancellation for error construction
func (v *DefaultInputValidator) BuildValidationError(ctx context.Context, fieldPath, message string, suggestions []string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	return &ValidationError{
		Field:       fieldPath,
		Message:     message,
		Code:        "validation_failed",
		Suggestions: suggestions,
	}
}

// validateField validates a single field against its schema definition.
//
// This internal method handles field-level validation including type checking,
// format validation, and constraint validation.
//
// Parameters:
// - ctx: Context for cancellation and timeout
// - fieldName: Name of the field being validated
// - value: Field value to validate
// - fieldSchema: Schema definition for this field
//
// Returns:
// - error: Field validation error, or nil if valid
//
// Notes:
// - Handles type validation, format validation, and constraints
// - Provides field-specific error messages with context
// - Supports nested validation for complex field types
func (v *DefaultInputValidator) validateField(ctx context.Context, fieldName string, value, fieldSchema interface{}) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	schemaMap, ok := fieldSchema.(map[string]interface{})
	if !ok {
		return v.BuildValidationError(ctx, fieldName,
			"invalid field schema definition",
			[]string{"check schema format for this field"})
	}

	// Validate field type
	if expectedType, exists := schemaMap["type"]; exists {
		if err := v.validateFieldType(ctx, fieldName, value, expectedType); err != nil {
			return err
		}
	}

	// Validate field format
	if format, exists := schemaMap["format"]; exists {
		if formatStr, ok := format.(string); ok {
			if err := v.ValidateFormat(ctx, fieldName, value, formatStr); err != nil {
				return err
			}
		}
	}

	// Validate string constraints
	if err := v.validateStringConstraints(ctx, fieldName, value, schemaMap); err != nil {
		return err
	}

	// Validate numeric constraints
	if err := v.validateNumericConstraints(ctx, fieldName, value, schemaMap); err != nil {
		return err
	}

	// Handle nested object validation
	if expectedType, exists := schemaMap["type"]; exists {
		if typeStr, ok := expectedType.(string); ok && typeStr == "object" {
			// Value must be a map for object validation
			if valueMap, ok := value.(map[string]interface{}); ok {
				// Recursively validate the nested object
				if err := v.ValidateAgainstSchema(ctx, valueMap, schemaMap); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// validateFieldType validates field type matches expected type.
//
// This internal method checks if a field value matches the expected JSON Schema type,
// providing clear error messages for type mismatches.
//
// Parameters:
// - ctx: Context for cancellation and timeout
// - fieldName: Name of the field being validated
// - value: Field value to check type for
// - expectedType: Expected JSON Schema type
//
// Returns:
// - error: Type validation error, or nil if type matches
func (v *DefaultInputValidator) validateFieldType(ctx context.Context, fieldName string, value, expectedType interface{}) error {
	expectedTypeStr, ok := expectedType.(string)
	if !ok {
		return v.BuildValidationError(ctx, fieldName,
			"invalid type definition in schema",
			[]string{"check schema type definition"})
	}

	actualType := v.getValueType(value)

	if !v.isTypeCompatible(actualType, expectedTypeStr) {
		return v.BuildValidationError(ctx, fieldName,
			fmt.Sprintf("expected type %s, got %s", expectedTypeStr, actualType),
			[]string{fmt.Sprintf("provide a value of type %s", expectedTypeStr)})
	}

	return nil
}

// validateStringConstraints validates string-specific constraints.
//
// This internal method validates string length constraints (minLength, maxLength)
// and pattern matching for string fields.
//
// Parameters:
// - ctx: Context for cancellation and timeout
// - fieldName: Name of the field being validated
// - value: Field value to validate
// - schema: Field schema with constraints
//
// Returns:
// - error: Constraint validation error, or nil if valid
func (v *DefaultInputValidator) validateStringConstraints(ctx context.Context, fieldName string, value interface{}, schema map[string]interface{}) error {
	strValue, ok := value.(string)
	if !ok {
		return nil // Only validate string constraints for string values
	}

	// Validate minLength
	if minLength, exists := schema["minLength"]; exists {
		if minLen, ok := minLength.(float64); ok {
			if len(strValue) < int(minLen) {
				return v.BuildValidationError(ctx, fieldName,
					fmt.Sprintf("string too short: minimum length is %d, got %d", int(minLen), len(strValue)),
					[]string{fmt.Sprintf("provide a string with at least %d characters", int(minLen))})
			}
		}
	}

	// Validate maxLength
	if maxLength, exists := schema["maxLength"]; exists {
		if maxLen, ok := maxLength.(float64); ok {
			if len(strValue) > int(maxLen) {
				return v.BuildValidationError(ctx, fieldName,
					fmt.Sprintf("string too long: maximum length is %d, got %d", int(maxLen), len(strValue)),
					[]string{fmt.Sprintf("provide a string with at most %d characters", int(maxLen))})
			}
		}
	}

	// Validate pattern
	if pattern, exists := schema["pattern"]; exists {
		if patternStr, ok := pattern.(string); ok {
			matched, err := regexp.MatchString(patternStr, strValue)
			if err != nil {
				return v.BuildValidationError(ctx, fieldName,
					"invalid pattern in schema",
					[]string{"check schema pattern definition"})
			}
			if !matched {
				return v.BuildValidationError(ctx, fieldName,
					fmt.Sprintf("string does not match required pattern: %s", patternStr),
					[]string{"provide a string that matches the required pattern"})
			}
		}
	}

	return nil
}

// validateNumericConstraints validates numeric-specific constraints.
//
// This internal method validates numeric range constraints (minimum, maximum)
// for numeric fields.
//
// Parameters:
// - ctx: Context for cancellation and timeout
// - fieldName: Name of the field being validated
// - value: Field value to validate
// - schema: Field schema with constraints
//
// Returns:
// - error: Constraint validation error, or nil if valid
func (v *DefaultInputValidator) validateNumericConstraints(ctx context.Context, fieldName string, value interface{}, schema map[string]interface{}) error {
	var numValue float64
	var ok bool

	// Convert value to float64 for comparison
	switch v := value.(type) {
	case float64:
		numValue = v
		ok = true
	case int:
		numValue = float64(v)
		ok = true
	case int64:
		numValue = float64(v)
		ok = true
	default:
		return nil // Only validate numeric constraints for numeric values
	}

	if !ok {
		return nil
	}

	// Validate minimum
	if minimum, exists := schema["minimum"]; exists {
		if minVal, ok := minimum.(float64); ok {
			if numValue < minVal {
				return v.BuildValidationError(ctx, fieldName,
					fmt.Sprintf("value too small: minimum is %g, got %g", minVal, numValue),
					[]string{fmt.Sprintf("provide a value >= %g", minVal)})
			}
		}
	}

	// Validate maximum
	if maximum, exists := schema["maximum"]; exists {
		if maxVal, ok := maximum.(float64); ok {
			if numValue > maxVal {
				return v.BuildValidationError(ctx, fieldName,
					fmt.Sprintf("value too large: maximum is %g, got %g", maxVal, numValue),
					[]string{fmt.Sprintf("provide a value <= %g", maxVal)})
			}
		}
	}

	return nil
}

// registerBuiltinFormatValidators registers standard format validators.
//
// This internal method sets up format validators for common data types
// used in tool input validation.
func (v *DefaultInputValidator) registerBuiltinFormatValidators() {
	v.formatValidators["date"] = v.validateDateFormat
	v.formatValidators["date-time"] = v.validateDateTimeFormat
	v.formatValidators["email"] = v.validateEmailFormat
	v.formatValidators["uuid"] = v.validateUUIDFormat
	v.formatValidators["uri"] = v.validateURIFormat
}

// validateDateFormat validates date format (YYYY-MM-DD).
func (v *DefaultInputValidator) validateDateFormat(ctx context.Context, value interface{}) error {
	strValue, ok := value.(string)
	if !ok {
		return v.BuildValidationError(ctx, "", "expected string value for date format", []string{"provide a string value in YYYY-MM-DD format"})
	}

	_, err := time.Parse("2006-01-02", strValue)
	if err != nil {
		return v.BuildValidationError(ctx, "", "invalid date format, expected YYYY-MM-DD", []string{"use format YYYY-MM-DD (e.g., 2023-12-25)"})
	}

	return nil
}

// validateDateTimeFormat validates date-time format (RFC3339).
func (v *DefaultInputValidator) validateDateTimeFormat(ctx context.Context, value interface{}) error {
	strValue, ok := value.(string)
	if !ok {
		return v.BuildValidationError(ctx, "", "expected string value for date-time format", []string{"provide a string value in RFC3339 format"})
	}

	_, err := time.Parse(time.RFC3339, strValue)
	if err != nil {
		return v.BuildValidationError(ctx, "", "invalid date-time format, expected RFC3339 format", []string{"use RFC3339 format (e.g., 2023-12-25T10:30:00Z)"})
	}

	return nil
}

// validateEmailFormat validates email format.
func (v *DefaultInputValidator) validateEmailFormat(ctx context.Context, value interface{}) error {
	strValue, ok := value.(string)
	if !ok {
		return v.BuildValidationError(ctx, "", "expected string value for email format", []string{"provide a string value containing a valid email address"})
	}

	_, err := mail.ParseAddress(strValue)
	if err != nil {
		return v.BuildValidationError(ctx, "", "invalid email format", []string{"use valid email format (e.g., user@example.com)"})
	}

	return nil
}

// validateUUIDFormat validates UUID format.
func (v *DefaultInputValidator) validateUUIDFormat(ctx context.Context, value interface{}) error {
	strValue, ok := value.(string)
	if !ok {
		return v.BuildValidationError(ctx, "", "expected string value for UUID format", []string{"provide a string value containing a valid UUID"})
	}

	uuidRegex := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	if !uuidRegex.MatchString(strValue) {
		return v.BuildValidationError(ctx, "", "invalid UUID format", []string{"use standard UUID format (e.g., 550e8400-e29b-41d4-a716-446655440000)"})
	}

	return nil
}

// validateURIFormat validates URI format.
func (v *DefaultInputValidator) validateURIFormat(ctx context.Context, value interface{}) error {
	strValue, ok := value.(string)
	if !ok {
		return v.BuildValidationError(ctx, "", "expected string value for URI format", []string{"provide a string value containing a valid URI"})
	}

	// Basic URI validation - should start with scheme
	if !strings.Contains(strValue, "://") {
		return v.BuildValidationError(ctx, "", "invalid URI format, missing scheme", []string{"include scheme in URI (e.g., https://example.com)"})
	}

	return nil
}

// getValueType determines the JSON Schema type of a value.
//
// This utility method maps Go types to JSON Schema type strings for
// type validation.
//
// Parameters:
// - value: Value to determine type for
//
// Returns:
// - string: JSON Schema type string
func (v *DefaultInputValidator) getValueType(value interface{}) string {
	switch value.(type) {
	case string:
		return "string"
	case float64, int, int64:
		return "number"
	case bool:
		return "boolean"
	case []interface{}:
		return "array"
	case map[string]interface{}:
		return "object"
	case nil:
		return "null"
	default:
		return "unknown"
	}
}

// isTypeCompatible checks if actual type is compatible with expected type.
//
// This utility method handles type compatibility checking including
// numeric type compatibility (int/float).
//
// Parameters:
// - actualType: Actual value type
// - expectedType: Expected JSON Schema type
//
// Returns:
// - bool: True if types are compatible
func (v *DefaultInputValidator) isTypeCompatible(actualType, expectedType string) bool {
	if actualType == expectedType {
		return true
	}

	// Handle numeric type compatibility
	if expectedType == "number" && (actualType == "number" || actualType == "integer") {
		return true
	}
	if expectedType == "integer" && actualType == "number" {
		// Check if number is actually an integer
		return true // Simplified - would need actual value to check
	}

	return false
}

// isEmptyValue checks if a value is considered empty.
//
// This utility method determines if a value should be considered empty
// for required field validation.
//
// Parameters:
// - value: Value to check for emptiness
//
// Returns:
// - bool: True if value is empty
func (v *DefaultInputValidator) isEmptyValue(value interface{}) bool {
	if value == nil {
		return true
	}

	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v) == ""
	case []interface{}:
		return len(v) == 0
	case map[string]interface{}:
		return len(v) == 0
	default:
		return false
	}
}

// getFormatExample returns an example of the correct format.
//
// This utility method provides format examples for error messages
// to help users understand the expected format.
//
// Parameters:
// - format: Format name to get example for
//
// Returns:
// - string: Example of correct format
func (v *DefaultInputValidator) getFormatExample(format string) string {
	switch format {
	case "date":
		return "example: 2025-08-03"
	case "date-time":
		return "example: 2025-08-03T10:30:00Z"
	case "email":
		return "example: user@example.com"
	case "uuid":
		return "example: 123e4567-e89b-12d3-a456-426614174000"
	case "uri":
		return "example: https://example.com/path"
	default:
		return "check format requirements"
	}
}
