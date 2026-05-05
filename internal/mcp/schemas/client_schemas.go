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
		keyType: keyObject,
		keyProperties: map[string]interface{}{
			keyName: map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Client name or company name. This will be displayed on invoices and used for client identification.",
				keyMinLength:   1.0,
				keyMaxLength:   200.0,
				keyExamples:    []interface{}{"Acme Corporation", "John Smith Consulting", "Tech Solutions Inc"},
			},
			keyEmail: map[string]interface{}{
				keyType:        typeString,
				keyFormat:      keyEmail,
				keyDescription: "Primary email address for the client. Must be unique across all clients and will be used for invoice delivery.",
				keyExamples:    []interface{}{"contact@acme.com", "john@smithconsulting.com", "billing@techsolutions.com"},
			},
			"phone": map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Phone number for the client. Can include country codes and various formats.",
				keyMinLength:   10.0,
				keyMaxLength:   20.0,
				"pattern":      "^[\\d\\s\\+\\-\\(\\)\\.\\/ext]+$",
				keyExamples:    []interface{}{"+1-555-123-4567", "(555) 987-6543", "+49-30-12345678"},
			},
			"address": map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Physical or mailing address for the client. Used for formal correspondence and invoicing.",
				keyMaxLength:   500.0,
				keyExamples: []interface{}{
					"123 Business Ave, Suite 200, Metro City, MC 12345",
					"456 Main Street, Small Town, ST 67890",
					"Hauptstraße 42, 10117 Berlin, Germany",
				},
			},
			"tax_id": map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Tax identification number (EIN, VAT number, etc.) for business clients. Used for tax reporting and compliance.",
				keyMaxLength:   50.0,
				keyExamples:    []interface{}{"EIN-12-3456789", "DE123456789", "VAT-GB123456789"},
			},
			"approver_contacts": map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Approver contacts (names or departments) who should be shown on the invoice. Can be multiple people separated by commas.",
				keyMaxLength:   500.0,
				keyExamples:    []interface{}{"John Doe, Finance Dept", "Jane Smith", "HR Department, Accounting Team"},
			},
		},
		keyRequired:             []interface{}{keyName, keyEmail},
		keyAdditionalProperties: false,
	}
}

// ClientListSchema defines the JSON schema for listing/filtering clients.
//
// This schema provides comprehensive filtering options for client discovery
// with natural language-friendly parameter names and flexible search criteria.
func ClientListSchema() map[string]interface{} {
	return map[string]interface{}{
		keyType: keyObject,
		keyProperties: map[string]interface{}{
			"active_only": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     true,
				keyDescription: "Show only active clients. Set to false to include deactivated clients.",
			},
			"inactive_only": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Show only inactive clients. Useful for reactivation campaigns and cleanup.",
			},
			"name_search": map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Search clients by name (partial matches supported). Case-insensitive search.",
				keyExamples:    []interface{}{"Acme", "Tech", "John", "Corp"},
			},
			"email_search": map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Search clients by email address (partial matches supported).",
				keyExamples:    []interface{}{"@acme.com", "contact", "billing"},
			},
			"sort_by": map[string]interface{}{
				keyType:        typeString,
				keyEnum:        []interface{}{keyName, keyEmail, "created_date", "last_invoice_date", "total_invoiced"},
				keyDefault:     keyName,
				keyDescription: "Field to sort results by.",
			},
			"sort_order": map[string]interface{}{
				keyType:        typeString,
				keyEnum:        []interface{}{"asc", "desc"},
				keyDefault:     "asc",
				keyDescription: "Sort order: ascending or descending.",
			},
			"limit": map[string]interface{}{
				keyType:        typeNumber,
				keyMinimum:     1.0,
				keyMaximum:     1000,
				keyDefault:     50,
				keyDescription: "Maximum number of clients to return.",
			},
			"offset": map[string]interface{}{
				keyType:        typeNumber,
				keyMinimum:     0,
				keyDefault:     0,
				keyDescription: "Number of clients to skip (for pagination).",
			},
			"output_format": map[string]interface{}{
				keyType:        typeString,
				keyEnum:        []interface{}{"table", typeJSON, "csv"},
				keyDefault:     "table",
				keyDescription: "Output format for the results.",
			},
			"include_invoices": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Include invoice count and total amounts for each client.",
			},
			"include_totals": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Include summary statistics and total amounts across all clients.",
			},
			"show_contact_info": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     true,
				keyDescription: "Include contact information (phone, address) in the output.",
			},
		},
		keyAdditionalProperties: false,
	}
}

// ClientShowSchema defines the JSON schema for showing client details.
//
// This schema supports client lookup by multiple identifiers and provides
// options for different levels of detail and output formats.
func ClientShowSchema() map[string]interface{} {
	return map[string]interface{}{
		keyType: keyObject,
		keyProperties: map[string]interface{}{
			keyClientID: map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Client ID to display details for.",
				keyMinLength:   1.0,
				keyExamples:    []interface{}{"client_123", "acme-corp-id"},
			},
			keyClientName: map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Client name to display details for (alternative to client_id).",
				keyMinLength:   1.0,
				keyExamples:    []interface{}{"Acme Corporation", "John Smith", "Tech Solutions"},
			},
			"client_email": map[string]interface{}{
				keyType:        typeString,
				keyFormat:      keyEmail,
				keyDescription: "Client email to display details for (alternative to client_id or name).",
				keyExamples:    []interface{}{"contact@acme.com", "john@example.com"},
			},
			"output_format": map[string]interface{}{
				keyType:        typeString,
				keyEnum:        []interface{}{"text", typeJSON, "yaml"},
				keyDefault:     "text",
				keyDescription: "Output format for client details.",
			},
			"include_invoices": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     true,
				keyDescription: "Include detailed invoice history in the output.",
			},
			"include_totals": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     true,
				keyDescription: "Include financial summary totals and statistics.",
			},
			"include_overdue": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Highlight overdue invoices and payment issues.",
			},
			"summary_only": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Show only summary information without detailed invoice listing.",
			},
			"invoice_limit": map[string]interface{}{
				keyType:        typeNumber,
				keyMinimum:     1.0,
				keyMaximum:     100.0,
				keyDefault:     10,
				keyDescription: "Maximum number of recent invoices to include.",
			},
		},
		keyAdditionalProperties: false,
	}
}

// ClientUpdateSchema defines the JSON schema for updating clients.
//
// This schema supports selective updates of client properties with validation
// to ensure business rules are enforced and data integrity is maintained.
func ClientUpdateSchema() map[string]interface{} {
	return map[string]interface{}{
		keyType: keyObject,
		keyProperties: map[string]interface{}{
			keyClientID: map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Client ID to update.",
				keyMinLength:   1.0,
				keyExamples:    []interface{}{"client_123", "acme-corp-id"},
			},
			keyClientName: map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Client name to update (alternative to client_id).",
				keyMinLength:   1.0,
				keyExamples:    []interface{}{"Acme Corporation", "John Smith"},
			},
			"client_email": map[string]interface{}{
				keyType:        typeString,
				keyFormat:      keyEmail,
				keyDescription: "Client email to identify the client to update (alternative to client_id or name).",
				keyExamples:    []interface{}{"contact@acme.com", "john@example.com"},
			},
			keyName: map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Update client name or company name.",
				keyMinLength:   1.0,
				keyMaxLength:   200.0,
				keyExamples:    []interface{}{"Updated Company Name", "John Smith LLC"},
			},
			keyEmail: map[string]interface{}{
				keyType:        typeString,
				keyFormat:      keyEmail,
				keyDescription: "Update client email address. Must be unique across all clients.",
				keyExamples:    []interface{}{"newemail@company.com", "updated@example.com"},
			},
			"phone": map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Update client phone number. Can include country codes and various formats.",
				keyMinLength:   10.0,
				keyMaxLength:   20.0,
				"pattern":      "^[\\d\\s\\+\\-\\(\\)\\.\\/ext]+$",
				keyExamples:    []interface{}{"+1-555-999-8888", "(555) 111-2222"},
			},
			"address": map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Update client address information.",
				keyMaxLength:   500.0,
				keyExamples: []interface{}{
					"789 New Business Plaza, Suite 500, Metro City, MC 67890",
					"Updated Address, New City, NC 12345",
				},
			},
			"tax_id": map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Update tax identification number for the client.",
				keyMaxLength:   50.0,
				keyExamples:    []interface{}{"EIN-98-7654321", "VAT-GB987654321"},
			},
			"approver_contacts": map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Update approver contacts (names or departments) who should be shown on the invoice.",
				keyMaxLength:   500.0,
				keyExamples:    []interface{}{"John Doe, Finance Dept", "Jane Smith", "HR Department, Accounting Team"},
			},
			"activate": map[string]interface{}{
				keyType:        typeBoolean,
				keyDescription: "Reactivate a deactivated client. Cannot be used with deactivate.",
			},
			"deactivate": map[string]interface{}{
				keyType:        typeBoolean,
				keyDescription: "Deactivate the client (soft delete). Cannot be used with activate.",
			},
		},
		keyAdditionalProperties: false,
	}
}

// ClientDeleteSchema defines the JSON schema for deleting clients.
//
// This schema supports both soft and hard deletion with appropriate
// safety confirmations and business rule validation.
func ClientDeleteSchema() map[string]interface{} {
	return map[string]interface{}{
		keyType: keyObject,
		keyProperties: map[string]interface{}{
			keyClientID: map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Client ID to delete.",
				keyMinLength:   1.0,
				keyExamples:    []interface{}{"client_123", "test-client"},
			},
			keyClientName: map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Client name to delete (alternative to client_id).",
				keyMinLength:   1.0,
				keyExamples:    []interface{}{"Test Client", "Inactive Corp"},
			},
			"client_email": map[string]interface{}{
				keyType:        typeString,
				keyFormat:      keyEmail,
				keyDescription: "Client email to identify the client to delete (alternative to client_id or name).",
				keyExamples:    []interface{}{"test@example.com", "inactive@corp.com"},
			},
			"soft_delete": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     true,
				keyDescription: "Deactivate client instead of permanent deletion (recommended). Preserves audit trail and historical data.",
			},
			"hard_delete": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Permanently delete client data (cannot be undone). Use with extreme caution.",
			},
			"force": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Skip confirmation prompts (use with caution in automated scripts).",
			},
			"preserve_data": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     true,
				keyDescription: "Preserve all client data and invoice history even with soft delete.",
			},
			"cascade_invoices": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Also delete associated invoices (only available with hard_delete and force).",
			},
		},
		keyAdditionalProperties: false,
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
