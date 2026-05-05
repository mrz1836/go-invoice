package mcp

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type HandlersTestSuite struct {
	suite.Suite

	handler MCPHandler
	logger  *TestLogger
	bridge  *MockCLIBridge
	config  *Config
}

func TestHandlersSuite(t *testing.T) {
	suite.Run(t, new(HandlersTestSuite))
}

func (s *HandlersTestSuite) SetupTest() {
	s.logger = NewTestLogger()
	s.bridge = NewMockCLIBridge()
	s.config = &Config{
		CLI: CLIConfig{
			Path:       defaultCLIName,
			WorkingDir: "/tmp",
			MaxTimeout: 60 * time.Second,
		},
	}
	s.handler = NewMCPHandler(s.logger, s.bridge, s.config)
}

func (s *HandlersTestSuite) TestHandleInitialize() {
	ctx := context.Background()

	req := &MCPRequest{
		JSONRPC: jsonRPCVersion,
		ID:      1,
		Method:  methodInitialize,
	}

	resp, err := s.handler.HandleInitialize(ctx, req)
	s.Require().NoError(err)
	s.NotNil(resp)
	s.Equal(jsonRPCVersion, resp.JSONRPC) //nolint:testifylint // JSONRPC is a version string, not JSON-encoded data
	s.Equal(1, resp.ID)
	s.Nil(resp.Error)

	result, ok := resp.Result.(InitializeResult)
	s.Require().True(ok)
	s.Equal("2024-11-05", result.ProtocolVersion)
	s.Equal("go-invoice-mcp", result.ServerInfo.Name)
	s.Equal("1.0.0", result.ServerInfo.Version)
}

func (s *HandlersTestSuite) TestHandlePing() {
	ctx := context.Background()

	req := &MCPRequest{
		JSONRPC: jsonRPCVersion,
		ID:      2,
		Method:  methodPing,
	}

	resp, err := s.handler.HandlePing(ctx, req)
	s.Require().NoError(err)
	s.NotNil(resp)
	s.Equal(jsonRPCVersion, resp.JSONRPC) //nolint:testifylint // JSONRPC is a version string, not JSON-encoded data
	s.Equal(2, resp.ID)
	s.Nil(resp.Error)

	result, ok := resp.Result.(map[string]string)
	s.Require().True(ok)
	s.Equal("ok", result["status"])
}

func (s *HandlersTestSuite) TestHandleToolsList() {
	ctx := context.Background()

	req := &MCPRequest{
		JSONRPC: jsonRPCVersion,
		ID:      3,
		Method:  methodToolsList,
	}

	resp, err := s.handler.HandleToolsList(ctx, req)
	s.Require().NoError(err)
	s.NotNil(resp)
	s.Equal(jsonRPCVersion, resp.JSONRPC) //nolint:testifylint // JSONRPC is a version string, not JSON-encoded data
	s.Equal(3, resp.ID)
	s.Nil(resp.Error)

	result, ok := resp.Result.(ToolListResult)
	s.Require().True(ok)
	s.NotEmpty(result.Tools)

	// Verify basic tools are present
	toolNames := make([]string, len(result.Tools))
	for i, tool := range result.Tools {
		toolNames[i] = tool.Name
	}
	s.Contains(toolNames, methodPing)
	s.Contains(toolNames, "version")
}

func (s *HandlersTestSuite) TestHandleToolCallPing() {
	ctx := context.Background()

	// Configure mock bridge for success
	s.bridge.SetResponse(&CommandResponse{
		ExitCode: 0,
		Stdout:   "go-invoice help output",
		Stderr:   "",
		Duration: 100 * time.Millisecond,
	}, nil)

	req := &MCPRequest{
		JSONRPC: jsonRPCVersion,
		ID:      4,
		Method:  methodToolsCall,
		Params: ToolCallParams{
			Name:      methodPing,
			Arguments: map[string]interface{}{},
		},
	}

	resp, err := s.handler.HandleToolCall(ctx, req)
	s.Require().NoError(err)
	s.NotNil(resp)
	s.Nil(resp.Error)

	result, ok := resp.Result.(ToolCallResult)
	s.Require().True(ok)
	s.False(result.IsError)
	s.NotEmpty(result.Content)
	s.Contains(result.Content[0].Text, "CLI connectivity test successful")
}

func (s *HandlersTestSuite) TestHandleToolCallVersion() {
	ctx := context.Background()

	// Configure mock bridge for success
	s.bridge.SetResponse(&CommandResponse{
		ExitCode: 0,
		Stdout:   "go-invoice version 1.0.0",
		Stderr:   "",
		Duration: 50 * time.Millisecond,
	}, nil)

	req := &MCPRequest{
		JSONRPC: jsonRPCVersion,
		ID:      5,
		Method:  methodToolsCall,
		Params: ToolCallParams{
			Name:      "version",
			Arguments: map[string]interface{}{},
		},
	}

	resp, err := s.handler.HandleToolCall(ctx, req)
	s.Require().NoError(err)
	s.NotNil(resp)
	s.Nil(resp.Error)

	result, ok := resp.Result.(ToolCallResult)
	s.Require().True(ok)
	s.False(result.IsError)
	s.NotEmpty(result.Content)
	s.Contains(result.Content[0].Text, "go-invoice version 1.0.0")
}

func (s *HandlersTestSuite) TestHandleToolCallUnknownTool() {
	ctx := context.Background()

	req := &MCPRequest{
		JSONRPC: jsonRPCVersion,
		ID:      6,
		Method:  methodToolsCall,
		Params: ToolCallParams{
			Name:      "unknown-tool",
			Arguments: map[string]interface{}{},
		},
	}

	resp, err := s.handler.HandleToolCall(ctx, req)
	s.Require().NoError(err)
	s.NotNil(resp)
	s.NotNil(resp.Error)
	s.Equal(-32602, resp.Error.Code)
	s.Contains(resp.Error.Data, "Unknown tool")
}

func (s *HandlersTestSuite) TestHandleContextCancellation() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Immediately cancel

	req := &MCPRequest{
		JSONRPC: jsonRPCVersion,
		ID:      7,
		Method:  methodPing,
	}

	_, err := s.handler.HandlePing(ctx, req)
	s.Equal(context.Canceled, err)

	_, err = s.handler.HandleInitialize(ctx, req)
	s.Equal(context.Canceled, err)

	_, err = s.handler.HandleToolsList(ctx, req)
	s.Equal(context.Canceled, err)

	_, err = s.handler.HandleToolCall(ctx, req)
	s.Equal(context.Canceled, err)
}

// Benchmark tests for performance validation
func BenchmarkHandleInitialize(b *testing.B) {
	logger := NewTestLogger()
	bridge := NewMockCLIBridge()
	config := &Config{
		CLI: CLIConfig{Path: defaultCLIName},
	}
	handler := NewMCPHandler(logger, bridge, config)

	ctx := context.Background()
	req := &MCPRequest{
		JSONRPC: jsonRPCVersion,
		ID:      1,
		Method:  methodInitialize,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := handler.HandleInitialize(ctx, req)
		require.NoError(b, err)
	}
}

func BenchmarkHandleToolCall(b *testing.B) {
	logger := NewTestLogger()
	bridge := NewMockCLIBridge()
	bridge.SetResponse(&CommandResponse{
		ExitCode: 0,
		Stdout:   "benchmark output",
		Duration: 1 * time.Millisecond,
	}, nil)

	config := &Config{
		CLI: CLIConfig{Path: defaultCLIName},
	}
	handler := NewMCPHandler(logger, bridge, config)

	ctx := context.Background()
	req := &MCPRequest{
		JSONRPC: jsonRPCVersion,
		ID:      1,
		Method:  methodToolsCall,
		Params: ToolCallParams{
			Name:      methodPing,
			Arguments: map[string]interface{}{},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := handler.HandleToolCall(ctx, req)
		require.NoError(b, err)
	}
}
