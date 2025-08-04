// Package executor implements secure command execution for the MCP server.
package executor

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Error represents command execution errors.
type Error struct {
	Op  string
	Msg string
}

func (e *Error) Error() string {
	return fmt.Sprintf("executor %s: %s", e.Op, e.Msg)
}

// Common errors
var (
	ErrCommandNotAllowed = &Error{Op: "validate", Msg: "command not allowed"}
	ErrInvalidPath       = &Error{Op: "validate", Msg: "invalid path"}
	ErrTimeout           = &Error{Op: "execute", Msg: "command execution timeout"}
	ErrOutputTooLarge    = &Error{Op: "execute", Msg: "output exceeds maximum size"}
	ErrInvalidWorkDir    = &Error{Op: "validate", Msg: "invalid working directory"}
)

// SecureExecutor implements the CommandExecutor interface with security features.
type SecureExecutor struct {
	logger      Logger
	validator   CommandValidator
	fileHandler FileHandler
	sandbox     SandboxConfig
	allowedCmds map[string]bool
	mu          sync.RWMutex
}

// NewSecureExecutor creates a new secure command executor.
func NewSecureExecutor(logger Logger, validator CommandValidator, sandbox SandboxConfig, fileHandler FileHandler) *SecureExecutor {
	if logger == nil {
		panic("logger is required")
	}
	if validator == nil {
		panic("validator is required")
	}
	if fileHandler == nil {
		panic("fileHandler is required")
	}

	// Build allowed commands map for fast lookup
	allowedCmds := make(map[string]bool)
	for _, cmd := range sandbox.AllowedCommands {
		allowedCmds[cmd] = true
	}

	return &SecureExecutor{
		logger:      logger,
		validator:   validator,
		fileHandler: fileHandler,
		sandbox:     sandbox,
		allowedCmds: allowedCmds,
	}
}

// Execute runs a command with the given request parameters.
func (e *SecureExecutor) Execute(ctx context.Context, req *ExecutionRequest) (*ExecutionResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Validate the command
	if err := e.ValidateCommand(ctx, req.Command, req.Args); err != nil {
		return nil, fmt.Errorf("command validation failed: %w", err)
	}

	// Validate environment
	if err := e.validator.ValidateEnvironment(ctx, req.Environment); err != nil {
		return nil, fmt.Errorf("environment validation failed: %w", err)
	}

	// Set default timeout if not specified
	timeout := req.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	if timeout > e.sandbox.MaxExecutionTime {
		timeout = e.sandbox.MaxExecutionTime
	}

	// Create execution context with timeout
	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Prepare workspace if input files are provided
	var workDir string
	var cleanup func()
	var err error

	if len(req.InputFiles) > 0 {
		workDir, cleanup, err = e.fileHandler.PrepareWorkspace(execCtx, req.InputFiles)
		if err != nil {
			return nil, fmt.Errorf("workspace preparation failed: %w", err)
		}
		defer cleanup()
	} else if req.WorkingDir != "" {
		// Validate provided working directory
		if validateErr := e.validator.ValidatePath(execCtx, req.WorkingDir); validateErr != nil {
			return nil, fmt.Errorf("%w: %w", ErrInvalidWorkDir, validateErr)
		}
		workDir = req.WorkingDir
	} else {
		// Use temporary directory
		workDir, err = os.MkdirTemp("", "mcp-exec-*")
		if err != nil {
			return nil, fmt.Errorf("failed to create temp dir: %w", err)
		}
		defer func() {
			// Best effort cleanup of temp directory
			_ = os.RemoveAll(workDir)
		}()
	}

	// Build and execute the command
	cmd := exec.CommandContext(execCtx, req.Command, req.Args...) //nolint:gosec // Command execution is the intended functionality
	cmd.Dir = workDir
	cmd.Env = e.buildEnvironment(req.Environment)

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Log command execution
	e.logger.Info("executing command",
		"command", req.Command,
		"args", req.Args,
		"workDir", workDir,
		"timeout", timeout,
	)

	// Start timer
	start := time.Now()

	// Report progress if callback is provided
	if req.ProgressCallback != nil {
		go e.reportProgress(execCtx, req.ProgressCallback, start, timeout)
	}

	// Execute the command
	err = cmd.Run()
	duration := time.Since(start)

	// Check output size limits
	if int64(stdout.Len()) > e.sandbox.MaxOutputSize || int64(stderr.Len()) > e.sandbox.MaxOutputSize {
		return nil, ErrOutputTooLarge
	}

	// Build response
	response := &ExecutionResponse{
		ExitCode: 0,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Duration: duration,
		Metadata: make(map[string]interface{}),
	}

	// Handle execution error
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			response.ExitCode = exitErr.ExitCode()
		} else if errors.Is(err, context.DeadlineExceeded) {
			response.ExitCode = -1
			response.Error = ErrTimeout.Error()
		} else {
			response.ExitCode = -1
			response.Error = err.Error()
		}
		e.logger.Error("command execution failed",
			"command", req.Command,
			"exitCode", response.ExitCode,
			"error", err,
		)
	} else {
		e.logger.Info("command executed successfully",
			"command", req.Command,
			"duration", duration,
		)
	}

	// Output file collection would be implemented when tool expectations are defined

	return response, nil
}

// ValidateCommand checks if a command is allowed to execute.
func (e *SecureExecutor) ValidateCommand(ctx context.Context, command string, args []string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Check if command is in allowed list
	e.mu.RLock()
	allowed := e.allowedCmds[command]
	e.mu.RUnlock()

	if !allowed {
		// Check if it's a path to an allowed command
		base := filepath.Base(command)
		e.mu.RLock()
		allowed = e.allowedCmds[base]
		e.mu.RUnlock()

		if !allowed {
			return fmt.Errorf("%w: %s", ErrCommandNotAllowed, command)
		}
	}

	// Additional validation via validator
	return e.validator.ValidateCommand(ctx, command, args)
}

// GetAllowedCommands returns the list of allowed commands.
func (e *SecureExecutor) GetAllowedCommands(ctx context.Context) ([]string, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	commands := make([]string, 0, len(e.allowedCmds))
	for cmd := range e.allowedCmds {
		commands = append(commands, cmd)
	}

	return commands, nil
}

// buildEnvironment builds the environment for command execution.
func (e *SecureExecutor) buildEnvironment(additional map[string]string) []string {
	// Start with a clean environment
	env := make([]string, 0)

	// Add whitelisted environment variables from current environment
	for _, key := range e.sandbox.EnvironmentWhitelist {
		if value, exists := os.LookupEnv(key); exists {
			env = append(env, fmt.Sprintf("%s=%s", key, value))
		}
	}

	// Add additional environment variables
	for key, value := range additional {
		// Check if key is in whitelist
		allowed := false
		for _, whitelisted := range e.sandbox.EnvironmentWhitelist {
			if key == whitelisted {
				allowed = true
				break
			}
		}
		if allowed {
			env = append(env, fmt.Sprintf("%s=%s", key, value))
		}
	}

	// Always add PATH if not already present
	hasPath := false
	for _, e := range env {
		if strings.HasPrefix(e, "PATH=") {
			hasPath = true
			break
		}
	}
	if !hasPath {
		if path, exists := os.LookupEnv("PATH"); exists {
			env = append(env, fmt.Sprintf("PATH=%s", path))
		}
	}

	return env
}

// reportProgress sends progress updates during command execution.
func (e *SecureExecutor) reportProgress(ctx context.Context, callback ProgressFunc, start time.Time, timeout time.Duration) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			elapsed := time.Since(start)
			percent := int((elapsed.Seconds() / timeout.Seconds()) * 100)
			if percent > 100 {
				percent = 100
			}

			update := &ProgressUpdate{
				Stage:     "executing",
				Percent:   percent,
				Message:   fmt.Sprintf("Command running... %v elapsed", elapsed.Round(time.Second)),
				Timestamp: time.Now(),
			}
			callback(update)
		}
	}
}
