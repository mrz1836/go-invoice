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
		keyType: keyObject,
		keyProperties: map[string]interface{}{
			keyClientName: map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Client name or partial name for invoice. If multiple clients match, you'll be prompted to be more specific.",
				keyMinLength:   1,
				keyMaxLength:   200,
				keyExamples:    []string{"Acme Corp", "John Smith", "Tech Solutions"},
			},
			keyClientID: map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Exact client ID for invoice. Use this when you know the specific client identifier.",
				"pattern":      "^[A-Za-z0-9_-]+$",
				keyExamples:    []string{"client_123", "acme-corp-id"},
			},
			"client_email": map[string]interface{}{
				keyType:        typeString,
				keyFormat:      keyEmail,
				keyDescription: "Client email address to identify the client for this invoice.",
				keyExamples:    []string{"contact@acme.com", "john@techsolutions.com"},
			},
			"invoice_date": map[string]interface{}{
				keyType:        typeString,
				keyFormat:      formatDate,
				keyDescription: "Invoice date in YYYY-MM-DD format. Defaults to today if not specified.",
				keyExamples:    []string{time.Now().Format("2006-01-02"), "2025-01-15", "2025-08-03"},
			},
			"due_date": map[string]interface{}{
				keyType:        typeString,
				keyFormat:      formatDate,
				keyDescription: "Payment due date in YYYY-MM-DD format. If not provided, calculated based on default payment terms.",
				keyExamples:    []string{time.Now().AddDate(0, 0, 30).Format("2006-01-02"), "2025-02-15", "2025-09-02"},
			},
			keyDescription: map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Optional description for the invoice (e.g., project name, billing period).",
				keyMaxLength:   500,
				keyExamples:    []string{"January 2025 consulting services", "Website development project", "Monthly retainer"},
			},
			"create_client_if_missing": map[string]interface{}{
				keyType:        typeBoolean,
				keyDescription: "Whether to create a new client if the specified client is not found.",
				keyDefault:     false,
			},
			"new_client_email": map[string]interface{}{
				keyType:        typeString,
				keyFormat:      keyEmail,
				keyDescription: "Email address for new client (required when create_client_if_missing is true).",
				keyExamples:    []string{"newclient@company.com"},
			},
			"new_client_address": map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Address for new client (optional when creating new client).",
				keyMaxLength:   500,
				keyExamples:    []string{"123 Main St, City, State 12345"},
			},
			"new_client_phone": map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Phone number for new client (optional when creating new client).",
				"pattern":      "^[\\d\\s\\+\\-\\(\\)\\.\\/ext]+$",
				keyExamples:    []string{"+1-555-123-4567", "(555) 123-4567"},
			},
			"usdc_address": map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Override USDC cryptocurrency address for this specific invoice. If not provided, uses the global USDC address from configuration. Useful when you want a unique payment address for this invoice.",
				keyExamples:    []string{"0x1234567890abcdef1234567890abcdef12345678"},
			},
			"bsv_address": map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Override BSV cryptocurrency address for this specific invoice. If not provided, uses the global BSV address from configuration. Useful when you want a unique payment address for this invoice.",
				keyExamples:    []string{"1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa"},
			},
			"work_items": map[string]interface{}{
				keyType:        typeArray,
				keyDescription: "Optional work items to add to the invoice upon creation.",
				"items": map[string]interface{}{
					keyType: keyObject,
					keyProperties: map[string]interface{}{
						formatDate: map[string]interface{}{
							keyType:        typeString,
							keyFormat:      formatDate,
							keyDescription: "Date when work was performed (YYYY-MM-DD).",
						},
						keyHours: map[string]interface{}{
							keyType:        typeNumber,
							keyMinimum:     0.01,
							keyMaximum:     24.0,
							keyDescription: "Number of hours worked (decimal allowed, e.g., 1.5 for 1 hour 30 minutes).",
						},
						keyRate: map[string]interface{}{
							keyType:        typeNumber,
							keyMinimum:     0.01,
							keyDescription: "Hourly rate for this work item.",
						},
						keyDescription: map[string]interface{}{
							keyType:        typeString,
							keyMinLength:   1,
							keyMaxLength:   500,
							keyDescription: "Description of work performed.",
						},
					},
					keyRequired: []string{formatDate, keyHours, keyRate, keyDescription},
				},
			},
		},
		keyRequired:             []string{}, // At least one client identifier (client_name, client_id, or client_email) is required but handled in validation logic
		keyAdditionalProperties: false,
	}
}

// InvoiceListSchema defines the JSON schema for listing/filtering invoices.
//
// This schema provides comprehensive filtering options for invoice discovery
// with natural language-friendly parameter names and flexible search criteria.
func InvoiceListSchema() map[string]interface{} {
	return map[string]interface{}{
		keyType: keyObject,
		keyProperties: map[string]interface{}{
			"status": map[string]interface{}{
				keyType:        typeString,
				keyEnum:        []string{"draft", "sent", "paid", "overdue", "voided"},
				keyDescription: "Filter invoices by status. Leave empty to show all statuses.",
				keyExamples:    []string{"paid", "overdue", "draft"},
			},
			keyClientName: map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Filter by client name (partial matches supported).",
				keyExamples:    []string{"Acme", "Tech Solutions", "John"},
			},
			keyClientID: map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Filter by exact client ID.",
				keyExamples:    []string{"client_123", "acme-corp"},
			},
			"from_date": map[string]interface{}{
				keyType:        typeString,
				keyFormat:      formatDate,
				keyDescription: "Show invoices from this date onwards (YYYY-MM-DD). Filters by invoice date.",
				keyExamples:    []string{"2025-01-01", "2025-07-01"},
			},
			"to_date": map[string]interface{}{
				keyType:        typeString,
				keyFormat:      formatDate,
				keyDescription: "Show invoices up to this date (YYYY-MM-DD). Filters by invoice date.",
				keyExamples:    []string{"2025-12-31", "2025-08-31"},
			},
			"sort_by": map[string]interface{}{
				keyType:        typeString,
				keyEnum:        []string{formatDate, "amount", "status", "client", "due_date"},
				keyDefault:     formatDate,
				keyDescription: "Field to sort results by.",
			},
			"sort_order": map[string]interface{}{
				keyType:        typeString,
				keyEnum:        []string{"asc", "desc"},
				keyDefault:     "desc",
				keyDescription: "Sort order: ascending or descending.",
			},
			"limit": map[string]interface{}{
				keyType:        typeNumber,
				keyMinimum:     1,
				keyMaximum:     1000,
				keyDescription: "Maximum number of invoices to return. Default is 50.",
				keyDefault:     50,
			},
			"output_format": map[string]interface{}{
				keyType:        typeString,
				keyEnum:        []string{"table", typeJSON, "csv"},
				keyDefault:     "table",
				keyDescription: "Output format for the results.",
			},
			"include_summary": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Include summary statistics (total amounts, counts by status).",
			},
		},
		keyAdditionalProperties: false,
	}
}

// InvoiceShowSchema defines the JSON schema for showing invoice details.
//
// This schema supports invoice lookup by multiple identifiers and provides
// options for different levels of detail and output formats.
func InvoiceShowSchema() map[string]interface{} {
	return map[string]interface{}{
		keyType: keyObject,
		keyProperties: map[string]interface{}{
			keyInvoiceID: map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Invoice ID to display details for.",
				keyMinLength:   1,
				keyExamples:    []string{exampleInvoiceID, "invoice_abc123"},
			},
			"invoice_number": map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Invoice number to display details for (alternative to invoice_id).",
				keyMinLength:   1,
				keyExamples:    []string{exampleInvoiceID, "2025-001"},
			},
			"output_format": map[string]interface{}{
				keyType:        typeString,
				keyEnum:        []string{"text", typeJSON, "yaml"},
				keyDefault:     "text",
				keyDescription: "Output format for invoice details.",
			},
			"show_work_items": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     true,
				keyDescription: "Include detailed work items in the output.",
			},
			"show_client_details": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     true,
				keyDescription: "Include full client information in the output.",
			},
		},
		keyRequired:             []string{}, // At least one invoice identifier (invoice_id or invoice_number) is required but handled in validation logic
		keyAdditionalProperties: false,
	}
}

// InvoiceUpdateSchema defines the JSON schema for updating invoices.
//
// This schema supports selective updates of invoice properties with validation
// to ensure business rules are enforced (e.g., no updates to paid invoices).
func InvoiceUpdateSchema() map[string]interface{} {
	return map[string]interface{}{
		keyType: keyObject,
		keyProperties: map[string]interface{}{
			keyInvoiceID: map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Invoice ID to update.",
				keyMinLength:   1,
				keyExamples:    []string{exampleInvoiceID, "invoice_abc123"},
			},
			"invoice_number": map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Invoice number to update (alternative to invoice_id).",
				keyMinLength:   1,
				keyExamples:    []string{exampleInvoiceID, "2025-001"},
			},
			"status": map[string]interface{}{
				keyType:        typeString,
				keyEnum:        []string{"draft", "sent", "paid", "overdue", "voided"},
				keyDescription: "Update invoice status. Note: certain status transitions may be restricted.",
				keyExamples:    []string{"sent", "paid"},
			},
			"due_date": map[string]interface{}{
				keyType:        typeString,
				keyFormat:      formatDate,
				keyDescription: "Update payment due date (YYYY-MM-DD).",
				keyExamples:    []string{"2025-02-28", "2025-09-15"},
			},
			keyDescription: map[string]interface{}{
				keyType:        typeString,
				keyMaxLength:   500,
				keyDescription: "Update invoice description.",
				keyExamples:    []string{"Updated: January 2025 consulting services", "Q1 2025 development work"},
			},
			"usdc_address": map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Override USDC cryptocurrency address for this invoice. Set to empty string to clear override and use global config.",
				keyExamples:    []string{"0x1234567890abcdef1234567890abcdef12345678", ""},
			},
			"bsv_address": map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Override BSV cryptocurrency address for this invoice. Set to empty string to clear override and use global config.",
				keyExamples:    []string{"1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa", ""},
			},
		},
		keyRequired:             []string{}, // Invoice ID (invoice_id or invoice_number) and at least one update field (status, due_date, description, usdc_address, bsv_address) required but handled in validation logic
		keyAdditionalProperties: false,
	}
}

// InvoiceDeleteSchema defines the JSON schema for deleting invoices.
//
// This schema supports both soft and hard deletion with appropriate
// safety confirmations and business rule validation.
func InvoiceDeleteSchema() map[string]interface{} {
	return map[string]interface{}{
		keyType: keyObject,
		keyProperties: map[string]interface{}{
			keyInvoiceID: map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Invoice ID to delete.",
				keyMinLength:   1,
				keyExamples:    []string{exampleInvoiceID, "invoice_abc123"},
			},
			"invoice_number": map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Invoice number to delete (alternative to invoice_id).",
				keyMinLength:   1,
				keyExamples:    []string{exampleInvoiceID, "2025-001"},
			},
			"hard_delete": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Permanently delete invoice (cannot be undone). Default is soft delete.",
			},
			"force": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Skip confirmation prompt (use with caution).",
			},
		},
		keyAdditionalProperties: false,
	}
}

// InvoiceAddItemSchema defines the JSON schema for adding work items to invoices.
//
// This schema supports adding single or multiple work items with comprehensive
// validation and natural date/time entry formats.
func InvoiceAddItemSchema() map[string]interface{} {
	return map[string]interface{}{
		keyType: keyObject,
		keyProperties: map[string]interface{}{
			keyInvoiceID: map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Invoice ID to add work items to.",
				keyMinLength:   1,
				keyExamples:    []string{exampleInvoiceID, "invoice_abc123"},
			},
			"invoice_number": map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Invoice number to add work items to (alternative to invoice_id).",
				keyMinLength:   1,
				keyExamples:    []string{exampleInvoiceID, "2025-001"},
			},
			"work_items": map[string]interface{}{
				keyType:        typeArray,
				"minItems":     1,
				keyDescription: "Work items to add to the invoice.",
				"items": map[string]interface{}{
					keyType: keyObject,
					keyProperties: map[string]interface{}{
						formatDate: map[string]interface{}{
							keyType:        typeString,
							keyFormat:      formatDate,
							keyDescription: "Date when work was performed (YYYY-MM-DD).",
							keyExamples:    []string{time.Now().Format("2006-01-02"), "2025-08-01"},
						},
						keyHours: map[string]interface{}{
							keyType:        typeNumber,
							keyMinimum:     0.01,
							keyMaximum:     24.0,
							keyDescription: "Number of hours worked (decimal allowed, e.g., 1.5 for 1 hour 30 minutes).",
							keyExamples:    []interface{}{8.0, 4.5, 2.25},
						},
						keyRate: map[string]interface{}{
							keyType:        typeNumber,
							keyMinimum:     0.01,
							keyDescription: "Hourly rate for this work item.",
							keyExamples:    []interface{}{75.0, 125.0, 200.0},
						},
						keyDescription: map[string]interface{}{
							keyType:        typeString,
							keyMinLength:   1,
							keyMaxLength:   500,
							keyDescription: "Description of work performed.",
							keyExamples:    []string{"Frontend development", "Bug fixes and testing", "Client meeting and planning"},
						},
					},
					keyRequired: []string{formatDate, keyHours, keyRate, keyDescription},
				},
			},
		},
		keyRequired:             []string{}, // Invoice ID and work items required but handled in validation logic
		keyAdditionalProperties: false,
	}
}

// InvoiceAddLineItemSchema defines the JSON schema for adding flexible line items to invoices.
//
// This schema supports adding line items with three billing types: hourly, fixed, and quantity.
// It provides a flexible way to add various types of charges to an invoice beyond simple hourly work.
func InvoiceAddLineItemSchema() map[string]interface{} {
	return map[string]interface{}{
		keyType: keyObject,
		keyProperties: map[string]interface{}{
			keyInvoiceID: map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Invoice ID to add line items to.",
				keyMinLength:   1,
				keyExamples:    []string{exampleInvoiceID, "invoice_abc123"},
			},
			"invoice_number": map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Invoice number to add line items to (alternative to invoice_id).",
				keyMinLength:   1,
				keyExamples:    []string{exampleInvoiceID, "2025-001"},
			},
			"line_items": map[string]interface{}{
				keyType:        typeArray,
				"minItems":     1,
				keyDescription: "Line items to add to the invoice. Supports hourly, fixed, and quantity-based billing.",
				"items": map[string]interface{}{
					keyType: keyObject,
					keyProperties: map[string]interface{}{
						keyType: map[string]interface{}{
							keyType:        typeString,
							keyEnum:        []string{"hourly", "fixed", "quantity"},
							keyDescription: "Type of line item: 'hourly' for time-based billing, 'fixed' for flat fees/retainers, 'quantity' for unit-based pricing.",
							keyDefault:     "hourly",
							keyExamples:    []string{"hourly", "fixed", "quantity"},
						},
						formatDate: map[string]interface{}{
							keyType:        typeString,
							keyFormat:      formatDate,
							keyDescription: "Date for this line item (YYYY-MM-DD). Defaults to today if not specified.",
							keyExamples:    []string{time.Now().Format("2006-01-02"), "2025-08-01"},
						},
						keyDescription: map[string]interface{}{
							keyType:        typeString,
							keyMinLength:   1,
							keyMaxLength:   1000,
							keyDescription: "Description of the line item (work performed, service provided, item sold, etc.).",
							keyExamples:    []string{"Development work", "Monthly retainer - August", "SSL certificates", "Project setup fee"},
						},
						// Hourly type fields
						keyHours: map[string]interface{}{
							keyType:        typeNumber,
							keyMinimum:     0.01,
							keyMaximum:     24.0,
							keyDescription: "Number of hours worked (required for 'hourly' type). Decimal values allowed (e.g., 1.5 for 1 hour 30 minutes).",
							keyExamples:    []interface{}{8.0, 4.5, 2.25},
						},
						keyRate: map[string]interface{}{
							keyType:        typeNumber,
							keyMinimum:     0.01,
							keyMaximum:     10000.0,
							keyDescription: "Hourly rate (required for 'hourly' type).",
							keyExamples:    []interface{}{75.0, 125.0, 200.0},
						},
						// Fixed type fields
						"amount": map[string]interface{}{
							keyType:        typeNumber,
							keyMinimum:     0.01,
							keyMaximum:     1000000.0,
							keyDescription: "Fixed amount (required for 'fixed' type). Use for retainers, flat fees, setup charges, etc.",
							keyExamples:    []interface{}{2000.0, 500.0, 150.0},
						},
						// Quantity type fields
						"quantity": map[string]interface{}{
							keyType:        typeNumber,
							keyMinimum:     0.01,
							keyMaximum:     10000.0,
							keyDescription: "Quantity of units (required for 'quantity' type). Use for materials, licenses, subscriptions, etc.",
							keyExamples:    []interface{}{2.0, 5.0, 10.0},
						},
						"unit_price": map[string]interface{}{
							keyType:        typeNumber,
							keyMinimum:     0.01,
							keyMaximum:     100000.0,
							keyDescription: "Price per unit (required for 'quantity' type).",
							keyExamples:    []interface{}{50.0, 25.0, 100.0},
						},
					},
					keyRequired: []string{keyDescription}, // Type defaults to hourly, other fields required based on type
					"oneOf": []interface{}{
						// Hourly type validation
						map[string]interface{}{
							keyProperties: map[string]interface{}{
								keyType: map[string]interface{}{"const": "hourly"},
							},
							keyRequired: []string{keyHours, keyRate},
						},
						// Fixed type validation
						map[string]interface{}{
							keyProperties: map[string]interface{}{
								keyType: map[string]interface{}{"const": "fixed"},
							},
							keyRequired: []string{"amount"},
						},
						// Quantity type validation
						map[string]interface{}{
							keyProperties: map[string]interface{}{
								keyType: map[string]interface{}{"const": "quantity"},
							},
							keyRequired: []string{"quantity", "unit_price"},
						},
					},
				},
			},
		},
		keyRequired:             []string{}, // Invoice ID and line items required but handled in validation logic
		keyAdditionalProperties: false,
	}
}

// InvoiceRemoveItemSchema defines the JSON schema for removing work items from invoices.
//
// This schema supports removing work items by ID or by matching criteria
// with appropriate safety validations.
func InvoiceRemoveItemSchema() map[string]interface{} {
	return map[string]interface{}{
		keyType: keyObject,
		keyProperties: map[string]interface{}{
			keyInvoiceID: map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Invoice ID to remove work items from.",
				keyMinLength:   1,
				keyExamples:    []string{exampleInvoiceID, "invoice_abc123"},
			},
			"invoice_number": map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Invoice number to remove work items from (alternative to invoice_id).",
				keyMinLength:   1,
				keyExamples:    []string{exampleInvoiceID, "2025-001"},
			},
			"work_item_id": map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Specific work item ID to remove.",
				keyMinLength:   1,
				keyExamples:    []string{"work_item_123", "wi_abc"},
			},
			"work_item_description": map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Remove work items matching this description (partial matches supported).",
				keyMinLength:   1,
				keyExamples:    []string{"Frontend development", "Bug fixes"},
			},
			"work_item_date": map[string]interface{}{
				keyType:        typeString,
				keyFormat:      formatDate,
				keyDescription: "Remove work items from this specific date (YYYY-MM-DD).",
				keyExamples:    []string{"2025-08-01", "2025-07-15"},
			},
			"remove_all_matching": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Remove all work items matching the criteria (vs. just the first match).",
			},
			"confirm": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Confirm removal without additional prompts.",
			},
		},
		keyRequired:             []string{}, // Invoice ID (invoice_id or invoice_number) and work item identifier (work_item_id, work_item_description, or work_item_date) required but handled in validation logic
		keyAdditionalProperties: false,
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
		"invoice_create":        InvoiceCreateSchema(),
		"invoice_list":          InvoiceListSchema(),
		"invoice_show":          InvoiceShowSchema(),
		"invoice_update":        InvoiceUpdateSchema(),
		"invoice_delete":        InvoiceDeleteSchema(),
		"invoice_add_item":      InvoiceAddItemSchema(),
		"invoice_add_line_item": InvoiceAddLineItemSchema(),
		"invoice_remove_item":   InvoiceRemoveItemSchema(),
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
