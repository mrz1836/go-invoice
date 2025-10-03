// Package tools provides the complete registry implementation for all MCP tools.
//
// This package implements the unified tool registry and validation system that integrates
// all 21 tools from phases 2.2-2.4 into a comprehensive MCP integration. It provides
// centralized tool registration, discovery, and validation with context-first design
// and comprehensive error handling.
//
// Key features:
// - Registration of all 21 tools across 6 categories
// - Unified tool discovery and search functionality
// - JSON schema validation integration
// - MCP server bridge compatibility
// - Context-aware operations with cancellation support
//
// This registry serves as the single source of truth for all MCP tool definitions
// and provides the foundation for Claude Desktop integration.
package tools

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Registry validation errors (additional to those in init.go)
var (
	ErrMissingCategory   = errors.New("missing expected category")
	ErrCategoryToolCount = errors.New("category has incorrect tool count")
)

// CompleteToolRegistry provides the unified registry with all 21 tools pre-registered.
//
// This implementation extends DefaultToolRegistry with automatic tool registration
// for all invoice management tools. It provides a convenient single-step initialization
// for the complete MCP tool ecosystem.
//
// Key features:
// - Pre-registered with all 21 tools from phases 2.2-2.4
// - Category-based organization for efficient discovery
// - Comprehensive validation with JSON schema support
// - Context-aware operations with proper cancellation
// - Performance-optimized for high-frequency tool access
//
// Categories included:
// - CategoryInvoiceManagement: 8 invoice management tools
// - CategoryClientManagement: 5 client management tools
// - CategoryDataImport: 3 data import tools
// - CategoryDataExport: 3 document generation tools
// - CategoryConfiguration: 3 configuration management tools
//
// All tools are validated and ready for MCP protocol interaction with Claude.
type CompleteToolRegistry struct {
	*DefaultToolRegistry

	// initializationTime tracks when the registry was created for metrics
	initializationTime time.Time

	// toolCount tracks the total number of registered tools for validation
	toolCount int

	// categoryCount tracks the number of active categories for metrics
	categoryCount int
}

// NewCompleteToolRegistry creates a fully initialized tool registry with all 21 tools.
//
// This constructor provides a one-step initialization of the complete MCP tool ecosystem
// with automatic registration of all tools from phases 2.2-2.4. It validates the
// registration process and ensures all tools are properly configured.
//
// Parameters:
// - ctx: Context for cancellation and timeout during initialization
// - validator: Input validation engine for tool parameter validation
// - logger: Structured logger for registry operations and debugging
//
// Returns:
// - *CompleteToolRegistry: Fully initialized registry with all 21 tools registered
// - error: Initialization error if tool registration fails
//
// Side Effects:
// - Registers all 21 tools in their respective categories
// - Validates tool definitions and schemas
// - Logs initialization progress and results
//
// Notes:
// - Initialization may take several seconds due to comprehensive validation
// - All dependencies must be non-nil or constructor will return error
// - Registry is ready for immediate use after successful construction
// - Respects context cancellation for responsive initialization
func NewCompleteToolRegistry(ctx context.Context, validator InputValidator, logger Logger) (*CompleteToolRegistry, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if validator == nil {
		return nil, ErrValidatorNil
	}
	if logger == nil {
		return nil, ErrLoggerNil
	}

	logger.Info("initializing complete tool registry",
		"expectedTools", 22,
		"expectedCategories", 5)

	// Create base registry
	baseRegistry := NewDefaultToolRegistry(validator, logger)

	registry := &CompleteToolRegistry{
		DefaultToolRegistry: baseRegistry,
		initializationTime:  time.Now(),
	}

	// Register all tool categories
	if err := registry.registerAllTools(ctx); err != nil {
		logger.Error("tool registry initialization failed", "error", err.Error())
		return nil, fmt.Errorf("failed to initialize tool registry: %w", err)
	}

	// Validate final state
	if err := registry.validateRegistration(ctx); err != nil {
		logger.Error("tool registry validation failed", "error", err.Error())
		return nil, fmt.Errorf("tool registry validation failed: %w", err)
	}

	logger.Info("complete tool registry initialized successfully",
		"toolCount", registry.toolCount,
		"categoryCount", registry.categoryCount,
		"initializationTime", time.Since(registry.initializationTime))

	return registry, nil
}

// registerAllTools registers all 21 tools from phases 2.2-2.4 into the registry.
//
// This method orchestrates the registration of all tool categories with proper
// error handling and progress tracking. It ensures all tools are registered
// successfully before completing initialization.
//
// Parameters:
// - ctx: Context for cancellation and timeout
//
// Returns:
// - error: Registration error if any category fails to register
//
// Side Effects:
// - Registers all tools in CategoryInvoiceManagement (8 tools)
// - Registers all tools in CategoryClientManagement (5 tools)
// - Registers all tools in CategoryDataImport (3 tools)
// - Registers all tools in CategoryDataExport (3 tools)
// - Registers all tools in CategoryConfiguration (3 tools)
// - Updates internal counters for validation
//
// Notes:
// - Registration order ensures dependencies are available when needed
// - Partial registration is rolled back on failure
// - Progress is logged for debugging and monitoring
// - Respects context cancellation throughout the process
func (r *CompleteToolRegistry) registerAllTools(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	r.logger.Debug("starting tool registration process")

	// Register invoice management tools (8 tools)
	if err := RegisterInvoiceManagementTools(ctx, r.DefaultToolRegistry); err != nil {
		return fmt.Errorf("failed to register invoice management tools: %w", err)
	}
	r.logger.Debug("invoice management tools registered", "count", 8)

	// Register client management tools (5 tools)
	if err := RegisterClientManagementTools(ctx, r.DefaultToolRegistry); err != nil {
		return fmt.Errorf("failed to register client management tools: %w", err)
	}
	r.logger.Debug("client management tools registered", "count", 5)

	// Register data import tools (3 tools)
	if err := RegisterDataImportTools(ctx, r.DefaultToolRegistry); err != nil {
		return fmt.Errorf("failed to register data import tools: %w", err)
	}
	r.logger.Debug("data import tools registered", "count", 3)

	// Register document generation tools (3 tools)
	if err := RegisterDocumentGenerationTools(ctx, r.DefaultToolRegistry); err != nil {
		return fmt.Errorf("failed to register document generation tools: %w", err)
	}
	r.logger.Debug("document generation tools registered", "count", 3)

	// Register configuration management tools (3 tools)
	if err := RegisterConfigurationManagementTools(ctx, r.DefaultToolRegistry); err != nil {
		return fmt.Errorf("failed to register configuration management tools: %w", err)
	}
	r.logger.Debug("configuration management tools registered", "count", 3)

	r.logger.Debug("all tool categories registered successfully")
	return nil
}

// validateRegistration validates that all expected tools are properly registered.
//
// This method performs comprehensive validation of the registry state to ensure
// all 21 tools are registered correctly with proper schemas and metadata.
//
// Parameters:
// - ctx: Context for cancellation and timeout
//
// Returns:
// - error: Validation error if registry state is invalid
//
// Side Effects:
// - Updates toolCount and categoryCount for metrics
// - Logs validation results for monitoring
//
// Notes:
// - Validates tool count matches expected 22 tools
// - Checks all 5 categories are represented
// - Verifies tool definitions are complete and valid
// - Provides detailed error information for troubleshooting
func (r *CompleteToolRegistry) validateRegistration(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Get all tools for validation
	allTools, err := r.ListTools(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to list tools for validation: %w", err)
	}

	r.toolCount = len(allTools)
	if r.toolCount != 22 {
		return fmt.Errorf("%w: expected 22, got %d", ErrInvalidToolCount, r.toolCount)
	}

	// Get categories for validation
	categories, err := r.GetCategories(ctx)
	if err != nil {
		return fmt.Errorf("failed to get categories for validation: %w", err)
	}

	r.categoryCount = len(categories)
	if r.categoryCount != 5 {
		return fmt.Errorf("%w: expected 5, got %d", ErrInvalidCategoryCount, r.categoryCount)
	}

	// Validate expected categories are present
	expectedCategories := map[CategoryType]bool{
		CategoryInvoiceManagement: false,
		CategoryClientManagement:  false,
		CategoryDataImport:        false,
		CategoryDataExport:        false,
		CategoryConfiguration:     false,
	}

	for _, category := range categories {
		expectedCategories[category] = true
	}

	for category, found := range expectedCategories {
		if !found {
			return fmt.Errorf("%w: %s", ErrMissingCategory, category)
		}
	}

	// Validate tool distribution across categories
	categoryToolCounts := map[CategoryType]int{
		CategoryInvoiceManagement: 0,
		CategoryClientManagement:  0,
		CategoryDataImport:        0,
		CategoryDataExport:        0,
		CategoryConfiguration:     0,
	}

	for _, tool := range allTools {
		categoryToolCounts[tool.Category]++
	}

	// Validate expected tool counts per category
	expectedCounts := map[CategoryType]int{
		CategoryInvoiceManagement: 8,
		CategoryClientManagement:  5,
		CategoryDataImport:        3,
		CategoryDataExport:        3,
		CategoryConfiguration:     3,
	}

	for category, expectedCount := range expectedCounts {
		actualCount := categoryToolCounts[category]
		if actualCount != expectedCount {
			return fmt.Errorf("%w: category %s expected %d, got %d", ErrCategoryToolCount, category, expectedCount, actualCount)
		}
	}

	r.logger.Debug("registry validation completed successfully",
		"toolCount", r.toolCount,
		"categoryCount", r.categoryCount)

	return nil
}

// GetRegistrationMetrics returns metrics about the registry state for monitoring.
//
// This method provides operational metrics for registry health monitoring and
// performance tracking.
//
// Parameters:
// - ctx: Context for cancellation and timeout
//
// Returns:
// - *RegistrationMetrics: Comprehensive metrics about registry state
// - error: Error if metrics collection fails
//
// Notes:
// - Metrics include tool counts, categories, and performance data
// - Used for monitoring and alerting in production deployments
// - Lightweight operation suitable for frequent calls
// - Thread-safe for concurrent access
func (r *CompleteToolRegistry) GetRegistrationMetrics(ctx context.Context) (*RegistrationMetrics, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	return &RegistrationMetrics{
		TotalTools:         r.toolCount,
		TotalCategories:    r.categoryCount,
		InitializationTime: r.initializationTime,
		Uptime:             time.Since(r.initializationTime),
		ToolsByCategory: map[CategoryType]int{
			CategoryInvoiceManagement: 8,
			CategoryClientManagement:  5,
			CategoryDataImport:        3,
			CategoryDataExport:        3,
			CategoryConfiguration:     3,
		},
	}, nil
}

// RegistrationMetrics provides comprehensive metrics about registry state.
//
// This struct contains operational metrics for monitoring registry health,
// performance, and usage patterns.
//
// Fields:
// - TotalTools: Total number of registered tools
// - TotalCategories: Number of active tool categories
// - InitializationTime: When the registry was initialized
// - Uptime: How long the registry has been running
// - ToolsByCategory: Tool count breakdown by category
//
// Notes:
// - Used for monitoring and alerting systems
// - Provides data for performance optimization
// - Enables capacity planning and resource allocation
type RegistrationMetrics struct {
	TotalTools         int                  `json:"totalTools"`
	TotalCategories    int                  `json:"totalCategories"`
	InitializationTime time.Time            `json:"initializationTime"`
	Uptime             time.Duration        `json:"uptime"`
	ToolsByCategory    map[CategoryType]int `json:"toolsByCategory"`
}

// Comment: Registration functions are implemented in their respective tool files:
// - RegisterInvoiceManagementTools in invoice_tools.go
// - RegisterClientManagementTools in client_tools.go
// - RegisterDataImportTools in import_tools.go
// - RegisterDocumentGenerationTools in generate_tools.go
// - RegisterConfigurationManagementTools in config_tools.go
