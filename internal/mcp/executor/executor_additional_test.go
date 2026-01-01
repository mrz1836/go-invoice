package executor

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// ExecutorTestSuite tests the SecureExecutor implementation
type ExecutorTestSuite struct {
	suite.Suite

	executor    *SecureExecutor
	logger      *MockLogger
	validator   *MockCommandValidator
	fileHandler *MockFileHandler
	sandbox     SandboxConfig
}

func (suite *ExecutorTestSuite) SetupTest() {
	suite.logger = new(MockLogger)
	suite.validator = new(MockCommandValidator)
	suite.fileHandler = new(MockFileHandler)

	suite.sandbox = SandboxConfig{
		AllowedCommands:      []string{"echo", "go-invoice", "true"},
		AllowedPaths:         []string{"/tmp", "/home"},
		BlockedPaths:         []string{"/etc", "/root"},
		EnvironmentWhitelist: []string{"PATH", "HOME", "USER"},
		MaxExecutionTime:     30 * time.Second,
		MaxOutputSize:        1024 * 1024,
	}

	// Setup logger expectations
	suite.logger.On("Debug", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()
	suite.logger.On("Info", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()
	suite.logger.On("Warn", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()
	suite.logger.On("Error", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()

	suite.executor = NewSecureExecutor(suite.logger, suite.validator, suite.sandbox, suite.fileHandler)
}

// TestExecuteSuccess tests successful command execution
func (suite *ExecutorTestSuite) TestExecuteSuccess() {
	ctx := context.Background()

	suite.validator.On("ValidateCommand", ctx, "echo", []string{"hello"}).Return(nil)
	suite.validator.On("ValidateEnvironment", ctx, mock.Anything).Return(nil)

	req := &ExecutionRequest{
		Command: "echo",
		Args:    []string{"hello"},
		Timeout: 5 * time.Second,
	}

	resp, err := suite.executor.Execute(ctx, req)

	suite.Require().NoError(err)
	suite.NotNil(resp)
	suite.Equal(0, resp.ExitCode)
	suite.Contains(resp.Stdout, "hello")
}

// TestExecuteContextCancellation tests context cancellation
func (suite *ExecutorTestSuite) TestExecuteContextCancellation() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	req := &ExecutionRequest{
		Command: "echo",
		Args:    []string{"hello"},
	}

	resp, err := suite.executor.Execute(ctx, req)

	suite.Equal(context.Canceled, err)
	suite.Nil(resp)
}

// TestExecuteCommandValidationFailed tests command validation failure
func (suite *ExecutorTestSuite) TestExecuteCommandValidationFailed() {
	ctx := context.Background()

	suite.validator.On("ValidateCommand", ctx, "rm", []string{"-rf", "/"}).
		Return(ErrDangerousPattern)

	req := &ExecutionRequest{
		Command: "rm",
		Args:    []string{"-rf", "/"},
	}

	resp, err := suite.executor.Execute(ctx, req)

	suite.Require().Error(err)
	suite.Nil(resp)
	suite.Contains(err.Error(), "command validation failed")
}

// TestExecuteEnvironmentValidationFailed tests environment validation failure
func (suite *ExecutorTestSuite) TestExecuteEnvironmentValidationFailed() {
	ctx := context.Background()

	suite.validator.On("ValidateCommand", ctx, "echo", []string{"test"}).Return(nil)
	suite.validator.On("ValidateEnvironment", ctx, mock.Anything).Return(ErrEnvNotAllowed)

	req := &ExecutionRequest{
		Command:     "echo",
		Args:        []string{"test"},
		Environment: map[string]string{"MALICIOUS": "value"},
	}

	resp, err := suite.executor.Execute(ctx, req)

	suite.Require().Error(err)
	suite.Nil(resp)
	suite.Contains(err.Error(), "environment validation failed")
}

// TestExecuteWithWorkingDir tests execution with specified working directory
func (suite *ExecutorTestSuite) TestExecuteWithWorkingDir() {
	ctx := context.Background()
	tmpDir := os.TempDir()

	suite.validator.On("ValidateCommand", ctx, "echo", []string{"test"}).Return(nil)
	suite.validator.On("ValidateEnvironment", ctx, mock.Anything).Return(nil)
	suite.validator.On("ValidatePath", mock.Anything, tmpDir).Return(nil)

	req := &ExecutionRequest{
		Command:    "echo",
		Args:       []string{"test"},
		WorkingDir: tmpDir,
		Timeout:    5 * time.Second,
	}

	resp, err := suite.executor.Execute(ctx, req)

	suite.Require().NoError(err)
	suite.NotNil(resp)
	suite.Equal(0, resp.ExitCode)
}

// TestExecuteWithInvalidWorkingDir tests execution with invalid working directory
func (suite *ExecutorTestSuite) TestExecuteWithInvalidWorkingDir() {
	ctx := context.Background()

	suite.validator.On("ValidateCommand", ctx, "echo", []string{"test"}).Return(nil)
	suite.validator.On("ValidateEnvironment", ctx, mock.Anything).Return(nil)
	suite.validator.On("ValidatePath", mock.Anything, "/etc/secret").Return(ErrPathTraversal)

	req := &ExecutionRequest{
		Command:    "echo",
		Args:       []string{"test"},
		WorkingDir: "/etc/secret",
	}

	resp, err := suite.executor.Execute(ctx, req)

	suite.Require().Error(err)
	suite.Nil(resp)
}

// TestExecuteCommandNotAllowed tests execution of disallowed command
func (suite *ExecutorTestSuite) TestExecuteCommandNotAllowed() {
	ctx := context.Background()

	req := &ExecutionRequest{
		Command: "forbidden_command",
		Args:    []string{},
	}

	resp, err := suite.executor.Execute(ctx, req)

	suite.Require().Error(err)
	suite.Nil(resp)
	suite.Contains(err.Error(), "command not allowed")
}

// TestExecuteWithDefaultTimeout tests that default timeout is applied
func (suite *ExecutorTestSuite) TestExecuteWithDefaultTimeout() {
	ctx := context.Background()

	suite.validator.On("ValidateCommand", ctx, "true", mock.Anything).Return(nil)
	suite.validator.On("ValidateEnvironment", ctx, mock.Anything).Return(nil)

	req := &ExecutionRequest{
		Command: "true",
		Args:    nil,
		// No timeout specified - should use default
	}

	resp, err := suite.executor.Execute(ctx, req)

	suite.Require().NoError(err)
	suite.NotNil(resp)
	suite.Equal(0, resp.ExitCode)
}

// TestExecuteWithInputFiles tests execution with input files
func (suite *ExecutorTestSuite) TestExecuteWithInputFiles() {
	ctx := context.Background()
	tmpDir := os.TempDir()

	suite.validator.On("ValidateCommand", mock.Anything, "echo", []string{"test"}).Return(nil)
	suite.validator.On("ValidateEnvironment", mock.Anything, mock.Anything).Return(nil)
	suite.fileHandler.On("PrepareWorkspace", mock.Anything, mock.Anything).
		Return(tmpDir, func() {}, nil)

	req := &ExecutionRequest{
		Command: "echo",
		Args:    []string{"test"},
		InputFiles: []FileReference{
			{Path: "/tmp/input.txt", ContentType: "text/plain"},
		},
		Timeout: 5 * time.Second,
	}

	resp, err := suite.executor.Execute(ctx, req)

	suite.Require().NoError(err)
	suite.NotNil(resp)
}

// TestExecuteWorkspacePreparationFailed tests workspace prep failure
func (suite *ExecutorTestSuite) TestExecuteWorkspacePreparationFailed() {
	ctx := context.Background()

	suite.validator.On("ValidateCommand", mock.Anything, "echo", []string{"test"}).Return(nil)
	suite.validator.On("ValidateEnvironment", mock.Anything, mock.Anything).Return(nil)
	// Must return a valid cleanup function (even if empty) when returning an error
	suite.fileHandler.On("PrepareWorkspace", mock.Anything, mock.Anything).
		Return("", func() {}, ErrInvalidPath)

	req := &ExecutionRequest{
		Command: "echo",
		Args:    []string{"test"},
		InputFiles: []FileReference{
			{Path: "/tmp/input.txt", ContentType: "text/plain"},
		},
	}

	resp, err := suite.executor.Execute(ctx, req)

	suite.Require().Error(err)
	suite.Nil(resp)
	suite.Contains(err.Error(), "workspace preparation failed")
}

// TestExecuteCommandExitError tests handling of non-zero exit code
func (suite *ExecutorTestSuite) TestExecuteCommandExitError() {
	ctx := context.Background()

	// Add "false" to allowed commands temporarily
	suite.executor.allowedCmds["false"] = true

	suite.validator.On("ValidateCommand", ctx, "false", mock.Anything).Return(nil)
	suite.validator.On("ValidateEnvironment", ctx, mock.Anything).Return(nil)

	req := &ExecutionRequest{
		Command: "false", // Always exits with code 1
		Args:    nil,
		Timeout: 5 * time.Second,
	}

	resp, err := suite.executor.Execute(ctx, req)

	suite.Require().NoError(err) // No error returned, but exit code is non-zero
	suite.NotNil(resp)
	suite.NotEqual(0, resp.ExitCode)
}

// TestNewSecureExecutorPanicsWithNilLogger tests constructor panics
func (suite *ExecutorTestSuite) TestNewSecureExecutorPanicsWithNilLogger() {
	suite.Panics(func() {
		NewSecureExecutor(nil, suite.validator, suite.sandbox, suite.fileHandler)
	})
}

// TestNewSecureExecutorPanicsWithNilValidator tests constructor panics
func (suite *ExecutorTestSuite) TestNewSecureExecutorPanicsWithNilValidator() {
	suite.Panics(func() {
		NewSecureExecutor(suite.logger, nil, suite.sandbox, suite.fileHandler)
	})
}

// TestNewSecureExecutorPanicsWithNilFileHandler tests constructor panics
func (suite *ExecutorTestSuite) TestNewSecureExecutorPanicsWithNilFileHandler() {
	suite.Panics(func() {
		NewSecureExecutor(suite.logger, suite.validator, suite.sandbox, nil)
	})
}

// TestExecuteWithProgressCallback tests execution with progress callback
func (suite *ExecutorTestSuite) TestExecuteWithProgressCallback() {
	ctx := context.Background()

	suite.validator.On("ValidateCommand", ctx, "echo", []string{"test"}).Return(nil)
	suite.validator.On("ValidateEnvironment", ctx, mock.Anything).Return(nil)

	callback := func(_ *ProgressUpdate) {
		// Progress callback runs asynchronously
	}

	req := &ExecutionRequest{
		Command:          "echo",
		Args:             []string{"test"},
		Timeout:          2 * time.Second,
		ProgressCallback: callback,
	}

	resp, err := suite.executor.Execute(ctx, req)

	suite.Require().NoError(err)
	suite.NotNil(resp)
	// Progress callback runs asynchronously, so it may or may not have been called
	// depending on timing. We just verify execution completes successfully.
}

// TestValidateCommandWithPath tests validation of command with path
func (suite *ExecutorTestSuite) TestValidateCommandWithPath() {
	ctx := context.Background()

	suite.validator.On("ValidateCommand", ctx, "/usr/bin/echo", mock.Anything).Return(nil)

	err := suite.executor.ValidateCommand(ctx, "/usr/bin/echo", nil)

	suite.NoError(err)
}

// TestValidateCommandContextCanceled tests validation with canceled context
func (suite *ExecutorTestSuite) TestValidateCommandContextCanceled() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := suite.executor.ValidateCommand(ctx, "echo", nil)

	suite.Equal(context.Canceled, err)
}

// TestGetAllowedCommandsContextCanceled tests listing with canceled context
func (suite *ExecutorTestSuite) TestGetAllowedCommandsContextCanceled() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cmds, err := suite.executor.GetAllowedCommands(ctx)

	suite.Equal(context.Canceled, err)
	suite.Nil(cmds)
}

func TestExecutorTestSuite(t *testing.T) {
	suite.Run(t, new(ExecutorTestSuite))
}

// TestExecutorErrorVariables tests that error variables are properly defined
func TestExecutorErrorVariables(t *testing.T) {
	tests := []struct {
		err     error
		name    string
		wantOp  string
		wantMsg string
	}{
		{ErrCommandNotAllowed, "ErrCommandNotAllowed", "validate", "command not allowed"},
		{ErrInvalidPath, "ErrInvalidPath", "validate", "invalid path"},
		{ErrTimeout, "ErrTimeout", "execute", "command execution timeout"},
		{ErrOutputTooLarge, "ErrOutputTooLarge", "execute", "output exceeds maximum size"},
		{ErrInvalidWorkDir, "ErrInvalidWorkDir", "validate", "invalid working directory"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var execErr *Error
			if !errors.As(tt.err, &execErr) {
				t.Errorf("%s is not an *Error", tt.name)
				return
			}
			if execErr.Op != tt.wantOp {
				t.Errorf("%s.Op = %v, want %v", tt.name, execErr.Op, tt.wantOp)
			}
			if execErr.Msg != tt.wantMsg {
				t.Errorf("%s.Msg = %v, want %v", tt.name, execErr.Msg, tt.wantMsg)
			}
		})
	}
}
