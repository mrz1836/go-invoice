package executor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Configuration errors
var (
	ErrInvalidConfig           = errors.New("invalid configuration")
	ErrConfigNotFound          = errors.New("configuration not found")
	ErrPolicyViolation         = errors.New("security policy violation")
	ErrAuditLogFailed          = errors.New("audit log failed")
	ErrNoAllowedCommands       = errors.New("no allowed commands specified")
	ErrInvalidMaxExecutionTime = errors.New("invalid max execution time")
	ErrInvalidMaxOutputSize    = errors.New("invalid max output size")
	ErrInvalidMaxFileSize      = errors.New("invalid max file size")
	ErrInvalidCPULimit         = errors.New("invalid CPU limit")
	ErrInvalidMemoryLimit      = errors.New("invalid memory limit")
	ErrInvalidMaxConcurrentOps = errors.New("invalid max concurrent operations")
	ErrInvalidCommandTimeout   = errors.New("invalid command timeout")
)

// SecurityConfig manages security configuration and policies.
type SecurityConfig struct {
	// Sandbox contains the sandbox configuration
	Sandbox SandboxConfig `json:"sandbox"`

	// AuditEnabled enables security audit logging
	AuditEnabled bool `json:"auditEnabled"`

	// AuditLogPath is the path for audit logs
	AuditLogPath string `json:"auditLogPath"`

	// PolicyFile is the path to security policy file
	PolicyFile string `json:"policyFile,omitempty"`

	// StrictMode enables strict security checks
	StrictMode bool `json:"strictMode"`

	// RequireSignedCommands requires command signatures
	RequireSignedCommands bool `json:"requireSignedCommands"`

	// CommandTimeout is the global command timeout
	CommandTimeout time.Duration `json:"commandTimeout"`

	// MaxConcurrentOps limits concurrent operations
	MaxConcurrentOps int `json:"maxConcurrentOps"`
}

// DefaultSecurityConfig returns a secure default configuration.
func DefaultSecurityConfig() *SecurityConfig {
	homeDir, _ := os.UserHomeDir()
	return &SecurityConfig{
		Sandbox: SandboxConfig{
			AllowedCommands: []string{
				"go-invoice",
				"invoice",
				"client",
				"import",
				"export",
				"generate",
				"config",
				"summary",
			},
			AllowedPaths: []string{
				filepath.Join(homeDir, ".go-invoice"),
				filepath.Join(homeDir, "Documents", "invoices"),
				"/tmp",
			},
			BlockedPaths: []string{
				"/etc",
				"/sys",
				"/proc",
				"/dev",
				"/root",
				"/..",
			},
			MaxExecutionTime: 5 * time.Minute,
			MaxOutputSize:    10 * 1024 * 1024, // 10MB
			MaxFileSize:      50 * 1024 * 1024, // 50MB
			EnvironmentWhitelist: []string{
				"HOME",
				"USER",
				"PATH",
				"TMPDIR",
				"TEMP",
				"TMP",
				"LANG",
				"LC_ALL",
			},
			EnableNetworkIsolation: true,
			ResourceLimits: &ResourceLimits{
				MaxCPUPercent: 80,
				MaxMemoryMB:   512,
				MaxProcesses:  10,
				MaxOpenFiles:  100,
			},
		},
		AuditEnabled:     true,
		AuditLogPath:     filepath.Join(homeDir, ".go-invoice", "audit.log"),
		StrictMode:       true,
		CommandTimeout:   5 * time.Minute,
		MaxConcurrentOps: 5,
	}
}

// SecurityConfigBuilder builds security configurations.
type SecurityConfigBuilder struct {
	config *SecurityConfig
}

// NewSecurityConfigBuilder creates a new configuration builder.
func NewSecurityConfigBuilder() *SecurityConfigBuilder {
	return &SecurityConfigBuilder{
		config: DefaultSecurityConfig(),
	}
}

// WithSandbox sets the sandbox configuration.
func (b *SecurityConfigBuilder) WithSandbox(sandbox SandboxConfig) *SecurityConfigBuilder {
	b.config.Sandbox = sandbox
	return b
}

// WithAuditLog enables audit logging.
func (b *SecurityConfigBuilder) WithAuditLog(enabled bool, path string) *SecurityConfigBuilder {
	b.config.AuditEnabled = enabled
	b.config.AuditLogPath = path
	return b
}

// WithStrictMode enables strict security mode.
func (b *SecurityConfigBuilder) WithStrictMode(enabled bool) *SecurityConfigBuilder {
	b.config.StrictMode = enabled
	return b
}

// WithCommandTimeout sets the command timeout.
func (b *SecurityConfigBuilder) WithCommandTimeout(timeout time.Duration) *SecurityConfigBuilder {
	b.config.CommandTimeout = timeout
	return b
}

// WithMaxConcurrentOps sets the maximum concurrent operations.
func (b *SecurityConfigBuilder) WithMaxConcurrentOps(maxOps int) *SecurityConfigBuilder {
	b.config.MaxConcurrentOps = maxOps
	return b
}

// AddAllowedCommand adds an allowed command.
func (b *SecurityConfigBuilder) AddAllowedCommand(command string) *SecurityConfigBuilder {
	b.config.Sandbox.AllowedCommands = append(b.config.Sandbox.AllowedCommands, command)
	return b
}

// AddAllowedPath adds an allowed path.
func (b *SecurityConfigBuilder) AddAllowedPath(path string) *SecurityConfigBuilder {
	b.config.Sandbox.AllowedPaths = append(b.config.Sandbox.AllowedPaths, path)
	return b
}

// Build validates and returns the configuration.
func (b *SecurityConfigBuilder) Build() (*SecurityConfig, error) {
	if err := b.validate(); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidConfig, err)
	}
	return b.config, nil
}

// validate checks the configuration for errors.
func (b *SecurityConfigBuilder) validate() error {
	cfg := b.config

	// Validate sandbox
	if len(cfg.Sandbox.AllowedCommands) == 0 {
		return ErrNoAllowedCommands
	}

	if cfg.Sandbox.MaxExecutionTime <= 0 {
		return ErrInvalidMaxExecutionTime
	}

	if cfg.Sandbox.MaxOutputSize <= 0 {
		return ErrInvalidMaxOutputSize
	}

	if cfg.Sandbox.MaxFileSize <= 0 {
		return ErrInvalidMaxFileSize
	}

	// Validate resource limits
	if cfg.Sandbox.ResourceLimits != nil {
		if cfg.Sandbox.ResourceLimits.MaxCPUPercent <= 0 || cfg.Sandbox.ResourceLimits.MaxCPUPercent > 100 {
			return ErrInvalidCPULimit
		}
		if cfg.Sandbox.ResourceLimits.MaxMemoryMB <= 0 {
			return ErrInvalidMemoryLimit
		}
	}

	// Validate general settings
	if cfg.MaxConcurrentOps <= 0 {
		return ErrInvalidMaxConcurrentOps
	}

	if cfg.CommandTimeout <= 0 {
		return ErrInvalidCommandTimeout
	}

	return nil
}

// AuditLogger logs security-relevant events.
type AuditLogger interface {
	// LogCommandExecution logs command execution attempts.
	LogCommandExecution(ctx context.Context, event *CommandAuditEvent) error

	// LogSecurityViolation logs security violations.
	LogSecurityViolation(ctx context.Context, event *SecurityViolationEvent) error

	// LogAccessAttempt logs file/path access attempts.
	LogAccessAttempt(ctx context.Context, event *AccessAuditEvent) error

	// Query retrieves audit logs based on criteria.
	Query(ctx context.Context, criteria *AuditCriteria) ([]*AuditEntry, error)
}

// CommandAuditEvent represents a command execution audit event.
type CommandAuditEvent struct {
	Timestamp   time.Time              `json:"timestamp"`
	UserID      string                 `json:"userId"`
	SessionID   string                 `json:"sessionId"`
	Command     string                 `json:"command"`
	Args        []string               `json:"args"`
	WorkingDir  string                 `json:"workingDir"`
	Environment map[string]string      `json:"environment,omitempty"`
	ExitCode    int                    `json:"exitCode"`
	Duration    time.Duration          `json:"duration"`
	Error       string                 `json:"error,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// SecurityViolationEvent represents a security violation audit event.
type SecurityViolationEvent struct {
	Timestamp     time.Time              `json:"timestamp"`
	UserID        string                 `json:"userId"`
	SessionID     string                 `json:"sessionId"`
	ViolationType string                 `json:"violationType"`
	Resource      string                 `json:"resource"`
	Action        string                 `json:"action"`
	Details       string                 `json:"details"`
	Blocked       bool                   `json:"blocked"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// AccessAuditEvent represents a file/path access audit event.
type AccessAuditEvent struct {
	Timestamp time.Time              `json:"timestamp"`
	UserID    string                 `json:"userId"`
	SessionID string                 `json:"sessionId"`
	Path      string                 `json:"path"`
	Operation string                 `json:"operation"`
	Allowed   bool                   `json:"allowed"`
	Reason    string                 `json:"reason,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// AuditEntry represents a generic audit log entry.
type AuditEntry struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Event     interface{} `json:"event"`
}

// AuditCriteria defines search criteria for audit logs.
type AuditCriteria struct {
	StartTime     time.Time
	EndTime       time.Time
	UserID        string
	SessionID     string
	EventTypes    []string
	ViolationType string
	Limit         int
}

// FileAuditLogger implements AuditLogger with file-based storage.
type FileAuditLogger struct {
	logger   Logger
	filePath string
	mu       sync.Mutex
}

// NewFileAuditLogger creates a new file-based audit logger.
func NewFileAuditLogger(logger Logger, filePath string) (*FileAuditLogger, error) {
	if logger == nil {
		panic("logger is required")
	}

	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("failed to create audit log directory: %w", err)
	}

	return &FileAuditLogger{
		logger:   logger,
		filePath: filePath,
	}, nil
}

// LogCommandExecution logs command execution attempts.
func (a *FileAuditLogger) LogCommandExecution(ctx context.Context, event *CommandAuditEvent) error {
	return a.writeEntry(ctx, "command_execution", event)
}

// LogSecurityViolation logs security violations.
func (a *FileAuditLogger) LogSecurityViolation(ctx context.Context, event *SecurityViolationEvent) error {
	return a.writeEntry(ctx, "security_violation", event)
}

// LogAccessAttempt logs file/path access attempts.
func (a *FileAuditLogger) LogAccessAttempt(ctx context.Context, event *AccessAuditEvent) error {
	return a.writeEntry(ctx, "access_attempt", event)
}

// Query retrieves audit logs based on criteria.
func (a *FileAuditLogger) Query(_ context.Context, criteria *AuditCriteria) ([]*AuditEntry, error) {
	// For simplicity, this reads the entire file
	// In production, use a proper database or indexed storage
	a.mu.Lock()
	defer a.mu.Unlock()

	file, err := os.Open(a.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to open audit log: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	var entries []*AuditEntry
	decoder := json.NewDecoder(file)

	for decoder.More() {
		var entry AuditEntry
		if err := decoder.Decode(&entry); err != nil {
			a.logger.Warn("skipping malformed audit entry", "error", err)
			continue
		}

		// Apply filters
		if a.matchesCriteria(&entry, criteria) {
			entries = append(entries, &entry)
			if criteria.Limit > 0 && len(entries) >= criteria.Limit {
				break
			}
		}
	}

	return entries, nil
}

// writeEntry writes an audit entry to the log file.
func (a *FileAuditLogger) writeEntry(_ context.Context, eventType string, event interface{}) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	entry := &AuditEntry{
		ID:        generateAuditID(),
		Type:      eventType,
		Timestamp: time.Now(),
		Event:     event,
	}

	// Open file in append mode
	file, err := os.OpenFile(a.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("%w: failed to open audit log: %w", ErrAuditLogFailed, err)
	}
	defer func() {
		_ = file.Close()
	}()

	// Write JSON entry with newline
	encoder := json.NewEncoder(file)
	if err := encoder.Encode(entry); err != nil {
		return fmt.Errorf("%w: failed to write audit entry: %w", ErrAuditLogFailed, err)
	}

	return nil
}

// matchesCriteria checks if an entry matches the search criteria.
func (a *FileAuditLogger) matchesCriteria(entry *AuditEntry, criteria *AuditCriteria) bool {
	// Time range filter
	if !criteria.StartTime.IsZero() && entry.Timestamp.Before(criteria.StartTime) {
		return false
	}
	if !criteria.EndTime.IsZero() && entry.Timestamp.After(criteria.EndTime) {
		return false
	}

	// Event type filter
	if len(criteria.EventTypes) > 0 {
		found := false
		for _, t := range criteria.EventTypes {
			if entry.Type == t {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Additional filters based on event type
	// This is simplified; in production, properly unmarshal and check
	eventJSON, err := json.Marshal(entry.Event)
	if err != nil {
		// Log error but continue processing
		return true // Include entry if we can't marshal for comparison
	}
	eventStr := string(eventJSON)

	if criteria.UserID != "" && !contains(eventStr, criteria.UserID) {
		return false
	}
	if criteria.SessionID != "" && !contains(eventStr, criteria.SessionID) {
		return false
	}
	if criteria.ViolationType != "" && !contains(eventStr, criteria.ViolationType) {
		return false
	}

	return true
}

// Helper functions

func generateAuditID() string {
	return fmt.Sprintf("audit_%d_%d", time.Now().UnixNano(), os.Getpid())
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
