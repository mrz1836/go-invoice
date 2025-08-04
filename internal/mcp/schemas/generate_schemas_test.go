package schemas

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// GenerateSchemasTestSuite provides comprehensive tests for generation schema definitions
type GenerateSchemasTestSuite struct {
	suite.Suite
}

func (s *GenerateSchemasTestSuite) TestGenerateHTMLSchema() {
	schema := GenerateHTMLSchema()

	s.Run("BasicStructure", func() {
		s.NotNil(schema, "Schema should not be nil")
		s.Equal("object", schema["type"], "Schema type should be object")

		// Should have properties
		properties, hasProperties := schema["properties"]
		s.True(hasProperties, "Schema should have properties")
		s.IsType(map[string]interface{}{}, properties, "Properties should be a map")
	})

	s.Run("InvoiceIdentifierFields", func() {
		properties, _ := schema["properties"].(map[string]interface{})

		// Should have invoice_id field
		if invoiceIDField, hasID := properties["invoice_id"]; hasID {
			if idMap, ok := invoiceIDField.(map[string]interface{}); ok {
				s.Equal("string", idMap["type"], "Invoice ID should be string type")
				s.Contains(idMap, "minLength", "Invoice ID should have minimum length")
				s.Contains(idMap, "examples", "Invoice ID should have examples")
			}
		}

		// Should have invoice_number field
		if invoiceNumField, hasNum := properties["invoice_number"]; hasNum {
			if numMap, ok := invoiceNumField.(map[string]interface{}); ok {
				s.Equal("string", numMap["type"], "Invoice number should be string type")
				s.Contains(numMap, "minLength", "Invoice number should have minimum length")
				s.Contains(numMap, "examples", "Invoice number should have examples")
			}
		}

		// Should have batch_invoices field
		if batchField, hasBatch := properties["batch_invoices"]; hasBatch {
			if batchMap, ok := batchField.(map[string]interface{}); ok {
				s.Equal("array", batchMap["type"], "Batch invoices should be array type")
				s.Contains(batchMap, "minItems", "Batch invoices should have minimum items")
				s.Contains(batchMap, "maxItems", "Batch invoices should have maximum items")

				// Check items schema
				if items, hasItems := batchMap["items"]; hasItems {
					if itemsMap, ok := items.(map[string]interface{}); ok {
						s.Equal("string", itemsMap["type"], "Batch invoice items should be strings")
						s.Contains(itemsMap, "minLength", "Batch invoice items should have minimum length")
					}
				}
			}
		}
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
					expectedTemplates := []string{"default", "professional", "minimal", "custom", "web"}
					for _, expectedTemplate := range expectedTemplates {
						s.Contains(enumSlice, expectedTemplate, "Should contain template: %s", expectedTemplate)
					}
				}
			}
		}
	})

	s.Run("CustomizationFields", func() {
		properties, _ := schema["properties"].(map[string]interface{})

		// Test custom_css field
		if cssField, hasCSS := properties["custom_css"]; hasCSS {
			if cssMap, ok := cssField.(map[string]interface{}); ok {
				s.Equal("string", cssMap["type"], "Custom CSS should be string type")
				s.Contains(cssMap, "examples", "Custom CSS should have examples")
			}
		}

		// Test company_name field
		if companyField, hasCompany := properties["company_name"]; hasCompany {
			if companyMap, ok := companyField.(map[string]interface{}); ok {
				s.Equal("string", companyMap["type"], "Company name should be string type")
				s.Contains(companyMap, "maxLength", "Company name should have max length")
				s.Contains(companyMap, "examples", "Company name should have examples")
			}
		}

		// Test footer_text field
		if footerField, hasFooter := properties["footer_text"]; hasFooter {
			if footerMap, ok := footerField.(map[string]interface{}); ok {
				s.Equal("string", footerMap["type"], "Footer text should be string type")
				s.Contains(footerMap, "maxLength", "Footer text should have max length")
				s.Contains(footerMap, "examples", "Footer text should have examples")
			}
		}
	})

	s.Run("BooleanFields", func() {
		properties, _ := schema["properties"].(map[string]interface{})
		booleanFields := []string{"include_logo", "include_notes", "auto_name", "web_preview", "return_html"}

		for _, fieldName := range booleanFields {
			if field, hasField := properties[fieldName]; hasField {
				if fieldMap, ok := field.(map[string]interface{}); ok {
					s.Equal("boolean", fieldMap["type"], "Field %s should be boolean type", fieldName)
					s.Contains(fieldMap, "description", "Field %s should have description", fieldName)
					// Some boolean fields should have defaults
					if fieldName == "include_logo" || fieldName == "include_notes" || fieldName == "auto_name" || fieldName == "web_preview" || fieldName == "return_html" {
						s.Contains(fieldMap, "default", "Field %s should have default", fieldName)
					}
				}
			}
		}
	})

	s.Run("OutputFields", func() {
		properties, _ := schema["properties"].(map[string]interface{})

		// Test output_path field
		if pathField, hasPath := properties["output_path"]; hasPath {
			if pathMap, ok := pathField.(map[string]interface{}); ok {
				s.Equal("string", pathMap["type"], "Output path should be string type")
				s.Contains(pathMap, "examples", "Output path should have examples")
			}
		}

		// Test output_dir field
		if dirField, hasDir := properties["output_dir"]; hasDir {
			if dirMap, ok := dirField.(map[string]interface{}); ok {
				s.Equal("string", dirMap["type"], "Output dir should be string type")
				s.Contains(dirMap, "examples", "Output dir should have examples")
			}
		}
	})

	s.Run("ValidationConstraints", func() {
		// Check for allOf constraints
		if allOf, hasAllOf := schema["allOf"]; hasAllOf {
			s.IsType([]map[string]interface{}{}, allOf, "AllOf should be array of objects")
			if allOfSlice, ok := allOf.([]map[string]interface{}); ok {
				s.NotEmpty(allOfSlice, "AllOf should not be empty")
				s.Len(allOfSlice, 2, "Should have two constraint groups")
			}
		}

		// Should not allow additional properties
		additionalProps, hasAdditional := schema["additionalProperties"]
		s.True(hasAdditional, "Should specify additionalProperties")
		s.False(additionalProps.(bool), "Should not allow additional properties")
	})
}

func (s *GenerateSchemasTestSuite) TestGenerateSummarySchema() {
	schema := GenerateSummarySchema()

	s.Run("BasicStructure", func() {
		s.NotNil(schema, "Schema should not be nil")
		s.Equal("object", schema["type"], "Schema type should be object")
		s.Contains(schema, "properties", "Schema should have properties")
	})

	s.Run("SummaryTypeField", func() {
		properties, _ := schema["properties"].(map[string]interface{})
		summaryTypeField, hasSummaryType := properties["summary_type"]
		s.True(hasSummaryType, "Should have summary_type field")

		if summaryMap, ok := summaryTypeField.(map[string]interface{}); ok {
			s.Equal("string", summaryMap["type"], "Summary type should be string type")
			s.Equal("revenue", summaryMap["default"], "Summary type should default to 'revenue'")

			// Should have valid summary type options
			if enum, hasEnum := summaryMap["enum"]; hasEnum {
				if enumSlice, ok := enum.([]string); ok {
					expectedTypes := []string{"revenue", "client", "overdue", "tax", "dashboard", "custom"}
					for _, expectedType := range expectedTypes {
						s.Contains(enumSlice, expectedType, "Should contain summary type: %s", expectedType)
					}
				}
			}
		}
	})

	s.Run("PeriodField", func() {
		properties, _ := schema["properties"].(map[string]interface{})
		periodField, hasPeriod := properties["period"]
		s.True(hasPeriod, "Should have period field")

		if periodMap, ok := periodField.(map[string]interface{}); ok {
			s.Equal("string", periodMap["type"], "Period should be string type")
		}
	})

	s.Run("ValidationStructure", func() {
		// Should follow same validation patterns as other schemas
		additionalProps, hasAdditional := schema["additionalProperties"]
		s.True(hasAdditional, "Should specify additionalProperties")
		s.False(additionalProps.(bool), "Should not allow additional properties")
	})
}

func (s *GenerateSchemasTestSuite) TestExportDataSchema() {
	schema := ExportDataSchema()

	s.Run("BasicStructure", func() {
		s.NotNil(schema, "Schema should not be nil")
		s.Equal("object", schema["type"], "Schema type should be object")
		s.Contains(schema, "properties", "Schema should have properties")
	})

	s.Run("SchemaValidation", func() {
		// Should follow same validation patterns as other schemas
		additionalProps, hasAdditional := schema["additionalProperties"]
		s.True(hasAdditional, "Should specify additionalProperties")
		s.False(additionalProps.(bool), "Should not allow additional properties")
	})

	s.Run("PropertyValidation", func() {
		properties, hasProps := schema["properties"]
		s.True(hasProps, "Schema should have properties")

		if propsMap, ok := properties.(map[string]interface{}); ok {
			s.NotEmpty(propsMap, "Properties should not be empty")

			// Each property should have proper structure
			for fieldName, fieldDef := range propsMap {
				s.NotEmpty(fieldName, "Field name should not be empty")
				s.IsType(map[string]interface{}{}, fieldDef, "Field %s should have definition map", fieldName)

				if fieldMap, ok := fieldDef.(map[string]interface{}); ok {
					s.Contains(fieldMap, "type", "Field %s should have type", fieldName)
					s.Contains(fieldMap, "description", "Field %s should have description", fieldName)
				}
			}
		}
	})
}

func (s *GenerateSchemasTestSuite) TestGetAllGenerationSchemas() {
	schemas := GetAllGenerationSchemas()

	s.Run("AllSchemasPresent", func() {
		expectedSchemas := []string{"generate_html", "generate_summary", "export_data"}
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
		schemas1 := GetAllGenerationSchemas()
		schemas2 := GetAllGenerationSchemas()

		s.Len(schemas1, len(schemas2), "Schema count should be consistent")

		for schemaName := range schemas1 {
			s.Contains(schemas2, schemaName, "Schema %s should exist in both calls", schemaName)

			schema1 := schemas1[schemaName]
			schema2 := schemas2[schemaName]
			s.Equal(schema1["type"], schema2["type"], "Schema %s type should be consistent", schemaName)
		}
	})
}

func (s *GenerateSchemasTestSuite) TestGetGenerationToolSchema() {
	s.Run("ValidSchemas", func() {
		validSchemas := []string{"generate_html", "generate_summary", "export_data"}

		for _, schemaName := range validSchemas {
			schema, exists := GetGenerationToolSchema(schemaName)
			s.True(exists, "Schema %s should exist", schemaName)
			s.NotNil(schema, "Schema %s should not be nil", schemaName)
			s.Equal("object", schema["type"], "Schema %s should be object type", schemaName)
		}
	})

	s.Run("InvalidSchema", func() {
		schema, exists := GetGenerationToolSchema("nonexistent_schema")
		s.False(exists, "Nonexistent schema should not exist")
		s.Nil(schema, "Nonexistent schema should return nil")
	})

	s.Run("EmptySchemaName", func() {
		schema, exists := GetGenerationToolSchema("")
		s.False(exists, "Empty schema name should not exist")
		s.Nil(schema, "Empty schema name should return nil")
	})

	s.Run("CaseSensitivity", func() {
		schema, exists := GetGenerationToolSchema("GENERATE_HTML")
		s.False(exists, "Schema names should be case sensitive")
		s.Nil(schema, "Case mismatch should return nil")
	})
}

func (s *GenerateSchemasTestSuite) TestSchemaValidation() {
	schemas := map[string]func() map[string]interface{}{
		"GenerateHTML":    GenerateHTMLSchema,
		"GenerateSummary": GenerateSummarySchema,
		"ExportData":      ExportDataSchema,
	}

	for schemaName, schemaFunc := range schemas {
		s.Run(schemaName+"Validation", func() {
			schema := schemaFunc()
			s.validateJSONSchema(schema, schemaName)
		})
	}
}

func (s *GenerateSchemasTestSuite) TestSchemaConsistency() {
	s.Run("TemplateFieldConsistency", func() {
		// HTML generation schema should have template field with consistent structure
		htmlSchema := GenerateHTMLSchema()
		props, _ := htmlSchema["properties"].(map[string]interface{})

		if templateField, hasTemplate := props["template"]; hasTemplate {
			if templateMap, ok := templateField.(map[string]interface{}); ok {
				s.Equal("string", templateMap["type"], "Template type should be string")
				s.Contains(templateMap, "enum", "Template should have enum values")
				s.Contains(templateMap, "default", "Template should have default value")
			}
		}
	})

	s.Run("OutputFieldConsistency", func() {
		// Output-related fields should be consistent across schemas
		htmlSchema := GenerateHTMLSchema()
		props, _ := htmlSchema["properties"].(map[string]interface{})

		outputFields := []string{"output_path", "output_dir"}
		for _, fieldName := range outputFields {
			if field, hasField := props[fieldName]; hasField {
				if fieldMap, ok := field.(map[string]interface{}); ok {
					s.Equal("string", fieldMap["type"], "Output field %s should be string type", fieldName)
					s.Contains(fieldMap, "examples", "Output field %s should have examples", fieldName)
				}
			}
		}
	})

	s.Run("BooleanDefaultConsistency", func() {
		// Boolean fields should have consistent default values where logical
		htmlSchema := GenerateHTMLSchema()
		props, _ := htmlSchema["properties"].(map[string]interface{})

		// These boolean fields should default to false for safety
		falseDefaultFields := []string{"include_logo", "include_notes", "auto_name", "web_preview", "return_html"}
		for _, fieldName := range falseDefaultFields {
			if field, hasField := props[fieldName]; hasField {
				if fieldMap, ok := field.(map[string]interface{}); ok {
					s.Equal(false, fieldMap["default"], "Boolean field %s should default to false", fieldName)
				}
			}
		}
	})
}

func (s *GenerateSchemasTestSuite) TestSchemaEdgeCases() {
	s.Run("EmptySchemas", func() {
		// Ensure schemas are not empty
		schemas := []func() map[string]interface{}{
			GenerateHTMLSchema,
			GenerateSummarySchema,
			ExportDataSchema,
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
			"html":    GenerateHTMLSchema,
			"summary": GenerateSummarySchema,
			"export":  ExportDataSchema,
		}

		for schemaName, schemaFunc := range schemas {
			schema := schemaFunc()
			props, _ := schema["properties"].(map[string]interface{})
			s.Greater(len(props), 3, "Schema %s should have more than 3 properties", schemaName)
			s.Less(len(props), 50, "Schema %s should have fewer than 50 properties", schemaName)
		}
	})

	s.Run("DescriptionPresence", func() {
		// All properties should have descriptions
		schemas := []map[string]interface{}{
			GenerateHTMLSchema(),
			GenerateSummarySchema(),
			ExportDataSchema(),
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

	s.Run("EnumValidation", func() {
		// Test that enum values are valid and consistent
		htmlSchema := GenerateHTMLSchema()
		props, _ := htmlSchema["properties"].(map[string]interface{})

		if templateField, hasTemplate := props["template"]; hasTemplate {
			if templateMap, ok := templateField.(map[string]interface{}); ok {
				if enum, hasEnum := templateMap["enum"]; hasEnum {
					if enumSlice, ok := enum.([]string); ok {
						s.Require().NotEmpty(enumSlice, "Template enum should not be empty")
						s.Contains(enumSlice, "default", "Template enum should contain 'default'")

						// All enum values should be non-empty strings
						for _, enumVal := range enumSlice {
							s.NotEmpty(enumVal, "Enum value should not be empty")
						}
					}
				}
			}
		}
	})

	s.Run("LengthConstraints", func() {
		// Test that string fields with length constraints are properly configured
		htmlSchema := GenerateHTMLSchema()
		props, _ := htmlSchema["properties"].(map[string]interface{})

		// Fields that should have length constraints
		lengthFields := []string{"company_name", "footer_text"}
		for _, fieldName := range lengthFields {
			if field, hasField := props[fieldName]; hasField {
				if fieldMap, ok := field.(map[string]interface{}); ok {
					s.Contains(fieldMap, "maxLength", "Field %s should have maxLength constraint", fieldName)
					if maxLength, hasMax := fieldMap["maxLength"]; hasMax {
						s.IsType(0, maxLength, "MaxLength should be numeric for field %s", fieldName)
						s.Positive(maxLength.(int), "MaxLength should be positive for field %s", fieldName)
					}
				}
			}
		}
	})
}

// Helper method to validate JSON Schema structure
func (s *GenerateSchemasTestSuite) validateJSONSchema(schema map[string]interface{}, schemaName string) {
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

// TestGenerateSchemasTestSuite runs the complete generate schemas test suite
func TestGenerateSchemasTestSuite(t *testing.T) {
	suite.Run(t, new(GenerateSchemasTestSuite))
}

// Benchmark tests for schema creation
func BenchmarkGenerateHTMLSchema(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		schema := GenerateHTMLSchema()
		_ = schema
	}
}

func BenchmarkGenerateSummarySchema(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		schema := GenerateSummarySchema()
		_ = schema
	}
}

func BenchmarkExportDataSchema(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		schema := ExportDataSchema()
		_ = schema
	}
}

func BenchmarkGetAllGenerationSchemas(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		schemas := GetAllGenerationSchemas()
		_ = schemas
	}
}

// Unit tests for specific schema behaviors
func TestGenerateSchemas_Specific(t *testing.T) {
	t.Run("AllSchemasReturnValidMaps", func(t *testing.T) {
		schemas := []func() map[string]interface{}{
			GenerateHTMLSchema,
			GenerateSummarySchema,
			ExportDataSchema,
		}

		for _, schemaFunc := range schemas {
			schema := schemaFunc()
			assert.NotNil(t, schema)
			assert.IsType(t, map[string]interface{}{}, schema)
			assert.Contains(t, schema, "type")
		}
	})

	t.Run("GetGenerationToolSchemaEdgeCases", func(t *testing.T) {
		// Test with nil string
		schema, exists := GetGenerationToolSchema("")
		assert.False(t, exists)
		assert.Nil(t, schema)

		// Test with unknown schema
		schema, exists = GetGenerationToolSchema("unknown_generation_tool")
		assert.False(t, exists)
		assert.Nil(t, schema)

		// Test with valid schema
		schema, exists = GetGenerationToolSchema("generate_html")
		assert.True(t, exists)
		assert.NotNil(t, schema)
	})

	t.Run("HTMLSchemaConstraints", func(t *testing.T) {
		schema := GenerateHTMLSchema()

		// Should have allOf constraints for complex validation
		if allOf, hasAllOf := schema["allOf"]; hasAllOf {
			assert.IsType(t, []map[string]interface{}{}, allOf)
			allOfSlice := allOf.([]map[string]interface{})
			assert.NotEmpty(t, allOfSlice, "AllOf constraints should not be empty")
		}
	})

	t.Run("ArrayFieldValidation", func(t *testing.T) {
		schema := GenerateHTMLSchema()
		props, _ := schema["properties"].(map[string]interface{})

		// batch_invoices should have proper array constraints
		if batchField, hasBatch := props["batch_invoices"]; hasBatch {
			if batchMap, ok := batchField.(map[string]interface{}); ok {
				assert.Equal(t, "array", batchMap["type"])
				assert.Contains(t, batchMap, "minItems")
				assert.Contains(t, batchMap, "maxItems")
				assert.Contains(t, batchMap, "items")
			}
		}
	})
}

// Edge case tests
func TestGenerateSchemas_EdgeCases(t *testing.T) {
	t.Run("SchemasNotNil", func(t *testing.T) {
		assert.NotNil(t, GenerateHTMLSchema())
		assert.NotNil(t, GenerateSummarySchema())
		assert.NotNil(t, ExportDataSchema())
	})

	t.Run("SchemasHaveMinimalStructure", func(t *testing.T) {
		schemas := []map[string]interface{}{
			GenerateHTMLSchema(),
			GenerateSummarySchema(),
			ExportDataSchema(),
		}

		for i, schema := range schemas {
			assert.Contains(t, schema, "type", "Schema %d should have type", i)
			assert.Equal(t, "object", schema["type"], "Schema %d should be object type", i)
			assert.Contains(t, schema, "properties", "Schema %d should have properties", i)
		}
	})

	t.Run("DefaultValueTypes", func(t *testing.T) {
		// Test that default values match their field types
		htmlSchema := GenerateHTMLSchema()
		props, _ := htmlSchema["properties"].(map[string]interface{})

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

	t.Run("ExamplesValidation", func(t *testing.T) {
		// Test that example values are provided and valid
		htmlSchema := GenerateHTMLSchema()
		props, _ := htmlSchema["properties"].(map[string]interface{})

		fieldsWithExamples := []string{"invoice_id", "invoice_number", "custom_css", "company_name", "footer_text", "output_path", "output_dir"}
		for _, fieldName := range fieldsWithExamples {
			if field, hasField := props[fieldName]; hasField {
				if fieldMap, ok := field.(map[string]interface{}); ok {
					if examples, hasExamples := fieldMap["examples"]; hasExamples {
						assert.IsType(t, []string{}, examples, "Examples for field %s should be string array", fieldName)
						if exampleSlice, ok := examples.([]string); ok {
							assert.NotEmpty(t, exampleSlice, "Examples for field %s should not be empty", fieldName)
							for _, example := range exampleSlice {
								assert.NotEmpty(t, example, "Example value should not be empty for field %s", fieldName)
							}
						}
					}
				}
			}
		}
	})
}
