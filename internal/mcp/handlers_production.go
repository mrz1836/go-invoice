package mcp

import (
	"context"
	"fmt"

	"github.com/mrz/go-invoice/internal/mcp/executor"
	"github.com/mrz/go-invoice/internal/mcp/tools"
	"github.com/mrz/go-invoice/internal/mcp/types"
)

// ProductionMCPHandler implements MCPHandler with Phase 3 executor integration.
type ProductionMCPHandler struct {
	logger          Logger
	toolRegistry    *tools.DefaultToolRegistry
	toolCallHandler *executor.ToolCallHandler
	config          *Config
}

// NewProductionMCPHandler creates a new MCP handler with full Phase 3 integration.
func NewProductionMCPHandler(
	logger Logger,
	toolRegistry *tools.DefaultToolRegistry,
	toolCallHandler *executor.ToolCallHandler,
	config *Config,
) MCPHandler {
	if logger == nil {
		panic("logger is required")
	}
	if toolRegistry == nil {
		panic("toolRegistry is required")
	}
	if toolCallHandler == nil {
		panic("toolCallHandler is required")
	}
	if config == nil {
		panic("config is required")
	}

	return &ProductionMCPHandler{
		logger:          logger,
		toolRegistry:    toolRegistry,
		toolCallHandler: toolCallHandler,
		config:          config,
	}
}

// HandleInitialize handles the MCP initialize request.
func (h *ProductionMCPHandler) HandleInitialize(ctx context.Context, req *types.MCPRequest) (*types.MCPResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	h.logger.Info("handling initialize request",
		"version", "2024-11-05",
	)

	result := types.InitializeResult{
		ProtocolVersion: "2024-11-05",
		Capabilities: types.Capabilities{
			Tools: &types.ToolsCapability{
				ListChanged: false,
			},
		},
		ServerInfo: types.ServerInfo{
			Name:    "go-invoice-mcp",
			Version: "2.0.0", // Updated for Phase 3
		},
	}

	var reqID interface{}
	if req != nil {
		reqID = req.ID
	}

	return &types.MCPResponse{
		JSONRPC: "2.0",
		ID:      reqID,
		Result:  result,
	}, nil
}

// HandlePing handles the MCP ping request.
func (h *ProductionMCPHandler) HandlePing(ctx context.Context, req *types.MCPRequest) (*types.MCPResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	h.logger.Debug("handling ping request")

	return &types.MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  map[string]string{"status": "ok"},
	}, nil
}

// HandleToolsList handles the tools/list request using the tool registry.
func (h *ProductionMCPHandler) HandleToolsList(ctx context.Context, req *types.MCPRequest) (*types.MCPResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	h.logger.Debug("handling tools/list request")

	// Get all tools from registry
	allTools, err := h.toolRegistry.ListTools(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to list tools: %w", err)
	}

	// Convert to MCP tool format
	tools := make([]Tool, 0, len(allTools))
	for _, toolDef := range allTools {
		tool := Tool{
			Name:        toolDef.Name,
			Description: toolDef.Description,
			InputSchema: toolDef.InputSchema,
		}
		tools = append(tools, tool)
	}

	result := ToolListResult{
		Tools: tools,
	}

	h.logger.Info("tools list returned",
		"count", len(tools),
	)

	return &types.MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}, nil
}

// HandleToolCall handles the tools/call request using the executor.
func (h *ProductionMCPHandler) HandleToolCall(ctx context.Context, req *types.MCPRequest) (*types.MCPResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Delegate to the tool call handler
	return h.toolCallHandler.HandleToolCall(ctx, req)
}

// CreateProductionHandler creates a production-ready MCP handler with all integrations.
func CreateProductionHandler(config *Config) (MCPHandler, error) {
	ctx := context.Background()

	// Create logger
	logger := NewLogger(config.LogLevel)

	// Create security configuration
	securityConfig := executor.DefaultSecurityConfig()
	// Apply security settings from config
	securityConfig.Sandbox.AllowedCommands = config.Security.AllowedCommands
	securityConfig.Sandbox.AllowedPaths = []string{config.Security.WorkingDir}
	securityConfig.StrictMode = config.Security.SandboxEnabled

	// Create audit logger
	var auditLogger executor.AuditLogger
	if securityConfig.AuditEnabled {
		var err error
		auditLogger, err = executor.NewFileAuditLogger(logger, securityConfig.AuditLogPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create audit logger: %w", err)
		}
	}

	// Create validator
	validator := executor.NewDefaultCommandValidator(logger, securityConfig.Sandbox)

	// Create file handler
	fileHandler := executor.NewDefaultFileHandler(logger, validator, securityConfig.Sandbox)

	// Create executor
	secureExecutor := executor.NewSecureExecutor(logger, validator, securityConfig.Sandbox, fileHandler)

	// Create parser
	parser := executor.NewDefaultOutputParser(logger)

	// Create progress tracker
	tracker := executor.NewDefaultProgressTracker(logger)

	// Create tool registry
	inputValidator := tools.NewDefaultInputValidator(logger)
	toolRegistry := tools.NewDefaultToolRegistry(inputValidator, logger)

	// Register all tool categories
	if err := tools.RegisterInvoiceManagementTools(ctx, toolRegistry); err != nil {
		return nil, fmt.Errorf("failed to register invoice tools: %w", err)
	}
	if err := tools.RegisterClientManagementTools(ctx, toolRegistry); err != nil {
		return nil, fmt.Errorf("failed to register client tools: %w", err)
	}
	if err := tools.RegisterDataImportTools(ctx, toolRegistry); err != nil {
		return nil, fmt.Errorf("failed to register import tools: %w", err)
	}
	if err := tools.RegisterDocumentGenerationTools(ctx, toolRegistry); err != nil {
		return nil, fmt.Errorf("failed to register generation tools: %w", err)
	}
	if err := tools.RegisterConfigTools(ctx, toolRegistry); err != nil {
		return nil, fmt.Errorf("failed to register configuration tools: %w", err)
	}

	// Create MCP executor bridge
	bridge := executor.NewMCPExecutorBridge(
		logger,
		secureExecutor,
		parser,
		tracker,
		fileHandler,
		toolRegistry,
		auditLogger,
		securityConfig,
		config.CLI.Path,
	)

	// Create tool call handler
	toolCallHandler := executor.NewToolCallHandler(
		logger,
		bridge,
		toolRegistry,
		parser,
		tracker,
	)

	// Create handler
	handler := NewProductionMCPHandler(
		logger,
		toolRegistry,
		toolCallHandler,
		config,
	)

	// Get tool count safely
	toolList, err := toolRegistry.ListTools(ctx, "")
	toolCount := 0
	if err == nil {
		toolCount = len(toolList)
	}

	logger.Info("production MCP handler created",
		"auditEnabled", securityConfig.AuditEnabled,
		"strictMode", securityConfig.StrictMode,
		"toolCount", toolCount,
	)

	return handler, nil
}
