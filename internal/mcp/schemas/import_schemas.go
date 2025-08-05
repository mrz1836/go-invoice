// Package schemas provides JSON schema definitions for CSV import and data management MCP tools.
//
// This package contains comprehensive schema definitions optimized for natural language
// interaction with Claude. Each schema is designed to be intuitive and provides clear
// error messages for validation failures.
//
// Key design principles:
// - Parameter names map naturally to user requests ("file_path" vs "filePath")
// - Support for various CSV formats and field mapping configurations
// - Flexible import destination options with comprehensive validation
// - Comprehensive validation with helpful error messages
// - Real-world examples demonstrating various import scenarios
//
// All schemas follow JSON Schema Draft 7 specification and integrate with the
// validation system from the tools package.
package schemas

import (
	"time"
)

// ImportCSVSchema defines the JSON schema for CSV import operations.
//
// This schema supports comprehensive CSV import with flexible field mapping,
// destination control, and validation options. It provides natural language
// parameter names and extensive validation for reliable data import.
func ImportCSVSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"file_path": map[string]interface{}{
				"type":        "string",
				"description": "Path to the CSV file to import. Can be absolute or relative path.",
				"minLength":   1,
				"examples": []string{
					"/path/to/timesheet.csv",
					"./data/monthly-timesheet.csv",
					"~/Downloads/export.csv",
				},
			},
			"client_name": map[string]interface{}{
				"type":        "string",
				"description": "Client name for the import (used when creating new invoices). Partial matches supported.",
				"examples":    []string{"Acme Corporation", "Tech Solutions", "John Smith"},
			},
			"client_id": map[string]interface{}{
				"type":        "string",
				"description": "Exact client ID for the import (alternative to client_name).",
				"examples":    []string{"client_123", "acme-corp-id"},
			},
			"client_email": map[string]interface{}{
				"type":        "string",
				"format":      "email",
				"description": "Client email address to identify the client (alternative to client_name or client_id).",
				"examples":    []string{"contact@acme.com", "billing@techsolutions.com"},
			},
			"invoice_id": map[string]interface{}{
				"type":        "string",
				"description": "Existing invoice ID to append data to (used with append_invoice mode).",
				"examples":    []string{"INV-2025-001", "invoice_abc123"},
			},
			"import_mode": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"new_invoice", "append_invoice"},
				"default":     "new_invoice",
				"description": "Import destination: create new invoice or append to existing invoice.",
			},
			"description": map[string]interface{}{
				"type":        "string",
				"description": "Description for the new invoice (used with new_invoice mode).",
				"maxLength":   500,
				"examples": []string{
					"August 2025 Development Work",
					"Consulting Services - Q3 2025",
					"Monthly Timesheet Import",
				},
			},
			"due_days": map[string]interface{}{
				"type":        "number",
				"minimum":     1,
				"maximum":     365,
				"default":     30,
				"description": "Number of days from invoice date to due date (used with new_invoice mode).",
			},
			"invoice_date": map[string]interface{}{
				"type":        "string",
				"format":      "date",
				"description": "Invoice date for new invoice (YYYY-MM-DD). Defaults to today if not specified.",
				"examples":    []string{time.Now().Format("2006-01-02"), "2025-08-01", "2025-09-01"},
			},
			"field_mapping": map[string]interface{}{
				"type":        "object",
				"description": "Custom field mapping for non-standard CSV formats.",
				"properties": map[string]interface{}{
					"date": map[string]interface{}{
						"type":        "string",
						"description": "CSV column name for work date.",
						"examples":    []string{"date", "work_date", "day"},
					},
					"hours": map[string]interface{}{
						"type":        "string",
						"description": "CSV column name for hours worked.",
						"examples":    []string{"hours", "time_spent", "duration"},
					},
					"rate": map[string]interface{}{
						"type":        "string",
						"description": "CSV column name for hourly rate.",
						"examples":    []string{"rate", "hourly_rate", "bill_rate"},
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "CSV column name for work description.",
						"examples":    []string{"description", "task", "notes"},
					},
				},
				"additionalProperties": false,
			},
			"has_header": map[string]interface{}{
				"type":        "boolean",
				"default":     true,
				"description": "Whether the CSV file has a header row with column names.",
			},
			"delimiter": map[string]interface{}{
				"type":        "string",
				"enum":        []string{",", ";", "\t", "|"},
				"default":     ",",
				"description": "CSV field delimiter character.",
			},
			"default_rate": map[string]interface{}{
				"type":        "number",
				"minimum":     0.01,
				"description": "Default hourly rate to use when rate is missing or zero in CSV.",
				"examples":    []interface{}{75.0, 125.0, 200.0},
			},
			"rate_override": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Override all rates in CSV with the default_rate value.",
			},
			"currency": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"USD", "EUR", "GBP", "CAD", "AUD"},
				"default":     "USD",
				"description": "Currency for the imported amounts.",
			},
			"batch_size": map[string]interface{}{
				"type":        "number",
				"minimum":     1,
				"maximum":     1000,
				"default":     100,
				"description": "Number of rows to process in each batch (for large files).",
			},
			"validate_only": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Validate data structure without importing (dry run mode).",
			},
			"dry_run": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Process data and show results without making changes.",
			},
			"skip_weekends": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Skip work items that fall on weekends.",
			},
			"skip_duplicates": map[string]interface{}{
				"type":        "boolean",
				"default":     true,
				"description": "Skip work items that appear to be duplicates when appending to existing invoice.",
			},
		},
		"additionalProperties": false,
	}
}

// ImportValidateSchema defines the JSON schema for CSV validation operations.
//
// This schema supports comprehensive CSV validation with structure checking,
// data quality analysis, and business rule validation before import execution.
func ImportValidateSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"file_path": map[string]interface{}{
				"type":        "string",
				"description": "Path to the CSV file to validate.",
				"minLength":   1,
				"examples": []string{
					"/path/to/timesheet.csv",
					"./data/export.csv",
					"~/Downloads/timesheet.csv",
				},
			},
			"expected_fields": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "List of expected field names in the CSV.",
				"examples": [][]string{
					{"date", "hours", "rate", "description"},
					{"work_date", "time_spent", "hourly_rate", "task"},
				},
			},
			"field_mapping": map[string]interface{}{
				"type":        "object",
				"description": "Custom field mapping for validation with non-standard CSV formats.",
				"properties": map[string]interface{}{
					"date": map[string]interface{}{
						"type":        "string",
						"description": "CSV column name for work date.",
						"examples":    []string{"date", "work_date", "day"},
					},
					"hours": map[string]interface{}{
						"type":        "string",
						"description": "CSV column name for hours worked.",
						"examples":    []string{"hours", "time_spent", "duration"},
					},
					"rate": map[string]interface{}{
						"type":        "string",
						"description": "CSV column name for hourly rate.",
						"examples":    []string{"rate", "hourly_rate", "bill_rate"},
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "CSV column name for work description.",
						"examples":    []string{"description", "task", "notes"},
					},
				},
				"additionalProperties": false,
			},
			"has_header": map[string]interface{}{
				"type":        "boolean",
				"default":     true,
				"description": "Whether the CSV file has a header row with column names.",
			},
			"delimiter": map[string]interface{}{
				"type":        "string",
				"enum":        []string{",", ";", "\t", "|"},
				"default":     ",",
				"description": "CSV field delimiter character.",
			},
			"validate_rates": map[string]interface{}{
				"type":        "boolean",
				"default":     true,
				"description": "Validate that hourly rates are reasonable and positive.",
			},
			"validate_dates": map[string]interface{}{
				"type":        "boolean",
				"default":     true,
				"description": "Validate that dates are in correct format and reasonable range.",
			},
			"validate_business": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Apply business rule validation (hour limits, weekend work, etc.).",
			},
			"max_hours_per_day": map[string]interface{}{
				"type":        "number",
				"minimum":     1,
				"maximum":     24,
				"default":     12,
				"description": "Maximum allowed hours per day for business rule validation.",
			},
			"min_rate": map[string]interface{}{
				"type":        "number",
				"minimum":     0.01,
				"default":     10.0,
				"description": "Minimum reasonable hourly rate for validation.",
			},
			"max_rate": map[string]interface{}{
				"type":        "number",
				"minimum":     1.0,
				"default":     1000.0,
				"description": "Maximum reasonable hourly rate for validation.",
			},
			"check_duplicates": map[string]interface{}{
				"type":        "boolean",
				"default":     true,
				"description": "Check for duplicate entries within the CSV data.",
			},
			"check_weekends": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Flag work items that fall on weekends as potential issues.",
			},
			"quick_validate": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Perform quick validation using data sampling (faster for large files).",
			},
			"sample_size": map[string]interface{}{
				"type":        "number",
				"minimum":     10,
				"maximum":     1000,
				"default":     100,
				"description": "Number of rows to sample for quick validation.",
			},
			"client_name": map[string]interface{}{
				"type":        "string",
				"description": "Client name for context-aware validation (optional).",
				"examples":    []string{"Regular Client Corp", "Acme Corporation"},
			},
			"compare_historical": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Compare against historical data patterns for the client.",
			},
			"validate_consistency": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Check for consistency in rates, patterns, and work distribution.",
			},
			"flag_anomalies": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Flag unusual patterns or anomalies in the data.",
			},
		},
		"required":             []string{"file_path"},
		"additionalProperties": false,
	}
}

// ImportPreviewSchema defines the JSON schema for CSV import preview operations.
//
// This schema supports comprehensive import preview with detailed analysis
// and impact assessment without making any changes to the system.
func ImportPreviewSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"file_path": map[string]interface{}{
				"type":        "string",
				"description": "Path to the CSV file to preview.",
				"minLength":   1,
				"examples": []string{
					"/path/to/timesheet.csv",
					"./data/preview-data.csv",
					"~/Downloads/import.csv",
				},
			},
			"client_name": map[string]interface{}{
				"type":        "string",
				"description": "Client name for the preview (used when previewing new invoice creation).",
				"examples":    []string{"Preview Client Inc", "Acme Corporation"},
			},
			"invoice_id": map[string]interface{}{
				"type":        "string",
				"description": "Existing invoice ID to preview appending to.",
				"examples":    []string{"INV-2025-005", "invoice_preview"},
			},
			"import_mode": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"new_invoice", "append_invoice"},
				"default":     "new_invoice",
				"description": "Import destination for preview: create new invoice or append to existing invoice.",
			},
			"field_mapping": map[string]interface{}{
				"type":        "object",
				"description": "Custom field mapping for preview with non-standard CSV formats.",
				"properties": map[string]interface{}{
					"date": map[string]interface{}{
						"type":        "string",
						"description": "CSV column name for work date.",
						"examples":    []string{"date", "work_date", "day"},
					},
					"hours": map[string]interface{}{
						"type":        "string",
						"description": "CSV column name for hours worked.",
						"examples":    []string{"hours", "time_spent", "duration"},
					},
					"rate": map[string]interface{}{
						"type":        "string",
						"description": "CSV column name for hourly rate.",
						"examples":    []string{"rate", "hourly_rate", "bill_rate"},
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "CSV column name for work description.",
						"examples":    []string{"description", "task", "notes"},
					},
				},
				"additionalProperties": false,
			},
			"has_header": map[string]interface{}{
				"type":        "boolean",
				"default":     true,
				"description": "Whether the CSV file has a header row with column names.",
			},
			"delimiter": map[string]interface{}{
				"type":        "string",
				"enum":        []string{",", ";", "\t", "|"},
				"default":     ",",
				"description": "CSV field delimiter character.",
			},
			"default_rate": map[string]interface{}{
				"type":        "number",
				"minimum":     0.01,
				"description": "Default hourly rate to use in preview when rate is missing or zero.",
				"examples":    []interface{}{75.0, 125.0, 200.0},
			},
			"rate_override": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Override all rates in preview with the default_rate value.",
			},
			"show_totals": map[string]interface{}{
				"type":        "boolean",
				"default":     true,
				"description": "Include calculated totals and financial summary in preview.",
			},
			"show_details": map[string]interface{}{
				"type":        "boolean",
				"default":     true,
				"description": "Show detailed work item breakdown in preview.",
			},
			"show_existing": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Include existing invoice items when previewing append operations.",
			},
			"detect_conflicts": map[string]interface{}{
				"type":        "boolean",
				"default":     true,
				"description": "Detect and highlight potential conflicts or duplicates.",
			},
			"show_before_after": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Show before/after comparison for rate transformations.",
			},
			"summary_only": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Show only high-level summary without detailed item listing.",
			},
			"sample_preview": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Show preview using data sampling for very large files.",
			},
			"preview_limit": map[string]interface{}{
				"type":        "number",
				"minimum":     5,
				"maximum":     100,
				"default":     20,
				"description": "Maximum number of work items to show in detailed preview.",
			},
			"show_warnings": map[string]interface{}{
				"type":        "boolean",
				"default":     true,
				"description": "Include data quality warnings and potential issues in preview.",
			},
			"show_suggestions": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Include suggestions for data improvement and optimization.",
			},
			"quality_analysis": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Perform comprehensive data quality analysis in preview.",
			},
		},
		"additionalProperties": false,
	}
}

// GetAllImportSchemas returns all import-related schemas mapped by tool name.
//
// This function provides a centralized way to access all import tool schemas
// for registration with the MCP tool system.
//
// Returns:
// - map[string]map[string]interface{}: Map of tool names to their JSON schemas
//
// Notes:
// - Schema names match the corresponding tool names exactly
// - All schemas follow JSON Schema Draft 7 specification
// - Schemas are optimized for Claude natural language interaction
func GetAllImportSchemas() map[string]map[string]interface{} {
	return map[string]map[string]interface{}{
		"import_csv":      ImportCSVSchema(),
		"import_validate": ImportValidateSchema(),
		"import_preview":  ImportPreviewSchema(),
	}
}

// GetImportToolSchema returns the schema for a specific import tool.
//
// This function provides a way to get the schema for a specific import tool
// for use with external validation systems.
//
// Parameters:
// - toolName: Name of the import tool to get schema for
//
// Returns:
// - map[string]interface{}: JSON schema for the tool, or nil if not found
// - bool: True if tool exists, false otherwise
//
// Notes:
// - Returns nil schema and false for unknown tool names
// - Schema can be used with any JSON Schema Draft 7 validator
// - Tool names match the MCP tool names exactly
func GetImportToolSchema(toolName string) (map[string]interface{}, bool) {
	schemas := GetAllImportSchemas()
	schema, exists := schemas[toolName]
	return schema, exists
}
