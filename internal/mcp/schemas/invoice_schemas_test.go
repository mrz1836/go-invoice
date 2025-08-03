package schemas

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInvoiceCreateSchema tests the invoice creation schema
func TestInvoiceCreateSchema(t *testing.T) {
	schema := InvoiceCreateSchema()

	// Verify basic schema structure
	assert.Equal(t, "object", schema["type"])

	properties, ok := schema["properties"].(map[string]interface{})
	require.True(t, ok, "Properties should be a map")

	// Verify required client identification fields
	assert.Contains(t, properties, "client_name")
	assert.Contains(t, properties, "client_id")
	assert.Contains(t, properties, "client_email")

	// Verify optional fields
	assert.Contains(t, properties, "invoice_date")
	assert.Contains(t, properties, "due_date")
	assert.Contains(t, properties, "description")
	assert.Contains(t, properties, "work_items")

	// Verify anyOf requirements for client identification
	anyOfConstraints, ok := schema["anyOf"].([]map[string]interface{})
	require.True(t, ok, "anyOf should be present")
	assert.Len(t, anyOfConstraints, 3, "Should have 3 client identification options")
}

// TestInvoiceListSchema tests the invoice listing schema
func TestInvoiceListSchema(t *testing.T) {
	schema := InvoiceListSchema()

	// Verify basic schema structure
	assert.Equal(t, "object", schema["type"])

	properties, ok := schema["properties"].(map[string]interface{})
	require.True(t, ok, "Properties should be a map")

	// Verify filtering fields
	assert.Contains(t, properties, "status")
	assert.Contains(t, properties, "client_name")
	assert.Contains(t, properties, "from_date")
	assert.Contains(t, properties, "to_date")

	// Verify output control fields
	assert.Contains(t, properties, "output_format")
	assert.Contains(t, properties, "sort_by")
	assert.Contains(t, properties, "limit")

	// Check status enum values
	statusField := properties["status"].(map[string]interface{})
	statusEnum := statusField["enum"].([]string)
	expectedStatuses := []string{"draft", "sent", "paid", "overdue", "voided"}
	assert.ElementsMatch(t, expectedStatuses, statusEnum)
}

// TestInvoiceShowSchema tests the invoice display schema
func TestInvoiceShowSchema(t *testing.T) {
	schema := InvoiceShowSchema()

	// Verify basic schema structure
	assert.Equal(t, "object", schema["type"])

	properties, ok := schema["properties"].(map[string]interface{})
	require.True(t, ok, "Properties should be a map")

	// Verify identification fields
	assert.Contains(t, properties, "invoice_id")
	assert.Contains(t, properties, "invoice_number")

	// Verify display control fields
	assert.Contains(t, properties, "output_format")
	assert.Contains(t, properties, "show_work_items")
	assert.Contains(t, properties, "show_client_details")

	// Verify anyOf requirements for invoice identification
	anyOfConstraints, ok := schema["anyOf"].([]map[string]interface{})
	require.True(t, ok, "anyOf should be present")
	assert.Len(t, anyOfConstraints, 2, "Should have 2 invoice identification options")
}

// TestInvoiceUpdateSchema tests the invoice update schema
func TestInvoiceUpdateSchema(t *testing.T) {
	schema := InvoiceUpdateSchema()

	// Verify basic schema structure
	assert.Equal(t, "object", schema["type"])

	properties, ok := schema["properties"].(map[string]interface{})
	require.True(t, ok, "Properties should be a map")

	// Verify identification and update fields
	assert.Contains(t, properties, "invoice_id")
	assert.Contains(t, properties, "invoice_number")
	assert.Contains(t, properties, "status")
	assert.Contains(t, properties, "due_date")
	assert.Contains(t, properties, "description")

	// Verify allOf constraints for identification and updates
	allOfConstraints, ok := schema["allOf"].([]map[string]interface{})
	require.True(t, ok, "allOf should be present")
	assert.Len(t, allOfConstraints, 2, "Should have identification and update constraints")
}

// TestInvoiceDeleteSchema tests the invoice deletion schema
func TestInvoiceDeleteSchema(t *testing.T) {
	schema := InvoiceDeleteSchema()

	// Verify basic schema structure
	assert.Equal(t, "object", schema["type"])

	properties, ok := schema["properties"].(map[string]interface{})
	require.True(t, ok, "Properties should be a map")

	// Verify required and optional fields
	assert.Contains(t, properties, "invoice_id")
	assert.Contains(t, properties, "invoice_number")
	assert.Contains(t, properties, "hard_delete")
	assert.Contains(t, properties, "force")

	// Check default values
	hardDeleteField := properties["hard_delete"].(map[string]interface{})
	assert.Equal(t, false, hardDeleteField["default"])

	forceField := properties["force"].(map[string]interface{})
	assert.Equal(t, false, forceField["default"])
}

// TestInvoiceAddItemSchema tests the work item addition schema
func TestInvoiceAddItemSchema(t *testing.T) {
	schema := InvoiceAddItemSchema()

	// Verify basic schema structure
	assert.Equal(t, "object", schema["type"])

	properties, ok := schema["properties"].(map[string]interface{})
	require.True(t, ok, "Properties should be a map")

	// Verify identification and work items fields
	assert.Contains(t, properties, "invoice_id")
	assert.Contains(t, properties, "invoice_number")
	assert.Contains(t, properties, "work_items")

	// Verify work items array structure
	workItemsField := properties["work_items"].(map[string]interface{})
	assert.Equal(t, "array", workItemsField["type"])
	assert.Equal(t, 1, workItemsField["minItems"])

	// Check work item object structure
	items := workItemsField["items"].(map[string]interface{})
	itemProperties := items["properties"].(map[string]interface{})
	assert.Contains(t, itemProperties, "date")
	assert.Contains(t, itemProperties, "hours")
	assert.Contains(t, itemProperties, "rate")
	assert.Contains(t, itemProperties, "description")

	// Check required work item fields
	itemRequired := items["required"].([]string)
	expectedRequired := []string{"date", "hours", "rate", "description"}
	assert.ElementsMatch(t, expectedRequired, itemRequired)
}

// TestInvoiceRemoveItemSchema tests the work item removal schema
func TestInvoiceRemoveItemSchema(t *testing.T) {
	schema := InvoiceRemoveItemSchema()

	// Verify basic schema structure
	assert.Equal(t, "object", schema["type"])

	properties, ok := schema["properties"].(map[string]interface{})
	require.True(t, ok, "Properties should be a map")

	// Verify identification fields
	assert.Contains(t, properties, "invoice_id")
	assert.Contains(t, properties, "invoice_number")
	assert.Contains(t, properties, "work_item_id")
	assert.Contains(t, properties, "work_item_description")
	assert.Contains(t, properties, "work_item_date")

	// Verify control fields
	assert.Contains(t, properties, "remove_all_matching")
	assert.Contains(t, properties, "confirm")

	// Verify allOf constraints
	allOfConstraints, ok := schema["allOf"].([]map[string]interface{})
	require.True(t, ok, "allOf should be present")
	assert.Len(t, allOfConstraints, 2, "Should have invoice and work item identification constraints")
}

// TestGetAllInvoiceSchemas tests schema retrieval function
func TestGetAllInvoiceSchemas(t *testing.T) {
	schemas := GetAllInvoiceSchemas()

	// Verify all expected schemas are present
	expectedSchemas := []string{
		"invoice_create",
		"invoice_list",
		"invoice_show",
		"invoice_update",
		"invoice_delete",
		"invoice_add_item",
		"invoice_remove_item",
	}

	assert.Len(t, schemas, len(expectedSchemas), "Should have all expected schemas")

	for _, schemaName := range expectedSchemas {
		assert.Contains(t, schemas, schemaName, "Should contain schema: %s", schemaName)

		schema := schemas[schemaName]
		assert.NotNil(t, schema, "Schema %s should not be nil", schemaName)
		assert.Equal(t, "object", schema["type"], "Schema %s should be object type", schemaName)
		assert.Contains(t, schema, "properties", "Schema %s should have properties", schemaName)
	}
}

// TestGetInvoiceToolSchema tests individual schema retrieval
func TestGetInvoiceToolSchema(t *testing.T) {
	// Test existing schema
	schema, exists := GetInvoiceToolSchema("invoice_create")
	assert.True(t, exists, "invoice_create schema should exist")
	assert.NotNil(t, schema, "invoice_create schema should not be nil")
	assert.Equal(t, "object", schema["type"], "invoice_create should be object type")

	// Test non-existing schema
	schema, exists = GetInvoiceToolSchema("nonexistent_tool")
	assert.False(t, exists, "nonexistent tool should not exist")
	assert.Nil(t, schema, "nonexistent tool schema should be nil")
}

// TestSchemaValidation tests that schemas are valid JSON Schema Draft 7
func TestSchemaValidation(t *testing.T) {
	schemas := GetAllInvoiceSchemas()

	for schemaName, schema := range schemas {
		// Basic structure validation
		assert.Equal(t, "object", schema["type"], "Schema %s should be object type", schemaName)

		properties, ok := schema["properties"].(map[string]interface{})
		assert.True(t, ok, "Schema %s properties should be map", schemaName)
		assert.NotEmpty(t, properties, "Schema %s should have properties", schemaName)

		// Check additionalProperties is explicitly set to false
		if additionalProps, exists := schema["additionalProperties"]; exists {
			assert.False(t, additionalProps.(bool), "Schema %s should not allow additional properties", schemaName)
		}

		// Verify field descriptions exist
		for fieldName, fieldDef := range properties {
			fieldMap, ok := fieldDef.(map[string]interface{})
			if !ok {
				continue
			}

			description, hasDesc := fieldMap["description"]
			assert.True(t, hasDesc, "Field %s in schema %s should have description", fieldName, schemaName)
			assert.NotEmpty(t, description, "Field %s description in schema %s should not be empty", fieldName, schemaName)
		}
	}
}
