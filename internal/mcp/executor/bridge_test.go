package executor

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// CLIBridgeTestSuite tests the CLIBridge implementation
type CLIBridgeTestSuite struct {
	suite.Suite

	bridge      *CLIBridge
	logger      *MockLogger
	executor    *MockCommandExecutor
	fileHandler *MockFileHandler
}

func (suite *CLIBridgeTestSuite) SetupTest() {
	suite.logger = new(MockLogger)
	suite.executor = new(MockCommandExecutor)
	suite.fileHandler = new(MockFileHandler)

	// Setup logger expectations
	suite.logger.On("Debug", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()
	suite.logger.On("Info", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()
	suite.logger.On("Warn", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()
	suite.logger.On("Error", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()

	suite.bridge = NewCLIBridge(suite.logger, suite.executor, suite.fileHandler, "")
}

// TestNewCLIBridge tests the constructor
func (suite *CLIBridgeTestSuite) TestNewCLIBridge() {
	bridge := NewCLIBridge(suite.logger, suite.executor, suite.fileHandler, "")
	suite.NotNil(bridge)
	suite.Equal(suite.logger, bridge.logger)
	suite.Equal(suite.executor, bridge.executor)
	suite.Equal(suite.fileHandler, bridge.fileHandler)
	suite.Equal("go-invoice", bridge.cliPath)
}

// TestNewCLIBridgeWithCustomPath tests constructor with custom CLI path
func (suite *CLIBridgeTestSuite) TestNewCLIBridgeWithCustomPath() {
	bridge := NewCLIBridge(suite.logger, suite.executor, suite.fileHandler, "/usr/local/bin/go-invoice")
	suite.Equal("/usr/local/bin/go-invoice", bridge.cliPath)
}

// TestNewCLIBridgePanicsWithNilLogger tests constructor panics without logger
func (suite *CLIBridgeTestSuite) TestNewCLIBridgePanicsWithNilLogger() {
	suite.Panics(func() {
		NewCLIBridge(nil, suite.executor, suite.fileHandler, "")
	})
}

// TestNewCLIBridgePanicsWithNilExecutor tests constructor panics without executor
func (suite *CLIBridgeTestSuite) TestNewCLIBridgePanicsWithNilExecutor() {
	suite.Panics(func() {
		NewCLIBridge(suite.logger, nil, suite.fileHandler, "")
	})
}

// TestNewCLIBridgePanicsWithNilFileHandler tests constructor panics without file handler
func (suite *CLIBridgeTestSuite) TestNewCLIBridgePanicsWithNilFileHandler() {
	suite.Panics(func() {
		NewCLIBridge(suite.logger, suite.executor, nil, "")
	})
}

// TestExecuteToolCommandToolNotFound tests error for unknown tool
func (suite *CLIBridgeTestSuite) TestExecuteToolCommandToolNotFound() {
	ctx := context.Background()

	result, err := suite.bridge.ExecuteToolCommand(ctx, "unknown_tool", nil)

	suite.Require().ErrorIs(err, ErrToolNotFound)
	suite.Nil(result)
}

// TestExecuteToolCommandContextCancellation tests context cancellation
func (suite *CLIBridgeTestSuite) TestExecuteToolCommandContextCancellation() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result, err := suite.bridge.ExecuteToolCommand(ctx, "invoice_list", nil)

	suite.Equal(context.Canceled, err)
	suite.Nil(result)
}

// TestBridgeErrorError tests BridgeError.Error() method
func (suite *CLIBridgeTestSuite) TestBridgeErrorError() {
	err := &BridgeError{Op: "test", Msg: "test message"}
	suite.Equal("bridge test: test message", err.Error())
}

func TestCLIBridgeTestSuite(t *testing.T) {
	suite.Run(t, new(CLIBridgeTestSuite))
}

// TestGetFloatValue tests the getFloatValue helper
func TestGetFloatValue(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected float64
		ok       bool
	}{
		{"Float64", float64(3.14), 3.14, true},
		{"Float32", float32(2.5), 2.5, true},
		{"Int", int(42), 42.0, true},
		{"Int64", int64(100), 100.0, true},
		{"StringFloat", "3.14", 3.14, true},
		{"StringInt", "42", 42.0, true},
		{"EmptyString", "", 0, false},
		{"WhitespaceString", "   ", 0, false},
		{"InvalidString", "not a number", 0, false},
		{"NilValue", nil, 0, false},
		{"BoolValue", true, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := getFloatValue(tt.input)
			if ok != tt.ok {
				t.Errorf("getFloatValue(%v) ok = %v, want %v", tt.input, ok, tt.ok)
			}
			if ok && result != tt.expected {
				t.Errorf("getFloatValue(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestGetIntValue tests the getIntValue helper
func TestGetIntValue(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected int
		ok       bool
	}{
		{"Int", int(42), 42, true},
		{"Int64", int64(100), 100, true},
		{"Float64", float64(3.9), 3, true}, // Truncated
		{"StringInt", "42", 42, true},
		{"EmptyString", "", 0, false},
		{"WhitespaceString", "   ", 0, false},
		{"InvalidString", "not a number", 0, false},
		{"NilValue", nil, 0, false},
		{"BoolValue", true, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := getIntValue(tt.input)
			if ok != tt.ok {
				t.Errorf("getIntValue(%v) ok = %v, want %v", tt.input, ok, tt.ok)
			}
			if ok && result != tt.expected {
				t.Errorf("getIntValue(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestBridgeErrorVariables tests that error variables are properly defined
func TestBridgeErrorVariables(t *testing.T) {
	tests := []struct {
		err     error
		name    string
		wantOp  string
		wantMsg string
	}{
		{ErrToolNotFound, "ErrToolNotFound", "lookup", "tool not found"},
		{ErrInvalidToolInput, "ErrInvalidToolInput", "validate", "invalid tool input"},
		{ErrMissingRequired, "ErrMissingRequired", "validate", "missing required parameter"},
		{ErrCommandBuildFailed, "ErrCommandBuildFailed", "build", "failed to build command"},
		{ErrMissingUpdateFields, "ErrMissingUpdateFields", "validate", "at least one field to update must be provided"},
		{ErrMissingItemIdentifier, "ErrMissingItemIdentifier", "validate", "either item_id or item_index must be provided"},
		{ErrCommandFailed, "ErrCommandFailed", "execute", "command execution failed"},
		{ErrCollectionFailed, "ErrCollectionFailed", "collect", "file collection failed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bridgeErr *BridgeError
			if !errors.As(tt.err, &bridgeErr) {
				t.Errorf("%s is not a *BridgeError", tt.name)
				return
			}
			if bridgeErr.Op != tt.wantOp {
				t.Errorf("%s.Op = %v, want %v", tt.name, bridgeErr.Op, tt.wantOp)
			}
			if bridgeErr.Msg != tt.wantMsg {
				t.Errorf("%s.Msg = %v, want %v", tt.name, bridgeErr.Msg, tt.wantMsg)
			}
		})
	}
}
