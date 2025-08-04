package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ServerTestSuite struct {
	suite.Suite

	server Server
	logger *TestLogger
	bridge *MockCLIBridge
	config *Config
}

func TestServerSuite(t *testing.T) {
	suite.Run(t, new(ServerTestSuite))
}

func (s *ServerTestSuite) SetupTest() {
	s.logger = NewTestLogger()
	s.bridge = NewMockCLIBridge()
	s.config = &Config{
		Server: ServerConfig{
			Host:        "localhost",
			Port:        0,
			Timeout:     30 * time.Second,
			ReadTimeout: 10 * time.Second,
		},
		CLI: CLIConfig{
			Path:       "go-invoice",
			WorkingDir: "/tmp",
			MaxTimeout: 60 * time.Second,
		},
		Security: SecurityConfig{
			AllowedCommands: []string{"go-invoice"},
		},
		LogLevel: "debug",
	}
	s.server = NewServer(s.logger, s.bridge, s.config)
}

func (s *ServerTestSuite) TestHandleInitializeRequest() {
	ctx := context.Background()

	req := &MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo": map[string]interface{}{
				"name":    "test-client",
				"version": "1.0.0",
			},
		},
	}

	resp, err := s.server.HandleRequest(ctx, req)
	s.Require().NoError(err)
	s.NotNil(resp)
	s.Equal("2.0", resp.JSONRPC)
	s.Equal(1, resp.ID)
	s.Nil(resp.Error)

	// Verify result structure
	s.NotNil(resp.Result)
	result, ok := resp.Result.(InitializeResult)
	s.Require().True(ok)
	s.Equal("2024-11-05", result.ProtocolVersion)
	s.Equal("go-invoice-mcp", result.ServerInfo.Name)
	s.Equal("1.0.0", result.ServerInfo.Version)
}

func (s *ServerTestSuite) TestHandlePingRequest() {
	ctx := context.Background()

	req := &MCPRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "ping",
	}

	resp, err := s.server.HandleRequest(ctx, req)
	s.Require().NoError(err)
	s.NotNil(resp)
	s.Equal("2.0", resp.JSONRPC)
	s.Equal(2, resp.ID)
	s.Nil(resp.Error)

	result, ok := resp.Result.(map[string]string)
	s.Require().True(ok)
	s.Equal("ok", result["status"])
}

func (s *ServerTestSuite) TestHandleToolsListRequest() {
	ctx := context.Background()

	req := &MCPRequest{
		JSONRPC: "2.0",
		ID:      3,
		Method:  "tools/list",
	}

	resp, err := s.server.HandleRequest(ctx, req)
	s.Require().NoError(err)
	s.NotNil(resp)
	s.Equal("2.0", resp.JSONRPC)
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
	s.Contains(toolNames, "ping")
	s.Contains(toolNames, "version")
}

func (s *ServerTestSuite) TestHandleToolCallPing() {
	ctx := context.Background()

	// Configure mock bridge for success
	s.bridge.SetResponse(&CommandResponse{
		ExitCode: 0,
		Stdout:   "go-invoice help output",
		Stderr:   "",
		Duration: 100 * time.Millisecond,
	}, nil)

	req := &MCPRequest{
		JSONRPC: "2.0",
		ID:      4,
		Method:  "tools/call",
		Params: ToolCallParams{
			Name:      "ping",
			Arguments: map[string]interface{}{},
		},
	}

	resp, err := s.server.HandleRequest(ctx, req)
	s.Require().NoError(err)
	s.NotNil(resp)
	s.Nil(resp.Error)

	result, ok := resp.Result.(ToolCallResult)
	s.Require().True(ok)
	s.False(result.IsError)
	s.NotEmpty(result.Content)
	s.Contains(result.Content[0].Text, "CLI connectivity test successful")
}

func (s *ServerTestSuite) TestHandleToolCallVersion() {
	ctx := context.Background()

	// Configure mock bridge for success
	s.bridge.SetResponse(&CommandResponse{
		ExitCode: 0,
		Stdout:   "go-invoice version 1.0.0",
		Stderr:   "",
		Duration: 50 * time.Millisecond,
	}, nil)

	req := &MCPRequest{
		JSONRPC: "2.0",
		ID:      5,
		Method:  "tools/call",
		Params: ToolCallParams{
			Name:      "version",
			Arguments: map[string]interface{}{},
		},
	}

	resp, err := s.server.HandleRequest(ctx, req)
	s.Require().NoError(err)
	s.NotNil(resp)
	s.Nil(resp.Error)

	result, ok := resp.Result.(ToolCallResult)
	s.Require().True(ok)
	s.False(result.IsError)
	s.NotEmpty(result.Content)
	s.Contains(result.Content[0].Text, "go-invoice version 1.0.0")
}

func (s *ServerTestSuite) TestHandleToolCallError() {
	ctx := context.Background()

	// Configure mock bridge for error
	s.bridge.SetResponse(nil, assert.AnError)

	req := &MCPRequest{
		JSONRPC: "2.0",
		ID:      6,
		Method:  "tools/call",
		Params: ToolCallParams{
			Name:      "ping",
			Arguments: map[string]interface{}{},
		},
	}

	resp, err := s.server.HandleRequest(ctx, req)
	s.Require().NoError(err)
	s.NotNil(resp)
	s.Nil(resp.Error) // Error should be in result, not response error

	result, ok := resp.Result.(ToolCallResult)
	s.Require().True(ok)
	s.True(result.IsError)
	s.NotEmpty(result.Content)
	s.Contains(result.Content[0].Text, "CLI connectivity test failed")
}

func (s *ServerTestSuite) TestHandleUnknownMethod() {
	ctx := context.Background()

	req := &MCPRequest{
		JSONRPC: "2.0",
		ID:      7,
		Method:  "unknown/method",
	}

	resp, err := s.server.HandleRequest(ctx, req)
	s.Require().NoError(err)
	s.NotNil(resp)
	s.NotNil(resp.Error)
	s.Equal(-32601, resp.Error.Code)
	s.Contains(resp.Error.Message, "Method not found")
}

func (s *ServerTestSuite) TestHandleUnknownTool() {
	ctx := context.Background()

	req := &MCPRequest{
		JSONRPC: "2.0",
		ID:      8,
		Method:  "tools/call",
		Params: ToolCallParams{
			Name:      "unknown-tool",
			Arguments: map[string]interface{}{},
		},
	}

	resp, err := s.server.HandleRequest(ctx, req)
	s.Require().NoError(err)
	s.NotNil(resp)
	s.NotNil(resp.Error)
	s.Equal(-32602, resp.Error.Code)
	s.Contains(resp.Error.Data, "Unknown tool")
}

func (s *ServerTestSuite) TestHandleRequestContextCancellation() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Immediately cancel

	req := &MCPRequest{
		JSONRPC: "2.0",
		ID:      9,
		Method:  "ping",
	}

	_, err := s.server.HandleRequest(ctx, req)
	s.Equal(context.Canceled, err)
}

func (s *ServerTestSuite) TestHTTPTransportHandler() {
	// Create test HTTP request
	reqBody := `{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "ping"
	}`

	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Get the server instance
	defaultServer := s.server.(*DefaultServer)

	// Call HTTP handler
	defaultServer.handleHTTPRequest(w, req)

	// Verify response
	s.Equal(http.StatusOK, w.Code)
	s.Equal("application/json", w.Header().Get("Content-Type"))

	var resp MCPResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	s.Require().NoError(err)
	s.Equal("2.0", resp.JSONRPC)
	// JSON unmarshaling converts numbers to float64, so we need to handle that
	s.InDelta(float64(1), resp.ID, 0.001)
}

func (s *ServerTestSuite) TestHTTPTransportInvalidMethod() {
	req := httptest.NewRequest(http.MethodGet, "/mcp", nil)
	w := httptest.NewRecorder()

	defaultServer := s.server.(*DefaultServer)
	defaultServer.handleHTTPRequest(w, req)

	s.Equal(http.StatusMethodNotAllowed, w.Code)
}

func (s *ServerTestSuite) TestHTTPTransportInvalidJSON() {
	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader("invalid json"))
	w := httptest.NewRecorder()

	defaultServer := s.server.(*DefaultServer)
	defaultServer.handleHTTPRequest(w, req)

	s.Equal(http.StatusBadRequest, w.Code)
}

// MockCLIBridge for testing
type MockCLIBridge struct {
	response *CommandResponse
	err      error
}

func NewMockCLIBridge() *MockCLIBridge {
	return &MockCLIBridge{}
}

func (m *MockCLIBridge) SetResponse(resp *CommandResponse, err error) {
	m.response = resp
	m.err = err
}

func (m *MockCLIBridge) ExecuteCommand(ctx context.Context, _ *CommandRequest) (*CommandResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if m.err != nil {
		return m.response, m.err
	}

	if m.response != nil {
		return m.response, nil
	}

	// Default successful response
	return &CommandResponse{
		ExitCode: 0,
		Stdout:   "mock output",
		Stderr:   "",
		Duration: 10 * time.Millisecond,
	}, nil
}

func (m *MockCLIBridge) ValidateCommand(ctx context.Context, _ string, _ []string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return nil
}

func (m *MockCLIBridge) GetAllowedCommands(ctx context.Context) ([]string, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	return []string{"go-invoice"}, nil
}

// Benchmark tests for server performance validation
func BenchmarkServerHandleRequest(b *testing.B) {
	logger := NewTestLogger()
	bridge := NewMockCLIBridge()
	config := &Config{
		CLI: CLIConfig{Path: "go-invoice"},
	}
	server := NewServer(logger, bridge, config)

	ctx := context.Background()
	req := &MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "ping",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := server.HandleRequest(ctx, req)
		require.NoError(b, err)
	}
}
