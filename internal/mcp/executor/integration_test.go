package executor

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mrz/go-invoice/internal/mcp/tools"
	"github.com/mrz/go-invoice/internal/mcp/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

var (
	errTestCommandNotAllowed = errors.New("command not allowed")
	errTestExtracted         = errors.New("extracted error")
)

// MockInputValidator is a mock implementation of InputValidator for testing.
type MockInputValidator struct {
	mock.Mock
}

func (m *MockInputValidator) ValidateAgainstSchema(ctx context.Context, input, schema map[string]interface{}) error {
	args := m.Called(ctx, input, schema)
	return args.Error(0)
}

func (m *MockInputValidator) ValidateRequired(ctx context.Context, input map[string]interface{}, requiredFields []string) error {
	args := m.Called(ctx, input, requiredFields)
	return args.Error(0)
}

func (m *MockInputValidator) ValidateFormat(ctx context.Context, fieldName string, value interface{}, format string) error {
	args := m.Called(ctx, fieldName, value, format)
	return args.Error(0)
}

func (m *MockInputValidator) BuildValidationError(ctx context.Context, fieldPath, message string, suggestions []string) error {
	args := m.Called(ctx, fieldPath, message, suggestions)
	return args.Error(0)
}

// MockAuditLogger is a mock implementation of AuditLogger for testing.
type MockAuditLogger struct {
	mock.Mock
}

func (m *MockAuditLogger) LogCommandExecution(ctx context.Context, event *CommandAuditEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockAuditLogger) LogSecurityViolation(ctx context.Context, event *SecurityViolationEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockAuditLogger) LogAccessAttempt(ctx context.Context, event *AccessAuditEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockAuditLogger) Query(ctx context.Context, criteria *AuditCriteria) ([]*AuditEntry, error) {
	args := m.Called(ctx, criteria)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*AuditEntry), args.Error(1)
}

// MCPExecutorBridgeTestSuite tests the MCPExecutorBridge functionality.
type MCPExecutorBridgeTestSuite struct {
	suite.Suite

	logger        *MockLogger
	executor      *MockCommandExecutor
	parser        *MockOutputParser
	tracker       *MockProgressTracker
	fileHandler   *MockFileHandler
	toolRegistry  *tools.DefaultToolRegistry
	auditLogger   *MockAuditLogger
	mockValidator *MockInputValidator
}

func (s *MCPExecutorBridgeTestSuite) SetupTest() {
	s.logger = new(MockLogger)
	s.executor = new(MockCommandExecutor)
	s.parser = new(MockOutputParser)
	s.tracker = new(MockProgressTracker)
	s.fileHandler = new(MockFileHandler)
	s.auditLogger = new(MockAuditLogger)
	s.mockValidator = new(MockInputValidator)

	// Create a real tool registry with a mock validator and logger
	s.toolRegistry = tools.NewDefaultToolRegistry(s.mockValidator, s.logger)
}

func (s *MCPExecutorBridgeTestSuite) TearDownTest() {
	s.logger.AssertExpectations(s.T())
	s.executor.AssertExpectations(s.T())
	s.parser.AssertExpectations(s.T())
	s.tracker.AssertExpectations(s.T())
	s.fileHandler.AssertExpectations(s.T())
	s.auditLogger.AssertExpectations(s.T())
}

func (s *MCPExecutorBridgeTestSuite) TestNewMCPExecutorBridgePanicOnNilLogger() {
	s.Panics(func() {
		NewMCPExecutorBridge(nil, s.executor, s.parser, s.tracker, s.fileHandler, s.toolRegistry, s.auditLogger, nil, "test-cli")
	})
}

func (s *MCPExecutorBridgeTestSuite) TestNewMCPExecutorBridgePanicOnNilExecutor() {
	s.Panics(func() {
		NewMCPExecutorBridge(s.logger, nil, s.parser, s.tracker, s.fileHandler, s.toolRegistry, s.auditLogger, nil, "test-cli")
	})
}

func (s *MCPExecutorBridgeTestSuite) TestNewMCPExecutorBridgePanicOnNilParser() {
	s.Panics(func() {
		NewMCPExecutorBridge(s.logger, s.executor, nil, s.tracker, s.fileHandler, s.toolRegistry, s.auditLogger, nil, "test-cli")
	})
}

func (s *MCPExecutorBridgeTestSuite) TestNewMCPExecutorBridgePanicOnNilTracker() {
	s.Panics(func() {
		NewMCPExecutorBridge(s.logger, s.executor, s.parser, nil, s.fileHandler, s.toolRegistry, s.auditLogger, nil, "test-cli")
	})
}

func (s *MCPExecutorBridgeTestSuite) TestNewMCPExecutorBridgePanicOnNilFileHandler() {
	s.Panics(func() {
		NewMCPExecutorBridge(s.logger, s.executor, s.parser, s.tracker, nil, s.toolRegistry, s.auditLogger, nil, "test-cli")
	})
}

func (s *MCPExecutorBridgeTestSuite) TestNewMCPExecutorBridgePanicOnNilToolRegistry() {
	s.Panics(func() {
		NewMCPExecutorBridge(s.logger, s.executor, s.parser, s.tracker, s.fileHandler, nil, s.auditLogger, nil, "test-cli")
	})
}

func (s *MCPExecutorBridgeTestSuite) TestNewMCPExecutorBridgeSuccessWithNilConfig() {
	// Should not panic - config will use defaults
	bridge := NewMCPExecutorBridge(s.logger, s.executor, s.parser, s.tracker, s.fileHandler, s.toolRegistry, nil, nil, "test-cli")
	s.NotNil(bridge)
	s.NotNil(bridge.securityConfig)
}

func (s *MCPExecutorBridgeTestSuite) TestNewMCPExecutorBridgeSuccessWithConfig() {
	config := DefaultSecurityConfig()
	bridge := NewMCPExecutorBridge(s.logger, s.executor, s.parser, s.tracker, s.fileHandler, s.toolRegistry, s.auditLogger, config, "test-cli")
	s.NotNil(bridge)
	s.Equal(config, bridge.securityConfig)
}

func (s *MCPExecutorBridgeTestSuite) TestExecuteCommandContextCanceled() {
	bridge := NewMCPExecutorBridge(s.logger, s.executor, s.parser, s.tracker, s.fileHandler, s.toolRegistry, nil, nil, "test-cli")

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	req := &types.CommandRequest{
		Command: "echo",
		Args:    []string{"test"},
	}

	resp, err := bridge.ExecuteCommand(ctx, req)
	s.Nil(resp)
	s.ErrorIs(err, context.Canceled)
}

func (s *MCPExecutorBridgeTestSuite) TestExecuteCommandSuccess() {
	config := &SecurityConfig{AuditEnabled: false}
	bridge := NewMCPExecutorBridge(s.logger, s.executor, s.parser, s.tracker, s.fileHandler, s.toolRegistry, nil, config, "test-cli")

	ctx := context.Background()
	req := &types.CommandRequest{
		Command:    "echo",
		Args:       []string{"hello"},
		WorkingDir: "/tmp",
		Timeout:    30 * time.Second,
	}

	execResp := &ExecutionResponse{
		ExitCode: 0,
		Stdout:   "hello\n",
		Stderr:   "",
		Duration: 100 * time.Millisecond,
		OutputFiles: []FileReference{
			{Path: "/tmp/output.txt", Size: 100},
		},
	}

	s.executor.On("Execute", ctx, mock.AnythingOfType("*executor.ExecutionRequest")).Return(execResp, nil).Once()

	resp, err := bridge.ExecuteCommand(ctx, req)
	s.Require().NoError(err)
	s.NotNil(resp)
	s.Equal(0, resp.ExitCode)
	s.Equal("hello\n", resp.Stdout)
	s.Len(resp.Files, 1)
	s.Equal("/tmp/output.txt", resp.Files[0])
}

func (s *MCPExecutorBridgeTestSuite) TestExecuteCommandWithInputFiles() {
	config := &SecurityConfig{AuditEnabled: false}
	bridge := NewMCPExecutorBridge(s.logger, s.executor, s.parser, s.tracker, s.fileHandler, s.toolRegistry, nil, config, "test-cli")

	ctx := context.Background()
	req := &types.CommandRequest{
		Command:    "cat",
		Args:       []string{},
		InputFiles: []string{"/tmp/input.txt"},
	}

	execResp := &ExecutionResponse{
		ExitCode: 0,
		Stdout:   "file contents",
	}

	s.executor.On("Execute", ctx, mock.MatchedBy(func(r *ExecutionRequest) bool {
		return len(r.InputFiles) == 1 && r.InputFiles[0].Path == "/tmp/input.txt"
	})).Return(execResp, nil).Once()

	resp, err := bridge.ExecuteCommand(ctx, req)
	s.Require().NoError(err)
	s.NotNil(resp)
	s.Equal("file contents", resp.Stdout)
}

func (s *MCPExecutorBridgeTestSuite) TestExecuteCommandExecutorError() {
	config := &SecurityConfig{AuditEnabled: false}
	bridge := NewMCPExecutorBridge(s.logger, s.executor, s.parser, s.tracker, s.fileHandler, s.toolRegistry, nil, config, "test-cli")

	ctx := context.Background()
	req := &types.CommandRequest{
		Command: "invalid",
	}

	s.executor.On("Execute", ctx, mock.AnythingOfType("*executor.ExecutionRequest")).Return((*ExecutionResponse)(nil), errTestCommandNotAllowed).Once()

	resp, err := bridge.ExecuteCommand(ctx, req)
	s.Nil(resp)
	s.Require().Error(err)
	s.Contains(err.Error(), "command execution failed")
}

func (s *MCPExecutorBridgeTestSuite) TestValidateFileReturnsNil() {
	bridge := NewMCPExecutorBridge(s.logger, s.executor, s.parser, s.tracker, s.fileHandler, s.toolRegistry, nil, nil, "test-cli")

	err := bridge.ValidateFile(context.Background(), "/any/path")
	s.NoError(err)
}

func (s *MCPExecutorBridgeTestSuite) TestPrepareWorkspaceReturnsEmptyValues() {
	bridge := NewMCPExecutorBridge(s.logger, s.executor, s.parser, s.tracker, s.fileHandler, s.toolRegistry, nil, nil, "test-cli")

	workDir, cleanup, err := bridge.PrepareWorkspace(context.Background(), "/any/dir")
	s.Require().NoError(err)
	s.Empty(workDir)
	s.NotNil(cleanup)
	cleanup() // Should not panic
}

func TestMCPExecutorBridgeTestSuite(t *testing.T) {
	suite.Run(t, new(MCPExecutorBridgeTestSuite))
}

// ToolCallHandlerTestSuite tests the ToolCallHandler functionality.
type ToolCallHandlerTestSuite struct {
	suite.Suite

	logger        *MockLogger
	executor      *MockCommandExecutor
	parser        *MockOutputParser
	tracker       *MockProgressTracker
	fileHandler   *MockFileHandler
	toolRegistry  *tools.DefaultToolRegistry
	mockValidator *MockInputValidator
	bridge        *MCPExecutorBridge
}

func (s *ToolCallHandlerTestSuite) SetupTest() {
	s.logger = new(MockLogger)
	s.executor = new(MockCommandExecutor)
	s.parser = new(MockOutputParser)
	s.tracker = new(MockProgressTracker)
	s.fileHandler = new(MockFileHandler)
	s.mockValidator = new(MockInputValidator)

	// Create a real tool registry
	s.toolRegistry = tools.NewDefaultToolRegistry(s.mockValidator, s.logger)

	s.bridge = NewMCPExecutorBridge(s.logger, s.executor, s.parser, s.tracker, s.fileHandler, s.toolRegistry, nil, nil, "test-cli")
}

func (s *ToolCallHandlerTestSuite) TearDownTest() {
	s.logger.AssertExpectations(s.T())
	s.parser.AssertExpectations(s.T())
	s.tracker.AssertExpectations(s.T())
}

func (s *ToolCallHandlerTestSuite) TestNewToolCallHandlerPanicOnNilLogger() {
	s.Panics(func() {
		NewToolCallHandler(nil, s.bridge, s.toolRegistry, s.parser, s.tracker)
	})
}

func (s *ToolCallHandlerTestSuite) TestNewToolCallHandlerPanicOnNilBridge() {
	s.Panics(func() {
		NewToolCallHandler(s.logger, nil, s.toolRegistry, s.parser, s.tracker)
	})
}

func (s *ToolCallHandlerTestSuite) TestNewToolCallHandlerPanicOnNilToolRegistry() {
	s.Panics(func() {
		NewToolCallHandler(s.logger, s.bridge, nil, s.parser, s.tracker)
	})
}

func (s *ToolCallHandlerTestSuite) TestNewToolCallHandlerPanicOnNilParser() {
	s.Panics(func() {
		NewToolCallHandler(s.logger, s.bridge, s.toolRegistry, nil, s.tracker)
	})
}

func (s *ToolCallHandlerTestSuite) TestNewToolCallHandlerPanicOnNilTracker() {
	s.Panics(func() {
		NewToolCallHandler(s.logger, s.bridge, s.toolRegistry, s.parser, nil)
	})
}

func (s *ToolCallHandlerTestSuite) TestNewToolCallHandlerSuccess() {
	handler := NewToolCallHandler(s.logger, s.bridge, s.toolRegistry, s.parser, s.tracker)
	s.NotNil(handler)
	s.Equal(s.logger, handler.logger)
	s.Equal(s.bridge, handler.bridge)
	s.Equal(s.toolRegistry, handler.toolRegistry)
}

func (s *ToolCallHandlerTestSuite) TestHandleToolCallContextCanceled() {
	handler := NewToolCallHandler(s.logger, s.bridge, s.toolRegistry, s.parser, s.tracker)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	req := &types.MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":      "test_tool",
			"arguments": map[string]interface{}{},
		},
	}

	resp, err := handler.HandleToolCall(ctx, req)
	s.Nil(resp)
	s.ErrorIs(err, context.Canceled)
}

func (s *ToolCallHandlerTestSuite) TestHandleToolCallInvalidParams() {
	handler := NewToolCallHandler(s.logger, s.bridge, s.toolRegistry, s.parser, s.tracker)

	ctx := context.Background()
	req := &types.MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params:  make(chan int), // Cannot be marshaled to JSON
	}

	resp, err := handler.HandleToolCall(ctx, req)
	s.Nil(resp)
	s.Require().Error(err)
	s.Contains(err.Error(), "failed to parse tool call params")
}

func (s *ToolCallHandlerTestSuite) TestHandleToolCallUnknownTool() {
	handler := NewToolCallHandler(s.logger, s.bridge, s.toolRegistry, s.parser, s.tracker)

	ctx := context.Background()
	req := &types.MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":      "unknown_tool",
			"arguments": map[string]interface{}{},
		},
	}

	// Use mock.AnythingOfType to match variadic args
	s.logger.On("Debug", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()

	resp, err := handler.HandleToolCall(ctx, req)
	s.Require().Error(err) // Tool not found returns error
	s.NotNil(resp)
	s.NotNil(resp.Error)
	s.Equal(-32602, resp.Error.Code)
	s.Contains(resp.Error.Data.(string), "Unknown tool")
}

func TestToolCallHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(ToolCallHandlerTestSuite))
}

// TestConvertParams tests the convertParams helper function.
func TestConvertParams(t *testing.T) {
	t.Run("ValidConversion", func(t *testing.T) {
		from := map[string]interface{}{
			"name": "test_tool",
			"arguments": map[string]interface{}{
				"key": "value",
			},
		}

		var to types.ToolCallParams
		err := convertParams(from, &to)
		require.NoError(t, err)
		assert.Equal(t, "test_tool", to.Name)
		assert.Equal(t, "value", to.Arguments["key"])
	})

	t.Run("EmptyInput", func(t *testing.T) {
		from := map[string]interface{}{}

		var to types.ToolCallParams
		err := convertParams(from, &to)
		require.NoError(t, err)
		assert.Empty(t, to.Name)
	})

	t.Run("NilArguments", func(t *testing.T) {
		from := map[string]interface{}{
			"name": "test_tool",
		}

		var to types.ToolCallParams
		err := convertParams(from, &to)
		require.NoError(t, err)
		assert.Nil(t, to.Arguments)
	})

	t.Run("InvalidInputUnmarshalable", func(t *testing.T) {
		// Create a type that can't be marshaled
		from := make(chan int)

		var to types.ToolCallParams
		err := convertParams(from, &to)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to marshal params")
	})

	t.Run("ComplexNestedStructure", func(t *testing.T) {
		from := map[string]interface{}{
			"name": "complex_tool",
			"arguments": map[string]interface{}{
				"nested": map[string]interface{}{
					"level2": "deep_value",
				},
				"array": []interface{}{"a", "b", "c"},
			},
		}

		var to types.ToolCallParams
		err := convertParams(from, &to)
		require.NoError(t, err)
		assert.Equal(t, "complex_tool", to.Name)
		assert.NotNil(t, to.Arguments["nested"])
		assert.NotNil(t, to.Arguments["array"])
	})
}

// TestParseToolOutput tests the parseToolOutput method.
func TestParseToolOutput(t *testing.T) {
	logger := new(MockLogger)
	parser := new(MockOutputParser)
	tracker := new(MockProgressTracker)
	fileHandler := new(MockFileHandler)
	executor := new(MockCommandExecutor)

	mockValidator := new(MockInputValidator)
	toolRegistry := tools.NewDefaultToolRegistry(mockValidator, logger)

	bridge := NewMCPExecutorBridge(logger, executor, parser, tracker, fileHandler, toolRegistry, nil, nil, "test-cli")
	handler := NewToolCallHandler(logger, bridge, toolRegistry, parser, tracker)

	testTool := &tools.MCPTool{
		Name:        "test_tool",
		Description: "Test tool",
	}

	t.Run("SuccessWithStdout", func(t *testing.T) {
		resp := &ExecutionResponse{
			ExitCode: 0,
			Stdout:   "output text",
		}

		content, err := handler.parseToolOutput(context.Background(), testTool, resp)
		require.NoError(t, err)
		require.Len(t, content, 1)
		assert.Equal(t, "text", content[0].Type)
		assert.Equal(t, "output text", content[0].Text)
	})

	t.Run("SuccessWithEmptyOutput", func(t *testing.T) {
		resp := &ExecutionResponse{
			ExitCode: 0,
			Stdout:   "",
			Stderr:   "",
		}

		content, err := handler.parseToolOutput(context.Background(), testTool, resp)
		require.NoError(t, err)
		assert.Empty(t, content)
	})

	t.Run("ErrorWithStderr", func(t *testing.T) {
		resp := &ExecutionResponse{
			ExitCode: 1,
			Stdout:   "",
			Stderr:   "error message",
		}

		parser.On("ExtractError", mock.Anything, "", "error message", 1).Return(errTestExtracted).Once()

		content, err := handler.parseToolOutput(context.Background(), testTool, resp)
		require.NoError(t, err)
		require.Len(t, content, 1)
		assert.Contains(t, content[0].Text, "Error:")
		parser.AssertExpectations(t)
	})

	t.Run("ErrorWithNilExtraction", func(t *testing.T) {
		resp := &ExecutionResponse{
			ExitCode: 1,
			Stdout:   "some output",
			Stderr:   "error output",
		}

		parser.On("ExtractError", mock.Anything, "some output", "error output", 1).Return(nil).Once()

		content, err := handler.parseToolOutput(context.Background(), testTool, resp)
		require.NoError(t, err)
		// Should have stdout and stderr content
		assert.GreaterOrEqual(t, len(content), 1)
		parser.AssertExpectations(t)
	})

	t.Run("WithOutputFiles", func(t *testing.T) {
		resp := &ExecutionResponse{
			ExitCode: 0,
			Stdout:   "done",
			OutputFiles: []FileReference{
				{Path: "/tmp/output.html", Size: 1024},
				{Path: "/tmp/output.pdf", Size: 2048},
			},
		}

		content, err := handler.parseToolOutput(context.Background(), testTool, resp)
		require.NoError(t, err)
		require.Len(t, content, 3)
		assert.Contains(t, content[1].Text, "output.html")
		assert.Contains(t, content[1].Text, "1024 bytes")
		assert.Contains(t, content[2].Text, "output.pdf")
	})

	t.Run("ContextCanceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		resp := &ExecutionResponse{
			ExitCode: 0,
			Stdout:   "output",
		}

		content, err := handler.parseToolOutput(ctx, testTool, resp)
		require.Error(t, err)
		require.ErrorIs(t, err, context.Canceled)
		assert.Nil(t, content)
	})
}
