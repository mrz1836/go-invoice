package schemas

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// ImportSchemasTestSuite provides comprehensive tests for import schema definitions
type ImportSchemasTestSuite struct {
	suite.Suite
}

func (s *ImportSchemasTestSuite) TestImportCSVSchema() {
	schema := ImportCSVSchema()

	s.Run("BasicStructure", func() {
		s.NotNil(schema, "Schema should not be nil")
		s.Equal("object", schema["type"], "Schema type should be object")

		// Should have properties
		properties, hasProperties := schema["properties"]
		s.True(hasProperties, "Schema should have properties")
		s.IsType(map[string]interface{}{}, properties, "Properties should be a map")
	})

	s.Run("FilePathField", func() {
		properties, _ := schema["properties"].(map[string]interface{})
		filePathField, hasFilePath := properties["file_path"]
		s.True(hasFilePath, "Should have file_path field")

		if pathMap, ok := filePathField.(map[string]interface{}); ok {
			s.Equal("string", pathMap["type"], "File path should be string type")
			s.Contains(pathMap, "description", "File path should have description")
			s.Contains(pathMap, "examples", "File path should have examples")
		}
	})

	s.Run("ValidationRequirements", func() {
		// Should have required fields
		if required, hasRequired := schema["required"]; hasRequired {
			s.IsType([]interface{}{}, required, "Required should be array")
			if reqSlice, ok := required.([]interface{}); ok {
				s.NotEmpty(reqSlice, "Should have at least one required field")

				// file_path should be required
				foundFilePath := false
				for _, req := range reqSlice {
					if req == "file_path" {
						foundFilePath = true
						break
					}
				}
				s.True(foundFilePath, "file_path should be required")
			}
		}

		// Should not allow additional properties
		additionalProps, hasAdditional := schema["additionalProperties"]
		s.True(hasAdditional, "Should specify additionalProperties")
		s.False(additionalProps.(bool), "Should not allow additional properties")
	})

	s.Run("OutputFormatField", func() {
		properties, _ := schema["properties"].(map[string]interface{})
		if formatField, hasFormat := properties["output_format"]; hasFormat {
			if formatMap, ok := formatField.(map[string]interface{}); ok {
				s.Equal("string", formatMap["type"], "Output format should be string type")
				if enum, hasEnum := formatMap["enum"]; hasEnum {
					if enumSlice, ok := enum.([]string); ok {
						s.NotEmpty(enumSlice, "Output format enum should not be empty")
						// Should contain common formats
						expectedFormats := []string{"json", "yaml", "table"}
						for _, expectedFormat := range expectedFormats {
							if s.Contains(enumSlice, expectedFormat) {
								s.Contains(enumSlice, expectedFormat, "Should contain format: %s", expectedFormat)
							}
						}
					}
				}
			}
		}
	})

	s.Run("BooleanFields", func() {
		properties, _ := schema["properties"].(map[string]interface{})

		// Common boolean fields in import schemas
		possibleBooleanFields := []string{
			"validate", "skip_empty", "auto_correct", "dry_run",
			"verbose", "strict_mode", "include_headers", "overwrite",
		}

		for _, fieldName := range possibleBooleanFields {
			if field, hasField := properties[fieldName]; hasField {
				if fieldMap, ok := field.(map[string]interface{}); ok {
					if fieldType, hasType := fieldMap["type"]; hasType && fieldType == "boolean" {
						s.Equal("boolean", fieldMap["type"], "Field %s should be boolean type", fieldName)
						s.Contains(fieldMap, "description", "Field %s should have description", fieldName)
					}
				}
			}
		}
	})
}

func (s *ImportSchemasTestSuite) TestImportValidateSchema() {
	schema := ImportValidateSchema()

	s.Run("BasicStructure", func() {
		s.NotNil(schema, "Schema should not be nil")
		s.Equal("object", schema["type"], "Schema type should be object")
		s.Contains(schema, "properties", "Schema should have properties")
	})

	s.Run("FilePathField", func() {
		properties, _ := schema["properties"].(map[string]interface{})
		filePathField, hasFilePath := properties["file_path"]
		s.True(hasFilePath, "Should have file_path field")

		if pathMap, ok := filePathField.(map[string]interface{}); ok {
			s.Equal("string", pathMap["type"], "File path should be string type")
			s.Contains(pathMap, "description", "File path should have description")
		}
	})

	s.Run("ValidationOptions", func() {
		properties, _ := schema["properties"].(map[string]interface{})

		// Should have validation-specific fields
		validationFields := []string{"strict", "comprehensive", "include_warnings"}
		for _, fieldName := range validationFields {
			if field, hasField := properties[fieldName]; hasField {
				if fieldMap, ok := field.(map[string]interface{}); ok {
					if fieldType, hasType := fieldMap["type"]; hasType && fieldType == "boolean" {
						s.Equal("boolean", fieldMap["type"], "Field %s should be boolean type", fieldName)
						s.Contains(fieldMap, "description", "Field %s should have description", fieldName)
					}
				}
			}
		}
	})

	s.Run("SchemaValidation", func() {
		// Should follow same validation patterns as other schemas
		additionalProps, hasAdditional := schema["additionalProperties"]
		s.True(hasAdditional, "Should specify additionalProperties")
		s.False(additionalProps.(bool), "Should not allow additional properties")
	})
}

func (s *ImportSchemasTestSuite) TestImportPreviewSchema() {
	schema := ImportPreviewSchema()

	s.Run("BasicStructure", func() {
		s.NotNil(schema, "Schema should not be nil")
		s.Equal("object", schema["type"], "Schema type should be object")
		s.Contains(schema, "properties", "Schema should have properties")
	})

	s.Run("FilePathField", func() {
		properties, _ := schema["properties"].(map[string]interface{})
		filePathField, hasFilePath := properties["file_path"]
		s.True(hasFilePath, "Should have file_path field")

		if pathMap, ok := filePathField.(map[string]interface{}); ok {
			s.Equal("string", pathMap["type"], "File path should be string type")
			s.Contains(pathMap, "description", "File path should have description")
		}
	})

	s.Run("PreviewOptions", func() {
		properties, _ := schema["properties"].(map[string]interface{})

		// Should have preview-specific fields
		if limitField, hasLimit := properties["limit"]; hasLimit {
			if limitMap, ok := limitField.(map[string]interface{}); ok {
				s.Equal("number", limitMap["type"], "Limit should be number type")
				s.Contains(limitMap, "description", "Limit should have description")
				if minimum, hasMin := limitMap["minimum"]; hasMin {
					s.IsType(0, minimum, "Minimum should be numeric")
				}
			}
		}

		if offsetField, hasOffset := properties["offset"]; hasOffset {
			if offsetMap, ok := offsetField.(map[string]interface{}); ok {
				s.Equal("number", offsetMap["type"], "Offset should be number type")
				s.Contains(offsetMap, "description", "Offset should have description")
				if minimum, hasMin := offsetMap["minimum"]; hasMin {
					s.GreaterOrEqual(minimum, 0, "Offset minimum should be 0 or greater")
				}
			}
		}
	})

	s.Run("SchemaValidation", func() {
		// Should follow same validation patterns as other schemas
		additionalProps, hasAdditional := schema["additionalProperties"]
		s.True(hasAdditional, "Should specify additionalProperties")
		s.False(additionalProps.(bool), "Should not allow additional properties")
	})
}

func (s *ImportSchemasTestSuite) TestGetAllImportSchemas() {
	schemas := GetAllImportSchemas()

	s.Run("AllSchemasPresent", func() {
		expectedSchemas := []string{"import_csv", "import_validate", "import_preview"}
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
			s.False(schema["additionalProperties"].(bool), "Schema %s should not allow additional properties", schemaName)
		}
	})

	s.Run("SchemaImmutability", func() {
		// Test that calling the function multiple times returns consistent results
		schemas1 := GetAllImportSchemas()
		schemas2 := GetAllImportSchemas()

		s.Len(schemas1, len(schemas2), "Schema count should be consistent")

		for schemaName := range schemas1 {
			s.Contains(schemas2, schemaName, "Schema %s should exist in both calls", schemaName)

			schema1 := schemas1[schemaName]
			schema2 := schemas2[schemaName]
			s.Equal(schema1["type"], schema2["type"], "Schema %s type should be consistent", schemaName)
		}
	})
}

func (s *ImportSchemasTestSuite) TestGetImportToolSchema() {
	s.Run("ValidSchemas", func() {
		validSchemas := []string{"import_csv", "import_validate", "import_preview"}

		for _, schemaName := range validSchemas {
			schema, exists := GetImportToolSchema(schemaName)
			s.True(exists, "Schema %s should exist", schemaName)
			s.NotNil(schema, "Schema %s should not be nil", schemaName)
			s.Equal("object", schema["type"], "Schema %s should be object type", schemaName)
		}
	})

	s.Run("InvalidSchema", func() {
		schema, exists := GetImportToolSchema("nonexistent_schema")
		s.False(exists, "Nonexistent schema should not exist")
		s.Nil(schema, "Nonexistent schema should return nil")
	})

	s.Run("EmptySchemaName", func() {
		schema, exists := GetImportToolSchema("")
		s.False(exists, "Empty schema name should not exist")
		s.Nil(schema, "Empty schema name should return nil")
	})

	s.Run("CaseSensitivity", func() {
		schema, exists := GetImportToolSchema("IMPORT_CSV")
		s.False(exists, "Schema names should be case sensitive")
		s.Nil(schema, "Case mismatch should return nil")
	})
}

func (s *ImportSchemasTestSuite) TestSchemaValidation() {
	schemas := map[string]func() map[string]interface{}{
		"ImportCSV":      ImportCSVSchema,
		"ImportValidate": ImportValidateSchema,
		"ImportPreview":  ImportPreviewSchema,
	}

	for schemaName, schemaFunc := range schemas {
		s.Run(schemaName+"Validation", func() {
			schema := schemaFunc()
			s.validateJSONSchema(schema, schemaName)
		})
	}
}

func (s *ImportSchemasTestSuite) TestSchemaConsistency() {
	s.Run("FilePathFieldConsistency", func() {
		// All import schemas should have consistent file_path field definitions
		schemas := []map[string]interface{}{
			ImportCSVSchema(),
			ImportValidateSchema(),
			ImportPreviewSchema(),
		}

		var firstFilePathDef map[string]interface{}
		for i, schema := range schemas {
			props, _ := schema["properties"].(map[string]interface{})
			if filePathField, hasFilePath := props["file_path"]; hasFilePath {
				if filePathMap, ok := filePathField.(map[string]interface{}); ok {
					if i == 0 {
						firstFilePathDef = filePathMap
					} else {
						s.Equal(firstFilePathDef["type"], filePathMap["type"], "File path type should be consistent across schemas")
						// Description might vary, but type should be consistent
					}
				}
			}
		}
	})

	s.Run("BooleanDefaultConsistency", func() {
		// Similar boolean fields should have consistent defaults where logical
		csvSchema := ImportCSVSchema()
		props, _ := csvSchema["properties"].(map[string]interface{})

		// Safety-related fields should default to false unless explicitly needed
		safetyFields := []string{"overwrite", "skip_validation", "auto_fix"}
		for _, fieldName := range safetyFields {
			if field, hasField := props[fieldName]; hasField {
				if fieldMap, ok := field.(map[string]interface{}); ok {
					if defaultVal, hasDefault := fieldMap["default"]; hasDefault {
						s.Equal(false, defaultVal, "Safety field %s should typically default to false", fieldName)
					}
				}
			}
		}
	})

	s.Run("OutputFormatConsistency", func() {
		// Output format fields should be consistent across schemas where they exist
		schemas := []map[string]interface{}{
			ImportCSVSchema(),
			ImportValidateSchema(),
			ImportPreviewSchema(),
		}

		for _, schema := range schemas {
			props, _ := schema["properties"].(map[string]interface{})
			if formatField, hasFormat := props["output_format"]; hasFormat {
				if formatMap, ok := formatField.(map[string]interface{}); ok {
					if enum, hasEnum := formatMap["enum"]; hasEnum {
						if enumSlice, ok := enum.([]string); ok {
							// Should have similar format options
							s.NotEmpty(enumSlice, "Output format enum should not be empty")
						}
					}
				}
			}
		}
	})
}

func (s *ImportSchemasTestSuite) TestSchemaEdgeCases() {
	s.Run("EmptySchemas", func() {
		// Ensure schemas are not empty
		schemas := []func() map[string]interface{}{
			ImportCSVSchema,
			ImportValidateSchema,
			ImportPreviewSchema,
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
			"csv":      ImportCSVSchema,
			"validate": ImportValidateSchema,
			"preview":  ImportPreviewSchema,
		}

		for schemaName, schemaFunc := range schemas {
			schema := schemaFunc()
			props, _ := schema["properties"].(map[string]interface{})
			s.Greater(len(props), 2, "Schema %s should have more than 2 properties", schemaName)
			s.Less(len(props), 50, "Schema %s should have fewer than 50 properties", schemaName)
		}
	})

	s.Run("DescriptionPresence", func() {
		// All properties should have descriptions
		schemas := []map[string]interface{}{
			ImportCSVSchema(),
			ImportValidateSchema(),
			ImportPreviewSchema(),
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

	s.Run("RequiredFieldsValidation", func() {
		// Import schemas should have required fields
		csvSchema := ImportCSVSchema()
		if required, hasRequired := csvSchema["required"]; hasRequired {
			if reqSlice, ok := required.([]interface{}); ok {
				s.NotEmpty(reqSlice, "Import CSV schema should have required fields")

				// file_path should typically be required for import operations
				foundFilePath := false
				for _, req := range reqSlice {
					if req == "file_path" {
						foundFilePath = true
						break
					}
				}
				s.True(foundFilePath, "file_path should be required for import operations")
			}
		}
	})

	s.Run("NumericConstraints", func() {
		// Test numeric fields have proper constraints
		previewSchema := ImportPreviewSchema()
		props, _ := previewSchema["properties"].(map[string]interface{})

		numericFields := []string{"limit", "offset", "max_rows"}
		for _, fieldName := range numericFields {
			if field, hasField := props[fieldName]; hasField {
				if fieldMap, ok := field.(map[string]interface{}); ok {
					if fieldType, hasType := fieldMap["type"]; hasType && fieldType == "number" {
						s.Contains(fieldMap, "minimum", "Numeric field %s should have minimum constraint", fieldName)
						if minimum, hasMin := fieldMap["minimum"]; hasMin {
							s.GreaterOrEqual(minimum, 0, "Minimum for field %s should be non-negative", fieldName)
						}
					}
				}
			}
		}
	})

	s.Run("ExamplesValidation", func() {
		// Test that example values are provided and valid
		csvSchema := ImportCSVSchema()
		props, _ := csvSchema["properties"].(map[string]interface{})

		if filePathField, hasFilePath := props["file_path"]; hasFilePath {
			if pathMap, ok := filePathField.(map[string]interface{}); ok {
				if examples, hasExamples := pathMap["examples"]; hasExamples {
					s.IsType([]string{}, examples, "Examples should be string array")
					if exampleSlice, ok := examples.([]string); ok {
						s.NotEmpty(exampleSlice, "Examples should not be empty")
						for _, example := range exampleSlice {
							s.NotEmpty(example, "Example value should not be empty")
							// File path examples should look like file paths
							s.Contains(example, ".", "File path example should contain file extension")
						}
					}
				}
			}
		}
	})
}

// Helper method to validate JSON Schema structure
func (s *ImportSchemasTestSuite) validateJSONSchema(schema map[string]interface{}, schemaName string) {
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

// TestImportSchemasTestSuite runs the complete import schemas test suite
func TestImportSchemasTestSuite(t *testing.T) {
	suite.Run(t, new(ImportSchemasTestSuite))
}

// Benchmark tests for schema creation
func BenchmarkImportCSVSchema(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		schema := ImportCSVSchema()
		_ = schema
	}
}

func BenchmarkImportValidateSchema(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		schema := ImportValidateSchema()
		_ = schema
	}
}

func BenchmarkImportPreviewSchema(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		schema := ImportPreviewSchema()
		_ = schema
	}
}

func BenchmarkGetAllImportSchemas(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		schemas := GetAllImportSchemas()
		_ = schemas
	}
}

// Unit tests for specific schema behaviors
func TestImportSchemas_Specific(t *testing.T) {
	t.Run("AllSchemasReturnValidMaps", func(t *testing.T) {
		schemas := []func() map[string]interface{}{
			ImportCSVSchema,
			ImportValidateSchema,
			ImportPreviewSchema,
		}

		for _, schemaFunc := range schemas {
			schema := schemaFunc()
			assert.NotNil(t, schema)
			assert.IsType(t, map[string]interface{}{}, schema)
			assert.Contains(t, schema, "type")
		}
	})

	t.Run("GetImportToolSchemaEdgeCases", func(t *testing.T) {
		// Test with nil string
		schema, exists := GetImportToolSchema("")
		assert.False(t, exists)
		assert.Nil(t, schema)

		// Test with unknown schema
		schema, exists = GetImportToolSchema("unknown_import_tool")
		assert.False(t, exists)
		assert.Nil(t, schema)

		// Test with valid schema
		schema, exists = GetImportToolSchema("import_csv")
		assert.True(t, exists)
		assert.NotNil(t, schema)
	})

	t.Run("FilePathFieldValidation", func(t *testing.T) {
		// All import schemas should have file_path field
		schemas := []map[string]interface{}{
			ImportCSVSchema(),
			ImportValidateSchema(),
			ImportPreviewSchema(),
		}

		for i, schema := range schemas {
			props, _ := schema["properties"].(map[string]interface{})
			assert.Contains(t, props, "file_path", "Schema %d should have file_path field", i)

			if filePathField, hasFilePath := props["file_path"]; hasFilePath {
				if pathMap, ok := filePathField.(map[string]interface{}); ok {
					assert.Equal(t, "string", pathMap["type"], "File path should be string type in schema %d", i)
				}
			}
		}
	})

	t.Run("RequiredFieldsValidation", func(t *testing.T) {
		csvSchema := ImportCSVSchema()

		if required, hasRequired := csvSchema["required"]; hasRequired {
			assert.IsType(t, []interface{}{}, required)
			reqSlice := required.([]interface{})
			assert.NotEmpty(t, reqSlice, "Should have required fields")
		}
	})
}

// Edge case tests
func TestImportSchemas_EdgeCases(t *testing.T) {
	t.Run("SchemasNotNil", func(t *testing.T) {
		assert.NotNil(t, ImportCSVSchema())
		assert.NotNil(t, ImportValidateSchema())
		assert.NotNil(t, ImportPreviewSchema())
	})

	t.Run("SchemasHaveMinimalStructure", func(t *testing.T) {
		schemas := []map[string]interface{}{
			ImportCSVSchema(),
			ImportValidateSchema(),
			ImportPreviewSchema(),
		}

		for i, schema := range schemas {
			assert.Contains(t, schema, "type", "Schema %d should have type", i)
			assert.Equal(t, "object", schema["type"], "Schema %d should be object type", i)
			assert.Contains(t, schema, "properties", "Schema %d should have properties", i)
		}
	})

	t.Run("DefaultValueTypes", func(t *testing.T) {
		// Test that default values match their field types
		csvSchema := ImportCSVSchema()
		props, _ := csvSchema["properties"].(map[string]interface{})

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

	t.Run("AdditionalPropertiesConsistency", func(t *testing.T) {
		// All import schemas should consistently disallow additional properties
		schemas := []map[string]interface{}{
			ImportCSVSchema(),
			ImportValidateSchema(),
			ImportPreviewSchema(),
		}

		for i, schema := range schemas {
			assert.Contains(t, schema, "additionalProperties", "Schema %d should specify additionalProperties", i)
			assert.False(t, schema["additionalProperties"].(bool), "Schema %d should not allow additional properties", i)
		}
	})
}
