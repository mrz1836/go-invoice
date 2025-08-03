package executor

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

// Common security errors
var (
	ErrPathTraversal    = errors.New("path traversal detected")
	ErrInvalidCharacter = errors.New("invalid character in input")
	ErrCommandInjection = errors.New("potential command injection detected")
	ErrEnvNotAllowed    = errors.New("environment variable not allowed")
)

// DefaultCommandValidator implements CommandValidator with security checks.
type DefaultCommandValidator struct {
	logger       Logger
	sandbox      SandboxConfig
	allowedPaths map[string]bool
	blockedPaths map[string]bool
}

// NewDefaultCommandValidator creates a new command validator.
func NewDefaultCommandValidator(logger Logger, sandbox SandboxConfig) *DefaultCommandValidator {
	if logger == nil {
		panic("logger is required")
	}

	// Build path maps for fast lookup
	allowedPaths := make(map[string]bool)
	for _, path := range sandbox.AllowedPaths {
		allowedPaths[filepath.Clean(path)] = true
	}

	blockedPaths := make(map[string]bool)
	for _, path := range sandbox.BlockedPaths {
		blockedPaths[filepath.Clean(path)] = true
	}

	return &DefaultCommandValidator{
		logger:       logger,
		sandbox:      sandbox,
		allowedPaths: allowedPaths,
		blockedPaths: blockedPaths,
	}
}

// ValidateCommand checks if a command is safe to execute.
func (v *DefaultCommandValidator) ValidateCommand(ctx context.Context, command string, args []string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Check for command injection attempts
	if err := v.checkCommandInjection(command); err != nil {
		v.logger.Warn("command injection attempt detected",
			"command", command,
			"error", err,
		)
		return err
	}

	// Validate each argument
	for _, arg := range args {
		if err := v.validateArgument(arg); err != nil {
			v.logger.Warn("invalid argument detected",
				"arg", arg,
				"error", err,
			)
			return fmt.Errorf("invalid argument %q: %w", arg, err)
		}
	}

	// Check for specific dangerous patterns
	fullCommand := fmt.Sprintf("%s %s", command, strings.Join(args, " "))
	if err := v.checkDangerousPatterns(fullCommand); err != nil {
		v.logger.Warn("dangerous pattern detected",
			"command", fullCommand,
			"error", err,
		)
		return err
	}

	v.logger.Debug("command validated successfully",
		"command", command,
		"argCount", len(args),
	)

	return nil
}

// ValidateEnvironment checks if environment variables are safe.
func (v *DefaultCommandValidator) ValidateEnvironment(ctx context.Context, env map[string]string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Check each environment variable
	for key, value := range env {
		// Check if key is in whitelist
		allowed := false
		for _, whitelisted := range v.sandbox.EnvironmentWhitelist {
			if key == whitelisted {
				allowed = true
				break
			}
		}

		if !allowed {
			v.logger.Warn("environment variable not in whitelist",
				"key", key,
			)
			return fmt.Errorf("%w: %s", ErrEnvNotAllowed, key)
		}

		// Validate the value doesn't contain dangerous content
		if err := v.validateArgument(value); err != nil {
			v.logger.Warn("invalid environment value",
				"key", key,
				"error", err,
			)
			return fmt.Errorf("invalid environment value for %s: %w", key, err)
		}
	}

	return nil
}

// ValidatePath checks if a file path is safe to access.
func (v *DefaultCommandValidator) ValidatePath(ctx context.Context, path string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Clean and resolve the path
	cleanPath := filepath.Clean(path)
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Check for path traversal
	if strings.Contains(path, "..") {
		v.logger.Warn("path traversal attempt",
			"path", path,
		)
		return ErrPathTraversal
	}

	// Check if path is in blocked list
	for blockedPath := range v.blockedPaths {
		if strings.HasPrefix(absPath, blockedPath) {
			v.logger.Warn("access to blocked path attempted",
				"path", absPath,
				"blockedPath", blockedPath,
			)
			return fmt.Errorf("%w: path is blocked", ErrInvalidPath)
		}
	}

	// Check if path is in allowed list (if any are specified)
	if len(v.allowedPaths) > 0 {
		allowed := false
		for allowedPath := range v.allowedPaths {
			if strings.HasPrefix(absPath, allowedPath) {
				allowed = true
				break
			}
		}
		if !allowed {
			v.logger.Warn("access to non-allowed path attempted",
				"path", absPath,
			)
			return fmt.Errorf("%w: path not in allowed list", ErrInvalidPath)
		}
	}

	return nil
}

// checkCommandInjection checks for command injection attempts.
func (v *DefaultCommandValidator) checkCommandInjection(command string) error {
	// Check for shell metacharacters
	dangerousChars := []string{";", "&", "|", "`", "$", "(", ")", "{", "}", "<", ">", "\n", "\r"}
	for _, char := range dangerousChars {
		if strings.Contains(command, char) {
			return fmt.Errorf("%w: contains %q", ErrCommandInjection, char)
		}
	}

	// Check for null bytes
	if strings.Contains(command, "\x00") {
		return fmt.Errorf("%w: contains null byte", ErrCommandInjection)
	}

	return nil
}

// validateArgument validates a single command argument.
func (v *DefaultCommandValidator) validateArgument(arg string) error {
	// Check for null bytes
	if strings.Contains(arg, "\x00") {
		return fmt.Errorf("%w: contains null byte", ErrInvalidCharacter)
	}

	// Check for control characters (except tab and newline which might be valid)
	for i, r := range arg {
		if r < 32 && r != '\t' && r != '\n' {
			return fmt.Errorf("%w: control character at position %d", ErrInvalidCharacter, i)
		}
	}

	// Check length
	if len(arg) > 4096 {
		return fmt.Errorf("argument too long: %d bytes", len(arg))
	}

	return nil
}

// checkDangerousPatterns checks for dangerous command patterns.
func (v *DefaultCommandValidator) checkDangerousPatterns(fullCommand string) error {
	// List of dangerous patterns to check
	dangerousPatterns := []string{
		"rm -rf /",
		"rm -fr /",
		"dd if=/dev/zero",
		"mkfs.",
		":(){ :|:& };:", // Fork bomb
		"> /dev/sda",
		"wget http",
		"curl http",
		"nc -l",
		"/etc/passwd",
		"/etc/shadow",
		"sudo ",
		"su -",
	}

	lowerCommand := strings.ToLower(fullCommand)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(lowerCommand, strings.ToLower(pattern)) {
			return fmt.Errorf("dangerous pattern detected: %s", pattern)
		}
	}

	return nil
}
