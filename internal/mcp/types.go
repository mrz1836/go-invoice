package mcp

import (
	"context"
	"time"
)

// TransportType represents the MCP transport mechanism
type TransportType string

const (
	// TransportStdio represents stdio transport for MCP communication
	TransportStdio TransportType = "stdio"
	// TransportHTTP represents HTTP transport for MCP communication
	TransportHTTP TransportType = "http"
)

// MCPRequest represents an incoming MCP protocol request
//
//nolint:revive // MCPRequest is intentionally prefixed to distinguish from generic requests
type MCPRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// MCPResponse represents an outgoing MCP protocol response
//
//nolint:revive // MCPResponse is intentionally prefixed to distinguish from generic responses
type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

// MCPError represents an MCP protocol error
//
//nolint:revive // MCPError is intentionally prefixed to distinguish from generic errors
type MCPError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// InitializeParams represents MCP initialize request parameters
type InitializeParams struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	ClientInfo      ClientInfo             `json:"clientInfo"`
}

// ClientInfo represents client information in initialize request
type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// InitializeResult represents MCP initialize response result
type InitializeResult struct {
	ProtocolVersion string       `json:"protocolVersion"`
	Capabilities    Capabilities `json:"capabilities"`
	ServerInfo      ServerInfo   `json:"serverInfo"`
}

// Capabilities represents server capabilities
type Capabilities struct {
	Tools []string `json:"tools,omitempty"`
}

// ServerInfo represents server information
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ToolListResult represents tools/list response
type ToolListResult struct {
	Tools []Tool `json:"tools"`
}

// Tool represents an MCP tool definition
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// ToolCallParams represents tools/call request parameters
type ToolCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

// ToolCallResult represents tools/call response result
type ToolCallResult struct {
	Content []Content `json:"content"`
	IsError bool      `json:"isError,omitempty"`
}

// Content represents MCP content block
type Content struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	Resource string `json:"resource,omitempty"`
	MimeType string `json:"mimeType,omitempty"`
}

// CommandRequest represents a request to execute a CLI command
type CommandRequest struct {
	Command    string            `json:"command"`
	Args       []string          `json:"args"`
	WorkingDir string            `json:"workingDir,omitempty"`
	Env        map[string]string `json:"env,omitempty"`
	Timeout    time.Duration     `json:"timeout,omitempty"`
}

// CommandResponse represents the result of a CLI command execution
type CommandResponse struct {
	ExitCode int           `json:"exitCode"`
	Stdout   string        `json:"stdout"`
	Stderr   string        `json:"stderr"`
	Duration time.Duration `json:"duration"`
	Error    string        `json:"error,omitempty"`
}

// Server interface defines the MCP server contract (consumer-driven)
type Server interface {
	Start(ctx context.Context, transport TransportType) error
	Shutdown(ctx context.Context) error
	HandleRequest(ctx context.Context, req *MCPRequest) (*MCPResponse, error)
}

// CLIBridge interface defines the CLI command execution contract (consumer-driven)
type CLIBridge interface {
	ExecuteCommand(ctx context.Context, req *CommandRequest) (*CommandResponse, error)
	ValidateCommand(ctx context.Context, command string, args []string) error
	GetAllowedCommands(ctx context.Context) ([]string, error)
}

// Logger interface defines logging contract (consumer-driven)
type Logger interface {
	Debug(msg string, keysAndValues ...interface{})
	Info(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
}

// CommandValidator interface defines command validation contract (consumer-driven)
type CommandValidator interface {
	ValidateCommand(ctx context.Context, command string, args []string) error
	IsCommandAllowed(ctx context.Context, command string) (bool, error)
}

// FileHandler interface defines file handling contract (consumer-driven)
type FileHandler interface {
	PrepareWorkspace(ctx context.Context, workingDir string) (string, func(), error)
	ValidatePath(ctx context.Context, path string) error
}
