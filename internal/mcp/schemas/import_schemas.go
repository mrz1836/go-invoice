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
		keyType: keyObject,
		keyProperties: map[string]interface{}{
			"file_path": map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Path to the CSV file to import. Can be absolute or relative path.",
				keyMinLength:   1,
				keyExamples: []string{
					"/path/to/timesheet.csv",
					"./data/monthly-timesheet.csv",
					"~/Downloads/export.csv",
				},
			},
			keyClientName: map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Client name for the import (used when creating new invoices). Partial matches supported.",
				keyExamples:    []string{"Acme Corporation", "Tech Solutions", "John Smith"},
			},
			keyClientID: map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Exact client ID for the import (alternative to client_name).",
				keyExamples:    []string{"client_123", "acme-corp-id"},
			},
			"client_email": map[string]interface{}{
				keyType:        typeString,
				keyFormat:      keyEmail,
				keyDescription: "Client email address to identify the client (alternative to client_name or client_id).",
				keyExamples:    []string{"contact@acme.com", "billing@techsolutions.com"},
			},
			keyInvoiceID: map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Existing invoice ID to append data to (used with append_invoice mode).",
				keyExamples:    []string{"INV-2025-001", "invoice_abc123"},
			},
			"import_mode": map[string]interface{}{
				keyType:        typeString,
				keyEnum:        []string{"new_invoice", "append_invoice"},
				keyDefault:     "new_invoice",
				keyDescription: "Import destination: create new invoice or append to existing invoice.",
			},
			keyDescription: map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Description for the new invoice (used with new_invoice mode).",
				keyMaxLength:   500,
				keyExamples: []string{
					"August 2025 Development Work",
					"Consulting Services - Q3 2025",
					"Monthly Timesheet Import",
				},
			},
			"due_days": map[string]interface{}{
				keyType:        typeNumber,
				keyMinimum:     1,
				keyMaximum:     365,
				keyDefault:     30,
				keyDescription: "Number of days from invoice date to due date (used with new_invoice mode).",
			},
			"invoice_date": map[string]interface{}{
				keyType:        typeString,
				keyFormat:      formatDate,
				keyDescription: "Invoice date for new invoice (YYYY-MM-DD). Defaults to today if not specified.",
				keyExamples:    []string{time.Now().Format("2006-01-02"), "2025-08-01", "2025-09-01"},
			},
			"field_mapping": map[string]interface{}{
				keyType:        keyObject,
				keyDescription: "Custom field mapping for non-standard CSV formats.",
				keyProperties: map[string]interface{}{
					formatDate: map[string]interface{}{
						keyType:        typeString,
						keyDescription: "CSV column name for work date.",
						keyExamples:    []string{formatDate, "work_date", "day"},
					},
					keyHours: map[string]interface{}{
						keyType:        typeString,
						keyDescription: "CSV column name for hours worked.",
						keyExamples:    []string{keyHours, "time_spent", "duration"},
					},
					keyRate: map[string]interface{}{
						keyType:        typeString,
						keyDescription: "CSV column name for hourly rate.",
						keyExamples:    []string{keyRate, "hourly_rate", "bill_rate"},
					},
					keyDescription: map[string]interface{}{
						keyType:        typeString,
						keyDescription: "CSV column name for work description.",
						keyExamples:    []string{keyDescription, "task", "notes"},
					},
				},
				keyAdditionalProperties: false,
			},
			"has_header": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     true,
				keyDescription: "Whether the CSV file has a header row with column names.",
			},
			"delimiter": map[string]interface{}{
				keyType:        typeString,
				keyEnum:        []string{",", ";", "\t", "|"},
				keyDefault:     ",",
				keyDescription: "CSV field delimiter character.",
			},
			"default_rate": map[string]interface{}{
				keyType:        typeNumber,
				keyMinimum:     0.01,
				keyDescription: "Default hourly rate to use when rate is missing or zero in CSV.",
				keyExamples:    []interface{}{75.0, 125.0, 200.0},
			},
			"rate_override": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Override all rates in CSV with the default_rate value.",
			},
			"currency": map[string]interface{}{
				keyType:        typeString,
				keyEnum:        []string{"USD", "EUR", "GBP", "CAD", "AUD"},
				keyDefault:     "USD",
				keyDescription: "Currency for the imported amounts.",
			},
			"batch_size": map[string]interface{}{
				keyType:        typeNumber,
				keyMinimum:     1,
				keyMaximum:     1000,
				keyDefault:     100,
				keyDescription: "Number of rows to process in each batch (for large files).",
			},
			"validate_only": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Validate data structure without importing (dry run mode).",
			},
			"dry_run": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Process data and show results without making changes.",
			},
			"skip_weekends": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Skip work items that fall on weekends.",
			},
			"skip_duplicates": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     true,
				keyDescription: "Skip work items that appear to be duplicates when appending to existing invoice.",
			},
		},
		keyAdditionalProperties: false,
	}
}

// ImportValidateSchema defines the JSON schema for CSV validation operations.
//
// This schema supports comprehensive CSV validation with structure checking,
// data quality analysis, and business rule validation before import execution.
func ImportValidateSchema() map[string]interface{} {
	return map[string]interface{}{
		keyType: keyObject,
		keyProperties: map[string]interface{}{
			"file_path": map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Path to the CSV file to validate.",
				keyMinLength:   1,
				keyExamples: []string{
					"/path/to/timesheet.csv",
					"./data/export.csv",
					"~/Downloads/timesheet.csv",
				},
			},
			"expected_fields": map[string]interface{}{
				keyType:        typeArray,
				"items":        map[string]interface{}{keyType: typeString},
				keyDescription: "List of expected field names in the CSV.",
				keyExamples: [][]string{
					{formatDate, keyHours, keyRate, keyDescription},
					{"work_date", "time_spent", "hourly_rate", "task"},
				},
			},
			"field_mapping": map[string]interface{}{
				keyType:        keyObject,
				keyDescription: "Custom field mapping for validation with non-standard CSV formats.",
				keyProperties: map[string]interface{}{
					formatDate: map[string]interface{}{
						keyType:        typeString,
						keyDescription: "CSV column name for work date.",
						keyExamples:    []string{formatDate, "work_date", "day"},
					},
					keyHours: map[string]interface{}{
						keyType:        typeString,
						keyDescription: "CSV column name for hours worked.",
						keyExamples:    []string{keyHours, "time_spent", "duration"},
					},
					keyRate: map[string]interface{}{
						keyType:        typeString,
						keyDescription: "CSV column name for hourly rate.",
						keyExamples:    []string{keyRate, "hourly_rate", "bill_rate"},
					},
					keyDescription: map[string]interface{}{
						keyType:        typeString,
						keyDescription: "CSV column name for work description.",
						keyExamples:    []string{keyDescription, "task", "notes"},
					},
				},
				keyAdditionalProperties: false,
			},
			"has_header": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     true,
				keyDescription: "Whether the CSV file has a header row with column names.",
			},
			"delimiter": map[string]interface{}{
				keyType:        typeString,
				keyEnum:        []string{",", ";", "\t", "|"},
				keyDefault:     ",",
				keyDescription: "CSV field delimiter character.",
			},
			"validate_rates": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     true,
				keyDescription: "Validate that hourly rates are reasonable and positive.",
			},
			"validate_dates": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     true,
				keyDescription: "Validate that dates are in correct format and reasonable range.",
			},
			"validate_business": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Apply business rule validation (hour limits, weekend work, etc.).",
			},
			"max_hours_per_day": map[string]interface{}{
				keyType:        typeNumber,
				keyMinimum:     1,
				keyMaximum:     24,
				keyDefault:     12,
				keyDescription: "Maximum allowed hours per day for business rule validation.",
			},
			"min_rate": map[string]interface{}{
				keyType:        typeNumber,
				keyMinimum:     0.01,
				keyDefault:     10.0,
				keyDescription: "Minimum reasonable hourly rate for validation.",
			},
			"max_rate": map[string]interface{}{
				keyType:        typeNumber,
				keyMinimum:     1.0,
				keyDefault:     1000.0,
				keyDescription: "Maximum reasonable hourly rate for validation.",
			},
			"check_duplicates": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     true,
				keyDescription: "Check for duplicate entries within the CSV data.",
			},
			"check_weekends": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Flag work items that fall on weekends as potential issues.",
			},
			"quick_validate": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Perform quick validation using data sampling (faster for large files).",
			},
			"sample_size": map[string]interface{}{
				keyType:        typeNumber,
				keyMinimum:     10,
				keyMaximum:     1000,
				keyDefault:     100,
				keyDescription: "Number of rows to sample for quick validation.",
			},
			keyClientName: map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Client name for context-aware validation (optional).",
				keyExamples:    []string{"Regular Client Corp", "Acme Corporation"},
			},
			"compare_historical": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Compare against historical data patterns for the client.",
			},
			"validate_consistency": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Check for consistency in rates, patterns, and work distribution.",
			},
			"flag_anomalies": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Flag unusual patterns or anomalies in the data.",
			},
		},
		keyRequired:             []string{"file_path"},
		keyAdditionalProperties: false,
	}
}

// ImportPreviewSchema defines the JSON schema for CSV import preview operations.
//
// This schema supports comprehensive import preview with detailed analysis
// and impact assessment without making any changes to the system.
func ImportPreviewSchema() map[string]interface{} {
	return map[string]interface{}{
		keyType: keyObject,
		keyProperties: map[string]interface{}{
			"file_path": map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Path to the CSV file to preview.",
				keyMinLength:   1,
				keyExamples: []string{
					"/path/to/timesheet.csv",
					"./data/preview-data.csv",
					"~/Downloads/import.csv",
				},
			},
			keyClientName: map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Client name for the preview (used when previewing new invoice creation).",
				keyExamples:    []string{"Preview Client Inc", "Acme Corporation"},
			},
			keyInvoiceID: map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Existing invoice ID to preview appending to.",
				keyExamples:    []string{"INV-2025-005", "invoice_preview"},
			},
			"import_mode": map[string]interface{}{
				keyType:        typeString,
				keyEnum:        []string{"new_invoice", "append_invoice"},
				keyDefault:     "new_invoice",
				keyDescription: "Import destination for preview: create new invoice or append to existing invoice.",
			},
			"field_mapping": map[string]interface{}{
				keyType:        keyObject,
				keyDescription: "Custom field mapping for preview with non-standard CSV formats.",
				keyProperties: map[string]interface{}{
					formatDate: map[string]interface{}{
						keyType:        typeString,
						keyDescription: "CSV column name for work date.",
						keyExamples:    []string{formatDate, "work_date", "day"},
					},
					keyHours: map[string]interface{}{
						keyType:        typeString,
						keyDescription: "CSV column name for hours worked.",
						keyExamples:    []string{keyHours, "time_spent", "duration"},
					},
					keyRate: map[string]interface{}{
						keyType:        typeString,
						keyDescription: "CSV column name for hourly rate.",
						keyExamples:    []string{keyRate, "hourly_rate", "bill_rate"},
					},
					keyDescription: map[string]interface{}{
						keyType:        typeString,
						keyDescription: "CSV column name for work description.",
						keyExamples:    []string{keyDescription, "task", "notes"},
					},
				},
				keyAdditionalProperties: false,
			},
			"has_header": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     true,
				keyDescription: "Whether the CSV file has a header row with column names.",
			},
			"delimiter": map[string]interface{}{
				keyType:        typeString,
				keyEnum:        []string{",", ";", "\t", "|"},
				keyDefault:     ",",
				keyDescription: "CSV field delimiter character.",
			},
			"default_rate": map[string]interface{}{
				keyType:        typeNumber,
				keyMinimum:     0.01,
				keyDescription: "Default hourly rate to use in preview when rate is missing or zero.",
				keyExamples:    []interface{}{75.0, 125.0, 200.0},
			},
			"rate_override": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Override all rates in preview with the default_rate value.",
			},
			"show_totals": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     true,
				keyDescription: "Include calculated totals and financial summary in preview.",
			},
			"show_details": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     true,
				keyDescription: "Show detailed work item breakdown in preview.",
			},
			"show_existing": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Include existing invoice items when previewing append operations.",
			},
			"detect_conflicts": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     true,
				keyDescription: "Detect and highlight potential conflicts or duplicates.",
			},
			"show_before_after": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Show before/after comparison for rate transformations.",
			},
			"summary_only": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Show only high-level summary without detailed item listing.",
			},
			"sample_preview": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Show preview using data sampling for very large files.",
			},
			"preview_limit": map[string]interface{}{
				keyType:        typeNumber,
				keyMinimum:     5,
				keyMaximum:     100,
				keyDefault:     20,
				keyDescription: "Maximum number of work items to show in detailed preview.",
			},
			"show_warnings": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     true,
				keyDescription: "Include data quality warnings and potential issues in preview.",
			},
			"show_suggestions": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Include suggestions for data improvement and optimization.",
			},
			"quality_analysis": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Perform comprehensive data quality analysis in preview.",
			},
		},
		keyAdditionalProperties: false,
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
