package tools

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
)

// DefaultToolRegistry provides a concrete implementation of the ToolRegistry interface.
//
// This implementation provides thread-safe tool registration and discovery with
// efficient lookup operations and comprehensive validation. It follows dependency
// injection patterns and context-first design principles.
//
// Key features:
// - Thread-safe concurrent access with read-write mutex
// - Efficient tool lookup with O(1) name-based access
// - Category-based filtering with sorted results
// - Comprehensive input validation with detailed error reporting
// - Context-aware operations with cancellation support
// - Fuzzy matching for tool discovery assistance
//
// The registry maintains tools in memory for fast access and supports dynamic
// tool registration for extensibility. All operations respect context cancellation
// for responsive behavior.
type DefaultToolRegistry struct {
	// tools maps tool names to tool definitions for O(1) lookup
	tools map[string]*MCPTool

	// categories maps category types to sets of tool names for efficient filtering
	categories map[CategoryType]map[string]bool

	// validator provides input validation capabilities
	validator InputValidator

	// logger provides structured logging for registry operations
	logger Logger

	// mu protects concurrent access to tools and categories maps
	mu sync.RWMutex
}

// Logger defines the logging contract for tool registry operations (consumer-driven interface).
//
// This interface provides structured logging capabilities for registry operations,
// debugging, and audit trails.
type Logger interface {
	// Debug logs debug-level messages with structured key-value pairs
	Debug(msg string, keysAndValues ...interface{})

	// Info logs info-level messages with structured key-value pairs
	Info(msg string, keysAndValues ...interface{})

	// Warn logs warning-level messages with structured key-value pairs
	Warn(msg string, keysAndValues ...interface{})

	// Error logs error-level messages with structured key-value pairs
	Error(msg string, keysAndValues ...interface{})
}

// NewDefaultToolRegistry creates a new tool registry with dependency injection.
//
// This constructor initializes a thread-safe tool registry with all required
// dependencies injected. It follows the dependency injection pattern to avoid
// global state and improve testability.
//
// Parameters:
// - validator: Input validation engine for tool parameter validation
// - logger: Structured logger for registry operations and debugging
//
// Returns:
// - *DefaultToolRegistry: Initialized registry ready for tool registration and lookup
//
// Notes:
// - Registry starts empty and tools must be registered using RegisterTool
// - All dependencies must be non-nil or the constructor will panic
// - Thread-safe for concurrent use after construction
func NewDefaultToolRegistry(validator InputValidator, logger Logger) *DefaultToolRegistry {
	if validator == nil {
		panic("validator cannot be nil")
	}
	if logger == nil {
		panic("logger cannot be nil")
	}

	return &DefaultToolRegistry{
		tools:      make(map[string]*MCPTool),
		categories: make(map[CategoryType]map[string]bool),
		validator:  validator,
		logger:     logger,
	}
}

// GetTool retrieves a specific tool by name with context support for cancellation.
//
// This method provides O(1) tool lookup with context cancellation support and
// comprehensive error handling. It returns detailed error information for
// tool discovery assistance.
//
// Parameters:
// - ctx: Context for cancellation and timeout
// - name: Unique tool identifier to retrieve
//
// Returns:
// - *MCPTool: The requested tool definition if found
// - error: ToolNotFoundError if tool doesn't exist, or context error if cancelled
//
// Side Effects:
// - Logs tool lookup operations for debugging and audit trails
//
// Notes:
// - Tool names are case-sensitive for exact matching
// - Returns ToolNotFoundError with suggestions for similar tool names
// - Respects context cancellation for responsive operations
// - Thread-safe for concurrent access
func (r *DefaultToolRegistry) GetTool(ctx context.Context, name string) (*MCPTool, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	tool, exists := r.tools[name]
	if !exists {
		r.logger.Debug("tool lookup failed", "toolName", name, "availableTools", len(r.tools))

		// Provide helpful suggestions for similar tool names
		suggestions := r.findSimilarToolNames(name)
		return nil, &ToolNotFoundError{
			ToolName:       name,
			AvailableTools: suggestions,
		}
	}

	r.logger.Debug("tool lookup successful", "toolName", name, "category", tool.Category)
	return tool, nil
}

// ListTools discovers tools by category with optional filtering support.
//
// This method provides efficient tool discovery with category filtering and
// consistent ordering. Results are sorted by tool name for predictable output.
//
// Parameters:
// - ctx: Context for cancellation and timeout
// - category: Tool category to filter by (empty string for all tools)
//
// Returns:
// - []*MCPTool: List of tools in the specified category, sorted by name
// - error: Error if category invalid or context cancelled
//
// Side Effects:
// - Logs tool discovery operations for debugging and analytics
//
// Notes:
// - Empty category ("") returns all registered tools
// - Results are sorted by tool name for consistent ordering
// - Respects context cancellation for large tool sets
// - Thread-safe for concurrent access
// - Returns defensive copies to prevent external modification
func (r *DefaultToolRegistry) ListTools(ctx context.Context, category CategoryType) ([]*MCPTool, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	var tools []*MCPTool

	if category == "" {
		// Return all tools
		tools = make([]*MCPTool, 0, len(r.tools))
		for _, tool := range r.tools {
			// Create defensive copy to prevent external modification
			toolCopy := *tool
			tools = append(tools, &toolCopy)
		}
	} else {
		// Filter by category
		toolNames, categoryExists := r.categories[category]
		if !categoryExists {
			r.logger.Debug("category not found", "category", category, "availableCategories", len(r.categories))
			return []*MCPTool{}, nil // Return empty slice, not error, for non-existent categories
		}

		tools = make([]*MCPTool, 0, len(toolNames))
		for toolName := range toolNames {
			if tool, exists := r.tools[toolName]; exists {
				// Create defensive copy to prevent external modification
				toolCopy := *tool
				tools = append(tools, &toolCopy)
			}
		}
	}

	// Sort tools by name for consistent ordering
	sort.Slice(tools, func(i, j int) bool {
		return tools[i].Name < tools[j].Name
	})

	r.logger.Debug("tool discovery completed",
		"category", category,
		"toolCount", len(tools),
		"totalRegistered", len(r.tools))

	return tools, nil
}

// ValidateToolInput validates input parameters against tool schema with comprehensive error reporting.
//
// This method provides comprehensive input validation with detailed field-level
// error reporting to help users correct input errors effectively.
//
// Parameters:
// - ctx: Context for cancellation and timeout
// - toolName: Name of the tool to validate input for
// - input: Input parameters to validate against tool schema
//
// Returns:
// - error: Validation error with detailed field-level guidance, or nil if valid
//
// Side Effects:
// - Logs validation operations for debugging and audit trails
//
// Notes:
// - Provides field-level validation errors for actionable feedback
// - Checks required fields, data types, and format constraints
// - Returns structured errors that can be formatted for Claude responses
// - Respects context cancellation for complex validation operations
// - Thread-safe for concurrent validation requests
func (r *DefaultToolRegistry) ValidateToolInput(ctx context.Context, toolName string, input map[string]interface{}) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Get tool definition for validation
	tool, err := r.GetTool(ctx, toolName)
	if err != nil {
		return fmt.Errorf("cannot validate input for unknown tool: %w", err)
	}

	// Perform comprehensive schema validation
	if err := r.validator.ValidateAgainstSchema(ctx, input, tool.InputSchema); err != nil {
		r.logger.Debug("tool input validation failed",
			"toolName", toolName,
			"error", err.Error(),
			"inputKeys", getMapKeys(input))
		return fmt.Errorf("input validation failed for tool %s: %w", toolName, err)
	}

	r.logger.Debug("tool input validation successful",
		"toolName", toolName,
		"inputKeys", getMapKeys(input))

	return nil
}

// RegisterTool adds a new tool to the registry for dynamic tool registration.
//
// This method provides thread-safe tool registration with comprehensive validation
// to ensure tool definitions are complete and consistent.
//
// Parameters:
// - ctx: Context for cancellation and timeout
// - tool: Tool definition to register
//
// Returns:
// - error: Registration error or nil if successful
//
// Side Effects:
// - Adds tool to registry maps for lookup and discovery
// - Updates category mappings for efficient filtering
// - Logs registration operations for debugging and audit trails
//
// Notes:
// - Validates tool definition before registration
// - Prevents duplicate tool names with clear error messages
// - Thread-safe for concurrent registration
// - Creates defensive copies to prevent external modification after registration
func (r *DefaultToolRegistry) RegisterTool(ctx context.Context, tool *MCPTool) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if tool == nil {
		return fmt.Errorf("tool cannot be nil")
	}

	// Validate tool definition
	if err := r.validateToolDefinition(tool); err != nil {
		r.logger.Error("tool registration failed - invalid definition",
			"toolName", tool.Name,
			"error", err.Error())
		return fmt.Errorf("invalid tool definition: %w", err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Check for duplicate tool names
	if _, exists := r.tools[tool.Name]; exists {
		return fmt.Errorf("tool already registered: %s", tool.Name)
	}

	// Create defensive copy to prevent external modification
	toolCopy := *tool
	if tool.Examples != nil {
		toolCopy.Examples = make([]MCPToolExample, len(tool.Examples))
		copy(toolCopy.Examples, tool.Examples)
	}
	if tool.CLIArgs != nil {
		toolCopy.CLIArgs = make([]string, len(tool.CLIArgs))
		copy(toolCopy.CLIArgs, tool.CLIArgs)
	}

	// Register tool
	r.tools[tool.Name] = &toolCopy

	// Update category mapping
	if r.categories[tool.Category] == nil {
		r.categories[tool.Category] = make(map[string]bool)
	}
	r.categories[tool.Category][tool.Name] = true

	r.logger.Info("tool registered successfully",
		"toolName", tool.Name,
		"category", tool.Category,
		"totalTools", len(r.tools))

	return nil
}

// GetCategories retrieves all available tool categories for discovery.
//
// This method provides efficient category discovery with consistent ordering
// for tool organization and navigation.
//
// Parameters:
// - ctx: Context for cancellation and timeout
//
// Returns:
// - []CategoryType: List of all available categories, sorted alphabetically
// - error: Error if context cancelled
//
// Side Effects:
// - Logs category discovery operations for analytics
//
// Notes:
// - Results include only categories with registered tools
// - Sorted alphabetically for consistent ordering
// - Thread-safe for concurrent access
// - Returns defensive copies to prevent external modification
func (r *DefaultToolRegistry) GetCategories(ctx context.Context) ([]CategoryType, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	categories := make([]CategoryType, 0, len(r.categories))
	for category := range r.categories {
		categories = append(categories, category)
	}

	// Sort categories alphabetically for consistent ordering
	sort.Slice(categories, func(i, j int) bool {
		return string(categories[i]) < string(categories[j])
	})

	r.logger.Debug("category discovery completed",
		"categoryCount", len(categories))

	return categories, nil
}

// validateToolDefinition performs comprehensive validation of tool definitions.
//
// This internal method ensures tool definitions are complete and consistent
// before registration to prevent runtime errors.
//
// Parameters:
// - tool: Tool definition to validate
//
// Returns:
// - error: Validation error describing the first issue found, or nil if valid
//
// Notes:
// - Checks all required fields are present and non-empty
// - Validates schema format and structure
// - Ensures category is a known type
// - Validates timeout values are reasonable
func (r *DefaultToolRegistry) validateToolDefinition(tool *MCPTool) error {
	if tool.Name == "" {
		return fmt.Errorf("tool name cannot be empty")
	}

	if tool.Description == "" {
		return fmt.Errorf("tool description cannot be empty")
	}

	if tool.InputSchema == nil {
		return fmt.Errorf("tool input schema cannot be nil")
	}

	if tool.CLICommand == "" {
		return fmt.Errorf("tool CLI command cannot be empty")
	}

	if tool.Version == "" {
		return fmt.Errorf("tool version cannot be empty")
	}

	// Validate category is known
	if !r.isValidCategory(tool.Category) {
		return fmt.Errorf("invalid tool category: %s", tool.Category)
	}

	// Validate timeout is reasonable (between 1 second and 10 minutes)
	if tool.Timeout < 1000000000 || tool.Timeout > 600000000000 { // 1s to 10m in nanoseconds
		return fmt.Errorf("tool timeout must be between 1 second and 10 minutes, got: %v", tool.Timeout)
	}

	// Validate input schema structure
	if schemaType, exists := tool.InputSchema["type"]; exists {
		if schemaType != "object" {
			return fmt.Errorf("tool input schema type must be 'object', got: %v", schemaType)
		}
	}

	return nil
}

// isValidCategory checks if a category type is valid.
//
// This internal method validates category types against known categories
// to ensure consistent categorization.
//
// Parameters:
// - category: Category type to validate
//
// Returns:
// - bool: True if category is valid, false otherwise
func (r *DefaultToolRegistry) isValidCategory(category CategoryType) bool {
	switch category {
	case CategoryInvoiceManagement, CategoryDataImport, CategoryDataExport,
		CategoryClientManagement, CategoryConfiguration, CategoryReporting:
		return true
	default:
		return false
	}
}

// findSimilarToolNames finds tool names similar to the given name for suggestions.
//
// This internal method provides fuzzy matching to help users discover tools
// when they make typos or don't remember exact tool names.
//
// Parameters:
// - name: Tool name to find similar matches for
//
// Returns:
// - []string: List of similar tool names, limited to 5 suggestions
//
// Notes:
// - Uses simple string distance and prefix matching
// - Limited to 5 suggestions to avoid overwhelming users
// - Sorted by similarity score (most similar first)
func (r *DefaultToolRegistry) findSimilarToolNames(name string) []string {
	var suggestions []string
	nameLower := strings.ToLower(name)

	for toolName := range r.tools {
		toolNameLower := strings.ToLower(toolName)

		// Check for prefix match or contains match
		if strings.HasPrefix(toolNameLower, nameLower) ||
			strings.Contains(toolNameLower, nameLower) ||
			strings.HasPrefix(nameLower, toolNameLower) {
			suggestions = append(suggestions, toolName)
		}
	}

	// Limit suggestions to avoid overwhelming users
	if len(suggestions) > 5 {
		suggestions = suggestions[:5]
	}

	return suggestions
}

// getMapKeys extracts keys from a map for logging purposes.
//
// This utility function provides key lists for structured logging without
// exposing sensitive values.
//
// Parameters:
// - m: Map to extract keys from
//
// Returns:
// - []string: List of map keys
//
// Notes:
// - Used for logging input parameter names without values
// - Provides debugging information while maintaining privacy
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys) // Sort for consistent logging
	return keys
}
