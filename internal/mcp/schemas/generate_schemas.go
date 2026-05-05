// Package schemas provides JSON schema definitions for document generation MCP tools.
//
// This package contains comprehensive schema definitions optimized for natural language
// interaction with Claude. Each schema is designed to be intuitive and provides clear
// error messages for validation failures.
//
// Key design principles:
// - Parameter names map naturally to user requests ("template" vs "templateID")
// - Support for multiple output formats and customization options
// - Flexible filtering and batch processing capabilities
// - Comprehensive validation with helpful error messages
// - Real-world examples demonstrating various usage patterns
//
// All schemas follow JSON Schema Draft 7 specification and integrate with the
// validation system from the tools package.
package schemas

// GenerateHTMLSchema defines the JSON schema for generating HTML invoices.
//
// This schema supports HTML generation with template selection, customization options,
// and batch processing capabilities for flexible invoice presentation.
func GenerateHTMLSchema() map[string]interface{} {
	return map[string]interface{}{
		keyType: keyObject,
		keyProperties: map[string]interface{}{
			keyInvoiceID: map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Invoice ID to generate HTML for.",
				keyMinLength:   1,
				keyExamples:    []string{exampleInvoiceID, "invoice_abc123"},
			},
			"invoice_number": map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Invoice number to generate HTML for (alternative to invoice_id).",
				keyMinLength:   1,
				keyExamples:    []string{exampleInvoiceID, "2025-001"},
			},
			"batch_invoices": map[string]interface{}{
				keyType:        typeArray,
				keyDescription: "List of invoice IDs or numbers for batch generation.",
				"items": map[string]interface{}{
					keyType:      typeString,
					keyMinLength: 1,
				},
				"minItems":  1,
				"maxItems":  50,
				keyExamples: []interface{}{[]string{exampleInvoiceID, "INV-002", "INV-003"}},
			},
			"template": map[string]interface{}{
				keyType:        typeString,
				keyEnum:        []string{"default", "minimal", "custom", "web"},
				keyDefault:     "default",
				keyDescription: "Template style for HTML generation.",
			},
			"custom_css": map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Path to custom CSS file for styling (used with 'custom' template).",
				keyExamples:    []string{"./assets/company-styles.css", "/path/to/custom.css"},
			},
			"include_logo": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Include company logo in the generated HTML.",
			},
			"company_name": map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Company name to display in the invoice header.",
				keyMaxLength:   200,
				keyExamples:    []string{"My Consulting LLC", "Tech Solutions Inc"},
			},
			"include_notes": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Include additional notes and terms in the invoice.",
			},
			"footer_text": map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Custom footer text for the invoice.",
				keyMaxLength:   500,
				keyExamples:    []string{"Thank you for your business!", "Payment terms: Net 30 days"},
			},
			"web_preview": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Generate HTML optimized for web preview instead of printing.",
			},
			"return_html": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Return HTML content in response instead of saving to file.",
			},
		},
		keyAdditionalProperties: false,
	}
}

// GenerateSummarySchema defines the JSON schema for generating invoice summaries.
//
// This schema supports comprehensive summary generation with analytics,
// reporting options, and multiple output formats for business intelligence.
func GenerateSummarySchema() map[string]interface{} {
	return map[string]interface{}{
		keyType: keyObject,
		keyProperties: map[string]interface{}{
			"summary_type": map[string]interface{}{
				keyType:        typeString,
				keyEnum:        []string{"revenue", "client", "overdue", "tax", "dashboard", "custom"},
				keyDefault:     "revenue",
				keyDescription: "Type of summary to generate.",
			},
			"period": map[string]interface{}{
				keyType:        typeString,
				keyEnum:        []string{"daily", "weekly", "monthly", "quarterly", "yearly", "custom"},
				keyDefault:     "monthly",
				keyDescription: "Time period for the summary.",
			},
			"from_date": map[string]interface{}{
				keyType:        typeString,
				keyFormat:      formatDate,
				keyDescription: "Start date for custom period (YYYY-MM-DD).",
				keyExamples:    []string{"2025-01-01", "2025-08-01"},
			},
			"to_date": map[string]interface{}{
				keyType:        typeString,
				keyFormat:      formatDate,
				keyDescription: "End date for custom period (YYYY-MM-DD).",
				keyExamples:    []string{"2025-12-31", "2025-08-31"},
			},
			"month": map[string]interface{}{
				keyType:        typeString,
				"pattern":      "^\\d{4}-\\d{2}$",
				keyDescription: "Specific month for monthly summaries (YYYY-MM).",
				keyExamples:    []string{"2025-08", "2025-12"},
			},
			"year": map[string]interface{}{
				keyType:        typeNumber,
				keyMinimum:     2020,
				keyMaximum:     2030,
				keyDescription: "Specific year for yearly summaries.",
				keyExamples:    []interface{}{2025, 2024},
			},
			keyClientName: map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Filter summary to specific client (for client-type summaries).",
				keyExamples:    []string{"Acme Corp", "Tech Solutions Inc"},
			},
			"include_charts": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Include visual charts and graphs in the summary.",
			},
			"include_details": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Include detailed invoice and transaction information.",
			},
			"show_trends": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Include trend analysis and comparisons with previous periods.",
			},
			"tax_categories": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Include tax category breakdown (for tax summaries).",
			},
			"payment_methods": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Include payment method analysis.",
			},
			keyFormat: map[string]interface{}{
				keyType:        typeString,
				keyEnum:        []string{"html", "pdf", "csv", "excel", typeJSON},
				keyDefault:     "html",
				keyDescription: "Output format for the summary report.",
			},
			"output_path": map[string]interface{}{
				keyType:        typeString,
				keyDescription: "File path where the summary should be saved. Required unless using return_data.",
				keyExamples:    []string{"./reports/monthly-summary.html", "/exports/summary.pdf"},
			},
			"return_data": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Return summary data in response instead of saving to file.",
			},
		},
		keyAdditionalProperties: false,
	}
}

// ExportDataSchema defines the JSON schema for exporting invoice data.
//
// This schema supports comprehensive data export with flexible filtering,
// multiple formats, and integration options for external systems.
func ExportDataSchema() map[string]interface{} {
	return map[string]interface{}{
		keyType: keyObject,
		keyProperties: map[string]interface{}{
			"export_type": map[string]interface{}{
				keyType:        typeString,
				keyEnum:        []string{"invoices", "clients", "work_items", "payments", "full_backup"},
				keyDefault:     "invoices",
				keyDescription: "Type of data to export.",
			},
			"status": map[string]interface{}{
				keyType:        typeString,
				keyEnum:        []string{"draft", "sent", "paid", "overdue", "voided"},
				keyDescription: "Filter by invoice status (for invoice exports).",
			},
			keyClientName: map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Filter by client name (partial matches supported).",
				keyExamples:    []string{"Acme Corp", "Tech Solutions"},
			},
			"from_date": map[string]interface{}{
				keyType:        typeString,
				keyFormat:      formatDate,
				keyDescription: "Export data from this date onwards (YYYY-MM-DD).",
				keyExamples:    []string{"2025-01-01", "2025-08-01"},
			},
			"to_date": map[string]interface{}{
				keyType:        typeString,
				keyFormat:      formatDate,
				keyDescription: "Export data up to this date (YYYY-MM-DD).",
				keyExamples:    []string{"2025-12-31", "2025-08-31"},
			},
			"date_range": map[string]interface{}{
				keyType:        typeString,
				keyEnum:        []string{"2025-Q1", "2025-Q2", "2025-Q3", "2025-Q4", "2025-YTD", "last-30-days", "last-90-days"},
				keyDescription: "Predefined date range for export (alternative to from_date/to_date).",
			},
			"include_items": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Include work items/line items in the export (for invoice exports).",
			},
			"include_stats": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Include statistics and summary data.",
			},
			"include_payments": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Include payment information and history.",
			},
			"include_rates": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     true,
				keyDescription: "Include hourly rates and pricing information.",
			},
			"group_by": map[string]interface{}{
				keyType:        typeString,
				keyEnum:        []string{"client", "project", formatDate, "status", "none"},
				keyDefault:     "none",
				keyDescription: "Group exported data by specified criteria.",
			},
			keyFormat: map[string]interface{}{
				keyType:        typeString,
				keyEnum:        []string{"csv", typeJSON, "xml", "excel", "yaml"},
				keyDefault:     "csv",
				keyDescription: "Format for the exported data.",
			},
			"limit": map[string]interface{}{
				keyType:        typeNumber,
				keyMinimum:     1,
				keyMaximum:     10000,
				keyDescription: "Maximum number of records to export.",
				keyDefault:     1000,
			},
			"output_path": map[string]interface{}{
				keyType:        typeString,
				keyDescription: "File path where the export should be saved. Required unless using return_data.",
				keyExamples:    []string{"./exports/invoices.csv", "/data/export.json"},
			},
			"return_data": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Return exported data in response instead of saving to file.",
			},
		},
		keyAdditionalProperties: false,
	}
}

// GetAllGenerationSchemas returns all generation-related schemas mapped by tool name.
//
// This function provides a centralized way to access all generation tool schemas
// for registration with the MCP tool system.
//
// Returns:
// - map[string]map[string]interface{}: Map of tool names to their JSON schemas
//
// Notes:
// - Schema names match the corresponding tool names exactly
// - All schemas follow JSON Schema Draft 7 specification
// - Schemas are optimized for Claude natural language interaction
func GetAllGenerationSchemas() map[string]map[string]interface{} {
	return map[string]map[string]interface{}{
		"generate_html":    GenerateHTMLSchema(),
		"generate_summary": GenerateSummarySchema(),
		"export_data":      ExportDataSchema(),
	}
}

// GetGenerationToolSchema returns the schema for a specific generation tool.
//
// This function provides a way to get the schema for a specific generation tool
// for use with external validation systems.
//
// Parameters:
// - toolName: Name of the generation tool to get schema for
//
// Returns:
// - map[string]interface{}: JSON schema for the tool, or nil if not found
// - bool: True if tool exists, false otherwise
//
// Notes:
// - Returns nil schema and false for unknown tool names
// - Schema can be used with any JSON Schema Draft 7 validator
// - Tool names match the MCP tool names exactly
func GetGenerationToolSchema(toolName string) (map[string]interface{}, bool) {
	schemas := GetAllGenerationSchemas()
	schema, exists := schemas[toolName]
	return schema, exists
}
