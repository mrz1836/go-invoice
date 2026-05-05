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
		keyType: keyObject,
		keyProperties: map[string]interface{}{
			"section": map[string]interface{}{
				keyType:        typeString,
				keyEnum:        []string{"all", "invoice_settings", "database", keyEmail, "templates", "security", "integrations"},
				keyDefault:     "all",
				keyDescription: "Specific configuration section to display.",
			},
			"output_format": map[string]interface{}{
				keyType:        typeString,
				keyEnum:        []string{"yaml", typeJSON, "table", "text"},
				keyDefault:     "yaml",
				keyDescription: "Format for displaying the configuration.",
			},
			"include_defaults": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Include default values alongside configured values.",
			},
			"show_descriptions": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Include field descriptions and help text.",
			},
			"exclude_sensitive": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     true,
				keyDescription: "Exclude sensitive data like passwords and API keys from output.",
			},
			"include_metadata": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Include metadata like source files and modification dates.",
			},
			"include_sources": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Show where each configuration value comes from (file, environment, default).",
			},
			"show_overrides": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Show environment-specific overrides and their precedence.",
			},
			"show_validation": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Include validation status for each configuration item.",
			},
			"environment": map[string]interface{}{
				keyType:        typeString,
				keyEnum:        []string{"development", "staging", "production", "test"},
				keyDescription: "Show configuration for specific environment.",
			},
			"summary_only": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Show only a summary of key configuration settings.",
			},
			"verbose": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Include detailed information and explanations.",
			},
			"api_format": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Format output for API consumption (structured, no decorative elements).",
			},
			"output_path": map[string]interface{}{
				keyType:        typeString,
				keyDescription: "File path to save configuration output. Optional - displays to console if not provided.",
				keyExamples:    []string{"./config-backup.yaml", "/exports/config.json"},
			},
			"return_data": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Return configuration data in response instead of displaying or saving.",
			},
		},
		keyAdditionalProperties: false,
	}
}

// ConfigValidateSchema defines the JSON schema for validating configuration.
//
// This schema supports comprehensive configuration validation with various
// levels of checking and detailed error reporting with fix suggestions.
func ConfigValidateSchema() map[string]interface{} {
	return map[string]interface{}{
		keyType: keyObject,
		keyProperties: map[string]interface{}{
			"section": map[string]interface{}{
				keyType:        typeString,
				keyEnum:        []string{"all", "invoice_settings", "database", keyEmail, "templates", "security", "integrations"},
				keyDefault:     "all",
				keyDescription: "Specific configuration section to validate.",
			},
			"comprehensive": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Perform comprehensive validation including all optional checks.",
			},
			"check_dependencies": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Validate dependencies and external service connections.",
			},
			"test_connections": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Test actual connections to databases and external services.",
			},
			"security_check": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Perform security-focused validation checks.",
			},
			"performance_check": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Check performance-related configuration settings.",
			},
			"compliance_check": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Validate compliance with best practices and standards.",
			},
			"deployment_check": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Validate configuration for deployment readiness.",
			},
			"environment": map[string]interface{}{
				keyType:        typeString,
				keyEnum:        []string{"development", "staging", "production", "test"},
				keyDescription: "Validate configuration for specific environment requirements.",
			},
			"fail_fast": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Stop validation at first critical error instead of checking everything.",
			},
			"include_suggestions": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     true,
				keyDescription: "Include suggestions for fixing validation errors.",
			},
			"auto_fix_safe": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Automatically fix safe/non-destructive validation issues.",
			},
			"backup_before_fix": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     true,
				keyDescription: "Create backup before applying automatic fixes.",
			},
			"verbose": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Include detailed validation information and explanations.",
			},
			"output_format": map[string]interface{}{
				keyType:        typeString,
				keyEnum:        []string{"detailed", "summary", typeJSON, "report", "interactive"},
				keyDefault:     "detailed",
				keyDescription: "Format for validation results.",
			},
			"include_metrics": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Include validation metrics and performance data.",
			},
			"output_path": map[string]interface{}{
				keyType:        typeString,
				keyDescription: "File path to save validation report. Optional - displays to console if not provided.",
				keyExamples:    []string{"./validation-report.html", "/reports/config-validation.json"},
			},
			"return_data": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Return validation results in response instead of displaying or saving.",
			},
		},
		keyAdditionalProperties: false,
	}
}

// ConfigInitSchema defines the JSON schema for initializing configuration.
//
// This schema supports configuration initialization with templates, guided setup,
// and comprehensive validation to ensure proper system configuration.
func ConfigInitSchema() map[string]interface{} {
	return map[string]interface{}{
		keyType: keyObject,
		keyProperties: map[string]interface{}{
			"template": map[string]interface{}{
				keyType:        typeString,
				keyEnum:        []string{"default", "minimal", "production", "enterprise", "custom"},
				keyDefault:     "default",
				keyDescription: "Configuration template to use as starting point.",
			},
			"environment": map[string]interface{}{
				keyType:        typeString,
				keyEnum:        []string{"development", "staging", "production", "test"},
				keyDefault:     "development",
				keyDescription: "Target environment for the configuration.",
			},
			"interactive": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Use interactive guided setup with prompts for configuration values.",
			},
			"company_name": map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Company name for invoice headers and branding.",
				keyMaxLength:   200,
				keyExamples:    []string{"My Consulting LLC", "Tech Solutions Inc"},
			},
			"default_tax_rate": map[string]interface{}{
				keyType:        typeNumber,
				keyMinimum:     0,
				keyMaximum:     100,
				keyDescription: "Default tax rate percentage for invoices.",
				keyExamples:    []interface{}{8.5, 10.0, 0},
			},
			"currency": map[string]interface{}{
				keyType:        typeString,
				keyEnum:        []string{"USD", "EUR", "GBP", "CAD", "AUD"},
				keyDefault:     "USD",
				keyDescription: "Default currency for invoices and reporting.",
			},
			"timezone": map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Default timezone for date/time operations.",
				keyExamples:    []string{"America/New_York", "Europe/London", "UTC"},
			},
			"custom_values": map[string]interface{}{
				keyType:                 keyObject,
				keyDescription:          "Custom configuration values to override template defaults.",
				keyAdditionalProperties: true,
				keyExamples: []interface{}{
					map[string]interface{}{
						"invoice_prefix":   "MYCO",
						"default_due_days": 45,
					},
				},
			},
			"enable_integrations": map[string]interface{}{
				keyType:        typeArray,
				keyDescription: "List of integrations to enable and configure.",
				"items": map[string]interface{}{
					keyType: typeString,
					keyEnum: []string{"quickbooks", "stripe", keyEmail, "slack", "webhooks"},
				},
				"uniqueItems": true,
			},
			"config_path": map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Path where the configuration file should be created.",
				keyExamples:    []string{"./config.yaml", "/etc/go-invoice/config.yaml"},
			},
			"config_file": map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Configuration file path (alias for config_path).",
				keyExamples:    []string{"./config.yaml", "/etc/go-invoice/config.yaml"},
			},
			"import_from": map[string]interface{}{
				keyType:        typeString,
				keyDescription: "Import configuration from existing file (for migration).",
				keyExamples:    []string{"./old-config.json", "/backup/config.yaml"},
			},
			"migrate_format": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Migrate configuration format during import (e.g., JSON to YAML).",
			},
			"update_schema": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Update configuration schema to latest version during import.",
			},
			"backup_existing": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     true,
				keyDescription: "Create backup of existing configuration before overwriting.",
			},
			"backup_original": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     true,
				keyDescription: "Create backup of original file when importing/migrating.",
			},
			"validate_after": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     true,
				keyDescription: "Validate configuration after initialization.",
			},
			"validate_custom": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     true,
				keyDescription: "Validate custom values against schema before applying.",
			},
			"test_connections": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Test external service connections after configuration.",
			},
			"setup_webhooks": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Set up webhooks for enabled integrations.",
			},
			"skip_optional": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Skip optional configuration items (for minimal setup).",
			},
			"auto_confirm": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Automatically confirm all prompts (non-interactive mode).",
			},
			"save_template": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Save custom configuration as reusable template.",
			},
			"return_data": map[string]interface{}{
				keyType:        typeBoolean,
				keyDefault:     false,
				keyDescription: "Return initialization results in response instead of just status.",
			},
		},
		keyAdditionalProperties: false,
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
