package mcp

import (
	"github.com/mrz/go-invoice/internal/mcp/types"
)

// Re-export types from the types package to maintain backward compatibility
type (
	TransportType    = types.TransportType
	MCPRequest       = types.MCPRequest
	MCPResponse      = types.MCPResponse
	MCPError         = types.MCPError
	InitializeParams = types.InitializeParams
	ClientInfo       = types.ClientInfo
	InitializeResult = types.InitializeResult
	Capabilities     = types.Capabilities
	ServerInfo       = types.ServerInfo
	ToolListResult   = types.ToolListResult
	Tool             = types.Tool
	ToolCallParams   = types.ToolCallParams
	ToolCallResult   = types.ToolCallResult
	Content          = types.Content
	CommandRequest   = types.CommandRequest
	CommandResponse  = types.CommandResponse
	Server           = types.Server
	CLIBridge        = types.CLIBridge
	Logger           = types.Logger
	CommandValidator = types.CommandValidator
	FileHandler      = types.FileHandler
)

// Re-export constants from the types package
const (
	TransportStdio = types.TransportStdio
	TransportHTTP  = types.TransportHTTP
)
