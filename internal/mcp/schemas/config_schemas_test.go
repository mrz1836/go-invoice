package schemas

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ConfigSchemasTestSuite provides comprehensive tests for configuration schema definitions
type ConfigSchemasTestSuite struct {
	suite.Suite
}

func (s *ConfigSchemasTestSuite) TestConfigShowSchema() {
	schema := ConfigShowSchema()

	s.Run("BasicStructure", func() {
		s.NotNil(schema, "Schema should not be nil")
		s.Equal("object", schema["type"], "Schema type should be object")

		// Should have properties
		properties, hasProperties := schema["properties"]
		s.True(hasProperties, "Schema should have properties")
		s.IsType(map[string]interface{}{}, properties, "Properties should be a map")
	})

	s.Run("SectionField", func() {
		properties, _ := schema["properties"].(map[string]interface{})
		sectionField, hasSection := properties["section"]
		s.True(hasSection, "Should have section field")

		if sectionMap, ok := sectionField.(map[string]interface{}); ok {
			s.Equal("string", sectionMap["type"], "Section should be string type")
			s.Equal("all", sectionMap["default"], "Section should default to 'all'")

			// Should have valid enum values
			if enum, hasEnum := sectionMap["enum"]; hasEnum {
				if enumSlice, ok := enum.([]string); ok {
					expectedSections := []string{"all", "invoice_settings", "database", "email", "templates", "security", "integrations"}
					for _, expectedSection := range expectedSections {
						s.Contains(enumSlice, expectedSection, "Should contain section: %s", expectedSection)
					}
				}
			}
		}
	})

	s.Run("OutputFormatField", func() {
		properties, _ := schema["properties"].(map[string]interface{})
		formatField, hasFormat := properties["output_format"]
		s.True(hasFormat, "Should have output_format field")

		if formatMap, ok := formatField.(map[string]interface{}); ok {
			s.Equal("string", formatMap["type"], "Output format should be string type")
			s.Equal("yaml", formatMap["default"], "Output format should default to 'yaml'")

			// Should have valid format options
			if enum, hasEnum := formatMap["enum"]; hasEnum {
				if enumSlice, ok := enum.([]string); ok {
					expectedFormats := []string{"yaml", "json", "table", "text"}
					for _, expectedFormat := range expectedFormats {
						s.Contains(enumSlice, expectedFormat, "Should contain format: %s", expectedFormat)
					}
				}
			}
		}
	})

	s.Run("BooleanFields", func() {
		properties, _ := schema["properties"].(map[string]interface{})
		booleanFields := []string{
			"include_defaults", "show_descriptions", "exclude_sensitive",
			"include_metadata", "include_sources", "show_overrides",
			"show_validation", "summary_only", "verbose", "api_format", "return_data",
		}

		for _, fieldName := range booleanFields {
			if field, hasField := properties[fieldName]; hasField {
				if fieldMap, ok := field.(map[string]interface{}); ok {
					s.Equal("boolean", fieldMap["type"], "Field %s should be boolean type", fieldName)
					s.Contains(fieldMap, "description", "Field %s should have description", fieldName)
				}
			}
		}
	})

	s.Run("EnvironmentField", func() {
		properties, _ := schema["properties"].(map[string]interface{})
		envField, hasEnv := properties["environment"]
		s.True(hasEnv, "Should have environment field")

		if envMap, ok := envField.(map[string]interface{}); ok {
			s.Equal("string", envMap["type"], "Environment should be string type")

			// Should have valid environment options
			if enum, hasEnum := envMap["enum"]; hasEnum {
				if enumSlice, ok := enum.([]string); ok {
					expectedEnvs := []string{"development", "staging", "production", "test"}
					for _, expectedEnv := range expectedEnvs {
						s.Contains(enumSlice, expectedEnv, "Should contain environment: %s", expectedEnv)
					}
				}
			}
		}
	})

	s.Run("OutputPathField", func() {
		properties, _ := schema["properties"].(map[string]interface{})
		pathField, hasPath := properties["output_path"]
		s.True(hasPath, "Should have output_path field")

		if pathMap, ok := pathField.(map[string]interface{}); ok {
			s.Equal("string", pathMap["type"], "Output path should be string type")
			s.Contains(pathMap, "description", "Output path should have description")

			// Should have examples
			if examples, hasExamples := pathMap["examples"]; hasExamples {
				s.IsType([]string{}, examples, "Examples should be string slice")
			}
		}
	})

	s.Run("NoAdditionalProperties", func() {
		additionalProps, hasAdditional := schema["additionalProperties"]
		s.True(hasAdditional, "Should specify additionalProperties")
		s.False(additionalProps.(bool), "Should not allow additional properties")
	})
}

func (s *ConfigSchemasTestSuite) TestConfigValidateSchema() {
	schema := ConfigValidateSchema()

	s.Run("BasicStructure", func() {
		s.NotNil(schema, "Schema should not be nil")
		s.Equal("object", schema["type"], "Schema type should be object")
		s.Contains(schema, "properties", "Schema should have properties")
	})

	s.Run("ValidationFeatures", func() {
		properties, _ := schema["properties"].(map[string]interface{})
		validationFields := []string{
			"comprehensive", "check_dependencies", "test_connections",
			"security_check", "performance_check", "compliance_check",
			"deployment_check", "fail_fast", "include_suggestions",
		}

		for _, fieldName := range validationFields {
			if field, hasField := properties[fieldName]; hasField {
				if fieldMap, ok := field.(map[string]interface{}); ok {
					s.Equal("boolean", fieldMap["type"], "Field %s should be boolean type", fieldName)
					s.Contains(fieldMap, "description", "Field %s should have description", fieldName)
				}
			}
		}
	})

	s.Run("AutoFixFeatures", func() {
		properties, _ := schema["properties"].(map[string]interface{})

		// Test auto_fix_safe field
		if autoFixField, hasAutoFix := properties["auto_fix_safe"]; hasAutoFix {
			if autoFixMap, ok := autoFixField.(map[string]interface{}); ok {
				s.Equal("boolean", autoFixMap["type"], "Auto fix should be boolean")
				s.Equal(false, autoFixMap["default"], "Auto fix should default to false")
			}
		}

		// Test backup_before_fix field
		if backupField, hasBackup := properties["backup_before_fix"]; hasBackup {
			if backupMap, ok := backupField.(map[string]interface{}); ok {
				s.Equal("boolean", backupMap["type"], "Backup should be boolean")
				s.Equal(true, backupMap["default"], "Backup should default to true")
			}
		}
	})

	s.Run("OutputFormatOptions", func() {
		properties, _ := schema["properties"].(map[string]interface{})
		formatField, hasFormat := properties["output_format"]
		s.True(hasFormat, "Should have output_format field")

		if formatMap, ok := formatField.(map[string]interface{}); ok {
			s.Equal("string", formatMap["type"], "Output format should be string type")
			s.Equal("detailed", formatMap["default"], "Output format should default to 'detailed'")

			// Should have valid format options
			if enum, hasEnum := formatMap["enum"]; hasEnum {
				if enumSlice, ok := enum.([]string); ok {
					expectedFormats := []string{"detailed", "summary", "json", "report", "interactive"}
					for _, expectedFormat := range expectedFormats {
						s.Contains(enumSlice, expectedFormat, "Should contain format: %s", expectedFormat)
					}
				}
			}
		}
	})

	s.Run("EnvironmentValidation", func() {
		properties, _ := schema["properties"].(map[string]interface{})
		envField, hasEnv := properties["environment"]
		s.True(hasEnv, "Should have environment field")

		if envMap, ok := envField.(map[string]interface{}); ok {
			s.Equal("string", envMap["type"], "Environment should be string type")

			if enum, hasEnum := envMap["enum"]; hasEnum {
				if enumSlice, ok := enum.([]string); ok {
					expectedEnvs := []string{"development", "staging", "production", "test"}
					s.Len(enumSlice, len(expectedEnvs), "Should have expected number of environments")
				}
			}
		}
	})
}

func (s *ConfigSchemasTestSuite) TestConfigInitSchema() {
	schema := ConfigInitSchema()

	s.Run("BasicStructure", func() {
		s.NotNil(schema, "Schema should not be nil")
		s.Equal("object", schema["type"], "Schema type should be object")
		s.Contains(schema, "properties", "Schema should have properties")
	})

	s.Run("TemplateField", func() {
		properties, _ := schema["properties"].(map[string]interface{})
		templateField, hasTemplate := properties["template"]
		s.True(hasTemplate, "Should have template field")

		if templateMap, ok := templateField.(map[string]interface{}); ok {
			s.Equal("string", templateMap["type"], "Template should be string type")
			s.Equal("default", templateMap["default"], "Template should default to 'default'")

			// Should have valid template options
			if enum, hasEnum := templateMap["enum"]; hasEnum {
				if enumSlice, ok := enum.([]string); ok {
					expectedTemplates := []string{"default", "minimal", "production", "enterprise", "custom"}
					for _, expectedTemplate := range expectedTemplates {
						s.Contains(enumSlice, expectedTemplate, "Should contain template: %s", expectedTemplate)
					}
				}
			}
		}
	})

	s.Run("CompanyConfigFields", func() {
		properties, _ := schema["properties"].(map[string]interface{})

		// Test company_name field
		if companyField, hasCompany := properties["company_name"]; hasCompany {
			if companyMap, ok := companyField.(map[string]interface{}); ok {
				s.Equal("string", companyMap["type"], "Company name should be string type")
				s.Contains(companyMap, "maxLength", "Company name should have max length")
				s.Contains(companyMap, "examples", "Company name should have examples")
			}
		}

		// Test default_tax_rate field
		if taxField, hasTax := properties["default_tax_rate"]; hasTax {
			if taxMap, ok := taxField.(map[string]interface{}); ok {
				s.Equal("number", taxMap["type"], "Tax rate should be number type")
				s.Contains(taxMap, "minimum", "Tax rate should have minimum")
				s.Contains(taxMap, "maximum", "Tax rate should have maximum")
				s.Equal(0, taxMap["minimum"], "Tax rate minimum should be 0")
				s.Equal(100, taxMap["maximum"], "Tax rate maximum should be 100")
			}
		}

		// Test currency field
		if currencyField, hasCurrency := properties["currency"]; hasCurrency {
			if currencyMap, ok := currencyField.(map[string]interface{}); ok {
				s.Equal("string", currencyMap["type"], "Currency should be string type")
				s.Equal("USD", currencyMap["default"], "Currency should default to USD")

				if enum, hasEnum := currencyMap["enum"]; hasEnum {
					if enumSlice, ok := enum.([]string); ok {
						expectedCurrencies := []string{"USD", "EUR", "GBP", "CAD", "AUD"}
						for _, expectedCurrency := range expectedCurrencies {
							s.Contains(enumSlice, expectedCurrency, "Should contain currency: %s", expectedCurrency)
						}
					}
				}
			}
		}
	})

	s.Run("CustomValuesField", func() {
		properties, _ := schema["properties"].(map[string]interface{})
		customField, hasCustom := properties["custom_values"]
		s.True(hasCustom, "Should have custom_values field")

		if customMap, ok := customField.(map[string]interface{}); ok {
			s.Equal("object", customMap["type"], "Custom values should be object type")
			s.Equal(true, customMap["additionalProperties"], "Custom values should allow additional properties")
			s.Contains(customMap, "examples", "Custom values should have examples")
		}
	})

	s.Run("IntegrationsField", func() {
		properties, _ := schema["properties"].(map[string]interface{})
		integrationsField, hasIntegrations := properties["enable_integrations"]
		s.True(hasIntegrations, "Should have enable_integrations field")

		if integrationsMap, ok := integrationsField.(map[string]interface{}); ok {
			s.Equal("array", integrationsMap["type"], "Integrations should be array type")
			s.Equal(true, integrationsMap["uniqueItems"], "Integrations should have unique items")

			// Check items schema
			if items, hasItems := integrationsMap["items"]; hasItems {
				if itemsMap, ok := items.(map[string]interface{}); ok {
					s.Equal("string", itemsMap["type"], "Integration items should be strings")

					if enum, hasEnum := itemsMap["enum"]; hasEnum {
						if enumSlice, ok := enum.([]string); ok {
							expectedIntegrations := []string{"quickbooks", "stripe", "email", "slack", "webhooks"}
							for _, expectedIntegration := range expectedIntegrations {
								s.Contains(enumSlice, expectedIntegration, "Should contain integration: %s", expectedIntegration)
							}
						}
					}
				}
			}
		}
	})

	s.Run("FilePathFields", func() {
		properties, _ := schema["properties"].(map[string]interface{})
		pathFields := []string{"config_path", "config_file", "import_from"}

		for _, fieldName := range pathFields {
			if field, hasField := properties[fieldName]; hasField {
				if fieldMap, ok := field.(map[string]interface{}); ok {
					s.Equal("string", fieldMap["type"], "Field %s should be string type", fieldName)
					s.Contains(fieldMap, "description", "Field %s should have description", fieldName)
					s.Contains(fieldMap, "examples", "Field %s should have examples", fieldName)
				}
			}
		}
	})

	s.Run("BackupAndValidationFields", func() {
		properties, _ := schema["properties"].(map[string]interface{})
		backupAndValidationFields := map[string]bool{
			"backup_existing":  true,  // should default to true
			"backup_original":  true,  // should default to true
			"validate_after":   true,  // should default to true
			"validate_custom":  true,  // should default to true
			"test_connections": false, // should default to false
			"setup_webhooks":   false, // should default to false
			"skip_optional":    false, // should default to false
			"auto_confirm":     false, // should default to false
			"save_template":    false, // should default to false
		}

		for fieldName, expectedDefault := range backupAndValidationFields {
			if field, hasField := properties[fieldName]; hasField {
				if fieldMap, ok := field.(map[string]interface{}); ok {
					s.Equal("boolean", fieldMap["type"], "Field %s should be boolean type", fieldName)
					s.Equal(expectedDefault, fieldMap["default"], "Field %s should default to %v", fieldName, expectedDefault)
				}
			}
		}
	})
}

func (s *ConfigSchemasTestSuite) TestGetAllConfigSchemas() {
	schemas := GetAllConfigSchemas()

	s.Run("AllSchemasPresent", func() {
		expectedSchemas := []string{"config_show", "config_validate", "config_init"}
		s.Len(schemas, len(expectedSchemas), "Should have expected number of schemas")

		for _, expectedSchema := range expectedSchemas {
			schema, exists := schemas[expectedSchema]
			s.True(exists, "Should have schema: %s", expectedSchema)
			s.NotNil(schema, "Schema %s should not be nil", expectedSchema)
			s.Equal("object", schema["type"], "Schema %s should be object type", expectedSchema)
		}
	})

	s.Run("SchemaConsistency", func() {
		// All schemas should have consistent structure
		for schemaName, schema := range schemas {
			s.Contains(schema, "type", "Schema %s should have type", schemaName)
			s.Contains(schema, "properties", "Schema %s should have properties", schemaName)
			s.Contains(schema, "additionalProperties", "Schema %s should specify additionalProperties", schemaName)
			s.Equal(false, schema["additionalProperties"], "Schema %s should not allow additional properties", schemaName)
		}
	})

	s.Run("SchemaImmutability", func() {
		// Test that calling the function multiple times returns consistent results
		schemas1 := GetAllConfigSchemas()
		schemas2 := GetAllConfigSchemas()

		s.Len(schemas1, len(schemas2), "Schema count should be consistent")

		for schemaName := range schemas1 {
			s.Contains(schemas2, schemaName, "Schema %s should exist in both calls", schemaName)

			schema1 := schemas1[schemaName]
			schema2 := schemas2[schemaName]
			s.Equal(schema1["type"], schema2["type"], "Schema %s type should be consistent", schemaName)
		}
	})
}

func (s *ConfigSchemasTestSuite) TestGetConfigToolSchema() {
	s.Run("ValidSchemas", func() {
		validSchemas := []string{"config_show", "config_validate", "config_init"}

		for _, schemaName := range validSchemas {
			schema, exists := GetConfigToolSchema(schemaName)
			s.True(exists, "Schema %s should exist", schemaName)
			s.NotNil(schema, "Schema %s should not be nil", schemaName)
			s.Equal("object", schema["type"], "Schema %s should be object type", schemaName)
		}
	})

	s.Run("InvalidSchema", func() {
		schema, exists := GetConfigToolSchema("nonexistent_schema")
		s.False(exists, "Nonexistent schema should not exist")
		s.Nil(schema, "Nonexistent schema should return nil")
	})

	s.Run("EmptySchemaName", func() {
		schema, exists := GetConfigToolSchema("")
		s.False(exists, "Empty schema name should not exist")
		s.Nil(schema, "Empty schema name should return nil")
	})

	s.Run("CaseSensitivity", func() {
		schema, exists := GetConfigToolSchema("CONFIG_SHOW")
		s.False(exists, "Schema names should be case sensitive")
		s.Nil(schema, "Case mismatch should return nil")
	})
}

func (s *ConfigSchemasTestSuite) TestSchemaValidation() {
	schemas := map[string]func() map[string]interface{}{
		"ConfigShow":     ConfigShowSchema,
		"ConfigValidate": ConfigValidateSchema,
		"ConfigInit":     ConfigInitSchema,
	}

	for schemaName, schemaFunc := range schemas {
		s.Run(schemaName+"Validation", func() {
			schema := schemaFunc()
			s.validateJSONSchema(schema, schemaName)
		})
	}
}

func (s *ConfigSchemasTestSuite) TestSchemaConsistency() {
	s.Run("SectionFieldConsistency", func() {
		// All schemas should have consistent section field definitions
		showSchema := ConfigShowSchema()
		validateSchema := ConfigValidateSchema()

		showProps, _ := showSchema["properties"].(map[string]interface{})
		validateProps, _ := validateSchema["properties"].(map[string]interface{})

		if showSection, hasShow := showProps["section"]; hasShow {
			if validateSection, hasValidate := validateProps["section"]; hasValidate {
				showSectionMap := showSection.(map[string]interface{})
				validateSectionMap := validateSection.(map[string]interface{})

				s.Equal(showSectionMap["type"], validateSectionMap["type"], "Section type should be consistent")
				s.Equal(showSectionMap["default"], validateSectionMap["default"], "Section default should be consistent")
				s.Equal(showSectionMap["enum"], validateSectionMap["enum"], "Section enum should be consistent")
			}
		}
	})

	s.Run("EnvironmentFieldConsistency", func() {
		// Environment fields should be consistent across schemas
		schemas := []map[string]interface{}{
			ConfigShowSchema(),
			ConfigValidateSchema(),
			ConfigInitSchema(),
		}

		var firstEnumValues []string
		for i, schema := range schemas {
			props, _ := schema["properties"].(map[string]interface{})
			if envField, hasEnv := props["environment"]; hasEnv {
				if envMap, ok := envField.(map[string]interface{}); ok {
					if enum, hasEnum := envMap["enum"]; hasEnum {
						if enumSlice, ok := enum.([]string); ok {
							if i == 0 {
								firstEnumValues = enumSlice
							} else {
								s.Equal(firstEnumValues, enumSlice, "Environment enum should be consistent across schemas")
							}
						}
					}
				}
			}
		}
	})

	s.Run("BooleanDefaultConsistency", func() {
		// Similar boolean fields should have consistent defaults where logical
		initSchema := ConfigInitSchema()
		props, _ := initSchema["properties"].(map[string]interface{})

		// Backup-related fields should default to true for safety
		backupFields := []string{"backup_existing", "backup_original"}
		for _, fieldName := range backupFields {
			if field, hasField := props[fieldName]; hasField {
				if fieldMap, ok := field.(map[string]interface{}); ok {
					s.Equal(true, fieldMap["default"], "Backup field %s should default to true for safety", fieldName)
				}
			}
		}

		// Validation fields should default to true for thoroughness
		validationFields := []string{"validate_after", "validate_custom"}
		for _, fieldName := range validationFields {
			if field, hasField := props[fieldName]; hasField {
				if fieldMap, ok := field.(map[string]interface{}); ok {
					s.Equal(true, fieldMap["default"], "Validation field %s should default to true", fieldName)
				}
			}
		}
	})
}

func (s *ConfigSchemasTestSuite) TestSchemaEdgeCases() {
	s.Run("EmptySchemas", func() {
		// Ensure schemas are not empty
		schemas := []func() map[string]interface{}{
			ConfigShowSchema,
			ConfigValidateSchema,
			ConfigInitSchema,
		}

		for _, schemaFunc := range schemas {
			schema := schemaFunc()
			s.NotEmpty(schema, "Schema should not be empty")
			s.Contains(schema, "type", "Schema should have type field")
			s.Contains(schema, "properties", "Schema should have properties field")
		}
	})

	s.Run("PropertyCountValidation", func() {
		// Schemas should have reasonable number of properties
		schemas := map[string]func() map[string]interface{}{
			"show":     ConfigShowSchema,
			"validate": ConfigValidateSchema,
			"init":     ConfigInitSchema,
		}

		for schemaName, schemaFunc := range schemas {
			schema := schemaFunc()
			props, _ := schema["properties"].(map[string]interface{})
			s.Greater(len(props), 5, "Schema %s should have more than 5 properties", schemaName)
			s.Less(len(props), 50, "Schema %s should have fewer than 50 properties", schemaName)
		}
	})

	s.Run("DescriptionPresence", func() {
		// All properties should have descriptions
		schemas := []map[string]interface{}{
			ConfigShowSchema(),
			ConfigValidateSchema(),
			ConfigInitSchema(),
		}

		for i, schema := range schemas {
			props, _ := schema["properties"].(map[string]interface{})
			for propName, propDef := range props {
				if propMap, ok := propDef.(map[string]interface{}); ok {
					s.Contains(propMap, "description", "Property %s in schema %d should have description", propName, i)
					if desc, hasDesc := propMap["description"]; hasDesc {
						s.NotEmpty(desc, "Description for property %s should not be empty", propName)
					}
				}
			}
		}
	})
}

// Helper method to validate JSON Schema structure
func (s *ConfigSchemasTestSuite) validateJSONSchema(schema map[string]interface{}, schemaName string) {
	// Basic JSON Schema requirements
	s.Contains(schema, "type", "%s should have type field", schemaName)
	s.Equal("object", schema["type"], "%s should be object type", schemaName)

	// Properties validation
	if properties, hasProps := schema["properties"]; hasProps {
		s.IsType(map[string]interface{}{}, properties, "%s properties should be map", schemaName)

		if propsMap, ok := properties.(map[string]interface{}); ok {
			for fieldName, fieldDef := range propsMap {
				s.NotEmpty(fieldName, "Field name should not be empty")
				s.IsType(map[string]interface{}{}, fieldDef, "Field %s should have definition map", fieldName)

				if fieldMap, ok := fieldDef.(map[string]interface{}); ok {
					// Each field should have a type
					if fieldType, hasType := fieldMap["type"]; hasType {
						validTypes := []string{"string", "number", "boolean", "array", "object", "null"}
						s.Contains(validTypes, fieldType, "Field %s should have valid type", fieldName)
					}

					// Each field should have a description
					s.Contains(fieldMap, "description", "Field %s should have description", fieldName)
				}
			}
		}
	}

	// Additional properties should be explicitly set
	s.Contains(schema, "additionalProperties", "%s should specify additionalProperties", schemaName)
}

// TestConfigSchemasTestSuite runs the complete config schemas test suite
func TestConfigSchemasTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigSchemasTestSuite))
}

// Benchmark tests for schema creation
func BenchmarkConfigShowSchema(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		schema := ConfigShowSchema()
		_ = schema
	}
}

func BenchmarkConfigValidateSchema(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		schema := ConfigValidateSchema()
		_ = schema
	}
}

func BenchmarkConfigInitSchema(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		schema := ConfigInitSchema()
		_ = schema
	}
}

func BenchmarkGetAllConfigSchemas(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		schemas := GetAllConfigSchemas()
		_ = schemas
	}
}

// Unit tests for specific schema behaviors
func TestConfigSchemas_Specific(t *testing.T) {
	t.Run("AllSchemasReturnValidMaps", func(t *testing.T) {
		schemas := []func() map[string]interface{}{
			ConfigShowSchema,
			ConfigValidateSchema,
			ConfigInitSchema,
		}

		for _, schemaFunc := range schemas {
			schema := schemaFunc()
			assert.NotNil(t, schema)
			assert.IsType(t, map[string]interface{}{}, schema)
			assert.Contains(t, schema, "type")
		}
	})

	t.Run("GetConfigToolSchemaEdgeCases", func(t *testing.T) {
		// Test with nil string
		schema, exists := GetConfigToolSchema("")
		assert.False(t, exists)
		assert.Nil(t, schema)

		// Test with unknown schema
		schema, exists = GetConfigToolSchema("unknown_config_tool")
		assert.False(t, exists)
		assert.Nil(t, schema)

		// Test with valid schema
		schema, exists = GetConfigToolSchema("config_show")
		assert.True(t, exists)
		assert.NotNil(t, schema)
	})

	t.Run("SchemaFieldValidation", func(t *testing.T) {
		// Test specific field validations
		initSchema := ConfigInitSchema()
		props, _ := initSchema["properties"].(map[string]interface{})

		// Tax rate field should have proper bounds
		if taxField, hasTax := props["default_tax_rate"]; hasTax {
			if taxMap, ok := taxField.(map[string]interface{}); ok {
				assert.Equal(t, "number", taxMap["type"])
				assert.Equal(t, 0, taxMap["minimum"])
				assert.Equal(t, 100, taxMap["maximum"])
			}
		}

		// Company name should have max length
		if companyField, hasCompany := props["company_name"]; hasCompany {
			if companyMap, ok := companyField.(map[string]interface{}); ok {
				assert.Equal(t, "string", companyMap["type"])
				assert.Contains(t, companyMap, "maxLength")
			}
		}
	})
}

// Edge case tests
func TestConfigSchemas_EdgeCases(t *testing.T) {
	t.Run("SchemasNotNil", func(t *testing.T) {
		assert.NotNil(t, ConfigShowSchema())
		assert.NotNil(t, ConfigValidateSchema())
		assert.NotNil(t, ConfigInitSchema())
	})

	t.Run("SchemasHaveMinimalStructure", func(t *testing.T) {
		schemas := []map[string]interface{}{
			ConfigShowSchema(),
			ConfigValidateSchema(),
			ConfigInitSchema(),
		}

		for i, schema := range schemas {
			assert.Contains(t, schema, "type", "Schema %d should have type", i)
			assert.Equal(t, "object", schema["type"], "Schema %d should be object type", i)
			assert.Contains(t, schema, "properties", "Schema %d should have properties", i)
		}
	})

	t.Run("EnumValidation", func(t *testing.T) {
		// Test that enum values are consistent and valid
		schemas := []map[string]interface{}{
			ConfigShowSchema(),
			ConfigValidateSchema(),
		}

		for _, schema := range schemas {
			props, _ := schema["properties"].(map[string]interface{})
			if sectionField, hasSection := props["section"]; hasSection {
				if sectionMap, ok := sectionField.(map[string]interface{}); ok {
					if enum, hasEnum := sectionMap["enum"]; hasEnum {
						if enumSlice, ok := enum.([]string); ok {
							require.NotEmpty(t, enumSlice, "Section enum should not be empty")
							assert.Contains(t, enumSlice, "all", "Section enum should contain 'all'")
						}
					}
				}
			}
		}
	})

	t.Run("DefaultValueTypes", func(t *testing.T) {
		// Test that default values match their field types
		initSchema := ConfigInitSchema()
		props, _ := initSchema["properties"].(map[string]interface{})

		for fieldName, fieldDef := range props {
			if fieldMap, ok := fieldDef.(map[string]interface{}); ok {
				if defaultVal, hasDefault := fieldMap["default"]; hasDefault {
					if fieldType, hasType := fieldMap["type"]; hasType {
						switch fieldType {
						case "boolean":
							assert.IsType(t, true, defaultVal, "Default for boolean field %s should be boolean", fieldName)
						case "string":
							assert.IsType(t, "", defaultVal, "Default for string field %s should be string", fieldName)
						case "number":
							// JSON unmarshaling can produce either int or float64
							assert.True(t,
								assert.IsType(t, 0, defaultVal) || assert.IsType(t, 0.0, defaultVal),
								"Default for number field %s should be numeric", fieldName)
						}
					}
				}
			}
		}
	})
}
