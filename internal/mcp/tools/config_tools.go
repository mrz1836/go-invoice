// Package tools provides MCP tool definitions for configuration management operations.
//
// This package implements comprehensive configuration management tools optimized for natural
// language interaction with Claude. Each tool provides intuitive parameter names,
// flexible configuration options, and detailed examples for various use cases.
//
// Key features:
// - Natural language-friendly parameter names and descriptions
// - Support for configuration viewing, validation, and initialization
// - Comprehensive error checking and validation reports
// - Safe configuration updates with backup options
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

// CreateConfigurationManagementTools creates all configuration management MCP tool definitions.
//
// This function initializes comprehensive configuration management tools optimized for
// conversational interaction with Claude. Each tool is designed to handle common
// configuration workflows naturally and provide clear guidance for parameter usage.
//
// Returns:
// - []*MCPTool: Complete set of configuration management tools ready for registration
//
// Tools created:
// 1. config_show - Display current configuration with formatting options
// 2. config_validate - Validate configuration integrity and report issues
// 3. config_init - Initialize new configuration with guided setup
//
// Notes:
// - All tools use the CategoryConfiguration category for organization
// - Schemas are optimized for natural language parameter entry
// - Examples cover both simple and complex use cases
// - CLI integration follows standard go-invoice command patterns
func CreateConfigurationManagementTools() []*MCPTool {
	return []*MCPTool{
		createConfigShowTool(),
		createConfigValidateTool(),
		createConfigInitTool(),
	}
}

// createConfigShowTool creates the configuration display tool definition.
//
// This tool supports displaying current configuration with various formatting
// options and filtering capabilities for different use cases.
func createConfigShowTool() *MCPTool {
	return &MCPTool{
		Name:        "config_show",
		Description: "Display current system configuration with flexible formatting options, filtering capabilities, and sensitive data protection for troubleshooting and review.",
		InputSchema: schemas.ConfigShowSchema(),
		Examples: []MCPToolExample{
			{
				Description: "Show complete configuration in readable format",
				Input: map[string]interface{}{
					"output_format":     "yaml",
					"include_defaults":  true,
					"show_descriptions": true,
				},
				ExpectedOutput: "Complete configuration in YAML format with defaults and field descriptions",
				UseCase:        "Configuration review and documentation",
			},
			{
				Description: "Show specific configuration section for focused review",
				Input: map[string]interface{}{
					"section":       "invoice_settings",
					"output_format": "json",
					"verbose":       true,
				},
				ExpectedOutput: "Detailed invoice settings configuration with metadata",
				UseCase:        "Focused configuration review for specific functionality",
			},
			{
				Description: "Export configuration for backup or sharing",
				Input: map[string]interface{}{
					"output_format":     "json",
					"exclude_sensitive": true,
					"include_metadata":  true,
					"output_path":       "./config-backup.json",
				},
				ExpectedOutput: "Configuration exported to file with sensitive data excluded",
				UseCase:        "Configuration backup and team sharing",
			},
			{
				Description: "Quick configuration summary for status checking",
				Input: map[string]interface{}{
					"summary_only":    true,
					"output_format":   "table",
					"show_validation": true,
				},
				ExpectedOutput: "Summary table showing key settings and validation status",
				UseCase:        "Quick health check and configuration status overview",
			},
			{
				Description: "Show configuration with environment-specific overrides",
				Input: map[string]interface{}{
					"environment":     "production",
					"show_overrides":  true,
					"output_format":   "yaml",
					"include_sources": true,
				},
				ExpectedOutput: "Configuration with production overrides and source information",
				UseCase:        "Environment-specific configuration debugging",
			},
			{
				Description: "Display configuration for API integration",
				Input: map[string]interface{}{
					"output_format": "json",
					"api_format":    true,
					"return_data":   true,
				},
				ExpectedOutput: "Configuration in API-friendly JSON format returned in memory",
				UseCase:        "Integration with configuration management APIs",
			},
		},
		Category:   CategoryConfiguration,
		CLICommand: "go-invoice",
		CLIArgs:    []string{"config", "show"},
		HelpText:   "Displays system configuration with flexible formatting and filtering. Supports sensitive data protection and environment-specific views for troubleshooting.",
		Version:    "1.0.0",
		Timeout:    20 * time.Second,
	}
}

// createConfigValidateTool creates the configuration validation tool definition.
//
// This tool provides comprehensive configuration validation with detailed
// error reporting and suggestions for fixing configuration issues.
func createConfigValidateTool() *MCPTool {
	return &MCPTool{
		Name:        "config_validate",
		Description: "Validate system configuration integrity with comprehensive error checking, dependency validation, and actionable recommendations for fixing issues.",
		InputSchema: schemas.ConfigValidateSchema(),
		Examples: []MCPToolExample{
			{
				Description: "Perform complete configuration validation check",
				Input: map[string]interface{}{
					"comprehensive":      true,
					"check_dependencies": true,
					"output_format":      "detailed",
				},
				ExpectedOutput: "Comprehensive validation report with all checks and dependency verification",
				UseCase:        "Complete system health check and configuration audit",
			},
			{
				Description: "Validate specific configuration section",
				Input: map[string]interface{}{
					"section":          "database",
					"test_connections": true,
					"verbose":          true,
				},
				ExpectedOutput: "Database configuration validation with connection testing results",
				UseCase:        "Focused validation after configuration changes",
			},
			{
				Description: "Quick validation for deployment readiness",
				Input: map[string]interface{}{
					"deployment_check": true,
					"fail_fast":        true,
					"output_format":    "summary",
				},
				ExpectedOutput: "Quick deployment readiness check with pass/fail status",
				UseCase:        "Pre-deployment configuration verification",
			},
			{
				Description: "Validate configuration with fix suggestions",
				Input: map[string]interface{}{
					"include_suggestions": true,
					"auto_fix_safe":       true,
					"backup_before_fix":   true,
					"output_format":       "interactive",
				},
				ExpectedOutput: "Validation report with automatic safe fixes and manual suggestions",
				UseCase:        "Configuration troubleshooting with guided repair options",
			},
			{
				Description: "Validate configuration against specific environment requirements",
				Input: map[string]interface{}{
					"environment":       "production",
					"security_check":    true,
					"performance_check": true,
					"compliance_check":  true,
				},
				ExpectedOutput: "Production-focused validation with security and performance analysis",
				UseCase:        "Production deployment validation and compliance checking",
			},
			{
				Description: "Export validation results for documentation",
				Input: map[string]interface{}{
					"comprehensive":   true,
					"output_format":   "report",
					"include_metrics": true,
					"output_path":     "./validation-report.html",
				},
				ExpectedOutput: "Detailed HTML validation report with metrics and recommendations",
				UseCase:        "Documentation and audit trail for configuration compliance",
			},
		},
		Category:   CategoryConfiguration,
		CLICommand: "go-invoice",
		CLIArgs:    []string{"config", "validate"},
		HelpText:   "Validates configuration integrity with comprehensive checks, dependency verification, and actionable fix suggestions. Supports deployment readiness verification.",
		Version:    "1.0.0",
		Timeout:    45 * time.Second,
	}
}

// createConfigInitTool creates the configuration initialization tool definition.
//
// This tool supports initializing new configurations with guided setup,
// templates, and validation to ensure proper system configuration.
func createConfigInitTool() *MCPTool {
	return &MCPTool{
		Name:        "config_init",
		Description: "Initialize new system configuration with guided setup, templates, and validation. Supports multiple deployment scenarios and customization options.",
		InputSchema: schemas.ConfigInitSchema(),
		Examples: []MCPToolExample{
			{
				Description: "Initialize configuration with interactive guided setup",
				Input: map[string]interface{}{
					"interactive":    true,
					"template":       "default",
					"environment":    "development",
					"validate_after": true,
				},
				ExpectedOutput: "Interactive setup process with prompts for all required configuration values",
				UseCase:        "First-time system setup with guided configuration",
			},
			{
				Description: "Initialize production configuration from template",
				Input: map[string]interface{}{
					"template":         "production",
					"environment":      "production",
					"company_name":     "My Consulting LLC",
					"default_tax_rate": 8.5,
					"backup_existing":  true,
					"config_path":      "./config/production.yaml",
				},
				ExpectedOutput: "Production configuration initialized with company-specific settings",
				UseCase:        "Production deployment setup with company branding",
			},
			{
				Description: "Initialize minimal configuration for testing",
				Input: map[string]interface{}{
					"template":      "minimal",
					"environment":   "test",
					"skip_optional": true,
					"auto_confirm":  true,
				},
				ExpectedOutput: "Minimal test configuration with only essential settings",
				UseCase:        "Test environment setup and CI/CD integration",
			},
			{
				Description: "Initialize configuration with custom values",
				Input: map[string]interface{}{
					"template": "custom",
					"custom_values": map[string]interface{}{
						"invoice_prefix":   "MYCO",
						"default_due_days": 45,
						"currency":         "USD",
						"timezone":         "America/New_York",
					},
					"validate_custom": true,
					"save_template":   true,
				},
				ExpectedOutput: "Custom configuration with specified values and template saved for reuse",
				UseCase:        "Customized setup for specific business requirements",
			},
			{
				Description: "Initialize configuration from existing file",
				Input: map[string]interface{}{
					"import_from":     "./old-config.json",
					"migrate_format":  true,
					"update_schema":   true,
					"backup_original": true,
					"config_path":     "./config/migrated.yaml",
				},
				ExpectedOutput: "Configuration migrated from existing file with format updates",
				UseCase:        "Configuration migration and format upgrades",
			},
			{
				Description: "Initialize configuration with external service integration",
				Input: map[string]interface{}{
					"template":            "enterprise",
					"enable_integrations": []string{"quickbooks", "stripe", "email"},
					"test_connections":    true,
					"setup_webhooks":      true,
				},
				ExpectedOutput: "Enterprise configuration with integrations tested and configured",
				UseCase:        "Enterprise deployment with external service integrations",
			},
		},
		Category:   CategoryConfiguration,
		CLICommand: "go-invoice",
		CLIArgs:    []string{"config", "init"},
		HelpText:   "Initializes system configuration with templates, guided setup, and validation. Supports migration from existing configurations and enterprise integrations.",
		Version:    "1.0.0",
		Timeout:    120 * time.Second,
	}
}

// CreateConfigTools is an alias for CreateConfigurationManagementTools for test compatibility.
func CreateConfigTools() []*MCPTool {
	return CreateConfigurationManagementTools()
}

// RegisterConfigTools is an alias for RegisterConfigurationManagementTools for test compatibility.
func RegisterConfigTools(ctx context.Context, registry ToolRegistry) error {
	return RegisterConfigurationManagementTools(ctx, registry)
}

// RegisterConfigurationManagementTools registers all configuration management tools with the provided registry.
//
// This function provides a convenient way to register all configuration management tools
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
// - Registers 3 configuration management tools in the CategoryConfiguration category
// - Tools become available for MCP client discovery and execution
//
// Notes:
// - All tools are registered with comprehensive validation schemas
// - Examples and help text are included for Claude interaction guidance
// - CLI command mappings enable direct go-invoice CLI integration
// - Respects context cancellation for responsive operations
func RegisterConfigurationManagementTools(ctx context.Context, registry ToolRegistry) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	tools := CreateConfigurationManagementTools()
	for _, tool := range tools {
		if err := registry.RegisterTool(ctx, tool); err != nil {
			return err
		}
	}

	return nil
}
