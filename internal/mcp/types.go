package mcp

import (
	"github.com/mrz/go-invoice/internal/mcp/types"
)

type (
	// TransportType represents the MCP transport mechanism (re-exported for compatibility)
	TransportType = types.TransportType
	// Request represents an MCP protocol request (renamed from MCPRequest to avoid stuttering)
	Request = types.MCPRequest
	// Response represents an MCP protocol response (renamed from MCPResponse to avoid stuttering)
	Response = types.MCPResponse
	// Error represents an MCP protocol error (renamed from MCPError to avoid stuttering)
	Error = types.MCPError
	// InitializeParams represents initialization parameters for MCP connections
	InitializeParams = types.InitializeParams
	// ClientInfo represents client identification information
	ClientInfo = types.ClientInfo
	// InitializeResult represents the result of initialization
	InitializeResult = types.InitializeResult
	// Capabilities represents server capabilities
	Capabilities = types.Capabilities
	// ServerInfo represents server identification information
	ServerInfo = types.ServerInfo
	// ToolListResult represents the result of listing tools
	ToolListResult = types.ToolListResult
	// Tool represents a tool definition
	Tool = types.Tool
	// ToolCallParams represents parameters for calling a tool
	ToolCallParams = types.ToolCallParams
	// ToolCallResult represents the result of calling a tool
	ToolCallResult = types.ToolCallResult
	// Content represents message content
	Content = types.Content
	// CommandRequest represents a command request
	CommandRequest = types.CommandRequest
	// CommandResponse represents a command response
	CommandResponse = types.CommandResponse
	// Server represents an MCP server instance
	Server = types.Server
	// CLIBridge represents a CLI bridge interface
	CLIBridge = types.CLIBridge
	// Logger represents a logging interface
	Logger = types.Logger
	// CommandValidator represents a command validation interface
	CommandValidator = types.CommandValidator
	// FileHandler represents a file handling interface
	FileHandler = types.FileHandler
)

// Backward compatibility type aliases for deprecated MCP types.
// These are provided for backward compatibility and should not be used in new code.
type (
	// MCPRequest is deprecated. Use Request instead.
	MCPRequest = Request
	// MCPResponse is deprecated. Use Response instead.
	MCPResponse = Response
	// MCPError is deprecated. Use Error instead.
	MCPError = Error
)

// Re-export constants from the types package
const (
	TransportStdio = types.TransportStdio
	TransportHTTP  = types.TransportHTTP
)
