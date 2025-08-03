// Package schemas provides JSON schema definitions for client management MCP tools.
//
// This package contains comprehensive schema definitions optimized for natural language
// interaction with Claude. Each schema is designed to be intuitive and provides clear
// error messages for validation failures.
//
// Key design principles:
// - Parameter names map naturally to user requests ("client_name" vs "clientID")
// - Support for multiple identification methods (ID, name, email)
// - Flexible contact information management with proper validation
// - Comprehensive validation with helpful error messages
// - Real-world examples demonstrating various usage patterns
//
// All schemas follow JSON Schema Draft 7 specification and integrate with the
// validation system from the tools package.
package schemas

// ClientCreateSchema defines the JSON schema for creating clients.
//
// This schema supports comprehensive client creation with contact information
// validation and business detail management. It provides natural language
// parameter names and extensive validation for professional client relationships.
func ClientCreateSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Client name or company name. This will be displayed on invoices and used for client identification.",
				"minLength":   1,
				"maxLength":   200,
				"examples":    []string{"Acme Corporation", "John Smith Consulting", "Tech Solutions Inc"},
			},
			"email": map[string]interface{}{
				"type":        "string",
				"format":      "email",
				"description": "Primary email address for the client. Must be unique across all clients and will be used for invoice delivery.",
				"examples":    []string{"contact@acme.com", "john@smithconsulting.com", "billing@techsolutions.com"},
			},
			"phone": map[string]interface{}{
				"type":        "string",
				"description": "Phone number for the client. Can include country codes and various formats.",
				"minLength":   10,
				"maxLength":   20,
				"pattern":     "^[\\d\\s\\+\\-\\(\\)\\.\\/ext]+$",
				"examples":    []string{"+1-555-123-4567", "(555) 987-6543", "+49-30-12345678"},
			},
			"address": map[string]interface{}{
				"type":        "string",
				"description": "Physical or mailing address for the client. Used for formal correspondence and invoicing.",
				"maxLength":   500,
				"examples": []string{
					"123 Business Ave, Suite 200, Metro City, MC 12345",
					"456 Main Street, Small Town, ST 67890",
					"Hauptstraße 42, 10117 Berlin, Germany",
				},
			},
			"tax_id": map[string]interface{}{
				"type":        "string",
				"description": "Tax identification number (EIN, VAT number, etc.) for business clients. Used for tax reporting and compliance.",
				"maxLength":   50,
				"examples":    []string{"EIN-12-3456789", "DE123456789", "VAT-GB123456789"},
			},
		},
		"required":             []string{"name", "email"},
		"additionalProperties": false,
	}
}

// ClientListSchema defines the JSON schema for listing/filtering clients.
//
// This schema provides comprehensive filtering options for client discovery
// with natural language-friendly parameter names and flexible search criteria.
func ClientListSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"active_only": map[string]interface{}{
				"type":        "boolean",
				"default":     true,
				"description": "Show only active clients. Set to false to include deactivated clients.",
			},
			"inactive_only": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Show only inactive clients. Useful for reactivation campaigns and cleanup.",
			},
			"name_search": map[string]interface{}{
				"type":        "string",
				"description": "Search clients by name (partial matches supported). Case-insensitive search.",
				"examples":    []string{"Acme", "Tech", "John", "Corp"},
			},
			"email_search": map[string]interface{}{
				"type":        "string",
				"description": "Search clients by email address (partial matches supported).",
				"examples":    []string{"@acme.com", "contact", "billing"},
			},
			"sort_by": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"name", "email", "created_date", "last_invoice_date", "total_invoiced"},
				"default":     "name",
				"description": "Field to sort results by.",
			},
			"sort_order": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"asc", "desc"},
				"default":     "asc",
				"description": "Sort order: ascending or descending.",
			},
			"limit": map[string]interface{}{
				"type":        "number",
				"minimum":     1,
				"maximum":     1000,
				"default":     50,
				"description": "Maximum number of clients to return.",
			},
			"offset": map[string]interface{}{
				"type":        "number",
				"minimum":     0,
				"default":     0,
				"description": "Number of clients to skip (for pagination).",
			},
			"output_format": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"table", "json", "csv"},
				"default":     "table",
				"description": "Output format for the results.",
			},
			"include_invoices": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Include invoice count and total amounts for each client.",
			},
			"include_totals": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Include summary statistics and total amounts across all clients.",
			},
			"show_contact_info": map[string]interface{}{
				"type":        "boolean",
				"default":     true,
				"description": "Include contact information (phone, address) in the output.",
			},
		},
		"additionalProperties": false,
	}
}

// ClientShowSchema defines the JSON schema for showing client details.
//
// This schema supports client lookup by multiple identifiers and provides
// options for different levels of detail and output formats.
func ClientShowSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"client_id": map[string]interface{}{
				"type":        "string",
				"description": "Client ID to display details for.",
				"minLength":   1,
				"examples":    []string{"client_123", "acme-corp-id"},
			},
			"client_name": map[string]interface{}{
				"type":        "string",
				"description": "Client name to display details for (alternative to client_id).",
				"minLength":   1,
				"examples":    []string{"Acme Corporation", "John Smith", "Tech Solutions"},
			},
			"client_email": map[string]interface{}{
				"type":        "string",
				"format":      "email",
				"description": "Client email to display details for (alternative to client_id or name).",
				"examples":    []string{"contact@acme.com", "john@example.com"},
			},
			"output_format": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"text", "json", "yaml"},
				"default":     "text",
				"description": "Output format for client details.",
			},
			"include_invoices": map[string]interface{}{
				"type":        "boolean",
				"default":     true,
				"description": "Include detailed invoice history in the output.",
			},
			"include_totals": map[string]interface{}{
				"type":        "boolean",
				"default":     true,
				"description": "Include financial summary totals and statistics.",
			},
			"include_overdue": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Highlight overdue invoices and payment issues.",
			},
			"summary_only": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Show only summary information without detailed invoice listing.",
			},
			"invoice_limit": map[string]interface{}{
				"type":        "number",
				"minimum":     1,
				"maximum":     100,
				"default":     10,
				"description": "Maximum number of recent invoices to include.",
			},
		},
		"anyOf": []map[string]interface{}{
			{"required": []string{"client_id"}},
			{"required": []string{"client_name"}},
			{"required": []string{"client_email"}},
		},
		"additionalProperties": false,
	}
}

// ClientUpdateSchema defines the JSON schema for updating clients.
//
// This schema supports selective updates of client properties with validation
// to ensure business rules are enforced and data integrity is maintained.
func ClientUpdateSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"client_id": map[string]interface{}{
				"type":        "string",
				"description": "Client ID to update.",
				"minLength":   1,
				"examples":    []string{"client_123", "acme-corp-id"},
			},
			"client_name": map[string]interface{}{
				"type":        "string",
				"description": "Client name to update (alternative to client_id).",
				"minLength":   1,
				"examples":    []string{"Acme Corporation", "John Smith"},
			},
			"client_email": map[string]interface{}{
				"type":        "string",
				"format":      "email",
				"description": "Client email to identify the client to update (alternative to client_id or name).",
				"examples":    []string{"contact@acme.com", "john@example.com"},
			},
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Update client name or company name.",
				"minLength":   1,
				"maxLength":   200,
				"examples":    []string{"Updated Company Name", "John Smith LLC"},
			},
			"email": map[string]interface{}{
				"type":        "string",
				"format":      "email",
				"description": "Update client email address. Must be unique across all clients.",
				"examples":    []string{"newemail@company.com", "updated@example.com"},
			},
			"phone": map[string]interface{}{
				"type":        "string",
				"description": "Update client phone number. Can include country codes and various formats.",
				"minLength":   10,
				"maxLength":   20,
				"pattern":     "^[\\d\\s\\+\\-\\(\\)\\.\\/ext]+$",
				"examples":    []string{"+1-555-999-8888", "(555) 111-2222"},
			},
			"address": map[string]interface{}{
				"type":        "string",
				"description": "Update client address information.",
				"maxLength":   500,
				"examples": []string{
					"789 New Business Plaza, Suite 500, Metro City, MC 67890",
					"Updated Address, New City, NC 12345",
				},
			},
			"tax_id": map[string]interface{}{
				"type":        "string",
				"description": "Update tax identification number for the client.",
				"maxLength":   50,
				"examples":    []string{"EIN-98-7654321", "VAT-GB987654321"},
			},
			"activate": map[string]interface{}{
				"type":        "boolean",
				"description": "Reactivate a deactivated client. Cannot be used with deactivate.",
			},
			"deactivate": map[string]interface{}{
				"type":        "boolean",
				"description": "Deactivate the client (soft delete). Cannot be used with activate.",
			},
		},
		"allOf": []map[string]interface{}{
			{
				"anyOf": []map[string]interface{}{
					{"required": []string{"client_id"}},
					{"required": []string{"client_name"}},
					{"required": []string{"client_email"}},
				},
			},
			{
				"anyOf": []map[string]interface{}{
					{"required": []string{"name"}},
					{"required": []string{"email"}},
					{"required": []string{"phone"}},
					{"required": []string{"address"}},
					{"required": []string{"tax_id"}},
					{"required": []string{"activate"}},
					{"required": []string{"deactivate"}},
				},
			},
			{
				"not": map[string]interface{}{
					"allOf": []map[string]interface{}{
						{"required": []string{"activate"}},
						{"required": []string{"deactivate"}},
					},
				},
			},
		},
		"additionalProperties": false,
	}
}

// ClientDeleteSchema defines the JSON schema for deleting clients.
//
// This schema supports both soft and hard deletion with appropriate
// safety confirmations and business rule validation.
func ClientDeleteSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"client_id": map[string]interface{}{
				"type":        "string",
				"description": "Client ID to delete.",
				"minLength":   1,
				"examples":    []string{"client_123", "test-client"},
			},
			"client_name": map[string]interface{}{
				"type":        "string",
				"description": "Client name to delete (alternative to client_id).",
				"minLength":   1,
				"examples":    []string{"Test Client", "Inactive Corp"},
			},
			"client_email": map[string]interface{}{
				"type":        "string",
				"format":      "email",
				"description": "Client email to identify the client to delete (alternative to client_id or name).",
				"examples":    []string{"test@example.com", "inactive@corp.com"},
			},
			"soft_delete": map[string]interface{}{
				"type":        "boolean",
				"default":     true,
				"description": "Deactivate client instead of permanent deletion (recommended). Preserves audit trail and historical data.",
			},
			"hard_delete": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Permanently delete client data (cannot be undone). Use with extreme caution.",
			},
			"force": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Skip confirmation prompts (use with caution in automated scripts).",
			},
			"preserve_data": map[string]interface{}{
				"type":        "boolean",
				"default":     true,
				"description": "Preserve all client data and invoice history even with soft delete.",
			},
			"cascade_invoices": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Also delete associated invoices (only available with hard_delete and force).",
			},
		},
		"allOf": []map[string]interface{}{
			{
				"anyOf": []map[string]interface{}{
					{"required": []string{"client_id"}},
					{"required": []string{"client_name"}},
					{"required": []string{"client_email"}},
				},
			},
			{
				"not": map[string]interface{}{
					"allOf": []map[string]interface{}{
						{"required": []string{"soft_delete"}},
						{"required": []string{"hard_delete"}},
					},
				},
			},
		},
		"additionalProperties": false,
	}
}

// GetAllClientSchemas returns all client-related schemas mapped by tool name.
//
// This function provides a centralized way to access all client tool schemas
// for registration with the MCP tool system.
//
// Returns:
// - map[string]map[string]interface{}: Map of tool names to their JSON schemas
//
// Notes:
// - Schema names match the corresponding tool names exactly
// - All schemas follow JSON Schema Draft 7 specification
// - Schemas are optimized for Claude natural language interaction
func GetAllClientSchemas() map[string]map[string]interface{} {
	return map[string]map[string]interface{}{
		"client_create": ClientCreateSchema(),
		"client_list":   ClientListSchema(),
		"client_show":   ClientShowSchema(),
		"client_update": ClientUpdateSchema(),
		"client_delete": ClientDeleteSchema(),
	}
}

// GetClientToolSchema returns the schema for a specific client tool.
//
// This function provides a way to get the schema for a specific client tool
// for use with external validation systems.
//
// Parameters:
// - toolName: Name of the client tool to get schema for
//
// Returns:
// - map[string]interface{}: JSON schema for the tool, or nil if not found
// - bool: True if tool exists, false otherwise
//
// Notes:
// - Returns nil schema and false for unknown tool names
// - Schema can be used with any JSON Schema Draft 7 validator
// - Tool names match the MCP tool names exactly
func GetClientToolSchema(toolName string) (map[string]interface{}, bool) {
	schemas := GetAllClientSchemas()
	schema, exists := schemas[toolName]
	return schema, exists
}
