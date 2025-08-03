package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mrz/go-invoice/internal/mcp/types"
)

// MCPHandler defines the interface for handling MCP requests (consumer-driven)
//
//nolint:revive // MCPHandler is intentionally prefixed to distinguish from generic handlers
type MCPHandler interface {
	HandleInitialize(ctx context.Context, req *types.MCPRequest) (*types.MCPResponse, error)
	HandlePing(ctx context.Context, req *types.MCPRequest) (*types.MCPResponse, error)
	HandleToolsList(ctx context.Context, req *types.MCPRequest) (*types.MCPResponse, error)
	HandleToolCall(ctx context.Context, req *types.MCPRequest) (*types.MCPResponse, error)
}

// DefaultMCPHandler implements the MCPHandler interface
type DefaultMCPHandler struct {
	logger Logger
	bridge CLIBridge
	config *Config
}

// NewMCPHandler creates a new MCP handler with dependency injection
func NewMCPHandler(logger Logger, bridge CLIBridge, config *Config) MCPHandler {
	return &DefaultMCPHandler{
		logger: logger,
		bridge: bridge,
		config: config,
	}
}

// HandleInitialize handles the MCP initialize request
func (h *DefaultMCPHandler) HandleInitialize(ctx context.Context, req *types.MCPRequest) (*types.MCPResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	h.logger.Info("Handling initialize request")

	result := types.InitializeResult{
		ProtocolVersion: "2024-11-05",
		Capabilities: types.Capabilities{
			Tools: &types.ToolsCapability{
				ListChanged: false,
			},
		},
		ServerInfo: types.ServerInfo{
			Name:    "go-invoice-mcp",
			Version: "1.0.0",
		},
	}

	return &types.MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}, nil
}

// HandlePing handles the MCP ping request
func (h *DefaultMCPHandler) HandlePing(ctx context.Context, req *types.MCPRequest) (*types.MCPResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	return &types.MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  map[string]string{"status": "ok"},
	}, nil
}

// HandleToolsList handles the tools/list request
func (h *DefaultMCPHandler) HandleToolsList(ctx context.Context, req *types.MCPRequest) (*types.MCPResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	h.logger.Debug("Handling tools/list request")

	// For Phase 1, return basic tools list
	// This will be expanded in Phase 2 with comprehensive tool definitions
	tools := []Tool{
		{
			Name:        "ping",
			Description: "Test connectivity to the go-invoice CLI",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "version",
			Description: "Get go-invoice CLI version information",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}

	result := ToolListResult{
		Tools: tools,
	}

	return &types.MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}, nil
}

// HandleToolCall handles the tools/call request
func (h *DefaultMCPHandler) HandleToolCall(ctx context.Context, req *types.MCPRequest) (*types.MCPResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Parse tool call parameters
	paramsBytes, err := json.Marshal(req.Params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal params: %w", err)
	}

	var params ToolCallParams
	if err := json.Unmarshal(paramsBytes, &params); err != nil {
		return nil, fmt.Errorf("failed to parse tool call params: %w", err)
	}

	h.logger.Debug("Handling tool call", "tool", params.Name)

	switch params.Name {
	case "ping":
		return h.handlePingTool(ctx, req, &params)
	case "version":
		return h.handleVersionTool(ctx, req, &params)
	default:
		return &types.MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &MCPError{
				Code:    -32602,
				Message: "Invalid params",
				Data:    fmt.Sprintf("Unknown tool: %s", params.Name),
			},
		}, nil
	}
}

// handlePingTool handles the ping tool call
func (h *DefaultMCPHandler) handlePingTool(ctx context.Context, req *types.MCPRequest, _ *ToolCallParams) (*types.MCPResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Test CLI connectivity
	cmdReq := &CommandRequest{
		Command: h.config.CLI.Path,
		Args:    []string{"--help"},
		Timeout: 5 * time.Second,
	}

	resp, err := h.bridge.ExecuteCommand(ctx, cmdReq)
	if err != nil {
		result := ToolCallResult{
			Content: []Content{
				{
					Type: "text",
					Text: fmt.Sprintf("CLI connectivity test failed: %v", err),
				},
			},
			IsError: true,
		}

		return &types.MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  result,
		}, nil
	}

	result := ToolCallResult{
		Content: []Content{
			{
				Type: "text",
				Text: fmt.Sprintf("CLI connectivity test successful (exit code: %d, duration: %v)", resp.ExitCode, resp.Duration),
			},
		},
		IsError: false,
	}

	return &types.MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}, nil
}

// handleVersionTool handles the version tool call
func (h *DefaultMCPHandler) handleVersionTool(ctx context.Context, req *types.MCPRequest, _ *ToolCallParams) (*types.MCPResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Get CLI version
	cmdReq := &CommandRequest{
		Command: h.config.CLI.Path,
		Args:    []string{"--version"},
		Timeout: 5 * time.Second,
	}

	resp, err := h.bridge.ExecuteCommand(ctx, cmdReq)
	if err != nil {
		result := ToolCallResult{
			Content: []Content{
				{
					Type: "text",
					Text: fmt.Sprintf("Failed to get CLI version: %v", err),
				},
			},
			IsError: true,
		}

		return &types.MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  result,
		}, nil
	}

	result := ToolCallResult{
		Content: []Content{
			{
				Type: "text",
				Text: fmt.Sprintf("go-invoice CLI version information:\n%s", resp.Stdout),
			},
		},
		IsError: resp.ExitCode != 0,
	}

	return &types.MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}, nil
}
