package executor

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestSecureExecutor_BasicFunctionality tests basic functionality without complex mocking
func TestSecureExecutor_BasicFunctionality(t *testing.T) {
	// Create a simple secure executor
	logger := new(MockLogger)
	validator := new(MockCommandValidator)
	fileHandler := new(MockFileHandler)

	sandbox := SandboxConfig{
		AllowedCommands:      []string{"echo", "ls", "cat"},
		MaxExecutionTime:     30000000000, // 30 seconds
		MaxOutputSize:        1024 * 1024, // 1MB
		MaxFileSize:          1024 * 1024, // 1MB
		EnvironmentWhitelist: []string{"PATH", "HOME"},
	}

	executor := NewSecureExecutor(logger, validator, sandbox, fileHandler)

	t.Run("Constructor", func(t *testing.T) {
		require.NotNil(t, executor)
		assert.NotNil(t, executor.allowedCmds)
		assert.True(t, executor.allowedCmds["echo"])
		assert.True(t, executor.allowedCmds["ls"])
		assert.True(t, executor.allowedCmds["cat"])
		assert.False(t, executor.allowedCmds["rm"])
	})

	t.Run("GetAllowedCommands", func(t *testing.T) {
		commands, err := executor.GetAllowedCommands(context.Background())
		require.NoError(t, err)
		assert.Len(t, commands, 3)
		assert.Contains(t, commands, "echo")
		assert.Contains(t, commands, "ls")
		assert.Contains(t, commands, "cat")
	})

	t.Run("ValidateCommand_Allowed", func(t *testing.T) {
		validator.On("ValidateCommand", context.Background(), "echo", []string{"test"}).Return(nil).Once()
		err := executor.ValidateCommand(context.Background(), "echo", []string{"test"})
		assert.NoError(t, err)
	})

	t.Run("ValidateCommand_NotAllowed", func(t *testing.T) {
		err := executor.ValidateCommand(context.Background(), "rm", []string{"-rf", "/"})
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrCommandNotAllowed)
	})

	t.Run("BuildEnvironment", func(t *testing.T) {
		env := executor.buildEnvironment(map[string]string{
			"HOME":      "/home/test",
			"DANGEROUS": "value", // Should be filtered out
		})

		// Convert to map for easier checking
		envMap := make(map[string]string)
		for _, e := range env {
			if len(e) > 0 && e[0] != '=' {
				parts := []string{"", ""}
				if idx := findEqual(e); idx != -1 {
					parts[0] = e[:idx]
					parts[1] = e[idx+1:]
				}
				if len(parts) == 2 && parts[0] != "" {
					envMap[parts[0]] = parts[1]
				}
			}
		}

		assert.Equal(t, "/home/test", envMap["HOME"])
		assert.NotContains(t, envMap, "DANGEROUS")
	})
}

// Helper function to find the equal sign in env string
func findEqual(s string) int {
	for i, r := range s {
		if r == '=' {
			return i
		}
	}
	return -1
}

// TestDefaultOutputParser_BasicFunctionality tests basic parsing functionality
func TestDefaultOutputParser_BasicFunctionality(t *testing.T) {
	logger := new(MockLogger)
	parser := NewDefaultOutputParser(logger)

	t.Run("Constructor", func(t *testing.T) {
		require.NotNil(t, parser)
		assert.Equal(t, logger, parser.logger)
	})

	t.Run("ParseJSON_Valid", func(t *testing.T) {
		logger.On("Debug", "JSON parsed successfully", "keys", 2).Once()

		result, err := parser.ParseJSON(context.Background(), `{"name": "John", "age": 30}`)
		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "John", result["name"])
		assert.InDelta(t, 30.0, result["age"], 0.001)
	})

	t.Run("ParseJSON_Invalid", func(t *testing.T) {
		logger.On("Debug", "JSON parsing failed", "error", mock.AnythingOfType("*json.SyntaxError"), "outputLen", 13).Once()

		result, err := parser.ParseJSON(context.Background(), `{"invalid": }`)
		require.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrInvalidJSON)
	})

	t.Run("ParseKeyValue_Valid", func(t *testing.T) {
		logger.On("Debug", "key-value pairs parsed", "count", 2).Once()

		result, err := parser.ParseKeyValue(context.Background(), "name: John\nage: 30")
		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "John", result["name"])
		assert.Equal(t, "30", result["age"])
	})

	t.Run("ExtractError_Success", func(t *testing.T) {
		err := parser.ExtractError(context.Background(), "success", "", 0)
		assert.NoError(t, err)
	})

	t.Run("ExtractError_WithError", func(t *testing.T) {
		err := parser.ExtractError(context.Background(), "", "Error: something failed", 1)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrCommandFailed)
		assert.Contains(t, err.Error(), "something failed")
	})
}

// TestCLIBridge_BasicFunctionality tests basic bridge functionality
func TestCLIBridge_BasicFunctionality(t *testing.T) {
	logger := new(MockLogger)
	executor := new(MockCommandExecutor)
	fileHandler := new(MockFileHandler)

	bridge := NewCLIBridge(logger, executor, fileHandler, "test-cli")

	t.Run("Constructor", func(t *testing.T) {
		require.NotNil(t, bridge)
		assert.Equal(t, "test-cli", bridge.cliPath)
		assert.NotEmpty(t, bridge.toolCommands)

		// Check that expected tools are registered
		expectedTools := []string{
			"invoice_create", "invoice_list", "invoice_show",
			"client_create", "client_list", "client_show",
			"import_csv", "generate_html", "config_show",
		}

		for _, tool := range expectedTools {
			assert.Contains(t, bridge.toolCommands, tool, "Tool %s should be registered", tool)
		}
	})

	t.Run("ToolNotFound", func(t *testing.T) {
		resp, err := bridge.ExecuteToolCommand(context.Background(), "nonexistent_tool", map[string]interface{}{})
		require.Error(t, err)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrToolNotFound)
	})
}

// TestSecurityConfig_BasicFunctionality tests security configuration
func TestSecurityConfig_BasicFunctionality(t *testing.T) {
	t.Run("DefaultConfig", func(t *testing.T) {
		config := DefaultSecurityConfig()
		require.NotNil(t, config)
		assert.True(t, config.AuditEnabled)
		assert.True(t, config.StrictMode)
		assert.NotEmpty(t, config.Sandbox.AllowedCommands)
		assert.Contains(t, config.Sandbox.AllowedCommands, "go-invoice")
		assert.NotEmpty(t, config.Sandbox.EnvironmentWhitelist)
		assert.Contains(t, config.Sandbox.EnvironmentWhitelist, "PATH")
	})

	t.Run("ConfigBuilder", func(t *testing.T) {
		builder := NewSecurityConfigBuilder()
		config, err := builder.
			WithStrictMode(false).
			WithMaxConcurrentOps(10).
			Build()

		require.NoError(t, err)
		assert.False(t, config.StrictMode)
		assert.Equal(t, 10, config.MaxConcurrentOps)
	})

	t.Run("ConfigValidation", func(t *testing.T) {
		builder := NewSecurityConfigBuilder()
		config, err := builder.
			WithMaxConcurrentOps(-1). // Invalid
			Build()

		require.Error(t, err)
		assert.Nil(t, config)
		assert.ErrorIs(t, err, ErrInvalidConfig)
	})
}

// TestProgressTracker_BasicFunctionality tests progress tracking
func TestProgressTracker_BasicFunctionality(t *testing.T) {
	logger := new(MockLogger)
	tracker := NewDefaultProgressTracker(logger)

	t.Run("Constructor", func(t *testing.T) {
		require.NotNil(t, tracker)
		assert.Equal(t, logger, tracker.logger)
	})

	t.Run("StartOperation", func(t *testing.T) {
		logger.On("Info", "operation started",
			"operationID", "test-op",
			"description", "Test operation",
			"totalSteps", 5).Once()

		ctx := context.Background()
		op, err := tracker.StartOperation(ctx, "test-op", "Test operation", 5)
		require.NoError(t, err)
		assert.Equal(t, "test-op", op.ID)
		assert.Equal(t, "Test operation", op.Description)
		assert.Equal(t, 5, op.TotalSteps)
		assert.False(t, op.IsComplete())
	})

	t.Run("GetOperation", func(t *testing.T) {
		logger.On("Info", "operation started",
			"operationID", "get-test",
			"description", "Get test",
			"totalSteps", 3).Once()

		ctx := context.Background()

		// Start an operation first
		_, err := tracker.StartOperation(ctx, "get-test", "Get test", 3)
		require.NoError(t, err)

		// Get the operation
		op, err := tracker.GetOperation(ctx, "get-test")
		require.NoError(t, err)
		assert.Equal(t, "get-test", op.ID)
	})

	t.Run("GetNonExistentOperation", func(t *testing.T) {
		ctx := context.Background()
		op, err := tracker.GetOperation(ctx, "nonexistent")
		require.Error(t, err)
		assert.Nil(t, op)
		assert.ErrorIs(t, err, ErrOperationNotFound)
	})
}

// TestCommandValidator_BasicFunctionality tests command validation
func TestCommandValidator_BasicFunctionality(t *testing.T) {
	logger := new(MockLogger)
	sandbox := SandboxConfig{
		AllowedCommands:      []string{"echo", "ls"},
		AllowedPaths:         []string{"/tmp", "/home"},
		BlockedPaths:         []string{"/etc", "/sys"},
		EnvironmentWhitelist: []string{"PATH", "HOME"},
	}

	validator := NewDefaultCommandValidator(logger, sandbox)

	t.Run("Constructor", func(t *testing.T) {
		require.NotNil(t, validator)
		assert.Equal(t, logger, validator.logger)
		assert.NotEmpty(t, validator.allowedPaths)
		assert.NotEmpty(t, validator.blockedPaths)
	})

	t.Run("ValidateCommand_Valid", func(t *testing.T) {
		logger.On("Debug", "command validated successfully",
			"command", "echo",
			"argCount", 1).Once()

		err := validator.ValidateCommand(context.Background(), "echo", []string{"hello"})
		assert.NoError(t, err)
	})

	t.Run("ValidateCommand_CommandInjection", func(t *testing.T) {
		logger.On("Warn", "command injection attempt detected",
			"command", "echo; rm -rf /",
			"error", mock.AnythingOfType("*fmt.wrapError")).Once()

		err := validator.ValidateCommand(context.Background(), "echo; rm -rf /", []string{})
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrCommandInjection)
	})

	t.Run("ValidateEnvironment_Valid", func(t *testing.T) {
		env := map[string]string{
			"PATH": "/usr/bin:/bin",
			"HOME": "/home/user",
		}
		err := validator.ValidateEnvironment(context.Background(), env)
		assert.NoError(t, err)
	})

	t.Run("ValidateEnvironment_NotAllowed", func(t *testing.T) {
		logger.On("Warn", "environment variable not in whitelist",
			"key", "LD_PRELOAD").Once()

		env := map[string]string{
			"LD_PRELOAD": "/malicious.so",
		}
		err := validator.ValidateEnvironment(context.Background(), env)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrEnvNotAllowed)
	})

	t.Run("ValidatePath_Blocked", func(t *testing.T) {
		logger.On("Warn", "access to blocked path attempted",
			"path", "/etc/passwd",
			"blockedPath", "/etc").Once()

		err := validator.ValidatePath(context.Background(), "/etc/passwd")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidPath)
	})

	t.Run("ValidatePath_Traversal", func(t *testing.T) {
		logger.On("Warn", "path traversal attempt",
			"path", "../../../etc/passwd").Once()

		err := validator.ValidatePath(context.Background(), "../../../etc/passwd")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrPathTraversal)
	})
}
