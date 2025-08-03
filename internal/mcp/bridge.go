// Package mcp provides the Model Context Protocol server implementation for go-invoice CLI integration.
// It includes command execution, validation, and secure file handling capabilities.
package mcp

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Static errors for err113 compliance
var (
	ErrCommandNotAllowed      = errors.New("command not allowed")
	ErrUnsafeArgumentDetected = errors.New("potentially unsafe argument detected")
	ErrPathTraversalDetected  = errors.New("path traversal attempt detected")
	ErrPathOutsideAllowedDir  = errors.New("path outside allowed directory")
)

// DefaultCLIBridge implements the CLIBridge interface with security and validation
type DefaultCLIBridge struct {
	logger      Logger
	validator   CommandValidator
	fileHandler FileHandler
	config      CLIConfig
}

// NewCLIBridge creates a new CLI bridge with dependency injection
func NewCLIBridge(logger Logger, validator CommandValidator, fileHandler FileHandler, config CLIConfig) CLIBridge {
	return &DefaultCLIBridge{
		logger:      logger,
		validator:   validator,
		fileHandler: fileHandler,
		config:      config,
	}
}

// ExecuteCommand executes a CLI command with proper validation and security
func (b *DefaultCLIBridge) ExecuteCommand(ctx context.Context, req *CommandRequest) (*CommandResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	b.logger.Debug("Executing CLI command", "command", req.Command, "args", req.Args)

	// Validate command is allowed
	if err := b.validator.ValidateCommand(ctx, req.Command, req.Args); err != nil {
		return nil, fmt.Errorf("command validation failed: %w", err)
	}

	// Set up execution timeout
	timeout := req.Timeout
	if timeout == 0 {
		timeout = b.config.MaxTimeout
	}
	if timeout > b.config.MaxTimeout {
		timeout = b.config.MaxTimeout
	}

	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Prepare workspace
	workingDir := req.WorkingDir
	if workingDir == "" {
		workingDir = b.config.WorkingDir
	}

	cleanWorkDir, cleanup, err := b.fileHandler.PrepareWorkspace(execCtx, workingDir)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare workspace: %w", err)
	}
	defer cleanup()

	// Create command
	// #nosec G204 -- Command and arguments are validated through CommandValidator.ValidateCommand()
	// which checks against an allowlist and validates arguments for security issues
	cmd := exec.CommandContext(execCtx, req.Command, req.Args...)
	cmd.Dir = cleanWorkDir

	// Set up environment
	cmd.Env = b.buildEnvironment(req.Env)

	// Execute command and capture output
	start := time.Now()
	stdout, stderr, exitCode, err := b.executeCommand(cmd)
	duration := time.Since(start)

	response := &CommandResponse{
		ExitCode: exitCode,
		Stdout:   string(stdout),
		Stderr:   string(stderr),
		Duration: duration,
	}

	if err != nil {
		response.Error = err.Error()
		b.logger.Error("Command execution failed",
			"command", req.Command,
			"args", req.Args,
			"exitCode", exitCode,
			"duration", duration,
			"error", err)
	} else {
		b.logger.Info("Command executed successfully",
			"command", req.Command,
			"args", req.Args,
			"exitCode", exitCode,
			"duration", duration)
	}

	return response, nil
}

// ValidateCommand validates that a command is allowed to be executed
func (b *DefaultCLIBridge) ValidateCommand(ctx context.Context, command string, args []string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	return b.validator.ValidateCommand(ctx, command, args)
}

// GetAllowedCommands returns the list of allowed commands
func (b *DefaultCLIBridge) GetAllowedCommands(ctx context.Context) ([]string, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// This is a simple implementation - in a real system, this might query the validator
	// For now, we'll return the base command
	return []string{b.config.Path}, nil
}

// executeCommand executes the command and captures output
func (b *DefaultCLIBridge) executeCommand(cmd *exec.Cmd) (stdout, stderr []byte, exitCode int, err error) {
	stdout, err = cmd.Output()
	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			stderr = exitError.Stderr
			exitCode = exitError.ExitCode()
		} else {
			exitCode = -1
		}
		return stdout, stderr, exitCode, err
	}

	return stdout, stderr, 0, nil
}

// buildEnvironment creates the environment for command execution
func (b *DefaultCLIBridge) buildEnvironment(extraEnv map[string]string) []string {
	env := os.Environ()

	// Add extra environment variables
	for key, value := range extraEnv {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}

	return env
}

// DefaultCommandValidator implements the CommandValidator interface
type DefaultCommandValidator struct {
	allowedCommands map[string]bool
}

// NewCommandValidator creates a new command validator
func NewCommandValidator(allowedCommands []string) CommandValidator {
	cmdMap := make(map[string]bool)
	for _, cmd := range allowedCommands {
		cmdMap[cmd] = true
	}

	return &DefaultCommandValidator{
		allowedCommands: cmdMap,
	}
}

// ValidateCommand validates that a command and its arguments are safe to execute
func (v *DefaultCommandValidator) ValidateCommand(ctx context.Context, command string, args []string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Check if command is in allowed list
	allowed, err := v.IsCommandAllowed(ctx, command)
	if err != nil {
		return fmt.Errorf("failed to check command allowlist: %w", err)
	}

	if !allowed {
		return fmt.Errorf("%w: %s", ErrCommandNotAllowed, command)
	}

	// Validate arguments for security
	if err := v.validateArgs(args); err != nil {
		return fmt.Errorf("argument validation failed: %w", err)
	}

	return nil
}

// IsCommandAllowed checks if a command is in the allowed list
func (v *DefaultCommandValidator) IsCommandAllowed(ctx context.Context, command string) (bool, error) {
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
	}

	// Check exact match first
	if v.allowedCommands[command] {
		return true, nil
	}

	// Check if it's a path to an allowed command
	baseCmd := filepath.Base(command)
	if v.allowedCommands[baseCmd] {
		return true, nil
	}

	return false, nil
}

// validateArgs validates command arguments for security issues
func (v *DefaultCommandValidator) validateArgs(args []string) error {
	for _, arg := range args {
		// Check for command injection attempts
		if strings.Contains(arg, ";") || strings.Contains(arg, "&") ||
			strings.Contains(arg, "|") || strings.Contains(arg, "`") ||
			strings.Contains(arg, "$(") || strings.Contains(arg, "${") {
			return fmt.Errorf("%w: %s", ErrUnsafeArgumentDetected, arg)
		}

		// Check for path traversal attempts
		if strings.Contains(arg, "../") || strings.Contains(arg, "..\\") {
			return fmt.Errorf("%w: %s", ErrPathTraversalDetected, arg)
		}
	}

	return nil
}

// DefaultFileHandler implements the FileHandler interface
type DefaultFileHandler struct {
	baseDir string
}

// NewFileHandler creates a new file handler
func NewFileHandler(baseDir string) FileHandler {
	return &DefaultFileHandler{
		baseDir: baseDir,
	}
}

// PrepareWorkspace prepares a secure workspace for command execution
func (f *DefaultFileHandler) PrepareWorkspace(ctx context.Context, workingDir string) (string, func(), error) {
	select {
	case <-ctx.Done():
		return "", nil, ctx.Err()
	default:
	}

	// Validate the working directory path
	if err := f.ValidatePath(ctx, workingDir); err != nil {
		return "", nil, fmt.Errorf("invalid working directory: %w", err)
	}

	// Ensure directory exists
	if err := os.MkdirAll(workingDir, 0o750); err != nil {
		return "", nil, fmt.Errorf("failed to create working directory: %w", err)
	}

	// Clean the path
	cleanPath, err := filepath.Abs(workingDir)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Cleanup function (currently no-op, but could be extended)
	cleanup := func() {
		// In a more sophisticated implementation, this might clean up temporary files
	}

	return cleanPath, cleanup, nil
}

// ValidatePath validates that a path is safe to use
func (f *DefaultFileHandler) ValidatePath(ctx context.Context, path string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Check for path traversal attempts
	if strings.Contains(path, "../") || strings.Contains(path, "..\\") {
		return fmt.Errorf("%w: %s", ErrPathTraversalDetected, path)
	}

	// Ensure path is within base directory (if set)
	if f.baseDir != "" {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("failed to get absolute path: %w", err)
		}

		absBaseDir, err := filepath.Abs(f.baseDir)
		if err != nil {
			return fmt.Errorf("failed to get absolute base directory: %w", err)
		}

		relPath, err := filepath.Rel(absBaseDir, absPath)
		if err != nil || strings.HasPrefix(relPath, "..") {
			return fmt.Errorf("%w: %s", ErrPathOutsideAllowedDir, path)
		}
	}

	return nil
}
