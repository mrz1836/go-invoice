package executor

import (
	"context"
	"fmt"
	"time"

	"github.com/mrz1836/go-invoice/internal/mcp"
	"github.com/mrz1836/go-invoice/internal/mcp/tools"
)

// MCPExecutorBridge integrates the Phase 3 executor with the MCP server.
// It adapts the MCP CLIBridge interface to use the new executor system.
type MCPExecutorBridge struct {
	logger         Logger
	bridge         *CLIBridge
	executor       CommandExecutor
	parser         OutputParser
	tracker        ProgressTracker
	toolRegistry   tools.Registry
	auditLogger    AuditLogger
	securityConfig *SecurityConfig
}

// NewMCPExecutorBridge creates a new bridge between MCP and the executor.
func NewMCPExecutorBridge(
	logger Logger,
	executor CommandExecutor,
	parser OutputParser,
	tracker ProgressTracker,
	fileHandler FileHandler,
	toolRegistry tools.Registry,
	auditLogger AuditLogger,
	config *SecurityConfig,
	cliPath string,
) *MCPExecutorBridge {
	if logger == nil {
		panic("logger is required")
	}
	if executor == nil {
		panic("executor is required")
	}
	if parser == nil {
		panic("parser is required")
	}
	if tracker == nil {
		panic("tracker is required")
	}
	if fileHandler == nil {
		panic("fileHandler is required")
	}
	if toolRegistry == nil {
		panic("toolRegistry is required")
	}
	if config == nil {
		config = DefaultSecurityConfig()
	}

	// Create the CLI bridge
	bridge := NewCLIBridge(logger, executor, fileHandler, cliPath)

	return &MCPExecutorBridge{
		logger:         logger,
		bridge:         bridge,
		executor:       executor,
		parser:         parser,
		tracker:        tracker,
		toolRegistry:   toolRegistry,
		auditLogger:    auditLogger,
		securityConfig: config,
	}
}

// ExecuteCommand implements the mcp.CLIBridge interface.
func (m *MCPExecutorBridge) ExecuteCommand(ctx context.Context, req *mcp.CommandRequest) (*mcp.CommandResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Convert MCP request to executor request
	execReq := &ExecutionRequest{
		Command:     req.Command,
		Args:        req.Args,
		WorkingDir:  req.WorkingDir,
		Environment: req.Env,
		Timeout:     req.Timeout,
		ExpectJSON:  req.ExpectJSON,
	}

	// Add input files if any
	for _, file := range req.InputFiles {
		execReq.InputFiles = append(execReq.InputFiles, FileReference{
			Path:        file,
			ContentType: "application/octet-stream",
		})
	}

	// Log command execution attempt
	if m.auditLogger != nil && m.securityConfig.AuditEnabled {
		auditEvent := &CommandAuditEvent{
			Timestamp:   time.Now(),
			UserID:      ctx.Value("userID").(string),
			SessionID:   ctx.Value("sessionID").(string),
			Command:     req.Command,
			Args:        req.Args,
			WorkingDir:  req.WorkingDir,
			Environment: req.Env,
		}
		_ = m.auditLogger.LogCommandExecution(ctx, auditEvent)
	}

	// Execute the command
	execResp, err := m.executor.Execute(ctx, execReq)
	if err != nil {
		return nil, fmt.Errorf("command execution failed: %w", err)
	}

	// Convert executor response to MCP response
	resp := &mcp.CommandResponse{
		ExitCode: execResp.ExitCode,
		Stdout:   execResp.Stdout,
		Stderr:   execResp.Stderr,
		Duration: execResp.Duration,
		Files:    []string{},
	}

	// Add output files
	for _, file := range execResp.OutputFiles {
		resp.Files = append(resp.Files, file.Path)
	}

	// Update audit log with result
	if m.auditLogger != nil && m.securityConfig.AuditEnabled {
		auditEvent := &CommandAuditEvent{
			Timestamp:  time.Now(),
			UserID:     ctx.Value("userID").(string),
			SessionID:  ctx.Value("sessionID").(string),
			Command:    req.Command,
			Args:       req.Args,
			WorkingDir: req.WorkingDir,
			ExitCode:   resp.ExitCode,
			Duration:   resp.Duration,
		}
		if execResp.Error != "" {
			auditEvent.Error = execResp.Error
		}
		_ = m.auditLogger.LogCommandExecution(ctx, auditEvent)
	}

	return resp, nil
}

// ValidateFile implements the mcp.FileHandler interface.
func (m *MCPExecutorBridge) ValidateFile(_ context.Context, _ string) error {
	// This is handled by the executor's file handler
	return nil
}

// PrepareWorkspace implements the mcp.FileHandler interface.
func (m *MCPExecutorBridge) PrepareWorkspace(_ context.Context, _ string) (string, func(), error) {
	// This is handled by the executor's file handler
	return "", func() {}, nil
}

// ToolCallHandler handles MCP tool calls using the executor.
type ToolCallHandler struct {
	logger       Logger
	bridge       *MCPExecutorBridge
	toolRegistry tools.Registry
	parser       OutputParser
	tracker      ProgressTracker
}

// NewToolCallHandler creates a new tool call handler.
func NewToolCallHandler(
	logger Logger,
	bridge *MCPExecutorBridge,
	toolRegistry tools.Registry,
	parser OutputParser,
	tracker ProgressTracker,
) *ToolCallHandler {
	if logger == nil {
		panic("logger is required")
	}
	if bridge == nil {
		panic("bridge is required")
	}
	if toolRegistry == nil {
		panic("toolRegistry is required")
	}
	if parser == nil {
		panic("parser is required")
	}
	if tracker == nil {
		panic("tracker is required")
	}

	return &ToolCallHandler{
		logger:       logger,
		bridge:       bridge,
		toolRegistry: toolRegistry,
		parser:       parser,
		tracker:      tracker,
	}
}

// HandleToolCall processes an MCP tool call request.
func (h *ToolCallHandler) HandleToolCall(ctx context.Context, req *mcp.MCPRequest) (*mcp.MCPResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Parse tool call parameters
	var params mcp.ToolCallParams
	if err := convertParams(req.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to parse tool call params: %w", err)
	}

	h.logger.Debug("handling tool call",
		"tool", params.Name,
		"hasArguments", params.Arguments != nil,
	)

	// Get tool definition from registry
	toolDef, err := h.toolRegistry.GetTool(ctx, params.Name)
	if err != nil {
		return &mcp.MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &mcp.MCPError{
				Code:    -32602,
				Message: "Invalid params",
				Data:    fmt.Sprintf("Unknown tool: %s", params.Name),
			},
		}, nil
	}

	// Validate tool arguments
	if err := h.toolRegistry.ValidateTool(ctx, params.Name, params.Arguments); err != nil {
		return &mcp.MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &mcp.MCPError{
				Code:    -32602,
				Message: "Invalid params",
				Data:    fmt.Sprintf("Invalid arguments for tool %s: %v", params.Name, err),
			},
		}, nil
	}

	// Create operation for progress tracking
	operationID := fmt.Sprintf("tool_%s_%d", params.Name, time.Now().UnixNano())
	operation, err := h.tracker.StartOperation(ctx, operationID, fmt.Sprintf("Executing %s", toolDef.Name), 0)
	if err != nil {
		h.logger.Warn("failed to start operation tracking",
			"error", err,
			"tool", params.Name,
		)
	}

	// Execute tool via bridge
	resp, err := h.bridge.bridge.ExecuteToolCommand(ctx, params.Name, params.Arguments)
	if err != nil {
		if operation != nil {
			operation.Complete(err)
		}
		return &mcp.MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &mcp.MCPError{
				Code:    -32603,
				Message: "Internal error",
				Data:    fmt.Sprintf("Tool execution failed: %v", err),
			},
		}, nil
	}

	// Complete operation
	if operation != nil {
		operation.Complete(nil)
	}

	// Parse output based on tool expectations
	content, err := h.parseToolOutput(ctx, toolDef, resp)
	if err != nil {
		h.logger.Warn("failed to parse tool output",
			"tool", params.Name,
			"error", err,
		)
		// Continue with raw output
		content = []mcp.Content{
			{
				Type: "text",
				Text: resp.Stdout,
			},
		}
	}

	result := mcp.ToolCallResult{
		Content: content,
		IsError: resp.ExitCode != 0,
	}

	return &mcp.MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}, nil
}

// parseToolOutput parses tool output based on the tool definition.
func (h *ToolCallHandler) parseToolOutput(ctx context.Context, tool *tools.Tool, resp *ExecutionResponse) ([]mcp.Content, error) {
	// Check if tool expects JSON output
	if tool.OutputSchema != nil {
		// Try to parse as JSON
		if resp.Stdout != "" {
			data, err := h.parser.ParseJSON(ctx, resp.Stdout)
			if err == nil {
				// Successfully parsed JSON
				return []mcp.Content{
					{
						Type: "text",
						Text: formatJSONOutput(data),
					},
				}, nil
			}
		}
	}

	// Check for error output
	if resp.ExitCode != 0 {
		err := h.parser.ExtractError(ctx, resp.Stdout, resp.Stderr, resp.ExitCode)
		if err != nil {
			return []mcp.Content{
				{
					Type: "text",
					Text: fmt.Sprintf("Error: %v\n\nStderr:\n%s", err, resp.Stderr),
				},
			}, nil
		}
	}

	// Return raw output
	content := []mcp.Content{}
	if resp.Stdout != "" {
		content = append(content, mcp.Content{
			Type: "text",
			Text: resp.Stdout,
		})
	}
	if resp.Stderr != "" && resp.ExitCode != 0 {
		content = append(content, mcp.Content{
			Type: "text",
			Text: fmt.Sprintf("Error output:\n%s", resp.Stderr),
		})
	}

	// Add file references if any
	for _, file := range resp.OutputFiles {
		content = append(content, mcp.Content{
			Type: "text",
			Text: fmt.Sprintf("Generated file: %s (size: %d bytes)", file.Path, file.Size),
		})
	}

	return content, nil
}

// Helper functions

func convertParams(from interface{}, to interface{}) error {
	// This is a simplified conversion
	// In production, use proper JSON marshaling/unmarshaling
	return nil
}

func formatJSONOutput(data map[string]interface{}) string {
	// Format JSON output for display
	// This is a placeholder implementation
	return fmt.Sprintf("%v", data)
}
