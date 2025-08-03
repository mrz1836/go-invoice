// Package schemas provides JSON schema definitions for configuration management MCP tools.
//
// This package contains comprehensive schema definitions optimized for natural language
// interaction with Claude. Each schema is designed to be intuitive and provides clear
// error messages for validation failures.
//
// Key design principles:
// - Parameter names map naturally to user requests ("section" vs "sectionID")
// - Support for various configuration operations and output formats
// - Flexible validation and initialization options
// - Comprehensive validation with helpful error messages
// - Real-world examples demonstrating various usage patterns
//
// All schemas follow JSON Schema Draft 7 specification and integrate with the
// validation system from the tools package.
package schemas

// ConfigShowSchema defines the JSON schema for displaying configuration.
//
// This schema supports configuration display with various formatting options,
// filtering capabilities, and sensitive data protection for troubleshooting.
func ConfigShowSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"section": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"all", "invoice_settings", "database", "email", "templates", "security", "integrations"},
				"default":     "all",
				"description": "Specific configuration section to display.",
			},
			"output_format": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"yaml", "json", "table", "text"},
				"default":     "yaml",
				"description": "Format for displaying the configuration.",
			},
			"include_defaults": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Include default values alongside configured values.",
			},
			"show_descriptions": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Include field descriptions and help text.",
			},
			"exclude_sensitive": map[string]interface{}{
				"type":        "boolean",
				"default":     true,
				"description": "Exclude sensitive data like passwords and API keys from output.",
			},
			"include_metadata": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Include metadata like source files and modification dates.",
			},
			"include_sources": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Show where each configuration value comes from (file, environment, default).",
			},
			"show_overrides": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Show environment-specific overrides and their precedence.",
			},
			"show_validation": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Include validation status for each configuration item.",
			},
			"environment": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"development", "staging", "production", "test"},
				"description": "Show configuration for specific environment.",
			},
			"summary_only": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Show only a summary of key configuration settings.",
			},
			"verbose": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Include detailed information and explanations.",
			},
			"api_format": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Format output for API consumption (structured, no decorative elements).",
			},
			"output_path": map[string]interface{}{
				"type":        "string",
				"description": "File path to save configuration output. Optional - displays to console if not provided.",
				"examples":    []string{"./config-backup.yaml", "/exports/config.json"},
			},
			"return_data": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Return configuration data in response instead of displaying or saving.",
			},
		},
		"additionalProperties": false,
	}
}

// ConfigValidateSchema defines the JSON schema for validating configuration.
//
// This schema supports comprehensive configuration validation with various
// levels of checking and detailed error reporting with fix suggestions.
func ConfigValidateSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"section": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"all", "invoice_settings", "database", "email", "templates", "security", "integrations"},
				"default":     "all",
				"description": "Specific configuration section to validate.",
			},
			"comprehensive": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Perform comprehensive validation including all optional checks.",
			},
			"check_dependencies": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Validate dependencies and external service connections.",
			},
			"test_connections": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Test actual connections to databases and external services.",
			},
			"security_check": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Perform security-focused validation checks.",
			},
			"performance_check": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Check performance-related configuration settings.",
			},
			"compliance_check": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Validate compliance with best practices and standards.",
			},
			"deployment_check": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Validate configuration for deployment readiness.",
			},
			"environment": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"development", "staging", "production", "test"},
				"description": "Validate configuration for specific environment requirements.",
			},
			"fail_fast": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Stop validation at first critical error instead of checking everything.",
			},
			"include_suggestions": map[string]interface{}{
				"type":        "boolean",
				"default":     true,
				"description": "Include suggestions for fixing validation errors.",
			},
			"auto_fix_safe": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Automatically fix safe/non-destructive validation issues.",
			},
			"backup_before_fix": map[string]interface{}{
				"type":        "boolean",
				"default":     true,
				"description": "Create backup before applying automatic fixes.",
			},
			"verbose": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Include detailed validation information and explanations.",
			},
			"output_format": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"detailed", "summary", "json", "report", "interactive"},
				"default":     "detailed",
				"description": "Format for validation results.",
			},
			"include_metrics": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Include validation metrics and performance data.",
			},
			"output_path": map[string]interface{}{
				"type":        "string",
				"description": "File path to save validation report. Optional - displays to console if not provided.",
				"examples":    []string{"./validation-report.html", "/reports/config-validation.json"},
			},
			"return_data": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Return validation results in response instead of displaying or saving.",
			},
		},
		"additionalProperties": false,
	}
}

// ConfigInitSchema defines the JSON schema for initializing configuration.
//
// This schema supports configuration initialization with templates, guided setup,
// and comprehensive validation to ensure proper system configuration.
func ConfigInitSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"template": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"default", "minimal", "production", "enterprise", "custom"},
				"default":     "default",
				"description": "Configuration template to use as starting point.",
			},
			"environment": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"development", "staging", "production", "test"},
				"default":     "development",
				"description": "Target environment for the configuration.",
			},
			"interactive": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Use interactive guided setup with prompts for configuration values.",
			},
			"company_name": map[string]interface{}{
				"type":        "string",
				"description": "Company name for invoice headers and branding.",
				"maxLength":   200,
				"examples":    []string{"My Consulting LLC", "Tech Solutions Inc"},
			},
			"default_tax_rate": map[string]interface{}{
				"type":        "number",
				"minimum":     0,
				"maximum":     100,
				"description": "Default tax rate percentage for invoices.",
				"examples":    []interface{}{8.5, 10.0, 0},
			},
			"currency": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"USD", "EUR", "GBP", "CAD", "AUD"},
				"default":     "USD",
				"description": "Default currency for invoices and reporting.",
			},
			"timezone": map[string]interface{}{
				"type":        "string",
				"description": "Default timezone for date/time operations.",
				"examples":    []string{"America/New_York", "Europe/London", "UTC"},
			},
			"custom_values": map[string]interface{}{
				"type":                 "object",
				"description":          "Custom configuration values to override template defaults.",
				"additionalProperties": true,
				"examples": []interface{}{
					map[string]interface{}{
						"invoice_prefix":   "MYCO",
						"default_due_days": 45,
					},
				},
			},
			"enable_integrations": map[string]interface{}{
				"type":        "array",
				"description": "List of integrations to enable and configure.",
				"items": map[string]interface{}{
					"type": "string",
					"enum": []string{"quickbooks", "stripe", "email", "slack", "webhooks"},
				},
				"uniqueItems": true,
			},
			"config_path": map[string]interface{}{
				"type":        "string",
				"description": "Path where the configuration file should be created.",
				"examples":    []string{"./config.yaml", "/etc/go-invoice/config.yaml"},
			},
			"import_from": map[string]interface{}{
				"type":        "string",
				"description": "Import configuration from existing file (for migration).",
				"examples":    []string{"./old-config.json", "/backup/config.yaml"},
			},
			"migrate_format": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Migrate configuration format during import (e.g., JSON to YAML).",
			},
			"update_schema": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Update configuration schema to latest version during import.",
			},
			"backup_existing": map[string]interface{}{
				"type":        "boolean",
				"default":     true,
				"description": "Create backup of existing configuration before overwriting.",
			},
			"backup_original": map[string]interface{}{
				"type":        "boolean",
				"default":     true,
				"description": "Create backup of original file when importing/migrating.",
			},
			"validate_after": map[string]interface{}{
				"type":        "boolean",
				"default":     true,
				"description": "Validate configuration after initialization.",
			},
			"validate_custom": map[string]interface{}{
				"type":        "boolean",
				"default":     true,
				"description": "Validate custom values against schema before applying.",
			},
			"test_connections": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Test external service connections after configuration.",
			},
			"setup_webhooks": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Set up webhooks for enabled integrations.",
			},
			"skip_optional": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Skip optional configuration items (for minimal setup).",
			},
			"auto_confirm": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Automatically confirm all prompts (non-interactive mode).",
			},
			"save_template": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Save custom configuration as reusable template.",
			},
			"return_data": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Return initialization results in response instead of just status.",
			},
		},
		"additionalProperties": false,
	}
}

// GetAllConfigSchemas returns all configuration-related schemas mapped by tool name.
//
// This function provides a centralized way to access all configuration tool schemas
// for registration with the MCP tool system.
//
// Returns:
// - map[string]map[string]interface{}: Map of tool names to their JSON schemas
//
// Notes:
// - Schema names match the corresponding tool names exactly
// - All schemas follow JSON Schema Draft 7 specification
// - Schemas are optimized for Claude natural language interaction
func GetAllConfigSchemas() map[string]map[string]interface{} {
	return map[string]map[string]interface{}{
		"config_show":     ConfigShowSchema(),
		"config_validate": ConfigValidateSchema(),
		"config_init":     ConfigInitSchema(),
	}
}

// GetConfigToolSchema returns the schema for a specific configuration tool.
//
// This function provides a way to get the schema for a specific configuration tool
// for use with external validation systems.
//
// Parameters:
// - toolName: Name of the configuration tool to get schema for
//
// Returns:
// - map[string]interface{}: JSON schema for the tool, or nil if not found
// - bool: True if tool exists, false otherwise
//
// Notes:
// - Returns nil schema and false for unknown tool names
// - Schema can be used with any JSON Schema Draft 7 validator
// - Tool names match the MCP tool names exactly
func GetConfigToolSchema(toolName string) (map[string]interface{}, bool) {
	schemas := GetAllConfigSchemas()
	schema, exists := schemas[toolName]
	return schema, exists
}
