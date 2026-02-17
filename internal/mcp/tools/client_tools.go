// Package tools provides MCP tool definitions for client management operations.
//
// This package implements comprehensive client management tools optimized for natural
// language interaction with Claude. Each tool provides intuitive parameter names,
// flexible identification methods, and detailed examples for various use cases.
//
// Key features:
// - Natural language-friendly parameter names and descriptions
// - Support for multiple identification methods (ID, name, email)
// - Flexible contact information management and validation
// - Comprehensive examples covering common business scenarios
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

// CreateClientManagementTools creates all client management MCP tool definitions.
//
// This function initializes comprehensive client management tools optimized for
// conversational interaction with Claude. Each tool is designed to handle common
// client workflows naturally and provide clear guidance for parameter usage.
//
// Returns:
// - []*MCPTool: Complete set of client management tools ready for registration
//
// Tools created:
// 1. client_create - Create new clients with contact information
// 2. client_list - List and filter clients with flexible criteria
// 3. client_show - Display detailed client information and invoice history
// 4. client_update - Modify client information and contact details
// 5. client_delete - Remove clients with dependency checking
//
// Notes:
// - All tools use the CategoryClientManagement category for organization
// - Schemas are optimized for natural language parameter entry
// - Examples cover both simple and complex use cases
// - CLI integration follows standard go-invoice command patterns
func CreateClientManagementTools() []*MCPTool {
	return []*MCPTool{
		createClientCreateTool(),
		createClientListTool(),
		createClientShowTool(),
		createClientUpdateTool(),
		createClientDeleteTool(),
	}
}

// createClientCreateTool creates the client creation tool definition.
//
// This tool supports creating clients with comprehensive contact information
// and business relationship details. It handles validation and uniqueness checking.
func createClientCreateTool() *MCPTool {
	return &MCPTool{
		Name:        "client_create",
		Description: "Create and register a new client with contact information and business details. This tool will create comprehensive client records and validate email uniqueness and contact information completeness.",
		InputSchema: schemas.ClientCreateSchema(),
		Examples: []MCPToolExample{
			{
				Description: "Create a simple client with basic contact information",
				Input: map[string]interface{}{
					"name":  "Acme Corporation",
					"email": "contact@acme.com",
				},
				ExpectedOutput: "Client created successfully with auto-generated ID and basic contact information",
				UseCase:        "Quick client setup for new business relationships",
			},
			{
				Description: "Create client with complete contact and business information",
				Input: map[string]interface{}{
					"name":    "Tech Solutions Inc",
					"email":   "billing@techsolutions.com",
					"phone":   "+1-555-123-4567",
					"address": "456 Business Ave, Suite 200, Tech City, TC 12345",
					"tax_id":  "EIN-98-7654321",
				},
				ExpectedOutput: "Complete client profile created with all contact details and tax information",
				UseCase:        "Comprehensive client onboarding for major accounts",
			},
			{
				Description: "Create individual contractor client",
				Input: map[string]interface{}{
					"name":    "John Smith Consulting",
					"email":   "john@smithconsulting.com",
					"phone":   "(555) 987-6543",
					"address": "123 Freelancer Lane, Remote City, RC 54321",
				},
				ExpectedOutput: "Individual contractor client created with professional contact details",
				UseCase:        "Setting up freelancers and individual service providers",
			},
			{
				Description: "Create international client with proper formatting",
				Input: map[string]interface{}{
					"name":    "European Tech Partners GmbH",
					"email":   "kontakt@eurotech.de",
					"phone":   "+49-30-12345678",
					"address": "Hauptstraße 42, 10117 Berlin, Germany",
					"tax_id":  "DE123456789",
				},
				ExpectedOutput: "International client created with European contact format and VAT number",
				UseCase:        "Managing international business relationships and compliance",
			},
		},
		Category:   CategoryClientManagement,
		CLICommand: "go-invoice",
		CLIArgs:    []string{"client", "create"},
		HelpText:   "Creates clients with automatic ID generation and contact validation. Ensures email uniqueness and validates contact information format for professional correspondence.",
		Version:    "1.0.0",
		Timeout:    30 * time.Second,
	}
}

// createClientListTool creates the client listing and filtering tool definition.
//
// This tool provides comprehensive client discovery with flexible filtering
// and multiple output formats for different business workflows.
func createClientListTool() *MCPTool {
	return &MCPTool{
		Name:        "client_list",
		Description: "List and filter clients with flexible search criteria. Supports filtering by status, name search, and provides multiple output formats with business intelligence.",
		InputSchema: schemas.ClientListSchema(),
		Examples: []MCPToolExample{
			{
				Description: "List all active clients with contact information",
				Input: map[string]interface{}{
					"active_only":    true,
					"output_format":  "table",
					"include_totals": true,
				},
				ExpectedOutput: "Table view of active clients with names, emails, phone numbers, and billing totals",
				UseCase:        "Client relationship management and contact directory maintenance",
			},
			{
				Description: "Search for clients by partial name match",
				Input: map[string]interface{}{
					"name_search":   "Tech",
					"output_format": "table",
					"sort_by":       "name",
					"sort_order":    "asc",
				},
				ExpectedOutput: "All clients with 'Tech' in their name, sorted alphabetically",
				UseCase:        "Finding specific clients or client groups for targeted communication",
			},
			{
				Description: "Export client data for CRM integration",
				Input: map[string]interface{}{
					"output_format":    "json",
					"include_invoices": true,
					"include_totals":   true,
					"active_only":      true,
				},
				ExpectedOutput: "JSON export with client details, invoice summaries, and financial totals",
				UseCase:        "CRM system integration and business intelligence reporting",
			},
			{
				Description: "Generate client contact list for marketing",
				Input: map[string]interface{}{
					"output_format": "csv",
					"active_only":   true,
					"limit":         100,
					"sort_by":       "created_date",
					"sort_order":    "desc",
				},
				ExpectedOutput: "CSV export of recent active clients for marketing campaigns",
				UseCase:        "Marketing automation and customer communication workflows",
			},
			{
				Description: "Review inactive clients for reactivation",
				Input: map[string]interface{}{
					"active_only":      false,
					"inactive_only":    true,
					"sort_by":          "last_invoice_date",
					"sort_order":       "desc",
					"include_invoices": true,
				},
				ExpectedOutput: "Inactive clients sorted by last activity for reactivation campaigns",
				UseCase:        "Customer retention and reactivation strategy planning",
			},
		},
		Category:   CategoryClientManagement,
		CLICommand: "go-invoice",
		CLIArgs:    []string{"client", "list"},
		HelpText:   "Provides comprehensive client discovery with business intelligence features. Supports name search, status filtering, and various export formats for CRM integration and business analysis.",
		Version:    "1.0.0",
		Timeout:    20 * time.Second,
	}
}

// createClientShowTool creates the client detail display tool definition.
//
// This tool provides comprehensive client details with invoice history,
// financial summaries, and relationship insights.
func createClientShowTool() *MCPTool {
	return &MCPTool{
		Name:        "client_show",
		Description: "Display comprehensive client details including contact information, invoice history, financial summary, and business relationship insights.",
		InputSchema: schemas.ClientShowSchema(),
		Examples: []MCPToolExample{
			{
				Description: "Show complete client profile with invoice history",
				Input: map[string]interface{}{
					"client_name":      "Acme Corporation",
					"include_invoices": true,
					"include_totals":   true,
					"output_format":    "text",
				},
				ExpectedOutput: "Formatted client profile with contact details, all invoices, payment history, and financial summary",
				UseCase:        "Client relationship review and account management preparation",
			},
			{
				Description: "Export client data for external CRM system",
				Input: map[string]interface{}{
					"client_email":     "contact@techsolutions.com",
					"output_format":    "json",
					"include_invoices": true,
					"include_totals":   true,
				},
				ExpectedOutput: "Complete client data in JSON format with invoice details and financial metrics",
				UseCase:        "CRM integration and external system data synchronization",
			},
			{
				Description: "Quick client lookup for support calls",
				Input: map[string]interface{}{
					"client_id":        "client_123",
					"include_invoices": false,
					"output_format":    "text",
				},
				ExpectedOutput: "Basic client information without invoice details for quick reference",
				UseCase:        "Customer support and quick client information lookup",
			},
			{
				Description: "Generate client summary for executive reporting",
				Input: map[string]interface{}{
					"client_name":    "Tech Solutions Inc",
					"output_format":  "yaml",
					"include_totals": true,
					"summary_only":   true,
				},
				ExpectedOutput: "Executive summary with key metrics and business relationship status",
				UseCase:        "Executive reporting and strategic account review",
			},
			{
				Description: "Detailed client analysis for collections workflow",
				Input: map[string]interface{}{
					"client_id":        "client_456",
					"include_invoices": true,
					"include_overdue":  true,
					"include_totals":   true,
					"output_format":    "text",
				},
				ExpectedOutput: "Client profile with emphasis on outstanding invoices and payment patterns",
				UseCase:        "Collections management and payment follow-up planning",
			},
		},
		Category:   CategoryClientManagement,
		CLICommand: "go-invoice",
		CLIArgs:    []string{"client", "show"},
		HelpText:   "Displays comprehensive client information with configurable detail levels and business intelligence. Supports lookup by client ID, name, or email with financial analytics.",
		Version:    "1.0.0",
		Timeout:    15 * time.Second,
	}
}

// createClientUpdateTool creates the client update tool definition.
//
// This tool supports selective updates of client properties with business
// rule validation and contact information management.
func createClientUpdateTool() *MCPTool {
	return &MCPTool{
		Name:        "client_update",
		Description: "Update client contact information, business details, and status. Validates email uniqueness and maintains contact information integrity.",
		InputSchema: schemas.ClientUpdateSchema(),
		Examples: []MCPToolExample{
			{
				Description: "Update client contact information after business move",
				Input: map[string]interface{}{
					"client_name": "Acme Corporation",
					"phone":       "+1-555-999-8888",
					"address":     "789 New Business Plaza, Suite 500, Metro City, MC 67890",
				},
				ExpectedOutput: "Client contact information updated with new office location and phone number",
				UseCase:        "Maintaining current contact information for business relationships",
			},
			{
				Description: "Change client email address after company rebrand",
				Input: map[string]interface{}{
					"client_id": "client_123",
					"name":      "Acme Technologies LLC",
					"email":     "billing@acmetech.com",
				},
				ExpectedOutput: "Client name and email updated to reflect company rebrand and new domain",
				UseCase:        "Managing client business changes and corporate restructuring",
			},
			{
				Description: "Add missing tax ID for client compliance reporting",
				Input: map[string]interface{}{
					"client_email": "contact@smallbiz.com",
					"tax_id":       "EIN-12-3456789",
					"email":        "compliance@smallbiz.com",
				},
				ExpectedOutput: "Tax ID added to client profile for compliance and invoicing requirements",
				UseCase:        "Compliance management and tax reporting preparation",
			},
			{
				Description: "Update international client with new VAT number",
				Input: map[string]interface{}{
					"client_name": "European Tech Partners",
					"tax_id":      "DE987654321",
					"address":     "Neue Straße 15, 10178 Berlin, Germany",
				},
				ExpectedOutput: "International client updated with new VAT number and registered address",
				UseCase:        "International compliance and regulatory requirement management",
			},
			{
				Description: "Complete contact information for partial client record",
				Input: map[string]interface{}{
					"client_id": "client_incomplete",
					"phone":     "+1-555-777-9999",
					"address":   "456 Complete Ave, Full City, FC 11111",
					"tax_id":    "EIN-55-5555555",
				},
				ExpectedOutput: "Client profile completed with all missing contact and business information",
				UseCase:        "Data completion and client relationship enhancement",
			},
		},
		Category:   CategoryClientManagement,
		CLICommand: "go-invoice",
		CLIArgs:    []string{"client", "update"},
		HelpText:   "Updates client information with validation and business rule enforcement. Supports contact information changes, business detail updates, and maintains data integrity with uniqueness checking.",
		Version:    "1.0.0",
		Timeout:    20 * time.Second,
	}
}

// createClientDeleteTool creates the client deletion tool definition.
//
// This tool provides safe client removal with dependency checking and
// business rule validation to prevent data integrity issues.
func createClientDeleteTool() *MCPTool {
	return &MCPTool{
		Name:        "client_delete",
		Description: "Remove clients with comprehensive dependency checking and business rule validation. Supports both deactivation (soft delete) and permanent removal with safety confirmations.",
		InputSchema: schemas.ClientDeleteSchema(),
		Examples: []MCPToolExample{
			{
				Description: "Deactivate client with no recent activity (recommended approach)",
				Input: map[string]interface{}{
					"client_name": "Inactive Corp",
					"soft_delete": true,
				},
				ExpectedOutput: "Client deactivated but data preserved for historical reporting and audit trail",
				UseCase:        "Managing inactive clients while preserving business history",
			},
			{
				Description: "Permanently remove test client data",
				Input: map[string]interface{}{
					"client_id":   "test_client_123",
					"hard_delete": true,
				},
				ExpectedOutput: "Confirmation prompt followed by complete client data removal",
				UseCase:        "Cleaning up test data and erroneous client records",
			},
			{
				Description: "Force delete without confirmation for automation",
				Input: map[string]interface{}{
					"client_email": "demo@example.com",
					"hard_delete":  true,
					"force":        true,
				},
				ExpectedOutput: "Immediate permanent deletion without confirmation prompts",
				UseCase:        "Automated client cleanup scripts and bulk data management",
			},
			{
				Description: "Attempt to delete client with active invoices (will be prevented)",
				Input: map[string]interface{}{
					"client_name": "Active Client Corp",
				},
				ExpectedOutput: "Deletion blocked with explanation of active invoices and business rule violation",
				UseCase:        "Demonstrates business rule protection for active client relationships",
			},
			{
				Description: "Deactivate client but preserve all historical data",
				Input: map[string]interface{}{
					"client_id":     "client_historical",
					"soft_delete":   true,
					"preserve_data": true,
				},
				ExpectedOutput: "Client marked inactive with complete data preservation for compliance",
				UseCase:        "Regulatory compliance and historical data preservation requirements",
			},
		},
		Category:   CategoryClientManagement,
		CLICommand: "go-invoice",
		CLIArgs:    []string{"client", "delete"},
		HelpText:   "Safely removes clients with comprehensive business rule validation. Default soft delete preserves audit trail and historical data. Hard delete requires confirmation and dependency checking.",
		Version:    "1.0.0",
		Timeout:    25 * time.Second,
	}
}

// RegisterClientManagementTools registers all client management tools with the provided registry.
//
// This function provides a convenient way to register all client management tools
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
// - Registers 5 client management tools in the CategoryClientManagement category
// - Tools become available for MCP client discovery and execution
//
// Notes:
// - All tools are registered with comprehensive validation schemas
// - Examples and help text are included for Claude interaction guidance
// - CLI command mappings enable direct go-invoice CLI integration
// - Respects context cancellation for responsive operations
func RegisterClientManagementTools(ctx context.Context, registry ToolRegistry) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	tools := CreateClientManagementTools()
	for _, tool := range tools {
		if err := registry.RegisterTool(ctx, tool); err != nil {
			return err
		}
	}

	return nil
}
