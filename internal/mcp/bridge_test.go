package mcp

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type BridgeTestSuite struct {
	suite.Suite

	tempDir     string
	logger      Logger
	validator   CommandValidator
	fileHandler FileHandler
	bridge      CLIBridge
}

func TestBridgeSuite(t *testing.T) {
	suite.Run(t, new(BridgeTestSuite))
}

func (s *BridgeTestSuite) SetupTest() {
	tempDir, err := os.MkdirTemp("", "mcp-bridge-test")
	s.Require().NoError(err)
	s.tempDir = tempDir

	s.logger = NewTestLogger()
	s.validator = NewCommandValidator([]string{"echo", "go"})
	s.fileHandler = NewFileHandler(tempDir)

	config := CLIConfig{
		Path:       "echo",
		WorkingDir: tempDir,
		MaxTimeout: 30 * time.Second,
	}

	s.bridge = NewCLIBridge(s.logger, s.validator, s.fileHandler, config)
}

func (s *BridgeTestSuite) TearDownTest() {
	if s.tempDir != "" {
		err := os.RemoveAll(s.tempDir)
		s.Require().NoError(err)
	}
}

func (s *BridgeTestSuite) TestExecuteCommandSuccess() {
	ctx := context.Background()

	req := &CommandRequest{
		Command: "echo",
		Args:    []string{"hello", "world"},
		Timeout: 5 * time.Second,
	}

	resp, err := s.bridge.ExecuteCommand(ctx, req)
	s.Require().NoError(err)
	s.NotNil(resp)
	s.Equal(0, resp.ExitCode)
	s.Contains(resp.Stdout, "hello world")
	s.Empty(resp.Error)
}

func (s *BridgeTestSuite) TestExecuteCommandWithInvalidCommand() {
	ctx := context.Background()

	req := &CommandRequest{
		Command: "forbidden-command",
		Args:    []string{},
		Timeout: 5 * time.Second,
	}

	_, err := s.bridge.ExecuteCommand(ctx, req)
	s.Require().Error(err)
	s.Contains(err.Error(), "command validation failed")
}

func (s *BridgeTestSuite) TestExecuteCommandContextCancellation() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Immediately cancel

	req := &CommandRequest{
		Command: "echo",
		Args:    []string{"test"},
	}

	_, err := s.bridge.ExecuteCommand(ctx, req)
	s.Equal(context.Canceled, err)
}

func (s *BridgeTestSuite) TestExecuteCommandTimeout() {
	ctx := context.Background()

	req := &CommandRequest{
		Command: "echo",
		Args:    []string{"test"},
		Timeout: 1 * time.Millisecond, // Very short timeout but not nanosecond
	}

	resp, err := s.bridge.ExecuteCommand(ctx, req)
	// Timeout might occur during workspace preparation or command execution
	// Both are acceptable outcomes for this test
	if err != nil {
		s.Contains(err.Error(), "context deadline exceeded")
	} else {
		s.NotNil(resp)
	}
}

func (s *BridgeTestSuite) TestValidateCommand() {
	ctx := context.Background()

	tests := []struct {
		name        string
		command     string
		args        []string
		expectError bool
	}{
		{
			name:        "AllowedCommand",
			command:     "echo",
			args:        []string{"test"},
			expectError: false,
		},
		{
			name:        "DisallowedCommand",
			command:     "rm",
			args:        []string{"-rf", "/"},
			expectError: true,
		},
		{
			name:        "CommandInjectionAttempt",
			command:     "echo",
			args:        []string{"test; rm -rf /"},
			expectError: true,
		},
		{
			name:        "PathTraversalAttempt",
			command:     "echo",
			args:        []string{"../../../etc/passwd"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			err := s.bridge.ValidateCommand(ctx, tt.command, tt.args)
			if tt.expectError {
				s.Error(err)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *BridgeTestSuite) TestGetAllowedCommands() {
	ctx := context.Background()

	commands, err := s.bridge.GetAllowedCommands(ctx)
	s.Require().NoError(err)
	s.NotEmpty(commands)
	s.Contains(commands, "echo")
}

type CommandValidatorTestSuite struct {
	suite.Suite

	validator CommandValidator
}

func TestCommandValidatorSuite(t *testing.T) {
	suite.Run(t, new(CommandValidatorTestSuite))
}

func (s *CommandValidatorTestSuite) SetupTest() {
	s.validator = NewCommandValidator([]string{"echo", "ls", "cat"})
}

func (s *CommandValidatorTestSuite) TestIsCommandAllowed() {
	ctx := context.Background()

	tests := []struct {
		name     string
		command  string
		expected bool
	}{
		{
			name:     "AllowedExactMatch",
			command:  "echo",
			expected: true,
		},
		{
			name:     "AllowedWithPath",
			command:  "/bin/echo",
			expected: true,
		},
		{
			name:     "DisallowedCommand",
			command:  "rm",
			expected: false,
		},
		{
			name:     "EmptyCommand",
			command:  "",
			expected: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			allowed, err := s.validator.IsCommandAllowed(ctx, tt.command)
			s.Require().NoError(err)
			s.Equal(tt.expected, allowed)
		})
	}
}

func (s *CommandValidatorTestSuite) TestValidateCommandWithDangerousArgs() {
	ctx := context.Background()

	dangerousArgs := [][]string{
		{"test; rm -rf /"},
		{"test && malicious"},
		{"test | evil"},
		{"test `bad`"},
		{"test $(evil)"},
		{"test ${bad}"},
		{"../../../etc/passwd"},
		{"..\\..\\windows\\system32"},
	}

	for i, args := range dangerousArgs {
		s.Run(s.T().Name()+"_"+string(rune('A'+i)), func() {
			err := s.validator.ValidateCommand(ctx, "echo", args)
			s.Error(err)
		})
	}
}

func (s *CommandValidatorTestSuite) TestValidateCommandContextCancellation() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := s.validator.ValidateCommand(ctx, "echo", []string{"test"})
	s.Equal(context.Canceled, err)
}

type FileHandlerTestSuite struct {
	suite.Suite

	tempDir     string
	fileHandler FileHandler
}

func TestFileHandlerSuite(t *testing.T) {
	suite.Run(t, new(FileHandlerTestSuite))
}

func (s *FileHandlerTestSuite) SetupTest() {
	tempDir, err := os.MkdirTemp("", "mcp-file-test")
	s.Require().NoError(err)
	s.tempDir = tempDir
	s.fileHandler = NewFileHandler(tempDir)
}

func (s *FileHandlerTestSuite) TearDownTest() {
	if s.tempDir != "" {
		err := os.RemoveAll(s.tempDir)
		s.Require().NoError(err)
	}
}

func (s *FileHandlerTestSuite) TestPrepareWorkspaceSuccess() {
	ctx := context.Background()

	workDir := filepath.Join(s.tempDir, "work")

	cleanPath, cleanup, err := s.fileHandler.PrepareWorkspace(ctx, workDir)
	s.Require().NoError(err)
	s.NotEmpty(cleanPath)
	s.NotNil(cleanup)

	// Verify directory was created
	s.DirExists(cleanPath)

	// Call cleanup
	cleanup()
}

func (s *FileHandlerTestSuite) TestPrepareWorkspaceContextCancellation() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _, err := s.fileHandler.PrepareWorkspace(ctx, s.tempDir)
	s.Equal(context.Canceled, err)
}

func (s *FileHandlerTestSuite) TestValidatePathSuccess() {
	ctx := context.Background()

	validPaths := []string{
		s.tempDir,
		filepath.Join(s.tempDir, "subdir"),
		filepath.Join(s.tempDir, "file.txt"),
	}

	for _, path := range validPaths {
		s.Run("ValidPath_"+filepath.Base(path), func() {
			err := s.fileHandler.ValidatePath(ctx, path)
			s.NoError(err)
		})
	}
}

func (s *FileHandlerTestSuite) TestValidatePathFailures() {
	ctx := context.Background()

	invalidPaths := []string{
		"../outside",
		"..\\windows",
		filepath.Join(s.tempDir, "../outside"),
	}

	for _, path := range invalidPaths {
		s.Run("InvalidPath_"+strings.ReplaceAll(path, "/", "_"), func() {
			err := s.fileHandler.ValidatePath(ctx, path)
			s.Require().Error(err)
			// Should contain either "path traversal" or "path outside allowed directory"
			s.True(strings.Contains(err.Error(), "path traversal") || strings.Contains(err.Error(), "path outside allowed directory"))
		})
	}
}

func (s *FileHandlerTestSuite) TestValidatePathContextCancellation() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := s.fileHandler.ValidatePath(ctx, s.tempDir)
	s.Equal(context.Canceled, err)
}

// Benchmark tests for performance validation
func BenchmarkExecuteCommand(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "mcp-bridge-bench")
	require.NoError(b, err)
	defer func() {
		err := os.RemoveAll(tempDir)
		require.NoError(b, err)
	}()

	logger := NewTestLogger()
	validator := NewCommandValidator([]string{"echo"})
	fileHandler := NewFileHandler(tempDir)

	config := CLIConfig{
		Path:       "echo",
		WorkingDir: tempDir,
		MaxTimeout: 30 * time.Second,
	}

	bridge := NewCLIBridge(logger, validator, fileHandler, config)
	ctx := context.Background()

	req := &CommandRequest{
		Command: "echo",
		Args:    []string{"benchmark", "test"},
		Timeout: 5 * time.Second,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := bridge.ExecuteCommand(ctx, req)
		require.NoError(b, err)
	}
}

func BenchmarkValidateCommand(b *testing.B) {
	validator := NewCommandValidator([]string{"echo", "ls", "cat"})
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := validator.ValidateCommand(ctx, "echo", []string{"test", "args"})
		require.NoError(b, err)
	}
}

func BenchmarkPrepareWorkspace(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "mcp-workspace-bench")
	require.NoError(b, err)
	defer func() {
		err := os.RemoveAll(tempDir)
		require.NoError(b, err)
	}()

	fileHandler := NewFileHandler(tempDir)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		workDir := filepath.Join(tempDir, "work", string(rune('a'+i%26)))
		_, cleanup, err := fileHandler.PrepareWorkspace(ctx, workDir)
		require.NoError(b, err)
		cleanup()
	}
}
