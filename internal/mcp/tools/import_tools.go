// Package tools provides MCP tool definitions for CSV import and data management operations.
//
// This package implements comprehensive data import tools optimized for natural
// language interaction with Claude. Each tool provides intuitive parameter names,
// flexible CSV handling options, and detailed examples for various import scenarios.
//
// Key features:
// - Natural language-friendly parameter names and descriptions
// - Support for various CSV formats and field mapping options
// - Comprehensive validation and preview capabilities before import
// - Flexible import destinations (new invoices, existing invoices)
// - Comprehensive examples covering common import workflows
// - Integration with go-invoice CLI commands
// - Context-aware operations with proper error handling
//
// All tools follow the core architecture from Phase 2.1 and integrate seamlessly
// with the tool registry, validation system, and category management.
package tools

import (
	"context"
	"time"

	"github.com/mrz1836/go-invoice/internal/mcp/schemas"
)

// CreateDataImportTools creates all data import MCP tool definitions.
//
// This function initializes comprehensive data import tools optimized for
// conversational interaction with Claude. Each tool is designed to handle common
// import workflows naturally and provide clear guidance for parameter usage.
//
// Returns:
// - []*MCPTool: Complete set of data import tools ready for registration
//
// Tools created:
// 1. import_csv - Import timesheet data with mapping options and destination control
// 2. import_validate - Validate CSV structure and data before import execution
// 3. import_preview - Preview import results without making any changes
//
// Notes:
// - All tools use the CategoryDataImport category for organization
// - Schemas are optimized for natural language parameter entry
// - Examples cover both simple and complex import scenarios
// - CLI integration follows standard go-invoice command patterns
func CreateDataImportTools() []*MCPTool {
	return []*MCPTool{
		createImportCSVTool(),
		createImportValidateTool(),
		createImportPreviewTool(),
	}
}

// createImportCSVTool creates the CSV import tool definition.
//
// This tool supports importing timesheet data from CSV files with flexible
// mapping options and destination control (new or existing invoices).
func createImportCSVTool() *MCPTool {
	return &MCPTool{
		Name:        "import_csv",
		Description: "Import timesheet data from CSV files with flexible field mapping and destination options. Supports creating new invoices or appending to existing ones with comprehensive validation.",
		InputSchema: schemas.ImportCSVSchema(),
		Examples: []MCPToolExample{
			{
				Description: "Import timesheet to create new invoice for existing client",
				Input: map[string]interface{}{
					"file_path":   "/path/to/timesheet.csv",
					"client_name": "Acme Corporation",
					"description": "August 2025 Development Work",
					"import_mode": "new_invoice",
					"due_days":    30,
				},
				ExpectedOutput: "New invoice created with imported timesheet data, automatic totals calculated",
				UseCase:        "Monthly billing from timesheet exports",
			},
			{
				Description: "Import timesheet with custom field mapping",
				Input: map[string]interface{}{
					"file_path":   "/path/to/custom-format.csv",
					"client_id":   "client_123",
					"import_mode": "new_invoice",
					"field_mapping": map[string]interface{}{
						"date":        "work_date",
						"hours":       "time_spent",
						"rate":        "hourly_rate",
						"description": "task_description",
					},
					"has_header": true,
				},
				ExpectedOutput: "Timesheet imported using custom field mappings with data validation",
				UseCase:        "Importing from custom timesheet formats and external systems",
			},
			{
				Description: "Append timesheet data to existing invoice",
				Input: map[string]interface{}{
					"file_path":     "/path/to/additional-work.csv",
					"invoice_id":    "INV-2025-001",
					"import_mode":   "append_invoice",
					"validate_only": false,
				},
				ExpectedOutput: "Additional work items added to existing invoice with automatic total recalculation",
				UseCase:        "Adding supplementary work to invoices in progress",
			},
			{
				Description: "Import with data transformation and rate override",
				Input: map[string]interface{}{
					"file_path":     "/path/to/contractor-hours.csv",
					"client_email":  "contractor@freelance.com",
					"import_mode":   "new_invoice",
					"default_rate":  150.0,
					"rate_override": true,
					"description":   "Consulting Services - September 2025",
					"currency":      "USD",
				},
				ExpectedOutput: "Contractor timesheet imported with standardized rate across all entries",
				UseCase:        "Standardizing rates for contractor and freelancer billing",
			},
			{
				Description: "Import large timesheet with batch processing",
				Input: map[string]interface{}{
					"file_path":     "/path/to/large-timesheet.csv",
					"client_name":   "Enterprise Client LLC",
					"import_mode":   "new_invoice",
					"batch_size":    100,
					"validate_only": false,
					"dry_run":       false,
				},
				ExpectedOutput: "Large timesheet processed in batches with progress tracking and error handling",
				UseCase:        "Processing large data exports from enterprise time tracking systems",
			},
		},
		Category:   CategoryDataImport,
		CLICommand: "go-invoice",
		CLIArgs:    []string{"import"},
		HelpText:   "Imports timesheet data from CSV files with comprehensive mapping and validation. Supports both new invoice creation and appending to existing invoices with flexible field mapping options.",
		Version:    "1.0.0",
		Timeout:    60 * time.Second,
	}
}

// createImportValidateTool creates the import validation tool definition.
//
// This tool validates CSV structure and data integrity before import execution
// to prevent errors and ensure data quality.
func createImportValidateTool() *MCPTool {
	return &MCPTool{
		Name:        "import_validate",
		Description: "Validate CSV file structure, data integrity, and field mappings before import execution. Provides detailed error reports and suggestions for data correction.",
		InputSchema: schemas.ImportValidateSchema(),
		Examples: []MCPToolExample{
			{
				Description: "Validate timesheet CSV structure and data quality",
				Input: map[string]interface{}{
					"file_path":       "/path/to/timesheet.csv",
					"has_header":      true,
					"expected_fields": []string{"date", "hours", "rate", "description"},
				},
				ExpectedOutput: "Validation report with field analysis, data quality metrics, and error identification",
				UseCase:        "Quality assurance before important data imports",
			},
			{
				Description: "Validate custom format with specific field mapping",
				Input: map[string]interface{}{
					"file_path": "/path/to/custom-export.csv",
					"field_mapping": map[string]interface{}{
						"date":        "work_date",
						"hours":       "duration",
						"rate":        "bill_rate",
						"description": "task_notes",
					},
					"validate_rates":   true,
					"validate_dates":   true,
					"check_duplicates": true,
				},
				ExpectedOutput: "Comprehensive validation including rate reasonableness and duplicate detection",
				UseCase:        "Validating complex data imports from external systems",
			},
			{
				Description: "Quick validation for standard timesheet format",
				Input: map[string]interface{}{
					"file_path":      "/path/to/standard-timesheet.csv",
					"quick_validate": true,
					"sample_size":    50,
				},
				ExpectedOutput: "Fast validation using data sampling for large files",
				UseCase:        "Quick validation of routine timesheet imports",
			},
			{
				Description: "Detailed validation with business rule checking",
				Input: map[string]interface{}{
					"file_path":         "/path/to/complex-timesheet.csv",
					"validate_business": true,
					"max_hours_per_day": 12,
					"min_rate":          25.0,
					"max_rate":          500.0,
					"check_weekends":    true,
				},
				ExpectedOutput: "Business rule validation including hour limits and rate reasonableness",
				UseCase:        "Comprehensive validation for compliance and quality control",
			},
			{
				Description: "Validate with client context and historical comparison",
				Input: map[string]interface{}{
					"file_path":            "/path/to/monthly-timesheet.csv",
					"client_name":          "Regular Client Corp",
					"compare_historical":   true,
					"validate_consistency": true,
					"flag_anomalies":       true,
				},
				ExpectedOutput: "Validation with anomaly detection and historical pattern comparison",
				UseCase:        "Advanced quality control with pattern recognition and anomaly detection",
			},
			{
				Description: "Validate malformed CSV file to identify format errors",
				Input: map[string]interface{}{
					"file_path":     "/path/to/corrupt-timesheet.csv",
					"has_header":    true,
					"strict_mode":   true,
					"error_details": true,
				},
				ExpectedOutput: "Detailed error report identifying malformed data, invalid fields, and corrupt entries with suggestions for correction",
				UseCase:        "Error detection and troubleshooting for problematic import files",
			},
		},
		Category:   CategoryDataImport,
		CLICommand: "go-invoice",
		CLIArgs:    []string{"import", "--validate"},
		HelpText:   "Validates CSV data structure and content before import execution. Provides comprehensive error reporting, business rule validation, and data quality analysis to ensure successful imports.",
		Version:    "1.0.0",
		Timeout:    45 * time.Second,
	}
}

// createImportPreviewTool creates the import preview tool definition.
//
// This tool provides a preview of import results without making any changes,
// allowing users to verify data and make adjustments before execution.
func createImportPreviewTool() *MCPTool {
	return &MCPTool{
		Name:        "import_preview",
		Description: "Preview import results without making any changes to the system. Shows exactly what data will be imported, calculated totals, and potential issues.",
		InputSchema: schemas.ImportPreviewSchema(),
		Examples: []MCPToolExample{
			{
				Description: "Preview timesheet import for new invoice creation",
				Input: map[string]interface{}{
					"file_path":    "/path/to/timesheet.csv",
					"client_name":  "Preview Client Inc",
					"import_mode":  "new_invoice",
					"show_totals":  true,
					"show_details": true,
				},
				ExpectedOutput: "Complete preview showing work items, calculated totals, and new invoice structure",
				UseCase:        "Verifying import data before creating client invoices",
			},
			{
				Description: "Preview append operation with conflict detection",
				Input: map[string]interface{}{
					"file_path":        "/path/to/additional-work.csv",
					"invoice_id":       "INV-2025-005",
					"import_mode":      "append_invoice",
					"detect_conflicts": true,
					"show_existing":    true,
				},
				ExpectedOutput: "Preview showing new items, existing items, and potential conflicts or duplicates",
				UseCase:        "Safely adding work to existing invoices without duplication",
			},
			{
				Description: "Preview with custom rate transformations",
				Input: map[string]interface{}{
					"file_path":         "/path/to/contractor-hours.csv",
					"client_id":         "contractor_client",
					"import_mode":       "new_invoice",
					"default_rate":      125.0,
					"rate_override":     true,
					"show_before_after": true,
				},
				ExpectedOutput: "Before/after comparison showing original rates and transformed rates",
				UseCase:        "Verifying rate standardization and billing calculations",
			},
			{
				Description: "Preview large import with statistical summary",
				Input: map[string]interface{}{
					"file_path":       "/path/to/large-dataset.csv",
					"client_email":    "enterprise@company.com",
					"import_mode":     "new_invoice",
					"summary_only":    true,
					"show_statistics": true,
					"sample_preview":  true,
				},
				ExpectedOutput: "Statistical summary with sample data preview for large imports",
				UseCase:        "Understanding large dataset structure before full import",
			},
			{
				Description: "Preview with validation warnings and recommendations",
				Input: map[string]interface{}{
					"file_path":        "/path/to/questionable-data.csv",
					"client_name":      "Quality Check Client",
					"import_mode":      "new_invoice",
					"show_warnings":    true,
					"show_suggestions": true,
					"quality_analysis": true,
				},
				ExpectedOutput: "Preview with data quality warnings and improvement recommendations",
				UseCase:        "Quality control and data improvement before critical imports",
			},
		},
		Category:   CategoryDataImport,
		CLICommand: "go-invoice",
		CLIArgs:    []string{"import", "--preview"},
		HelpText:   "Provides comprehensive preview of import operations without making changes. Shows calculated totals, identifies potential issues, and allows verification before execution.",
		Version:    "1.0.0",
		Timeout:    30 * time.Second,
	}
}

// RegisterDataImportTools registers all data import tools with the provided registry.
//
// This function provides a convenient way to register all data import tools
// in the tool registry for MCP server integration.
//
// Parameters:
// - ctx: Context for cancellation and timeout
// - registry: Tool registry to register tools with
//
// Returns:
// - error: Registration error if any tool fails to register, or nil if all successful
//
// Side Effects:
// - Registers 3 data import tools in the CategoryDataImport category
// - Tools become available for MCP client discovery and execution
//
// Notes:
// - All tools are registered with comprehensive validation schemas
// - Examples and help text are included for Claude interaction guidance
// - CLI command mappings enable direct go-invoice CLI integration
// - Respects context cancellation for responsive operations
func RegisterDataImportTools(ctx context.Context, registry ToolRegistry) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	tools := CreateDataImportTools()
	for _, tool := range tools {
		if err := registry.RegisterTool(ctx, tool); err != nil {
			return err
		}
	}

	return nil
}
