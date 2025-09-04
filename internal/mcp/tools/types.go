// Package tools provides the foundational architecture for MCP tool definitions and management.
//
// This package implements the core interfaces and types needed for MCP tool registration,
// validation, and execution within the go-invoice MCP server. It follows context-first design
// principles and consumer-driven interface patterns as defined in .github/AGENTS.md.
//
// Key features include:
// - MCPTool struct with schema, examples, and metadata support
// - ToolRegistry interface for consumer-driven tool discovery
// - InputValidator interface for comprehensive schema validation
// - Tool categorization system for natural language interaction
// - Context-aware operations with proper cancellation support
//
// The package is designed for conversational interaction with Claude, providing clear error
// messages and actionable guidance for tool usage. All operations accept context.Context
// for timeout and cancellation support.
//
// This package is part of the larger go-invoice MCP integration and interacts with the
// command execution bridge and response processing systems.
package tools

import (
	"context"
	"time"
)

// MCPTool represents a complete MCP tool definition with schema, examples, and metadata.
//
// This struct encapsulates all information needed for Claude to understand and interact
// with a specific go-invoice CLI command through the MCP protocol.
//
// Fields:
// - Name: Unique tool identifier for MCP protocol
// - Description: Human-readable description for Claude interaction
// - InputSchema: JSON schema definition for tool parameter validation
// - Examples: Usage examples for Claude to understand tool capabilities
// - Category: Categorization for tool organization and discovery
// - CLICommand: Underlying CLI command this tool executes
// - CLIArgs: Base arguments for the CLI command
// - HelpText: Additional guidance for tool usage
// - Version: Tool version for compatibility tracking
// - Timeout: Maximum execution time for this tool
//
// Notes:
// - All MCPTool instances should be immutable after creation
// - InputSchema must be valid JSON Schema Draft 7 format
// - Examples should cover common use cases and edge cases
// - Category should align with predefined CategoryType values
type MCPTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
	Examples    []MCPToolExample       `json:"examples,omitempty"`
	Category    CategoryType           `json:"category"`
	CLICommand  string                 `json:"cliCommand"`
	CLIArgs     []string               `json:"cliArgs"`
	HelpText    string                 `json:"helpText,omitempty"`
	Version     string                 `json:"version"`
	Timeout     time.Duration          `json:"timeout"`
}

// MCPToolExample provides usage examples for Claude to understand tool capabilities.
//
// Examples help Claude understand how to use tools effectively and provide context
// for parameter values and expected outcomes.
//
// Fields:
// - Description: Human-readable description of the example scenario
// - Input: Example input parameters that demonstrate tool usage
// - ExpectedOutput: Optional description of expected tool output
// - UseCase: The business scenario this example addresses
//
// Notes:
// - Input should be valid according to the tool's InputSchema
// - Examples should cover both simple and complex use cases
// - UseCase helps Claude understand when to use this tool
type MCPToolExample struct {
	Description    string                 `json:"description"`
	Input          map[string]interface{} `json:"input"`
	ExpectedOutput string                 `json:"expectedOutput,omitempty"`
	UseCase        string                 `json:"useCase,omitempty"`
}

// CategoryType represents tool categories for organization and discovery.
//
// Categories help organize tools into logical groups that align with user workflows
// and make tool discovery more intuitive for Claude.
type CategoryType string

const (
	// CategoryInvoiceManagement groups tools for creating, updating, and managing invoices
	CategoryInvoiceManagement CategoryType = "invoice_management"

	// CategoryDataImport groups tools for importing timesheet and client data
	CategoryDataImport CategoryType = "data_import"

	// CategoryDataExport groups tools for generating and exporting invoice documents
	CategoryDataExport CategoryType = "data_export"

	// CategoryClientManagement groups tools for managing client information
	CategoryClientManagement CategoryType = "client_management"

	// CategoryConfiguration groups tools for system configuration and validation
	CategoryConfiguration CategoryType = "configuration"

	// CategoryReporting groups tools for analytics and reporting
	CategoryReporting CategoryType = "reporting"
)

// ToolRegistry defines the contract for tool discovery and validation (consumer-driven interface).
//
// This interface is defined at the point of use to support consumer-driven design patterns.
// Implementations should provide efficient tool lookup and comprehensive validation.
//
// Methods require context.Context for cancellation and timeout support:
// - GetTool: Retrieve a specific tool by name with context awareness
// - ListTools: Discover tools by category with filtering support
// - ValidateToolInput: Validate input parameters against tool schema
// - RegisterTool: Add new tools to the registry (for dynamic registration)
// - GetCategories: Retrieve available tool categories for discovery
//
// All methods should handle context cancellation gracefully and provide clear error
// messages with actionable guidance for resolution.
type ToolRegistry interface {
	// GetTool retrieves a specific tool by name with context support for cancellation.
	//
	// Parameters:
	// - ctx: Context for cancellation and timeout
	// - name: Unique tool identifier to retrieve
	//
	// Returns:
	// - *MCPTool: The requested tool definition if found
	// - error: Error if tool not found or context canceled
	//
	// Notes:
	// - Returns ErrToolNotFound if tool doesn't exist
	// - Respects context cancellation for responsive operations
	GetTool(ctx context.Context, name string) (*MCPTool, error)

	// ListTools discovers tools by category with optional filtering support.
	//
	// Parameters:
	// - ctx: Context for cancellation and timeout
	// - category: Tool category to filter by (empty string for all tools)
	//
	// Returns:
	// - []*MCPTool: List of tools in the specified category
	// - error: Error if category invalid or context canceled
	//
	// Notes:
	// - Empty category returns all registered tools
	// - Results are sorted by tool name for consistent ordering
	// - Respects context cancellation for large tool sets
	ListTools(ctx context.Context, category CategoryType) ([]*MCPTool, error)

	// ValidateToolInput validates input parameters against tool schema with comprehensive error reporting.
	//
	// Parameters:
	// - ctx: Context for cancellation and timeout
	// - toolName: Name of the tool to validate input for
	// - input: Input parameters to validate against tool schema
	//
	// Returns:
	// - error: Validation error with detailed field-level guidance, or nil if valid
	//
	// Notes:
	// - Provides field-level validation errors for actionable feedback
	// - Checks required fields, data types, and format constraints
	// - Returns structured errors that can be formatted for Claude responses
	ValidateToolInput(ctx context.Context, toolName string, input map[string]interface{}) error

	// RegisterTool adds a new tool to the registry for dynamic tool registration.
	//
	// Parameters:
	// - ctx: Context for cancellation and timeout
	// - tool: Tool definition to register
	//
	// Returns:
	// - error: Registration error or nil if successful
	//
	// Notes:
	// - Validates tool definition before registration
	// - Prevents duplicate tool names
	// - Thread-safe for concurrent registration
	RegisterTool(ctx context.Context, tool *MCPTool) error

	// GetCategories retrieves all available tool categories for discovery.
	//
	// Parameters:
	// - ctx: Context for cancellation and timeout
	//
	// Returns:
	// - []CategoryType: List of all available categories
	// - error: Error if context canceled
	//
	// Notes:
	// - Results include only categories with registered tools
	// - Sorted alphabetically for consistent ordering
	GetCategories(ctx context.Context) ([]CategoryType, error)
}

// InputValidator defines the contract for comprehensive schema validation (consumer-driven interface).
//
// This interface handles JSON schema validation with detailed error reporting to provide
// actionable feedback for tool input validation.
//
// Methods support context cancellation and provide structured error information:
// - ValidateAgainstSchema: Core schema validation with detailed error reporting
// - ValidateRequired: Check required field presence and validity
// - ValidateFormat: Validate field formats (dates, emails, etc.)
// - BuildValidationError: Create structured validation errors
//
// All validation should provide clear, actionable error messages that can be
// presented to Claude for correction guidance.
type InputValidator interface {
	// ValidateAgainstSchema performs comprehensive JSON schema validation with detailed error reporting.
	//
	// This method validates input data against a JSON Schema Draft 7 specification,
	// providing field-level error details for precise error correction.
	//
	// Parameters:
	// - ctx: Context for cancellation and timeout
	// - input: Input data to validate
	// - schema: JSON Schema Draft 7 specification to validate against
	//
	// Returns:
	// - error: Validation error with field-level details, or nil if valid
	//
	// Side Effects:
	// - None - pure validation function
	//
	// Notes:
	// - Supports JSON Schema Draft 7 specification fully
	// - Provides field path information for nested validation errors
	// - Includes suggested corrections for common validation failures
	// - Respects context cancellation for large schema validation
	ValidateAgainstSchema(ctx context.Context, input, schema map[string]interface{}) error

	// ValidateRequired checks presence and validity of required fields with context-aware processing.
	//
	// Parameters:
	// - ctx: Context for cancellation and timeout
	// - input: Input data to check for required fields
	// - requiredFields: List of field names that must be present and non-empty
	//
	// Returns:
	// - error: Error listing missing required fields, or nil if all present
	//
	// Notes:
	// - Checks both presence and non-empty values for required fields
	// - Returns structured error with all missing fields listed
	// - Supports nested field paths using dot notation
	ValidateRequired(ctx context.Context, input map[string]interface{}, requiredFields []string) error

	// ValidateFormat validates field formats with context support for cancellation.
	//
	// Parameters:
	// - ctx: Context for cancellation and timeout
	// - fieldName: Name of the field being validated
	// - value: Field value to validate
	// - format: Expected format (date, email, uuid, etc.)
	//
	// Returns:
	// - error: Format validation error with correction guidance, or nil if valid
	//
	// Notes:
	// - Supports standard JSON Schema format validators
	// - Provides examples of correct format in error messages
	// - Extensible for custom format validation
	ValidateFormat(ctx context.Context, fieldName string, value interface{}, format string) error

	// BuildValidationError creates structured validation errors with actionable guidance.
	//
	// Parameters:
	// - ctx: Context for cancellation and timeout
	// - fieldPath: Path to the field with validation error
	// - message: Error message describing the validation failure
	// - suggestions: Optional suggestions for correcting the error
	//
	// Returns:
	// - error: Structured validation error with context and guidance
	//
	// Notes:
	// - Creates errors that can be formatted for Claude consumption
	// - Includes field path for precise error location
	// - Provides actionable suggestions when possible
	BuildValidationError(ctx context.Context, fieldPath, message string, suggestions []string) error
}

// ToolExecutionRequest represents a request to execute an MCP tool with full context.
//
// This struct encapsulates all information needed to execute a tool through the
// MCP protocol, including tool identification, input parameters, and execution context.
//
// Fields:
// - ToolName: Name of the tool to execute
// - Input: Input parameters for tool execution
// - RequestID: Unique identifier for this execution request
// - ClientInfo: Information about the client making the request
// - ExecutionTimeout: Maximum time allowed for tool execution
//
// Notes:
// - All executions should respect the ExecutionTimeout setting
// - RequestID should be unique for tracking and logging
// - Input should be pre-validated before execution
type ToolExecutionRequest struct {
	ToolName         string                 `json:"toolName"`
	Input            map[string]interface{} `json:"input"`
	RequestID        string                 `json:"requestId"`
	ClientInfo       *ClientInfo            `json:"clientInfo,omitempty"`
	ExecutionTimeout time.Duration          `json:"executionTimeout"`
}

// ToolExecutionResponse represents the result of tool execution with comprehensive details.
//
// This struct contains all information about tool execution results, including
// success/failure status, output data, and execution metadata.
//
// Fields:
// - Success: Whether tool execution completed successfully
// - Output: Tool execution output data
// - ErrorMessage: Human-readable error message if execution failed
// - ExecutionTime: Actual time taken for tool execution
// - ResourcesUsed: Resources used during execution (files, etc.)
// - Metadata: Additional execution metadata
//
// Notes:
// - Success should be false if any errors occurred during execution
// - Output should be structured data suitable for Claude consumption
// - ErrorMessage should provide actionable guidance for error resolution
type ToolExecutionResponse struct {
	Success       bool                   `json:"success"`
	Output        map[string]interface{} `json:"output"`
	ErrorMessage  string                 `json:"errorMessage,omitempty"`
	ExecutionTime time.Duration          `json:"executionTime"`
	ResourcesUsed []string               `json:"resourcesUsed,omitempty"`
	Metadata      map[string]string      `json:"metadata,omitempty"`
}

// ClientInfo represents information about the MCP client making requests.
//
// This struct contains metadata about the client for logging, security, and
// customization purposes.
//
// Fields:
// - Name: Client application name (e.g., "Claude Desktop")
// - Version: Client version information
// - Platform: Client platform information
//
// Notes:
// - Used for logging and audit trails
// - Can be used for client-specific behavior customization
// - Should not contain sensitive information
type ClientInfo struct {
	Name     string `json:"name"`
	Version  string `json:"version"`
	Platform string `json:"platform,omitempty"`
}

// ValidationError represents a structured validation error with actionable guidance.
//
// This struct provides detailed information about validation failures to help
// users correct input errors effectively.
//
// Fields:
// - Field: Path to the field that failed validation
// - Message: Human-readable error message
// - Code: Machine-readable error code for programmatic handling
// - Suggestions: Actionable suggestions for error correction
//
// Notes:
// - Field uses dot notation for nested field paths
// - Suggestions should provide concrete examples when possible
// - Code enables programmatic error handling by clients
type ValidationError struct {
	Field       string   `json:"field"`
	Message     string   `json:"message"`
	Code        string   `json:"code"`
	Suggestions []string `json:"suggestions,omitempty"`
}

// Error implements the error interface for ValidationError.
//
// Returns:
// - string: Formatted error message including field path and suggestions
//
// Notes:
// - Provides human-readable error format suitable for Claude consumption
// - Includes field context and actionable guidance
func (ve *ValidationError) Error() string {
	msg := ve.Message
	if ve.Field != "" {
		msg = ve.Field + ": " + msg
	}
	if len(ve.Suggestions) > 0 {
		msg += " (suggestions: " + ve.Suggestions[0]
		for _, suggestion := range ve.Suggestions[1:] {
			msg += ", " + suggestion
		}
		msg += ")"
	}
	return msg
}

// ToolNotFoundError represents an error when a requested tool cannot be found.
//
// This error type provides specific information about tool discovery failures
// with suggestions for correction.
//
// Fields:
// - ToolName: Name of the tool that was not found
// - AvailableTools: List of available tool names for suggestion
// - Category: Suggested category to search in
//
// Notes:
// - Used for tool registry lookup failures
// - Provides available alternatives for user guidance
// - Should include fuzzy matching suggestions when possible
type ToolNotFoundError struct {
	ToolName       string   `json:"toolName"`
	AvailableTools []string `json:"availableTools,omitempty"`
	Category       string   `json:"category,omitempty"`
}

// Error implements the error interface for ToolNotFoundError.
//
// Returns:
// - string: Formatted error message with tool name and suggestions
//
// Notes:
// - Provides clear error message with actionable guidance
// - Includes available alternatives when possible
func (tnfe *ToolNotFoundError) Error() string {
	msg := "tool not found: " + tnfe.ToolName
	if len(tnfe.AvailableTools) > 0 {
		msg += " (available tools: " + tnfe.AvailableTools[0]
		for _, tool := range tnfe.AvailableTools[1:] {
			msg += ", " + tool
		}
		msg += ")"
	}
	if tnfe.Category != "" {
		msg += " (try searching in category: " + tnfe.Category + ")"
	}
	return msg
}
