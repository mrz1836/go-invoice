// Package tools provides MCP tool definitions for invoice management operations.
//
// This package implements comprehensive invoice management tools optimized for natural
// language interaction with Claude. Each tool provides intuitive parameter names,
// flexible input options, and detailed examples for various use cases.
//
// Key features:
// - Natural language-friendly parameter names and descriptions
// - Support for multiple identification methods (ID, number, name)
// - Flexible date input formats and client resolution
// - Comprehensive examples covering common use cases
// - Integration with go-invoice CLI commands
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

// CreateInvoiceManagementTools creates all invoice management MCP tool definitions.
//
// This function initializes comprehensive invoice management tools optimized for
// conversational interaction with Claude. Each tool is designed to handle common
// invoice workflows naturally and provide clear guidance for parameter usage.
//
// Returns:
// - []*MCPTool: Complete set of invoice management tools ready for registration
//
// Tools created:
// 1. invoice_create - Create new invoices with optional work items
// 2. invoice_list - List and filter invoices with flexible criteria
// 3. invoice_show - Display detailed invoice information
// 4. invoice_update - Modify invoice metadata and settings
// 5. invoice_delete - Remove invoices with safety confirmations
// 6. invoice_add_item - Add work items to existing invoices
// 7. invoice_remove_item - Remove specific work items from invoices
//
// Notes:
// - All tools use the CategoryInvoiceManagement category for organization
// - Schemas are optimized for natural language parameter entry
// - Examples cover both simple and complex use cases
// - CLI integration follows standard go-invoice command patterns
func CreateInvoiceManagementTools() []*MCPTool {
	return []*MCPTool{
		createInvoiceCreateTool(),
		createInvoiceListTool(),
		createInvoiceShowTool(),
		createInvoiceUpdateTool(),
		createInvoiceDeleteTool(),
		createInvoiceAddItemTool(),
		createInvoiceAddLineItemTool(),
		createInvoiceRemoveItemTool(),
	}
}

// createInvoiceCreateTool creates the invoice creation tool definition.
//
// This tool supports creating invoices with flexible client resolution and optional
// work items. It handles client creation when needed and provides natural date entry.
func createInvoiceCreateTool() *MCPTool {
	return &MCPTool{
		Name:        "invoice_create",
		Description: "Create a new invoice for a client with optional work items. Supports flexible client identification and automatic client creation when needed.",
		InputSchema: schemas.InvoiceCreateSchema(),
		Examples: []MCPToolExample{
			{
				Description: "Create a simple invoice for an existing client",
				Input: map[string]interface{}{
					"client_name": "Acme Corp",
					"description": "January 2025 consulting services",
				},
				ExpectedOutput: "Invoice created successfully with auto-generated number and default due date",
				UseCase:        "Basic invoice creation for regular clients",
			},
			{
				Description: "Create invoice with specific dates and work items",
				Input: map[string]interface{}{
					"client_name":  "Tech Solutions Inc",
					"invoice_date": "2025-08-01",
					"due_date":     "2025-08-31",
					"description":  "Website development - Phase 1",
					"work_items": []map[string]interface{}{
						{
							"date":        "2025-08-01",
							"hours":       8.0,
							"rate":        125.0,
							"description": "Frontend development and UI design",
						},
						{
							"date":        "2025-08-02",
							"hours":       4.5,
							"rate":        125.0,
							"description": "Backend API integration",
						},
					},
				},
				ExpectedOutput: "Invoice created with specified dates and work items, totals calculated automatically",
				UseCase:        "Complete invoice creation with immediate work item entry",
			},
			{
				Description: "Create invoice and new client in one operation",
				Input: map[string]interface{}{
					"client_name":              "New Client LLC",
					"create_client_if_missing": true,
					"new_client_email":         "contact@newclient.com",
					"new_client_address":       "456 Business Ave, Suite 200, City, State 12345",
					"new_client_phone":         "+1-555-987-6543",
					"description":              "Initial consulting engagement",
				},
				ExpectedOutput: "New client created and invoice generated for first engagement",
				UseCase:        "Onboarding new clients with immediate invoicing",
			},
			{
				Description: "Create invoice using client email for identification",
				Input: map[string]interface{}{
					"client_email": "finance@acme.com",
					"description":  "Monthly retainer - August 2025",
				},
				ExpectedOutput: "Invoice created for client identified by email address",
				UseCase:        "Invoice creation when you know client email but not exact name",
			},
		},
		Category:   CategoryInvoiceManagement,
		CLICommand: "go-invoice",
		CLIArgs:    []string{"invoice", "create"},
		HelpText:   "Creates invoices with automatic number generation and flexible client resolution. Supports both simple creation and complex scenarios with work items and new client creation.",
		Version:    "1.0.0",
		Timeout:    30 * time.Second,
	}
}

// createInvoiceListTool creates the invoice listing and filtering tool definition.
//
// This tool provides comprehensive invoice discovery with flexible filtering
// and multiple output formats for different use cases.
func createInvoiceListTool() *MCPTool {
	return &MCPTool{
		Name:        "invoice_list",
		Description: "List and filter invoices with flexible search criteria. Supports filtering by status, client, date ranges, and provides multiple output formats.",
		InputSchema: schemas.InvoiceListSchema(),
		Examples: []MCPToolExample{
			{
				Description: "List all invoices with default formatting",
				Input: map[string]interface{}{
					"output_format": "table",
				},
				ExpectedOutput: "Table view of all invoices with number, client, date, status, and amount",
				UseCase:        "General invoice overview and status checking",
			},
			{
				Description: "Find all unpaid invoices with summary statistics",
				Input: map[string]interface{}{
					"status":          "sent",
					"include_summary": true,
					"sort_by":         "due_date",
					"sort_order":      "asc",
				},
				ExpectedOutput: "List of sent invoices sorted by due date with total unpaid amounts",
				UseCase:        "Collections management and overdue invoice tracking",
			},
			{
				Description: "Generate monthly invoice report for specific client",
				Input: map[string]interface{}{
					"client_name":     "Acme Corp",
					"from_date":       "2025-08-01",
					"to_date":         "2025-08-31",
					"output_format":   "json",
					"include_summary": true,
				},
				ExpectedOutput: "JSON export of all Acme Corp invoices for August 2025 with summary",
				UseCase:        "Client-specific reporting and data export",
			},
			{
				Description: "Find overdue invoices for follow-up",
				Input: map[string]interface{}{
					"status":     "overdue",
					"sort_by":    "due_date",
					"sort_order": "asc",
					"limit":      10,
				},
				ExpectedOutput: "Top 10 oldest overdue invoices requiring immediate attention",
				UseCase:        "Collections workflow and payment follow-up",
			},
			{
				Description: "Export invoice data for accounting system",
				Input: map[string]interface{}{
					"from_date":     "2025-01-01",
					"to_date":       "2025-12-31",
					"output_format": "csv",
					"status":        "paid",
				},
				ExpectedOutput: "CSV export of all paid invoices for tax year 2025",
				UseCase:        "Accounting integration and tax preparation",
			},
		},
		Category:   CategoryInvoiceManagement,
		CLICommand: "go-invoice",
		CLIArgs:    []string{"invoice", "list"},
		HelpText:   "Provides comprehensive invoice discovery with filtering by status, client, dates, and amounts. Supports table, JSON, and CSV output formats for different workflows.",
		Version:    "1.0.0",
		Timeout:    20 * time.Second,
	}
}

// createInvoiceShowTool creates the invoice detail display tool definition.
//
// This tool provides comprehensive invoice details with options for different
// output formats and levels of detail.
func createInvoiceShowTool() *MCPTool {
	return &MCPTool{
		Name:        "invoice_show",
		Description: "Display comprehensive details for a specific invoice including client information, work items, and financial breakdown.",
		InputSchema: schemas.InvoiceShowSchema(),
		Examples: []MCPToolExample{
			{
				Description: "Show complete invoice details in readable format",
				Input: map[string]interface{}{
					"invoice_number":      "INV-001",
					"show_work_items":     true,
					"show_client_details": true,
				},
				ExpectedOutput: "Formatted display with client info, all work items, totals, and timestamps",
				UseCase:        "Invoice review and client communication preparation",
			},
			{
				Description: "Export invoice data as JSON for integration",
				Input: map[string]interface{}{
					"invoice_id":    "invoice_abc123",
					"output_format": "json",
				},
				ExpectedOutput: "Complete invoice data in JSON format for API integration",
				UseCase:        "Data export for external systems and integrations",
			},
			{
				Description: "Quick invoice summary without work item details",
				Input: map[string]interface{}{
					"invoice_number":  "INV-025",
					"show_work_items": false,
					"output_format":   "text",
				},
				ExpectedOutput: "Concise invoice summary with totals and basic information",
				UseCase:        "Quick status checks and summary reviews",
			},
			{
				Description: "View invoice with YAML output for documentation",
				Input: map[string]interface{}{
					"invoice_id":    "INV-2025-001",
					"output_format": "yaml",
				},
				ExpectedOutput: "Invoice data in YAML format suitable for documentation",
				UseCase:        "Documentation and configuration file generation",
			},
		},
		Category:   CategoryInvoiceManagement,
		CLICommand: "go-invoice",
		CLIArgs:    []string{"invoice", "show"},
		HelpText:   "Displays comprehensive invoice information with configurable detail levels and output formats. Supports lookup by invoice ID or number.",
		Version:    "1.0.0",
		Timeout:    15 * time.Second,
	}
}

// createInvoiceUpdateTool creates the invoice update tool definition.
//
// This tool supports selective updates of invoice properties with business
// rule validation and status transition controls.
func createInvoiceUpdateTool() *MCPTool {
	return &MCPTool{
		Name:        "invoice_update",
		Description: "Update invoice details such as status, due date, or description. Validates business rules and prevents invalid status transitions.",
		InputSchema: schemas.InvoiceUpdateSchema(),
		Examples: []MCPToolExample{
			{
				Description: "Mark invoice as sent after delivery to client",
				Input: map[string]interface{}{
					"invoice_number": "INV-001",
					"status":         "sent",
				},
				ExpectedOutput: "Invoice status updated to 'sent' with timestamp recorded",
				UseCase:        "Invoice delivery workflow and status tracking",
			},
			{
				Description: "Extend payment due date for client accommodation",
				Input: map[string]interface{}{
					"invoice_id": "invoice_abc123",
					"due_date":   "2025-09-30",
				},
				ExpectedOutput: "Due date updated with change logged for audit trail",
				UseCase:        "Payment term adjustments and client relations",
			},
			{
				Description: "Update invoice description for clarity",
				Input: map[string]interface{}{
					"invoice_number": "INV-025",
					"description":    "Q3 2025 Development Services - Updated scope",
				},
				ExpectedOutput: "Invoice description updated to reflect scope changes",
				UseCase:        "Project scope clarification and invoice documentation",
			},
			{
				Description: "Mark invoice as paid after payment received",
				Input: map[string]interface{}{
					"invoice_id": "INV-2025-008",
					"status":     "paid",
				},
				ExpectedOutput: "Invoice marked as paid, affecting financial reporting",
				UseCase:        "Payment processing and accounts receivable management",
			},
			{
				Description: "Multiple field update in single operation",
				Input: map[string]interface{}{
					"invoice_number": "INV-042",
					"status":         "sent",
					"due_date":       "2025-09-15",
					"description":    "August 2025 Consulting - Terms Updated",
				},
				ExpectedOutput: "All specified fields updated with change summary provided",
				UseCase:        "Comprehensive invoice revision after client review",
			},
		},
		Category:   CategoryInvoiceManagement,
		CLICommand: "go-invoice",
		CLIArgs:    []string{"invoice", "update"},
		HelpText:   "Updates invoice properties with business rule validation. Supports status changes, due date adjustments, and description updates while maintaining audit trail.",
		Version:    "1.0.0",
		Timeout:    20 * time.Second,
	}
}

// createInvoiceDeleteTool creates the invoice deletion tool definition.
//
// This tool provides safe invoice deletion with confirmation prompts and
// business rule validation to prevent data loss.
func createInvoiceDeleteTool() *MCPTool {
	return &MCPTool{
		Name:        "invoice_delete",
		Description: "Delete invoices with safety confirmations and business rule validation. Supports both soft delete (default) and permanent removal.",
		InputSchema: schemas.InvoiceDeleteSchema(),
		Examples: []MCPToolExample{
			{
				Description: "Soft delete draft invoice (recommended approach)",
				Input: map[string]interface{}{
					"invoice_number": "INV-001",
				},
				ExpectedOutput: "Invoice marked as deleted but data preserved for audit trail",
				UseCase:        "Removing incorrect or duplicate draft invoices safely",
			},
			{
				Description: "Permanently delete invoice with confirmation",
				Input: map[string]interface{}{
					"invoice_id":  "invoice_test123",
					"hard_delete": true,
				},
				ExpectedOutput: "Confirmation prompt followed by permanent data removal",
				UseCase:        "Complete data removal for test or erroneous invoices",
			},
			{
				Description: "Force delete without confirmation (use with extreme caution)",
				Input: map[string]interface{}{
					"invoice_number": "DRAFT-999",
					"hard_delete":    true,
					"force":          true,
				},
				ExpectedOutput: "Immediate permanent deletion without confirmation prompts",
				UseCase:        "Automated cleanup of test data or bulk operations",
			},
			{
				Description: "Attempt to delete paid invoice (will be prevented)",
				Input: map[string]interface{}{
					"invoice_id": "INV-PAID-001",
				},
				ExpectedOutput: "Deletion blocked with explanation of business rule violation",
				UseCase:        "Demonstrates business rule protection for financial integrity",
			},
		},
		Category:   CategoryInvoiceManagement,
		CLICommand: "go-invoice",
		CLIArgs:    []string{"invoice", "delete"},
		HelpText:   "Safely removes invoices with business rule validation. Default soft delete preserves audit trail; hard delete permanently removes data. Paid invoices cannot be deleted.",
		Version:    "1.0.0",
		Timeout:    25 * time.Second,
	}
}

// createInvoiceAddItemTool creates the work item addition tool definition.
//
// This tool supports adding single or multiple work items to invoices with
// automatic total recalculation and validation.
func createInvoiceAddItemTool() *MCPTool {
	return &MCPTool{
		Name:        "invoice_add_item",
		Description: "Add work items to existing invoices with automatic total calculation. Supports single or batch work item entry for draft invoices.",
		InputSchema: schemas.InvoiceAddItemSchema(),
		Examples: []MCPToolExample{
			{
				Description: "Add single work item to invoice",
				Input: map[string]interface{}{
					"invoice_number": "INV-001",
					"work_items": []map[string]interface{}{
						{
							"date":        "2025-08-03",
							"hours":       6.0,
							"rate":        150.0,
							"description": "Bug fixes and performance optimization",
						},
					},
				},
				ExpectedOutput: "Work item added to invoice with totals automatically recalculated",
				UseCase:        "Adding newly completed work to existing invoices",
			},
			{
				Description: "Batch add multiple work items for a week's work",
				Input: map[string]interface{}{
					"invoice_id": "invoice_abc123",
					"work_items": []map[string]interface{}{
						{
							"date":        "2025-08-01",
							"hours":       8.0,
							"rate":        125.0,
							"description": "Frontend component development",
						},
						{
							"date":        "2025-08-02",
							"hours":       7.5,
							"rate":        125.0,
							"description": "API integration and testing",
						},
						{
							"date":        "2025-08-03",
							"hours":       4.0,
							"rate":        125.0,
							"description": "Code review and documentation",
						},
					},
				},
				ExpectedOutput: "Multiple work items added with comprehensive total recalculation",
				UseCase:        "Bulk entry of completed work from timesheets",
			},
			{
				Description: "Add premium rate consulting work",
				Input: map[string]interface{}{
					"invoice_number": "INV-CONSULTING-025",
					"work_items": []map[string]interface{}{
						{
							"date":        "2025-08-01",
							"hours":       2.5,
							"rate":        300.0,
							"description": "Executive strategy consulting session",
						},
					},
				},
				ExpectedOutput: "High-value consulting work added with appropriate billing rate",
				UseCase:        "Adding specialized or premium rate work items",
			},
			{
				Description: "Add partial hour work with decimal precision",
				Input: map[string]interface{}{
					"invoice_id": "INV-2025-042",
					"work_items": []map[string]interface{}{
						{
							"date":        "2025-08-03",
							"hours":       1.25,
							"rate":        100.0,
							"description": "Quick bug fix - payment gateway issue",
						},
					},
				},
				ExpectedOutput: "Precise time tracking with decimal hours correctly calculated",
				UseCase:        "Accurate billing for short-duration tasks",
			},
		},
		Category:   CategoryInvoiceManagement,
		CLICommand: "go-invoice",
		CLIArgs:    []string{"import", "--add-to-invoice"},
		HelpText:   "Adds work items to draft invoices with automatic total recalculation. Supports flexible time entry and maintains billing accuracy. Only works with draft status invoices.",
		Version:    "1.0.0",
		Timeout:    25 * time.Second,
	}
}

// createInvoiceAddLineItemTool creates the flexible line item addition tool definition.
//
// This tool supports adding various types of line items to invoices including hourly work,
// fixed fees, and quantity-based charges with automatic total recalculation.
func createInvoiceAddLineItemTool() *MCPTool {
	return &MCPTool{
		Name:        "invoice_add_line_item",
		Description: "Add flexible line items to invoices with support for hourly, fixed, and quantity-based billing. Automatically calculates totals and supports mixed billing types on the same invoice.",
		InputSchema: schemas.InvoiceAddLineItemSchema(),
		Examples: []MCPToolExample{
			{
				Description: "Add hourly work item to invoice",
				Input: map[string]interface{}{
					"invoice_number": "INV-001",
					"line_items": []map[string]interface{}{
						{
							"type":        "hourly",
							"date":        "2025-08-03",
							"hours":       8.0,
							"rate":        125.0,
							"description": "Development work on authentication module",
						},
					},
				},
				ExpectedOutput: "Hourly line item added to invoice with totals automatically recalculated",
				UseCase:        "Adding time-based work to invoices (same as traditional work items)",
			},
			{
				Description: "Add monthly retainer (fixed amount)",
				Input: map[string]interface{}{
					"invoice_number": "INV-001",
					"line_items": []map[string]interface{}{
						{
							"type":        "fixed",
							"date":        "2025-08-01",
							"amount":      2000.0,
							"description": "Monthly Retainer - August 2025",
						},
					},
				},
				ExpectedOutput: "Fixed amount line item added to invoice",
				UseCase:        "Adding retainers, flat fees, or fixed charges",
			},
			{
				Description: "Add quantity-based items (licenses, materials)",
				Input: map[string]interface{}{
					"invoice_number": "INV-001",
					"line_items": []map[string]interface{}{
						{
							"type":        "quantity",
							"date":        "2025-08-01",
							"quantity":    3.0,
							"unit_price":  50.0,
							"description": "SSL certificates",
						},
					},
				},
				ExpectedOutput: "Quantity-based line item added with automatic total calculation",
				UseCase:        "Adding materials, licenses, or other unit-priced items",
			},
			{
				Description: "Mix different billing types on same invoice",
				Input: map[string]interface{}{
					"invoice_number": "INV-001",
					"line_items": []map[string]interface{}{
						{
							"type":        "hourly",
							"date":        "2025-08-01",
							"hours":       40.0,
							"rate":        125.0,
							"description": "Development work - 40 hours",
						},
						{
							"type":        "fixed",
							"date":        "2025-08-01",
							"amount":      500.0,
							"description": "Project setup fee",
						},
						{
							"type":        "quantity",
							"date":        "2025-08-01",
							"quantity":    2.0,
							"unit_price":  25.0,
							"description": "Hosting licenses",
						},
					},
				},
				ExpectedOutput: "Multiple line items of different types added with comprehensive total calculation",
				UseCase:        "Creating invoices with mixed billing models (hourly work + fixed fees + materials)",
			},
		},
		Category:   CategoryInvoiceManagement,
		CLICommand: "go-invoice invoice add-line-item",
		CLIArgs:    []string{"invoice", "add-line-item"},
		HelpText:   "Adds flexible line items to draft invoices supporting hourly, fixed, and quantity-based billing. Enables mixed billing models on the same invoice with automatic total recalculation.",
		Version:    "1.0.0",
		Timeout:    25 * time.Second,
	}
}

// createInvoiceRemoveItemTool creates the work item removal tool definition.
//
// This tool supports removing work items from invoices using various identification
// methods with safety confirmations and total recalculation.
func createInvoiceRemoveItemTool() *MCPTool {
	return &MCPTool{
		Name:        "invoice_remove_item",
		Description: "Remove work items from invoices using flexible identification methods. Supports removal by ID, description match, or date with automatic total recalculation.",
		InputSchema: schemas.InvoiceRemoveItemSchema(),
		Examples: []MCPToolExample{
			{
				Description: "Remove specific work item by ID",
				Input: map[string]interface{}{
					"invoice_number": "INV-001",
					"work_item_id":   "work_item_123",
					"confirm":        true,
				},
				ExpectedOutput: "Specific work item removed with totals recalculated",
				UseCase:        "Removing incorrectly added or duplicate work items",
			},
			{
				Description: "Remove work items matching description",
				Input: map[string]interface{}{
					"invoice_id":            "invoice_abc123",
					"work_item_description": "Bug fixes",
					"remove_all_matching":   false,
				},
				ExpectedOutput: "First work item matching description removed with confirmation prompt",
				UseCase:        "Removing work by description when ID is unknown",
			},
			{
				Description: "Remove all work from specific date",
				Input: map[string]interface{}{
					"invoice_number":      "INV-025",
					"work_item_date":      "2025-08-01",
					"remove_all_matching": true,
					"confirm":             true,
				},
				ExpectedOutput: "All work items from August 1st removed with total adjustment",
				UseCase:        "Correcting timesheet errors for specific dates",
			},
			{
				Description: "Remove duplicate entries with partial description match",
				Input: map[string]interface{}{
					"invoice_id":            "INV-2025-008",
					"work_item_description": "Frontend",
					"remove_all_matching":   true,
				},
				ExpectedOutput: "Confirmation prompt for removing all items containing 'Frontend'",
				UseCase:        "Bulk removal of duplicate or incorrectly categorized items",
			},
			{
				Description: "Attempt to remove from non-draft invoice (will be prevented)",
				Input: map[string]interface{}{
					"invoice_number": "INV-SENT-001",
					"work_item_id":   "work_123",
				},
				ExpectedOutput: "Removal blocked with explanation of business rule",
				UseCase:        "Demonstrates protection against modifying sent invoices",
			},
		},
		Category:   CategoryInvoiceManagement,
		CLICommand: "go-invoice",
		CLIArgs:    []string{"invoice", "remove-item"},
		HelpText:   "Removes work items from draft invoices using flexible identification. Supports removal by ID, description pattern, or date. Automatically recalculates totals and maintains data integrity.",
		Version:    "1.0.0",
		Timeout:    20 * time.Second,
	}
}

// RegisterInvoiceManagementTools registers all invoice management tools with the provided registry.
//
// This function provides a convenient way to register all invoice management tools
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
// - Registers 7 invoice management tools in the CategoryInvoiceManagement category
// - Tools become available for MCP client discovery and execution
//
// Notes:
// - All tools are registered with comprehensive validation schemas
// - Examples and help text are included for Claude interaction guidance
// - CLI command mappings enable direct go-invoice CLI integration
// - Respects context cancellation for responsive operations
func RegisterInvoiceManagementTools(ctx context.Context, registry ToolRegistry) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	tools := CreateInvoiceManagementTools()
	for _, tool := range tools {
		if err := registry.RegisterTool(ctx, tool); err != nil {
			return err
		}
	}

	return nil
}
