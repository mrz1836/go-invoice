// Package schemas provides JSON schema definitions for invoice management MCP tools.
//
// This package contains comprehensive schema definitions optimized for natural language
// interaction with Claude. Each schema is designed to be intuitive and provides clear
// error messages for validation failures.
//
// Key design principles:
// - Parameter names map naturally to user requests ("client_name" vs "clientID")
// - Support for multiple input formats (dates, client references, etc.)
// - Flexible search and filtering options
// - Comprehensive validation with helpful error messages
// - Real-world examples demonstrating various usage patterns
//
// All schemas follow JSON Schema Draft 7 specification and integrate with the
// validation system from the tools package.
package schemas

import (
	"time"
)

// InvoiceCreateSchema defines the JSON schema for creating invoices.
//
// This schema supports both simple invoice creation and creation with work items.
// It provides flexible client resolution (by name, ID, or email) and supports
// multiple date formats for natural language interaction.
func InvoiceCreateSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"client_name": map[string]interface{}{
				"type":        "string",
				"description": "Client name or partial name for invoice. If multiple clients match, you'll be prompted to be more specific.",
				"minLength":   1,
				"maxLength":   200,
				"examples":    []string{"Acme Corp", "John Smith", "Tech Solutions"},
			},
			"client_id": map[string]interface{}{
				"type":        "string",
				"description": "Exact client ID for invoice. Use this when you know the specific client identifier.",
				"pattern":     "^[A-Za-z0-9_-]+$",
				"examples":    []string{"client_123", "acme-corp-id"},
			},
			"client_email": map[string]interface{}{
				"type":        "string",
				"format":      "email",
				"description": "Client email address to identify the client for this invoice.",
				"examples":    []string{"contact@acme.com", "john@techsolutions.com"},
			},
			"invoice_date": map[string]interface{}{
				"type":        "string",
				"format":      "date",
				"description": "Invoice date in YYYY-MM-DD format. Defaults to today if not specified.",
				"examples":    []string{time.Now().Format("2006-01-02"), "2025-01-15", "2025-08-03"},
			},
			"due_date": map[string]interface{}{
				"type":        "string",
				"format":      "date",
				"description": "Payment due date in YYYY-MM-DD format. If not provided, calculated based on default payment terms.",
				"examples":    []string{time.Now().AddDate(0, 0, 30).Format("2006-01-02"), "2025-02-15", "2025-09-02"},
			},
			"description": map[string]interface{}{
				"type":        "string",
				"description": "Optional description for the invoice (e.g., project name, billing period).",
				"maxLength":   500,
				"examples":    []string{"January 2025 consulting services", "Website development project", "Monthly retainer"},
			},
			"create_client_if_missing": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to create a new client if the specified client is not found.",
				"default":     false,
			},
			"new_client_email": map[string]interface{}{
				"type":        "string",
				"format":      "email",
				"description": "Email address for new client (required when create_client_if_missing is true).",
				"examples":    []string{"newclient@company.com"},
			},
			"new_client_address": map[string]interface{}{
				"type":        "string",
				"description": "Address for new client (optional when creating new client).",
				"maxLength":   500,
				"examples":    []string{"123 Main St, City, State 12345"},
			},
			"new_client_phone": map[string]interface{}{
				"type":        "string",
				"description": "Phone number for new client (optional when creating new client).",
				"pattern":     "^[\\d\\s\\+\\-\\(\\)\\.\\/ext]+$",
				"examples":    []string{"+1-555-123-4567", "(555) 123-4567"},
			},
			"work_items": map[string]interface{}{
				"type":        "array",
				"description": "Optional work items to add to the invoice upon creation.",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"date": map[string]interface{}{
							"type":        "string",
							"format":      "date",
							"description": "Date when work was performed (YYYY-MM-DD).",
						},
						"hours": map[string]interface{}{
							"type":        "number",
							"minimum":     0.01,
							"maximum":     24.0,
							"description": "Number of hours worked (decimal allowed, e.g., 1.5 for 1 hour 30 minutes).",
						},
						"rate": map[string]interface{}{
							"type":        "number",
							"minimum":     0.01,
							"description": "Hourly rate for this work item.",
						},
						"description": map[string]interface{}{
							"type":        "string",
							"minLength":   1,
							"maxLength":   500,
							"description": "Description of work performed.",
						},
					},
					"required": []string{"date", "hours", "rate", "description"},
				},
			},
		},
		"anyOf": []map[string]interface{}{
			{
				"required": []string{"client_name"},
			},
			{
				"required": []string{"client_id"},
			},
			{
				"required": []string{"client_email"},
			},
		},
		"additionalProperties": false,
	}
}

// InvoiceListSchema defines the JSON schema for listing/filtering invoices.
//
// This schema provides comprehensive filtering options for invoice discovery
// with natural language-friendly parameter names and flexible search criteria.
func InvoiceListSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"status": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"draft", "sent", "paid", "overdue", "voided"},
				"description": "Filter invoices by status. Leave empty to show all statuses.",
				"examples":    []string{"paid", "overdue", "draft"},
			},
			"client_name": map[string]interface{}{
				"type":        "string",
				"description": "Filter by client name (partial matches supported).",
				"examples":    []string{"Acme", "Tech Solutions", "John"},
			},
			"client_id": map[string]interface{}{
				"type":        "string",
				"description": "Filter by exact client ID.",
				"examples":    []string{"client_123", "acme-corp"},
			},
			"from_date": map[string]interface{}{
				"type":        "string",
				"format":      "date",
				"description": "Show invoices from this date onwards (YYYY-MM-DD). Filters by invoice date.",
				"examples":    []string{"2025-01-01", "2025-07-01"},
			},
			"to_date": map[string]interface{}{
				"type":        "string",
				"format":      "date",
				"description": "Show invoices up to this date (YYYY-MM-DD). Filters by invoice date.",
				"examples":    []string{"2025-12-31", "2025-08-31"},
			},
			"sort_by": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"date", "amount", "status", "client", "due_date"},
				"default":     "date",
				"description": "Field to sort results by.",
			},
			"sort_order": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"asc", "desc"},
				"default":     "desc",
				"description": "Sort order: ascending or descending.",
			},
			"limit": map[string]interface{}{
				"type":        "number",
				"minimum":     1,
				"maximum":     1000,
				"description": "Maximum number of invoices to return. Default is 50.",
				"default":     50,
			},
			"output_format": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"table", "json", "csv"},
				"default":     "table",
				"description": "Output format for the results.",
			},
			"include_summary": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Include summary statistics (total amounts, counts by status).",
			},
		},
		"additionalProperties": false,
	}
}

// InvoiceShowSchema defines the JSON schema for showing invoice details.
//
// This schema supports invoice lookup by multiple identifiers and provides
// options for different levels of detail and output formats.
func InvoiceShowSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"invoice_id": map[string]interface{}{
				"type":        "string",
				"description": "Invoice ID to display details for.",
				"minLength":   1,
				"examples":    []string{"INV-001", "invoice_abc123"},
			},
			"invoice_number": map[string]interface{}{
				"type":        "string",
				"description": "Invoice number to display details for (alternative to invoice_id).",
				"minLength":   1,
				"examples":    []string{"INV-001", "2025-001"},
			},
			"output_format": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"text", "json", "yaml"},
				"default":     "text",
				"description": "Output format for invoice details.",
			},
			"show_work_items": map[string]interface{}{
				"type":        "boolean",
				"default":     true,
				"description": "Include detailed work items in the output.",
			},
			"show_client_details": map[string]interface{}{
				"type":        "boolean",
				"default":     true,
				"description": "Include full client information in the output.",
			},
		},
		"anyOf": []map[string]interface{}{
			{
				"required": []string{"invoice_id"},
			},
			{
				"required": []string{"invoice_number"},
			},
		},
		"additionalProperties": false,
	}
}

// InvoiceUpdateSchema defines the JSON schema for updating invoices.
//
// This schema supports selective updates of invoice properties with validation
// to ensure business rules are enforced (e.g., no updates to paid invoices).
func InvoiceUpdateSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"invoice_id": map[string]interface{}{
				"type":        "string",
				"description": "Invoice ID to update.",
				"minLength":   1,
				"examples":    []string{"INV-001", "invoice_abc123"},
			},
			"invoice_number": map[string]interface{}{
				"type":        "string",
				"description": "Invoice number to update (alternative to invoice_id).",
				"minLength":   1,
				"examples":    []string{"INV-001", "2025-001"},
			},
			"status": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"draft", "sent", "paid", "overdue", "voided"},
				"description": "Update invoice status. Note: certain status transitions may be restricted.",
				"examples":    []string{"sent", "paid"},
			},
			"due_date": map[string]interface{}{
				"type":        "string",
				"format":      "date",
				"description": "Update payment due date (YYYY-MM-DD).",
				"examples":    []string{"2025-02-28", "2025-09-15"},
			},
			"description": map[string]interface{}{
				"type":        "string",
				"maxLength":   500,
				"description": "Update invoice description.",
				"examples":    []string{"Updated: January 2025 consulting services", "Q1 2025 development work"},
			},
		},
		"allOf": []map[string]interface{}{
			{
				"anyOf": []map[string]interface{}{
					{
						"required": []string{"invoice_id"},
					},
					{
						"required": []string{"invoice_number"},
					},
				},
			},
			{
				"anyOf": []map[string]interface{}{
					{
						"required": []string{"status"},
					},
					{
						"required": []string{"due_date"},
					},
					{
						"required": []string{"description"},
					},
				},
			},
		},
		"additionalProperties": false,
	}
}

// InvoiceDeleteSchema defines the JSON schema for deleting invoices.
//
// This schema supports both soft and hard deletion with appropriate
// safety confirmations and business rule validation.
func InvoiceDeleteSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"invoice_id": map[string]interface{}{
				"type":        "string",
				"description": "Invoice ID to delete.",
				"minLength":   1,
				"examples":    []string{"INV-001", "invoice_abc123"},
			},
			"invoice_number": map[string]interface{}{
				"type":        "string",
				"description": "Invoice number to delete (alternative to invoice_id).",
				"minLength":   1,
				"examples":    []string{"INV-001", "2025-001"},
			},
			"hard_delete": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Permanently delete invoice (cannot be undone). Default is soft delete.",
			},
			"force": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Skip confirmation prompt (use with caution).",
			},
		},
		"additionalProperties": false,
	}
}

// InvoiceAddItemSchema defines the JSON schema for adding work items to invoices.
//
// This schema supports adding single or multiple work items with comprehensive
// validation and natural date/time entry formats.
func InvoiceAddItemSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"invoice_id": map[string]interface{}{
				"type":        "string",
				"description": "Invoice ID to add work items to.",
				"minLength":   1,
				"examples":    []string{"INV-001", "invoice_abc123"},
			},
			"invoice_number": map[string]interface{}{
				"type":        "string",
				"description": "Invoice number to add work items to (alternative to invoice_id).",
				"minLength":   1,
				"examples":    []string{"INV-001", "2025-001"},
			},
			"work_items": map[string]interface{}{
				"type":        "array",
				"minItems":    1,
				"description": "Work items to add to the invoice.",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"date": map[string]interface{}{
							"type":        "string",
							"format":      "date",
							"description": "Date when work was performed (YYYY-MM-DD).",
							"examples":    []string{time.Now().Format("2006-01-02"), "2025-08-01"},
						},
						"hours": map[string]interface{}{
							"type":        "number",
							"minimum":     0.01,
							"maximum":     24.0,
							"description": "Number of hours worked (decimal allowed, e.g., 1.5 for 1 hour 30 minutes).",
							"examples":    []interface{}{8.0, 4.5, 2.25},
						},
						"rate": map[string]interface{}{
							"type":        "number",
							"minimum":     0.01,
							"description": "Hourly rate for this work item.",
							"examples":    []interface{}{75.0, 125.0, 200.0},
						},
						"description": map[string]interface{}{
							"type":        "string",
							"minLength":   1,
							"maxLength":   500,
							"description": "Description of work performed.",
							"examples":    []string{"Frontend development", "Bug fixes and testing", "Client meeting and planning"},
						},
					},
					"required": []string{"date", "hours", "rate", "description"},
				},
			},
		},
		"additionalProperties": false,
	}
}

// InvoiceRemoveItemSchema defines the JSON schema for removing work items from invoices.
//
// This schema supports removing work items by ID or by matching criteria
// with appropriate safety validations.
func InvoiceRemoveItemSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"invoice_id": map[string]interface{}{
				"type":        "string",
				"description": "Invoice ID to remove work items from.",
				"minLength":   1,
				"examples":    []string{"INV-001", "invoice_abc123"},
			},
			"invoice_number": map[string]interface{}{
				"type":        "string",
				"description": "Invoice number to remove work items from (alternative to invoice_id).",
				"minLength":   1,
				"examples":    []string{"INV-001", "2025-001"},
			},
			"work_item_id": map[string]interface{}{
				"type":        "string",
				"description": "Specific work item ID to remove.",
				"minLength":   1,
				"examples":    []string{"work_item_123", "wi_abc"},
			},
			"work_item_description": map[string]interface{}{
				"type":        "string",
				"description": "Remove work items matching this description (partial matches supported).",
				"minLength":   1,
				"examples":    []string{"Frontend development", "Bug fixes"},
			},
			"work_item_date": map[string]interface{}{
				"type":        "string",
				"format":      "date",
				"description": "Remove work items from this specific date (YYYY-MM-DD).",
				"examples":    []string{"2025-08-01", "2025-07-15"},
			},
			"remove_all_matching": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Remove all work items matching the criteria (vs. just the first match).",
			},
			"confirm": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Confirm removal without additional prompts.",
			},
		},
		"allOf": []map[string]interface{}{
			{
				"anyOf": []map[string]interface{}{
					{
						"required": []string{"invoice_id"},
					},
					{
						"required": []string{"invoice_number"},
					},
				},
			},
			{
				"anyOf": []map[string]interface{}{
					{
						"required": []string{"work_item_id"},
					},
					{
						"required": []string{"work_item_description"},
					},
					{
						"required": []string{"work_item_date"},
					},
				},
			},
		},
		"additionalProperties": false,
	}
}

// GetAllInvoiceSchemas returns all invoice-related schemas mapped by tool name.
//
// This function provides a centralized way to access all invoice tool schemas
// for registration with the MCP tool system.
//
// Returns:
// - map[string]map[string]interface{}: Map of tool names to their JSON schemas
//
// Notes:
// - Schema names match the corresponding tool names exactly
// - All schemas follow JSON Schema Draft 7 specification
// - Schemas are optimized for Claude natural language interaction
func GetAllInvoiceSchemas() map[string]map[string]interface{} {
	return map[string]map[string]interface{}{
		"invoice_create":      InvoiceCreateSchema(),
		"invoice_list":        InvoiceListSchema(),
		"invoice_show":        InvoiceShowSchema(),
		"invoice_update":      InvoiceUpdateSchema(),
		"invoice_delete":      InvoiceDeleteSchema(),
		"invoice_add_item":    InvoiceAddItemSchema(),
		"invoice_remove_item": InvoiceRemoveItemSchema(),
	}
}

// GetInvoiceToolSchema returns the schema for a specific invoice tool.
//
// This function provides a way to get the schema for a specific invoice tool
// for use with external validation systems.
//
// Parameters:
// - toolName: Name of the invoice tool to get schema for
//
// Returns:
// - map[string]interface{}: JSON schema for the tool, or nil if not found
// - bool: True if tool exists, false otherwise
//
// Notes:
// - Returns nil schema and false for unknown tool names
// - Schema can be used with any JSON Schema Draft 7 validator
// - Tool names match the MCP tool names exactly
func GetInvoiceToolSchema(toolName string) (map[string]interface{}, bool) {
	schemas := GetAllInvoiceSchemas()
	schema, exists := schemas[toolName]
	return schema, exists
}
