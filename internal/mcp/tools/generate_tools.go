// Package tools provides MCP tool definitions for document generation operations.
//
// This package implements comprehensive document generation tools optimized for natural
// language interaction with Claude. Each tool provides intuitive parameter names,
// flexible output options, and detailed examples for various use cases.
//
// Key features:
// - Natural language-friendly parameter names and descriptions
// - Support for multiple output formats (HTML, PDF, JSON, etc.)
// - Template selection and customization options
// - Batch processing capabilities for multiple documents
// - Context-aware operations with proper error handling
//
// All tools follow the core architecture from Phase 2.1 and integrate seamlessly
// with the tool registry, validation system, and category management.
package tools

import (
	"context"
	"time"

	"github.com/mrz/go-invoice/internal/mcp/schemas"
)

// CreateDocumentGenerationTools creates all document generation MCP tool definitions.
//
// This function initializes comprehensive document generation tools optimized for
// conversational interaction with Claude. Each tool is designed to handle common
// document generation workflows naturally and provide clear guidance for parameter usage.
//
// Returns:
// - []*MCPTool: Complete set of document generation tools ready for registration
//
// Tools created:
// 1. generate_html - Generate HTML invoices with template options
// 2. generate_summary - Create invoice summaries and reports
// 3. export_data - Export invoice data in various formats
//
// Notes:
// - All tools use the CategoryDataExport category for organization
// - Schemas are optimized for natural language parameter entry
// - Examples cover both simple and complex use cases
// - CLI integration follows standard go-invoice command patterns
func CreateDocumentGenerationTools() []*MCPTool {
	return []*MCPTool{
		createGenerateHTMLTool(),
		createGenerateSummaryTool(),
		createExportDataTool(),
	}
}

// createGenerateHTMLTool creates the HTML generation tool definition.
//
// This tool supports generating HTML invoices with template selection and customization
// options for client presentation and printing.
func createGenerateHTMLTool() *MCPTool {
	return &MCPTool{
		Name:        "generate_html",
		Description: "Generate HTML invoices with customizable templates for client presentation, printing, or web display. Supports batch generation and template selection.",
		InputSchema: schemas.GenerateHTMLSchema(),
		Examples: []MCPToolExample{
			{
				Description: "Generate HTML for a single invoice with default template",
				Input: map[string]interface{}{
					"invoice_number": "INV-001",
					"template":       "default",
					"output_path":    "./invoices/INV-001.html",
				},
				ExpectedOutput: "HTML invoice generated successfully with professional styling",
				UseCase:        "Creating presentable invoice for client email or printing",
			},
			{
				Description: "Generate HTML with professional template and company branding",
				Input: map[string]interface{}{
					"invoice_id":   "invoice_abc123",
					"template":     "professional",
					"include_logo": true,
					"company_name": "My Consulting LLC",
					"output_path":  "./client-invoices/professional-invoice.html",
				},
				ExpectedOutput: "Branded HTML invoice with company logo and professional styling",
				UseCase:        "High-quality client presentation with company branding",
			},
			{
				Description: "Batch generate HTML invoices for multiple clients",
				Input: map[string]interface{}{
					"batch_invoices": []string{"INV-001", "INV-002", "INV-003"},
					"template":       "minimal",
					"output_dir":     "./monthly-invoices/",
					"auto_name":      true,
				},
				ExpectedOutput: "Multiple HTML invoices generated with consistent formatting",
				UseCase:        "Monthly invoice delivery to multiple clients",
			},
			{
				Description: "Generate HTML with custom styling and additional notes",
				Input: map[string]interface{}{
					"invoice_number": "INV-CONSULTING-025",
					"template":       "custom",
					"custom_css":     "./assets/company-styles.css",
					"include_notes":  true,
					"footer_text":    "Thank you for your business! Payment terms: Net 30 days.",
					"output_path":    "./custom-invoice.html",
				},
				ExpectedOutput: "Custom-styled invoice with personalized notes and footer",
				UseCase:        "Highly customized invoice presentation for special clients",
			},
			{
				Description: "Generate HTML for web preview without saving to file",
				Input: map[string]interface{}{
					"invoice_id":  "INV-2025-042",
					"template":    "web",
					"web_preview": true,
					"return_html": true,
				},
				ExpectedOutput: "HTML content returned for web preview integration",
				UseCase:        "Web application integration and online invoice preview",
			},
		},
		Category:   CategoryDataExport,
		CLICommand: "go-invoice",
		CLIArgs:    []string{"generate", "html"},
		HelpText:   "Generates professional HTML invoices with template selection, branding options, and batch processing capabilities. Perfect for client presentation and printing.",
		Version:    "1.0.0",
		Timeout:    45 * time.Second,
	}
}

// createGenerateSummaryTool creates the summary generation tool definition.
//
// This tool provides comprehensive invoice summary generation with analytics
// and reporting capabilities for business intelligence.
func createGenerateSummaryTool() *MCPTool {
	return &MCPTool{
		Name:        "generate_summary",
		Description: "Create comprehensive invoice summaries and reports with analytics, statistics, and business intelligence data for decision making.",
		InputSchema: schemas.GenerateSummarySchema(),
		Examples: []MCPToolExample{
			{
				Description: "Generate monthly revenue summary for business overview",
				Input: map[string]interface{}{
					"period":         "monthly",
					"month":          "2025-08",
					"include_charts": true,
					"output_format":  "html",
					"output_path":    "./reports/august-2025-summary.html",
				},
				ExpectedOutput: "Comprehensive monthly report with revenue charts and statistics",
				UseCase:        "Monthly business review and performance tracking",
			},
			{
				Description: "Generate client-specific summary for account review",
				Input: map[string]interface{}{
					"client_name":     "Acme Corp",
					"from_date":       "2025-01-01",
					"to_date":         "2025-08-31",
					"include_details": true,
					"show_trends":     true,
					"output_format":   "pdf",
					"output_path":     "./client-reports/acme-corp-ytd-2025.pdf",
				},
				ExpectedOutput: "Detailed client summary with payment trends and project history",
				UseCase:        "Client relationship management and account reviews",
			},
			{
				Description: "Generate overdue invoices summary for collections",
				Input: map[string]interface{}{
					"summary_type":     "overdue",
					"aging_buckets":    true,
					"priority_ranking": true,
					"include_contacts": true,
					"output_format":    "csv",
					"output_path":      "./collections/overdue-summary.csv",
				},
				ExpectedOutput: "Collections-focused report with aging analysis and contact information",
				UseCase:        "Collections management and payment follow-up prioritization",
			},
			{
				Description: "Generate tax year summary for accounting",
				Input: map[string]interface{}{
					"period":              "yearly",
					"year":                2025,
					"tax_categories":      true,
					"payment_methods":     true,
					"quarterly_breakdown": true,
					"output_format":       "excel",
					"output_path":         "./tax-reports/2025-tax-summary.xlsx",
				},
				ExpectedOutput: "Comprehensive tax year report with quarterly breakdown and categorization",
				UseCase:        "Tax preparation and accounting reconciliation",
			},
			{
				Description: "Generate real-time dashboard data for web application",
				Input: map[string]interface{}{
					"summary_type":  "dashboard",
					"real_time":     true,
					"key_metrics":   true,
					"output_format": "json",
					"return_data":   true,
				},
				ExpectedOutput: "Real-time metrics and KPIs in JSON format for dashboard integration",
				UseCase:        "Web dashboard integration and real-time business monitoring",
			},
		},
		Category:   CategoryDataExport,
		CLICommand: "go-invoice",
		CLIArgs:    []string{"generate", "summary"},
		HelpText:   "Creates comprehensive business summaries with analytics, charts, and reports. Supports various time periods, client-specific views, and multiple output formats.",
		Version:    "1.0.0",
		Timeout:    60 * time.Second,
	}
}

// createExportDataTool creates the data export tool definition.
//
// This tool supports comprehensive data export in multiple formats for
// integration with accounting systems, spreadsheets, and external tools.
func createExportDataTool() *MCPTool {
	return &MCPTool{
		Name:        "export_data",
		Description: "Export invoice data in various formats (CSV, JSON, XML, Excel) for accounting systems, spreadsheets, and external integrations with flexible filtering and customization.",
		InputSchema: schemas.ExportDataSchema(),
		Examples: []MCPToolExample{
			{
				Description: "Export all paid invoices to CSV for accounting software",
				Input: map[string]interface{}{
					"export_type":   "invoices",
					"status":        "paid",
					"output_format": "csv",
					"include_items": true,
					"output_path":   "./exports/paid-invoices-2025.csv",
				},
				ExpectedOutput: "CSV file with all paid invoices and work items for accounting import",
				UseCase:        "Quarterly accounting reconciliation and tax preparation",
			},
			{
				Description: "Export client database to Excel for CRM import",
				Input: map[string]interface{}{
					"export_type":      "clients",
					"output_format":    "excel",
					"include_contacts": true,
					"include_stats":    true,
					"output_path":      "./exports/client-database.xlsx",
				},
				ExpectedOutput: "Excel workbook with client information and relationship statistics",
				UseCase:        "CRM system migration and client database management",
			},
			{
				Description: "Export work items data for project management analysis",
				Input: map[string]interface{}{
					"export_type":   "work_items",
					"from_date":     "2025-01-01",
					"to_date":       "2025-08-31",
					"group_by":      "project",
					"output_format": "json",
					"include_rates": true,
					"output_path":   "./analytics/work-items-ytd.json",
				},
				ExpectedOutput: "JSON export with work items grouped by project for analysis",
				UseCase:        "Project profitability analysis and time tracking review",
			},
			{
				Description: "Export filtered invoice data for specific client",
				Input: map[string]interface{}{
					"export_type":      "invoices",
					"client_name":      "Tech Solutions Inc",
					"output_format":    "xml",
					"date_range":       "2025-Q1",
					"include_payments": true,
					"output_path":      "./client-exports/techsolutions-q1-2025.xml",
				},
				ExpectedOutput: "XML export of all Tech Solutions invoices for Q1 2025",
				UseCase:        "Client-specific data export for contract reviews",
			},
			{
				Description: "Bulk export for data backup and archival",
				Input: map[string]interface{}{
					"export_type":      "full_backup",
					"output_format":    "json",
					"compress":         true,
					"include_metadata": true,
					"output_path":      "./backups/invoice-backup-2025-08.json.gz",
				},
				ExpectedOutput: "Compressed JSON backup with complete database export",
				UseCase:        "Data backup and archival for disaster recovery",
			},
			{
				Description: "Export to memory for API integration without file output",
				Input: map[string]interface{}{
					"export_type":   "invoices",
					"status":        "sent",
					"output_format": "json",
					"return_data":   true,
					"limit":         100,
				},
				ExpectedOutput: "JSON data returned in memory for API response",
				UseCase:        "API integration and real-time data serving",
			},
		},
		Category:   CategoryDataExport,
		CLICommand: "go-invoice",
		CLIArgs:    []string{"export"},
		HelpText:   "Exports invoice, client, and work item data in multiple formats with comprehensive filtering options. Supports both file output and in-memory data for integrations.",
		Version:    "1.0.0",
		Timeout:    90 * time.Second,
	}
}

// RegisterDocumentGenerationTools registers all document generation tools with the provided registry.
//
// This function provides a convenient way to register all document generation tools
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
// - Registers 3 document generation tools in the CategoryDataExport category
// - Tools become available for MCP client discovery and execution
//
// Notes:
// - All tools are registered with comprehensive validation schemas
// - Examples and help text are included for Claude interaction guidance
// - CLI command mappings enable direct go-invoice CLI integration
// - Respects context cancellation for responsive operations
func RegisterDocumentGenerationTools(ctx context.Context, registry ToolRegistry) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	tools := CreateDocumentGenerationTools()
	for _, tool := range tools {
		if err := registry.RegisterTool(ctx, tool); err != nil {
			return err
		}
	}

	return nil
}
