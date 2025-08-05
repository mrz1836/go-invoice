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
		"type": "object",
		"properties": map[string]interface{}{
			"invoice_id": map[string]interface{}{
				"type":        "string",
				"description": "Invoice ID to generate HTML for.",
				"minLength":   1,
				"examples":    []string{"INV-001", "invoice_abc123"},
			},
			"invoice_number": map[string]interface{}{
				"type":        "string",
				"description": "Invoice number to generate HTML for (alternative to invoice_id).",
				"minLength":   1,
				"examples":    []string{"INV-001", "2025-001"},
			},
			"batch_invoices": map[string]interface{}{
				"type":        "array",
				"description": "List of invoice IDs or numbers for batch generation.",
				"items": map[string]interface{}{
					"type":      "string",
					"minLength": 1,
				},
				"minItems": 1,
				"maxItems": 50,
				"examples": []interface{}{[]string{"INV-001", "INV-002", "INV-003"}},
			},
			"template": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"default", "professional", "minimal", "custom", "web"},
				"default":     "default",
				"description": "Template style for HTML generation.",
			},
			"custom_css": map[string]interface{}{
				"type":        "string",
				"description": "Path to custom CSS file for styling (used with 'custom' template).",
				"examples":    []string{"./assets/company-styles.css", "/path/to/custom.css"},
			},
			"include_logo": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Include company logo in the generated HTML.",
			},
			"company_name": map[string]interface{}{
				"type":        "string",
				"description": "Company name to display in the invoice header.",
				"maxLength":   200,
				"examples":    []string{"My Consulting LLC", "Tech Solutions Inc"},
			},
			"include_notes": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Include additional notes and terms in the invoice.",
			},
			"footer_text": map[string]interface{}{
				"type":        "string",
				"description": "Custom footer text for the invoice.",
				"maxLength":   500,
				"examples":    []string{"Thank you for your business!", "Payment terms: Net 30 days"},
			},
			"output_path": map[string]interface{}{
				"type":        "string",
				"description": "File path where the HTML should be saved. Required unless using web_preview or return_html.",
				"examples":    []string{"./invoices/INV-001.html", "/exports/invoice.html"},
			},
			"output_dir": map[string]interface{}{
				"type":        "string",
				"description": "Directory for batch invoice generation (used with batch_invoices).",
				"examples":    []string{"./monthly-invoices/", "/exports/batch/"},
			},
			"auto_name": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Automatically generate file names based on invoice numbers (for batch generation).",
			},
			"web_preview": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Generate HTML optimized for web preview instead of printing.",
			},
			"return_html": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Return HTML content in response instead of saving to file.",
			},
		},
		"additionalProperties": false,
	}
}

// GenerateSummarySchema defines the JSON schema for generating invoice summaries.
//
// This schema supports comprehensive summary generation with analytics,
// reporting options, and multiple output formats for business intelligence.
func GenerateSummarySchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"summary_type": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"revenue", "client", "overdue", "tax", "dashboard", "custom"},
				"default":     "revenue",
				"description": "Type of summary to generate.",
			},
			"period": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"daily", "weekly", "monthly", "quarterly", "yearly", "custom"},
				"default":     "monthly",
				"description": "Time period for the summary.",
			},
			"from_date": map[string]interface{}{
				"type":        "string",
				"format":      "date",
				"description": "Start date for custom period (YYYY-MM-DD).",
				"examples":    []string{"2025-01-01", "2025-08-01"},
			},
			"to_date": map[string]interface{}{
				"type":        "string",
				"format":      "date",
				"description": "End date for custom period (YYYY-MM-DD).",
				"examples":    []string{"2025-12-31", "2025-08-31"},
			},
			"month": map[string]interface{}{
				"type":        "string",
				"pattern":     "^\\d{4}-\\d{2}$",
				"description": "Specific month for monthly summaries (YYYY-MM).",
				"examples":    []string{"2025-08", "2025-12"},
			},
			"year": map[string]interface{}{
				"type":        "number",
				"minimum":     2020,
				"maximum":     2030,
				"description": "Specific year for yearly summaries.",
				"examples":    []interface{}{2025, 2024},
			},
			"client_name": map[string]interface{}{
				"type":        "string",
				"description": "Filter summary to specific client (for client-type summaries).",
				"examples":    []string{"Acme Corp", "Tech Solutions Inc"},
			},
			"include_charts": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Include visual charts and graphs in the summary.",
			},
			"include_details": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Include detailed invoice and transaction information.",
			},
			"show_trends": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Include trend analysis and comparisons with previous periods.",
			},
			"tax_categories": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Include tax category breakdown (for tax summaries).",
			},
			"payment_methods": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Include payment method analysis.",
			},
			"format": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"html", "pdf", "csv", "excel", "json"},
				"default":     "html",
				"description": "Output format for the summary report.",
			},
			"output_path": map[string]interface{}{
				"type":        "string",
				"description": "File path where the summary should be saved. Required unless using return_data.",
				"examples":    []string{"./reports/monthly-summary.html", "/exports/summary.pdf"},
			},
			"return_data": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Return summary data in response instead of saving to file.",
			},
		},
		"additionalProperties": false,
	}
}

// ExportDataSchema defines the JSON schema for exporting invoice data.
//
// This schema supports comprehensive data export with flexible filtering,
// multiple formats, and integration options for external systems.
func ExportDataSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"export_type": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"invoices", "clients", "work_items", "payments", "full_backup"},
				"default":     "invoices",
				"description": "Type of data to export.",
			},
			"status": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"draft", "sent", "paid", "overdue", "voided"},
				"description": "Filter by invoice status (for invoice exports).",
			},
			"client_name": map[string]interface{}{
				"type":        "string",
				"description": "Filter by client name (partial matches supported).",
				"examples":    []string{"Acme Corp", "Tech Solutions"},
			},
			"from_date": map[string]interface{}{
				"type":        "string",
				"format":      "date",
				"description": "Export data from this date onwards (YYYY-MM-DD).",
				"examples":    []string{"2025-01-01", "2025-08-01"},
			},
			"to_date": map[string]interface{}{
				"type":        "string",
				"format":      "date",
				"description": "Export data up to this date (YYYY-MM-DD).",
				"examples":    []string{"2025-12-31", "2025-08-31"},
			},
			"date_range": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"2025-Q1", "2025-Q2", "2025-Q3", "2025-Q4", "2025-YTD", "last-30-days", "last-90-days"},
				"description": "Predefined date range for export (alternative to from_date/to_date).",
			},
			"include_items": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Include work items/line items in the export (for invoice exports).",
			},
			"include_stats": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Include statistics and summary data.",
			},
			"include_payments": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Include payment information and history.",
			},
			"include_rates": map[string]interface{}{
				"type":        "boolean",
				"default":     true,
				"description": "Include hourly rates and pricing information.",
			},
			"group_by": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"client", "project", "date", "status", "none"},
				"default":     "none",
				"description": "Group exported data by specified criteria.",
			},
			"format": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"csv", "json", "xml", "excel", "yaml"},
				"default":     "csv",
				"description": "Format for the exported data.",
			},
			"limit": map[string]interface{}{
				"type":        "number",
				"minimum":     1,
				"maximum":     10000,
				"description": "Maximum number of records to export.",
				"default":     1000,
			},
			"output_path": map[string]interface{}{
				"type":        "string",
				"description": "File path where the export should be saved. Required unless using return_data.",
				"examples":    []string{"./exports/invoices.csv", "/data/export.json"},
			},
			"return_data": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Return exported data in response instead of saving to file.",
			},
		},
		"additionalProperties": false,
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
