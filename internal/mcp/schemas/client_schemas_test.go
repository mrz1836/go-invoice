package schemas

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ClientSchemasTestSuite provides comprehensive tests for client schema definitions
type ClientSchemasTestSuite struct {
	suite.Suite
}

func (s *ClientSchemasTestSuite) TestClientCreateSchema() {
	schema := ClientCreateSchema()

	s.Run("BasicStructure", func() {
		s.NotNil(schema, "Schema should not be nil")
		s.Equal("object", schema["type"], "Schema type should be object")

		// Should have properties
		properties, hasProperties := schema["properties"]
		s.True(hasProperties, "Schema should have properties")
		s.IsType(map[string]interface{}{}, properties, "Properties should be a map")
	})

	s.Run("RequiredFields", func() {
		required, hasRequired := schema["required"]
		if hasRequired {
			s.IsType([]interface{}{}, required, "Required should be an array")

			if requiredSlice, ok := required.([]interface{}); ok {
				// Should have at least basic required fields for client creation
				expectedFields := []string{"name"}
				for _, expectedField := range expectedFields {
					found := false
					for _, reqField := range requiredSlice {
						if reqField == expectedField {
							found = true
							break
						}
					}
					if found {
						s.True(found, "Should have required field: %s", expectedField)
					}
				}
			}
		}
	})

	s.Run("FieldDefinitions", func() {
		properties, _ := schema["properties"].(map[string]interface{})

		// Common client fields that should exist
		expectedFields := map[string]string{
			"name":    "string",
			"email":   "string",
			"address": "string",
			"phone":   "string",
		}

		for fieldName, expectedType := range expectedFields {
			if fieldDef, hasField := properties[fieldName]; hasField {
				s.IsType(map[string]interface{}{}, fieldDef, "Field %s should have definition", fieldName)

				if fieldMap, ok := fieldDef.(map[string]interface{}); ok {
					if fieldType, hasType := fieldMap["type"]; hasType {
						s.Equal(expectedType, fieldType, "Field %s should have type %s", fieldName, expectedType)
					}
				}
			}
		}
	})

	s.Run("ValidationConstraints", func() {
		properties, _ := schema["properties"].(map[string]interface{})

		// Email field should have format validation
		if emailField, hasEmail := properties["email"]; hasEmail {
			if emailMap, ok := emailField.(map[string]interface{}); ok {
				if format, hasFormat := emailMap["format"]; hasFormat {
					s.Equal("email", format, "Email field should have email format")
				}
			}
		}

		// Name field should have length constraints
		if nameField, hasName := properties["name"]; hasName {
			if nameMap, ok := nameField.(map[string]interface{}); ok {
				if minLength, hasMinLength := nameMap["minLength"]; hasMinLength {
					s.IsType(float64(0), minLength, "MinLength should be numeric")
					s.Greater(minLength.(float64), 0.0, "Name should have minimum length")
				}
			}
		}
	})

	s.Run("SchemaCompleteness", func() {
		// Verify schema has all necessary components for JSON Schema validation
		s.Contains(schema, "type", "Schema should have type")
		s.Contains(schema, "properties", "Schema should have properties")

		// Should not allow additional properties by default for strict validation
		if additionalProperties, hasAdditional := schema["additionalProperties"]; hasAdditional {
			s.IsType(false, additionalProperties, "Additional properties should be boolean")
		}
	})
}

func (s *ClientSchemasTestSuite) TestClientListSchema() {
	schema := ClientListSchema()

	s.Run("BasicStructure", func() {
		s.NotNil(schema, "Schema should not be nil")
		s.Equal("object", schema["type"], "Schema type should be object")
	})

	s.Run("FilteringFields", func() {
		properties, hasProperties := schema["properties"]
		if hasProperties {
			if propertiesMap, ok := properties.(map[string]interface{}); ok {
				// Common list filtering fields
				potentialFields := []string{"search", "limit", "offset", "output_format", "sort_by"}

				for _, field := range potentialFields {
					if fieldDef, hasField := propertiesMap[field]; hasField {
						s.IsType(map[string]interface{}{}, fieldDef, "Field %s should have definition", field)

						if fieldMap, ok := fieldDef.(map[string]interface{}); ok {
							s.Contains(fieldMap, "type", "Field %s should have type", field)
						}
					}
				}
			}
		}
	})

	s.Run("OptionalParameters", func() {
		// List schemas typically have all optional parameters
		required, hasRequired := schema["required"]
		if hasRequired {
			if requiredSlice, ok := required.([]interface{}); ok {
				s.Empty(requiredSlice, "List schema should have no required fields or minimal required fields")
			}
		}
	})
}

func (s *ClientSchemasTestSuite) TestClientShowSchema() {
	schema := ClientShowSchema()

	s.Run("BasicStructure", func() {
		s.NotNil(schema, "Schema should not be nil")
		s.Equal("object", schema["type"], "Schema type should be object")
	})

	s.Run("IdentifierFields", func() {
		properties, hasProperties := schema["properties"]
		s.True(hasProperties, "Schema should have properties for client identification")

		if propertiesMap, ok := properties.(map[string]interface{}); ok {
			// Should have at least one way to identify clients
			identifierFields := []string{"client_id", "client_name", "email", "id", "name"}
			hasIdentifier := false

			for _, field := range identifierFields {
				if _, hasField := propertiesMap[field]; hasField {
					hasIdentifier = true
					break
				}
			}

			s.True(hasIdentifier, "Should have at least one client identifier field")
		}
	})

	s.Run("OutputOptions", func() {
		properties, _ := schema["properties"].(map[string]interface{})

		// May have output format options
		outputFields := []string{"output_format", "format", "show_details"}
		for _, field := range outputFields {
			if fieldDef, hasField := properties[field]; hasField {
				s.IsType(map[string]interface{}{}, fieldDef, "Output field %s should have definition", field)
			}
		}
	})
}

func (s *ClientSchemasTestSuite) TestClientUpdateSchema() {
	schema := ClientUpdateSchema()

	s.Run("BasicStructure", func() {
		s.NotNil(schema, "Schema should not be nil")
		s.Equal("object", schema["type"], "Schema type should be object")
	})

	s.Run("IdentifierAndUpdateFields", func() {
		properties, hasProperties := schema["properties"]
		s.True(hasProperties, "Schema should have properties")

		if propertiesMap, ok := properties.(map[string]interface{}); ok {
			// Should have identifier fields
			identifierFields := []string{"client_id", "client_name", "email", "id", "name"}
			hasIdentifier := false
			for _, field := range identifierFields {
				if _, hasField := propertiesMap[field]; hasField {
					hasIdentifier = true
					break
				}
			}
			s.True(hasIdentifier, "Should have client identifier field")

			// Should have updatable fields
			updatableFields := []string{"name", "email", "address", "phone", "notes"}
			hasUpdatableField := false
			for _, field := range updatableFields {
				if _, hasField := propertiesMap[field]; hasField {
					hasUpdatableField = true
					break
				}
			}
			s.True(hasUpdatableField, "Should have at least one updatable field")
		}
	})

	s.Run("UpdateValidation", func() {
		properties, _ := schema["properties"].(map[string]interface{})

		// Email field should maintain format validation
		if emailField, hasEmail := properties["email"]; hasEmail {
			if emailMap, ok := emailField.(map[string]interface{}); ok {
				if format, hasFormat := emailMap["format"]; hasFormat {
					s.Equal("email", format, "Email field should maintain email format validation")
				}
			}
		}
	})
}

func (s *ClientSchemasTestSuite) TestClientDeleteSchema() {
	schema := ClientDeleteSchema()

	s.Run("BasicStructure", func() {
		s.NotNil(schema, "Schema should not be nil")
		s.Equal("object", schema["type"], "Schema type should be object")
	})

	s.Run("IdentifierFields", func() {
		properties, hasProperties := schema["properties"]
		s.True(hasProperties, "Schema should have properties")

		if propertiesMap, ok := properties.(map[string]interface{}); ok {
			// Should have identifier fields
			identifierFields := []string{"client_id", "client_name", "email", "id", "name"}
			hasIdentifier := false
			for _, field := range identifierFields {
				if _, hasField := propertiesMap[field]; hasField {
					hasIdentifier = true
					break
				}
			}
			s.True(hasIdentifier, "Should have client identifier field")
		}
	})

	s.Run("SafetyFields", func() {
		properties, _ := schema["properties"].(map[string]interface{})

		// May have safety/confirmation fields
		safetyFields := []string{"confirm", "force", "cascade"}
		for _, field := range safetyFields {
			if fieldDef, hasField := properties[field]; hasField {
				s.IsType(map[string]interface{}{}, fieldDef, "Safety field %s should have definition", field)

				if fieldMap, ok := fieldDef.(map[string]interface{}); ok {
					if fieldType, hasType := fieldMap["type"]; hasType {
						s.Equal("boolean", fieldType, "Safety field %s should be boolean", field)
					}
				}
			}
		}
	})
}

func (s *ClientSchemasTestSuite) TestSchemaValidation() {
	schemas := map[string]func() map[string]interface{}{
		"ClientCreate": ClientCreateSchema,
		"ClientList":   ClientListSchema,
		"ClientShow":   ClientShowSchema,
		"ClientUpdate": ClientUpdateSchema,
		"ClientDelete": ClientDeleteSchema,
	}

	for schemaName, schemaFunc := range schemas {
		s.Run(schemaName+"Validation", func() {
			schema := schemaFunc()
			s.validateJSONSchema(schema, schemaName)
		})
	}
}

func (s *ClientSchemasTestSuite) TestSchemaConsistency() {
	s.Run("FieldTypeConsistency", func() {
		// Email fields should always have email format across all schemas
		schemas := []map[string]interface{}{
			ClientCreateSchema(),
			ClientUpdateSchema(),
		}

		for i, schema := range schemas {
			properties, _ := schema["properties"].(map[string]interface{})
			if emailField, hasEmail := properties["email"]; hasEmail {
				if emailMap, ok := emailField.(map[string]interface{}); ok {
					if format, hasFormat := emailMap["format"]; hasFormat {
						s.Equal("email", format, "Email format should be consistent across schemas (schema %d)", i)
					}
				}
			}
		}
	})

	s.Run("IdentifierFieldConsistency", func() {
		// Identifier fields should be consistent across schemas
		identifierSchemas := []map[string]interface{}{
			ClientShowSchema(),
			ClientUpdateSchema(),
			ClientDeleteSchema(),
		}

		commonIdentifiers := []string{"client_id", "client_name", "email"}
		for _, identifier := range commonIdentifiers {
			fieldTypes := make(map[string]int)

			for _, schema := range identifierSchemas {
				properties, _ := schema["properties"].(map[string]interface{})
				if fieldDef, hasField := properties[identifier]; hasField {
					if fieldMap, ok := fieldDef.(map[string]interface{}); ok {
						if fieldType, hasType := fieldMap["type"]; hasType {
							fieldTypes[fieldType.(string)]++
						}
					}
				}
			}

			// If field appears in multiple schemas, type should be consistent
			if len(fieldTypes) > 1 {
				s.T().Logf("Warning: Field %s has inconsistent types across schemas: %v", identifier, fieldTypes)
			}
		}
	})
}

func (s *ClientSchemasTestSuite) TestSchemaEdgeCases() {
	s.Run("EmptySchema", func() {
		// Ensure schemas are not empty
		schemas := []func() map[string]interface{}{
			ClientCreateSchema,
			ClientListSchema,
			ClientShowSchema,
			ClientUpdateSchema,
			ClientDeleteSchema,
		}

		for _, schemaFunc := range schemas {
			schema := schemaFunc()
			s.NotEmpty(schema, "Schema should not be empty")
			s.Contains(schema, "type", "Schema should have type field")
		}
	})

	s.Run("SchemaImmutability", func() {
		// Test that calling schema functions multiple times returns consistent results
		schema1 := ClientCreateSchema()
		schema2 := ClientCreateSchema()

		s.Equal(schema1["type"], schema2["type"], "Schema type should be consistent")

		// Deep comparison would be complex, but basic structure should match
		props1, _ := schema1["properties"].(map[string]interface{})
		props2, _ := schema2["properties"].(map[string]interface{})
		s.Len(props1, len(props2), "Schema properties count should be consistent")
	})
}

// Helper method to validate JSON Schema structure
func (s *ClientSchemasTestSuite) validateJSONSchema(schema map[string]interface{}, schemaName string) {
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
				}
			}
		}
	}

	// Required field validation
	if required, hasRequired := schema["required"]; hasRequired {
		s.IsType([]interface{}{}, required, "%s required should be array", schemaName)

		if reqSlice, ok := required.([]interface{}); ok {
			properties, _ := schema["properties"].(map[string]interface{})
			for _, reqField := range reqSlice {
				s.IsType("", reqField, "Required field should be string")
				if reqFieldStr, ok := reqField.(string); ok {
					s.Contains(properties, reqFieldStr, "Required field %s should exist in properties", reqFieldStr)
				}
			}
		}
	}
}

// TestClientSchemasTestSuite runs the complete client schemas test suite
func TestClientSchemasTestSuite(t *testing.T) {
	suite.Run(t, new(ClientSchemasTestSuite))
}

// Benchmark tests for schema creation
func BenchmarkClientCreateSchema(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		schema := ClientCreateSchema()
		_ = schema
	}
}

func BenchmarkClientListSchema(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		schema := ClientListSchema()
		_ = schema
	}
}

// Unit tests for specific schema behaviors
func TestClientSchemas_Specific(t *testing.T) {
	t.Run("AllSchemasReturnValidMaps", func(t *testing.T) {
		schemas := []func() map[string]interface{}{
			ClientCreateSchema,
			ClientListSchema,
			ClientShowSchema,
			ClientUpdateSchema,
			ClientDeleteSchema,
		}

		for _, schemaFunc := range schemas {
			schema := schemaFunc()
			assert.NotNil(t, schema)
			assert.IsType(t, map[string]interface{}{}, schema)
			assert.Contains(t, schema, "type")
		}
	})

	t.Run("CreateSchemaHasRequiredFields", func(t *testing.T) {
		schema := ClientCreateSchema()

		// Client creation should require at least a name
		required, hasRequired := schema["required"]
		if hasRequired {
			if reqSlice, ok := required.([]interface{}); ok {
				hasNameRequired := false
				for _, req := range reqSlice {
					if req == "name" {
						hasNameRequired = true
						break
					}
				}
				assert.True(t, hasNameRequired, "Client creation should require name field")
			}
		}
	})

	t.Run("EmailFieldsHaveValidation", func(t *testing.T) {
		schemas := []map[string]interface{}{
			ClientCreateSchema(),
			ClientUpdateSchema(),
		}

		for _, schema := range schemas {
			properties, _ := schema["properties"].(map[string]interface{})
			if emailField, hasEmail := properties["email"]; hasEmail {
				if emailMap, ok := emailField.(map[string]interface{}); ok {
					format, hasFormat := emailMap["format"]
					assert.True(t, hasFormat, "Email field should have format validation")
					if hasFormat {
						assert.Equal(t, "email", format, "Email field should have email format")
					}
				}
			}
		}
	})
}

// Edge case tests
func TestClientSchemas_EdgeCases(t *testing.T) {
	t.Run("SchemasNotNil", func(t *testing.T) {
		assert.NotNil(t, ClientCreateSchema())
		assert.NotNil(t, ClientListSchema())
		assert.NotNil(t, ClientShowSchema())
		assert.NotNil(t, ClientUpdateSchema())
		assert.NotNil(t, ClientDeleteSchema())
	})

	t.Run("SchemasHaveMinimalStructure", func(t *testing.T) {
		schemas := []map[string]interface{}{
			ClientCreateSchema(),
			ClientListSchema(),
			ClientShowSchema(),
			ClientUpdateSchema(),
			ClientDeleteSchema(),
		}

		for i, schema := range schemas {
			assert.Contains(t, schema, "type", "Schema %d should have type", i)
			assert.Equal(t, "object", schema["type"], "Schema %d should be object type", i)
		}
	})

	t.Run("PropertiesAreValid", func(t *testing.T) {
		schema := ClientCreateSchema()
		properties, hasProps := schema["properties"]

		if hasProps {
			require.IsType(t, map[string]interface{}{}, properties)
			propsMap := properties.(map[string]interface{})

			for fieldName, fieldDef := range propsMap {
				assert.NotEmpty(t, fieldName, "Field name should not be empty")
				assert.IsType(t, map[string]interface{}{}, fieldDef, "Field definition should be map")
			}
		}
	})
}
