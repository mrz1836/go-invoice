package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/mrz/go-invoice/internal/mcp/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// Test errors
var (
	errMockHandler   = errors.New("mock handler error")
	errUnknownMethod = errors.New("unknown method")
)

type TransportTestSuite struct {
	suite.Suite

	logger  *TestLogger
	config  *TransportConfig
	tempDir string
}

func TestTransportSuite(t *testing.T) {
	suite.Run(t, new(TransportTestSuite))
}

func (s *TransportTestSuite) SetupTest() {
	s.logger = NewTestLogger()
	s.config = DefaultTransportConfig()

	tempDir, err := os.MkdirTemp("", "transport-test")
	s.Require().NoError(err)
	s.tempDir = tempDir
}

func (s *TransportTestSuite) TearDownTest() {
	if s.tempDir != "" {
		err := os.RemoveAll(s.tempDir)
		s.Require().NoError(err)
	}
}

// Test Transport Errors
func (s *TransportTestSuite) TestTransportErrors() {
	tests := []struct {
		name    string
		err     *TransportError
		message string
	}{
		{
			name:    "TransportNotInitialized",
			err:     ErrTransportNotInitialized,
			message: "transport init: transport not initialized",
		},
		{
			name:    "InvalidTransportType",
			err:     ErrInvalidTransportType,
			message: "transport validate: invalid transport type",
		},
		{
			name:    "TransportClosed",
			err:     ErrTransportClosed,
			message: "transport send: transport closed",
		},
		{
			name:    "MessageTooLarge",
			err:     ErrMessageTooLarge,
			message: "transport validate: message exceeds size limit",
		},
		{
			name:    "HandlerRequired",
			err:     ErrHandlerRequired,
			message: "transport create: handler required for HTTP transport",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Equal(tt.message, tt.err.Error())
		})
	}
}

// Test Default Transport Config
func (s *TransportTestSuite) TestDefaultTransportConfig() {
	config := DefaultTransportConfig()
	s.NotNil(config)
	s.Equal(types.TransportStdio, config.Type)
	s.Equal("localhost", config.Host)
	s.Equal(0, config.Port)
	s.Equal(30*time.Second, config.ReadTimeout)
	s.Equal(30*time.Second, config.WriteTimeout)
	s.Equal(int64(10*1024*1024), config.MaxMessageSize)
	s.True(config.EnableLogging)
	s.Equal("info", config.LogLevel)
}

// Stdio Transport Tests
func (s *TransportTestSuite) TestNewStdioTransport() {
	tests := []struct {
		name        string
		logger      Logger
		config      *TransportConfig
		expectPanic bool
	}{
		{
			name:        "ValidInputs",
			logger:      s.logger,
			config:      s.config,
			expectPanic: false,
		},
		{
			name:        "NilLogger",
			logger:      nil,
			config:      s.config,
			expectPanic: true,
		},
		{
			name:        "NilConfig",
			logger:      s.logger,
			config:      nil,
			expectPanic: false, // Should use default config
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			if tt.expectPanic {
				s.Panics(func() {
					NewStdioTransport(tt.logger, tt.config)
				})
			} else {
				transport := NewStdioTransport(tt.logger, tt.config)
				s.NotNil(transport)
				s.Equal(types.TransportStdio, transport.Type())
				s.Equal(tt.logger, transport.logger)
				if tt.config != nil {
					s.Equal(tt.config, transport.config)
				} else {
					s.NotNil(transport.config) // Should have default config
				}
			}
		})
	}
}

func (s *TransportTestSuite) TestStdioTransportStartStop() {
	transport := NewStdioTransport(s.logger, s.config)
	ctx := context.Background()

	// Test Start
	err := transport.Start(ctx)
	s.Require().NoError(err)
	s.True(transport.IsHealthy(ctx))

	// Test Stop
	err = transport.Stop(ctx)
	s.Require().NoError(err)
	s.False(transport.IsHealthy(ctx))

	// Test Start after stop should fail
	err = transport.Start(ctx)
	s.Equal(ErrTransportClosed, err)

	// Test Stop when already stopped
	err = transport.Stop(ctx)
	s.NoError(err) // Should not error when already stopped
}

func (s *TransportTestSuite) TestStdioTransportSendWhenClosed() {
	transport := NewStdioTransport(s.logger, s.config)
	ctx := context.Background()

	// Start and then stop the transport
	err := transport.Start(ctx)
	s.Require().NoError(err)
	err = transport.Stop(ctx)
	s.Require().NoError(err)

	// Try to send when closed
	response := &types.MCPResponse{
		JSONRPC: "2.0",
		ID:      1,
		Result:  map[string]string{"status": "ok"},
	}

	err = transport.Send(ctx, response)
	s.Equal(ErrTransportClosed, err)
}

func (s *TransportTestSuite) TestStdioTransportReceiveWhenClosed() {
	transport := NewStdioTransport(s.logger, s.config)
	ctx := context.Background()

	// Start and then stop the transport
	err := transport.Start(ctx)
	s.Require().NoError(err)
	err = transport.Stop(ctx)
	s.Require().NoError(err)

	// Try to receive when closed
	req, err := transport.Receive(ctx)
	s.Nil(req)
	s.Equal(ErrTransportClosed, err)
}

func (s *TransportTestSuite) TestStdioTransportContextCancellation() {
	transport := NewStdioTransport(s.logger, s.config)

	// Test context cancellation on Receive
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	req, err := transport.Receive(ctx)
	s.Nil(req)
	s.Equal(context.Canceled, err)

	// Test context cancellation on Send
	response := &types.MCPResponse{
		JSONRPC: "2.0",
		ID:      1,
		Result:  map[string]string{"status": "ok"},
	}

	err = transport.Send(ctx, response)
	s.Equal(context.Canceled, err)
}

// HTTP Transport Tests
func (s *TransportTestSuite) TestNewHTTPTransport() {
	handler := &mockMCPHandler{}

	tests := []struct {
		name        string
		logger      Logger
		config      *TransportConfig
		handler     types.MCPHandler
		expectPanic bool
	}{
		{
			name:        "ValidInputs",
			logger:      s.logger,
			config:      s.config,
			handler:     handler,
			expectPanic: false,
		},
		{
			name:        "NilLogger",
			logger:      nil,
			config:      s.config,
			handler:     handler,
			expectPanic: true,
		},
		{
			name:        "NilHandler",
			logger:      s.logger,
			config:      s.config,
			handler:     nil,
			expectPanic: true,
		},
		{
			name:        "NilConfig",
			logger:      s.logger,
			config:      nil,
			handler:     handler,
			expectPanic: false, // Should use default config
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			if tt.expectPanic {
				s.Panics(func() {
					NewHTTPTransport(tt.logger, tt.config, tt.handler)
				})
			} else {
				transport := NewHTTPTransport(tt.logger, tt.config, tt.handler)
				s.NotNil(transport)
				s.Equal(types.TransportHTTP, transport.Type())
				s.Equal(tt.logger, transport.logger)
			}
		})
	}
}

func (s *TransportTestSuite) TestHTTPTransportStartStop() {
	handler := &mockMCPHandler{}
	config := &TransportConfig{
		Type:           types.TransportHTTP,
		Host:           "localhost",
		Port:           0, // Auto-assign port
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		MaxMessageSize: 1024 * 1024,
	}

	transport := NewHTTPTransport(s.logger, config, handler)
	ctx := context.Background()

	// Test Start
	err := transport.Start(ctx)
	s.Require().NoError(err)
	s.True(transport.IsHealthy(ctx))

	// Test Stop
	err = transport.Stop(ctx)
	s.Require().NoError(err)
	s.False(transport.IsHealthy(ctx))

	// Test Start after stop should fail
	err = transport.Start(ctx)
	s.Equal(ErrTransportClosed, err)
}

func (s *TransportTestSuite) TestHTTPTransportHandleMCPRequest() {
	handler := &mockMCPHandler{
		response: &types.MCPResponse{
			JSONRPC: "2.0",
			ID:      1,
			Result:  map[string]string{"status": "ok"},
		},
	}

	config := &TransportConfig{
		Type:           types.TransportHTTP,
		Host:           "localhost",
		Port:           0,
		MaxMessageSize: 1024 * 1024,
	}

	transport := NewHTTPTransport(s.logger, config, handler)

	tests := []struct {
		name           string
		method         string
		contentType    string
		body           string
		expectedStatus int
		expectJSON     bool
	}{
		{
			name:           "ValidPOSTRequest",
			method:         "POST",
			contentType:    "application/json",
			body:           `{"jsonrpc":"2.0","id":1,"method":"ping"}`,
			expectedStatus: http.StatusOK,
			expectJSON:     true,
		},
		{
			name:           "InvalidMethod",
			method:         "GET",
			contentType:    "application/json",
			body:           `{"jsonrpc":"2.0","id":1,"method":"ping"}`,
			expectedStatus: http.StatusMethodNotAllowed,
			expectJSON:     false,
		},
		{
			name:           "InvalidContentType",
			method:         "POST",
			contentType:    "text/plain",
			body:           `{"jsonrpc":"2.0","id":1,"method":"ping"}`,
			expectedStatus: http.StatusBadRequest,
			expectJSON:     false,
		},
		{
			name:           "InvalidJSON",
			method:         "POST",
			contentType:    "application/json",
			body:           `invalid json`,
			expectedStatus: http.StatusBadRequest,
			expectJSON:     false,
		},
		{
			name:           "EmptyBody",
			method:         "POST",
			contentType:    "application/json",
			body:           "",
			expectedStatus: http.StatusBadRequest,
			expectJSON:     false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			req := httptest.NewRequest(tt.method, "/mcp", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", tt.contentType)
			w := httptest.NewRecorder()

			transport.handleMCPRequest(w, req)
			s.Equal(tt.expectedStatus, w.Code)

			if tt.expectJSON {
				s.Equal("application/json", w.Header().Get("Content-Type"))
				var response types.MCPResponse
				err := json.NewDecoder(w.Body).Decode(&response)
				s.Require().NoError(err)
				s.Equal("2.0", response.JSONRPC)
			}
		})
	}
}

func (s *TransportTestSuite) TestHTTPTransportHandleHealthCheck() {
	handler := &mockMCPHandler{}
	transport := NewHTTPTransport(s.logger, s.config, handler)

	// Start the transport to make it ready
	ctx := context.Background()
	err := transport.Start(ctx)
	s.Require().NoError(err)
	defer func() {
		s.Require().NoError(transport.Stop(ctx))
	}()

	tests := []struct {
		name           string
		method         string
		expectedStatus int
		expectJSON     bool
	}{
		{
			name:           "ValidHealthCheck",
			method:         "GET",
			expectedStatus: http.StatusOK, // Should be OK for healthy transport
			expectJSON:     true,
		},
		{
			name:           "InvalidMethod",
			method:         "POST",
			expectedStatus: http.StatusMethodNotAllowed,
			expectJSON:     false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			req := httptest.NewRequest(tt.method, "/health", nil)
			w := httptest.NewRecorder()

			transport.handleHealthCheck(w, req)
			s.Equal(tt.expectedStatus, w.Code)

			if tt.expectJSON {
				s.Equal("application/json", w.Header().Get("Content-Type"))
				var status map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&status)
				s.Require().NoError(err)
				s.Equal("ok", status["status"])
				s.Equal("http", status["transport"])
			}
		})
	}
}

func (s *TransportTestSuite) TestHTTPTransportMessageSizeLimit() {
	handler := &mockMCPHandler{}
	config := &TransportConfig{
		Type:           types.TransportHTTP,
		Host:           "localhost",
		Port:           0,
		MaxMessageSize: 10, // Very small limit for testing
	}

	transport := NewHTTPTransport(s.logger, config, handler)

	// Create a request larger than the limit
	largeBody := strings.Repeat("x", 20)
	req := httptest.NewRequest("POST", "/mcp", strings.NewReader(largeBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	transport.handleMCPRequest(w, req)
	s.Equal(http.StatusBadRequest, w.Code)
}

// Transport Factory Tests
func (s *TransportTestSuite) TestNewTransportFactory() {
	tests := []struct {
		name        string
		logger      Logger
		expectPanic bool
	}{
		{
			name:        "ValidLogger",
			logger:      s.logger,
			expectPanic: false,
		},
		{
			name:        "NilLogger",
			logger:      nil,
			expectPanic: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			if tt.expectPanic {
				s.Panics(func() {
					NewTransportFactory(tt.logger)
				})
			} else {
				factory := NewTransportFactory(tt.logger)
				s.NotNil(factory)
				s.Equal(tt.logger, factory.logger)
			}
		})
	}
}

func (s *TransportTestSuite) TestTransportFactoryCreateTransport() {
	factory := NewTransportFactory(s.logger)
	handler := &mockMCPHandler{}

	tests := []struct {
		name        string
		config      *TransportConfig
		handler     types.MCPHandler
		expectError bool
		expectType  types.TransportType
	}{
		{
			name: "CreateStdioTransport",
			config: &TransportConfig{
				Type: types.TransportStdio,
			},
			handler:     nil, // Not required for stdio
			expectError: false,
			expectType:  types.TransportStdio,
		},
		{
			name: "CreateHTTPTransport",
			config: &TransportConfig{
				Type: types.TransportHTTP,
			},
			handler:     handler,
			expectError: false,
			expectType:  types.TransportHTTP,
		},
		{
			name: "HTTPTransportWithoutHandler",
			config: &TransportConfig{
				Type: types.TransportHTTP,
			},
			handler:     nil,
			expectError: true,
		},
		{
			name: "InvalidTransportType",
			config: &TransportConfig{
				Type: types.TransportType("invalid"),
			},
			handler:     handler,
			expectError: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			transport, err := factory.CreateTransport(tt.config, tt.handler)

			if tt.expectError {
				s.Require().Error(err)
				s.Nil(transport)
			} else {
				s.Require().NoError(err)
				s.NotNil(transport)
				s.Equal(tt.expectType, transport.Type())
			}
		})
	}
}

// Transport Detection Tests
func (s *TransportTestSuite) TestDetectTransport() {
	// Save original args
	originalArgs := os.Args

	tests := []struct {
		name        string
		args        []string
		envVar      string
		expected    types.TransportType
		description string
	}{
		{
			name:        "StdioFromArgs",
			args:        []string{"program", "--stdio"},
			envVar:      "",
			expected:    types.TransportStdio,
			description: "Should detect stdio from command line argument",
		},
		{
			name:        "HTTPFromArgs",
			args:        []string{"program", "--http"},
			envVar:      "",
			expected:    types.TransportHTTP,
			description: "Should detect HTTP from command line argument",
		},
		{
			name:        "StdioFromEnv",
			args:        []string{"program"},
			envVar:      "stdio",
			expected:    types.TransportStdio,
			description: "Should detect stdio from environment variable",
		},
		{
			name:        "HTTPFromEnv",
			args:        []string{"program"},
			envVar:      "http",
			expected:    types.TransportHTTP,
			description: "Should detect HTTP from environment variable",
		},
		{
			name:        "DefaultToStdio",
			args:        []string{"program"},
			envVar:      "",
			expected:    types.TransportStdio,
			description: "Should default to stdio when no explicit transport specified",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// Set up test environment
			os.Args = tt.args
			if tt.envVar != "" {
				err := os.Setenv("MCP_TRANSPORT", tt.envVar)
				s.Require().NoError(err)
			} else {
				err := os.Unsetenv("MCP_TRANSPORT")
				s.Require().NoError(err)
			}

			result := DetectTransport()
			s.Equal(tt.expected, result, tt.description)
		})
	}

	// Restore original environment
	os.Args = originalArgs
	err := os.Unsetenv("MCP_TRANSPORT")
	s.Require().NoError(err)
}

// Transport Metrics Tests
func (s *TransportTestSuite) TestTransportMetrics() {
	metrics := NewTransportMetrics()
	s.NotNil(metrics)
	s.Equal(uint64(0), metrics.MessagesReceived())
	s.Equal(uint64(0), metrics.MessagesSent())
	s.Positive(metrics.Uptime())

	// Test recording messages
	metrics.RecordReceive(5)
	metrics.RecordSend(3)

	s.Equal(uint64(5), metrics.MessagesReceived())
	s.Equal(uint64(3), metrics.MessagesSent())

	// Test concurrent access
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			metrics.RecordReceive(1)
		}()
		go func() {
			defer wg.Done()
			metrics.RecordSend(1)
		}()
	}
	wg.Wait()

	s.Equal(uint64(15), metrics.MessagesReceived()) // 5 + 10
	s.Equal(uint64(13), metrics.MessagesSent())     // 3 + 10
}

// Security Tests
func (s *TransportTestSuite) TestHTTPTransportSecurity() {
	handler := &mockMCPHandler{}
	transport := NewHTTPTransport(s.logger, s.config, handler)

	tests := []struct {
		name           string
		setupRequest   func() *http.Request
		expectedStatus int
		description    string
	}{
		{
			name: "XSSAttempt",
			setupRequest: func() *http.Request {
				body := `{"jsonrpc":"2.0","id":"<script>alert('xss')</script>","method":"ping"}`
				req := httptest.NewRequest("POST", "/mcp", strings.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				return req
			},
			expectedStatus: http.StatusOK, // Should handle gracefully, not execute script
			description:    "Should handle XSS attempts safely",
		},
		{
			name: "SQLInjectionAttempt",
			setupRequest: func() *http.Request {
				body := `{"jsonrpc":"2.0","id":"1' OR '1'='1","method":"ping"}`
				req := httptest.NewRequest("POST", "/mcp", strings.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				return req
			},
			expectedStatus: http.StatusOK, // Should handle gracefully
			description:    "Should handle SQL injection attempts safely",
		},
		{
			name: "LargePayloadAttack",
			setupRequest: func() *http.Request {
				// This should be rejected due to size limits
				largePayload := strings.Repeat("A", int(s.config.MaxMessageSize)+1)
				body := fmt.Sprintf(`{"jsonrpc":"2.0","id":1,"method":"ping","data":"%s"}`, largePayload)
				req := httptest.NewRequest("POST", "/mcp", strings.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				return req
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Should reject oversized payloads",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			req := tt.setupRequest()
			w := httptest.NewRecorder()

			transport.handleMCPRequest(w, req)
			s.Equal(tt.expectedStatus, w.Code, tt.description)
		})
	}
}

// Error Handling Tests
func (s *TransportTestSuite) TestTransportErrorHandling() {
	// Test with a handler that returns errors
	handler := &mockMCPHandler{
		shouldError: true,
		err:         errMockHandler,
	}

	transport := NewHTTPTransport(s.logger, s.config, handler)

	req := httptest.NewRequest("POST", "/mcp", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"ping"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	transport.handleMCPRequest(w, req)

	// Should handle handler errors gracefully
	s.Equal(http.StatusInternalServerError, w.Code)
}

// Concurrent Access Tests
func (s *TransportTestSuite) TestConcurrentAccess() {
	handler := &mockMCPHandler{
		response: &types.MCPResponse{
			JSONRPC: "2.0",
			ID:      1,
			Result:  map[string]string{"status": "ok"},
		},
	}

	config := &TransportConfig{
		Type:           types.TransportHTTP,
		Host:           "localhost",
		Port:           0,
		MaxMessageSize: 1024 * 1024,
	}

	transport := NewHTTPTransport(s.logger, config, handler)
	ctx := context.Background()

	err := transport.Start(ctx)
	s.Require().NoError(err)
	defer func() {
		stopErr := transport.Stop(ctx)
		s.Require().NoError(stopErr)
	}()

	// Test concurrent health checks
	var wg sync.WaitGroup
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			healthy := transport.IsHealthy(ctx)
			s.True(healthy)
		}()
	}

	wg.Wait()
}

// Mock MCP Handler for testing
type mockMCPHandler struct {
	response    *types.MCPResponse
	shouldError bool
	err         error
	callCount   int
	mu          sync.Mutex
}

func (m *mockMCPHandler) HandleInitialize(_ context.Context, req *types.MCPRequest) (*types.MCPResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callCount++

	if m.shouldError {
		return nil, m.err
	}

	if m.response != nil {
		m.response.ID = req.ID
		return m.response, nil
	}

	return &types.MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  map[string]string{"method": "initialize"},
	}, nil
}

func (m *mockMCPHandler) HandlePing(_ context.Context, req *types.MCPRequest) (*types.MCPResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callCount++

	if m.shouldError {
		return nil, m.err
	}

	if m.response != nil {
		m.response.ID = req.ID
		return m.response, nil
	}

	return &types.MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  map[string]string{"status": "ok"},
	}, nil
}

func (m *mockMCPHandler) HandleToolsList(_ context.Context, req *types.MCPRequest) (*types.MCPResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callCount++

	if m.shouldError {
		return nil, m.err
	}

	if m.response != nil {
		m.response.ID = req.ID
		return m.response, nil
	}

	return &types.MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  types.ToolListResult{Tools: []Tool{}},
	}, nil
}

func (m *mockMCPHandler) HandleToolCall(_ context.Context, req *types.MCPRequest) (*types.MCPResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callCount++

	if m.shouldError {
		return nil, m.err
	}

	if m.response != nil {
		m.response.ID = req.ID
		return m.response, nil
	}

	return &types.MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  types.ToolCallResult{IsError: false, Content: []Content{{Type: "text", Text: "mock result"}}},
	}, nil
}

func (m *mockMCPHandler) GetCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.callCount
}

// HandleRequest implements the types.MCPHandler interface
func (m *mockMCPHandler) HandleRequest(ctx context.Context, req *types.MCPRequest) (*types.MCPResponse, error) {
	switch req.Method {
	case "initialize":
		return m.HandleInitialize(ctx, req)
	case "ping":
		return m.HandlePing(ctx, req)
	case "tools/list":
		return m.HandleToolsList(ctx, req)
	case "tools/call":
		return m.HandleToolCall(ctx, req)
	default:
		return nil, fmt.Errorf("%w: %s", errUnknownMethod, req.Method)
	}
}

// Benchmark Tests
func BenchmarkStdioTransportStart(b *testing.B) {
	logger := NewTestLogger()
	config := DefaultTransportConfig()
	transport := NewStdioTransport(logger, config)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := transport.Start(ctx)
		require.NoError(b, err)
		err = transport.Stop(ctx)
		require.NoError(b, err)

		// Reset transport for next iteration
		transport = NewStdioTransport(logger, config)
	}
}

func BenchmarkHTTPTransportHealthCheck(b *testing.B) {
	logger := NewTestLogger()
	config := DefaultTransportConfig()
	config.Type = types.TransportHTTP
	handler := &mockMCPHandler{}
	transport := NewHTTPTransport(logger, config, handler)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		transport.IsHealthy(ctx)
	}
}

func BenchmarkTransportMetrics(b *testing.B) {
	metrics := NewTransportMetrics()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metrics.RecordReceive(1)
		metrics.RecordSend(1)
		_ = metrics.MessagesReceived()
		_ = metrics.MessagesSent()
	}
}

// Integration Tests
func (s *TransportTestSuite) TestStdioTransportIntegration() {
	// This test simulates a full stdio transport lifecycle
	transport := NewStdioTransport(s.logger, s.config)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start transport
	err := transport.Start(ctx)
	s.Require().NoError(err)
	s.True(transport.IsHealthy(ctx))

	// Test metrics after start
	s.Positive(transport.metrics.Uptime())

	// Stop transport
	err = transport.Stop(ctx)
	s.Require().NoError(err)
	s.False(transport.IsHealthy(ctx))

	// Verify logging occurred
	s.True(s.logger.HasMessage("INFO", "stdio transport started"))
	s.True(s.logger.HasMessage("INFO", "stdio transport stopped"))
}

func (s *TransportTestSuite) TestHTTPTransportIntegration() {
	handler := &mockMCPHandler{
		response: &types.MCPResponse{
			JSONRPC: "2.0",
			ID:      1,
			Result:  map[string]string{"status": "ok"},
		},
	}

	config := &TransportConfig{
		Type:           types.TransportHTTP,
		Host:           "localhost",
		Port:           0, // Auto-assign
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		MaxMessageSize: 1024 * 1024,
	}

	transport := NewHTTPTransport(s.logger, config, handler)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Start transport
	err := transport.Start(ctx)
	s.Require().NoError(err)
	s.True(transport.IsHealthy(ctx))

	// Test that we can make HTTP requests to the transport
	// For this test, we're just verifying the transport lifecycle

	// Stop transport
	err = transport.Stop(ctx)
	s.Require().NoError(err)
	s.False(transport.IsHealthy(ctx))

	// Verify logging occurred
	s.True(s.logger.HasMessage("INFO", "starting HTTP transport"))
	s.True(s.logger.HasMessage("INFO", "HTTP transport started"))
	s.True(s.logger.HasMessage("INFO", "HTTP transport stopped"))
}

// Test transport type detection with various input scenarios
func (s *TransportTestSuite) TestDetectTransportEdgeCases() {
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Test with multiple flags
	os.Args = []string{"program", "--verbose", "--stdio", "--debug"}
	result := DetectTransport()
	s.Equal(types.TransportStdio, result)

	// Test with HTTP flag mixed with other flags
	os.Args = []string{"program", "--config", "test.json", "--http", "--verbose"}
	result = DetectTransport()
	s.Equal(types.TransportHTTP, result)

	// Test case sensitivity in environment variable
	err := os.Setenv("MCP_TRANSPORT", "HTTP")
	s.Require().NoError(err)
	os.Args = []string{"program"}
	result = DetectTransport()
	s.Equal(types.TransportType("http"), result) // Converted to lowercase

	err = os.Unsetenv("MCP_TRANSPORT")
	s.Require().NoError(err)
}
