package types //nolint:revive // types package contains core MCP protocol type definitions

import (
	"context"
)

// TransportType represents the MCP transport mechanism
type TransportType string

const (
	// TransportStdio represents stdio transport for MCP communication
	TransportStdio TransportType = "stdio"
	// TransportHTTP represents HTTP transport for MCP communication
	TransportHTTP TransportType = "http"
)

// MCPRequest represents an MCP protocol request.
type MCPRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// MCPResponse represents an MCP protocol response.
type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

// MCPError represents an MCP protocol error.
type MCPError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ToolCallParams represents parameters for an MCP tool call.
type ToolCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

// ToolCallResult represents the result of an MCP tool call.
type ToolCallResult struct {
	Content []Content `json:"content"`
	IsError bool      `json:"isError,omitempty"`
}

// Content represents MCP content.
type Content struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	Resource string `json:"resource,omitempty"`
	MimeType string `json:"mimeType,omitempty"`
}

// InitializeResult represents the result of MCP initialization.
type InitializeResult struct {
	ProtocolVersion string       `json:"protocolVersion"`
	ServerInfo      ServerInfo   `json:"serverInfo"`
	Capabilities    Capabilities `json:"capabilities"`
}

// ServerInfo represents MCP server information.
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Capabilities represents MCP server capabilities.
type Capabilities struct {
	Tools     *ToolsCapability     `json:"tools,omitempty"`
	Resources *ResourcesCapability `json:"resources,omitempty"`
	Prompts   *PromptsCapability   `json:"prompts,omitempty"`
	Logging   *LoggingCapability   `json:"logging,omitempty"`
}

// ToolsCapability represents tool capabilities.
type ToolsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// ResourcesCapability represents resource capabilities.
type ResourcesCapability struct {
	Subscribe   bool `json:"subscribe,omitempty"`
	ListChanged bool `json:"listChanged,omitempty"`
}

// PromptsCapability represents prompt capabilities.
type PromptsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// LoggingCapability represents logging capabilities.
type LoggingCapability struct {
	Level string `json:"level,omitempty"`
}

// Tool represents an MCP tool definition.
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// ToolListResult represents the result of listing MCP tools.
type ToolListResult struct {
	Tools []Tool `json:"tools"`
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

// MCPHandler interface defines the MCP request handling contract (consumer-driven)
type MCPHandler interface {
	HandleRequest(ctx context.Context, req *MCPRequest) (*MCPResponse, error)
}
