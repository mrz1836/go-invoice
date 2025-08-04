package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// Test errors
var (
	errCustomWarning     = errors.New("warning condition")
	errTestError         = errors.New("test error")
	errForcedFailure     = errors.New("forced failure")
	errCustomCheckFailed = errors.New("custom check failed")
)

type HealthTestSuite struct {
	suite.Suite

	tempDir    string
	logger     *mockLogger
	config     *Config
	checker    *DefaultHealthChecker
	testCLIDir string
}

func TestHealthSuite(t *testing.T) {
	suite.Run(t, new(HealthTestSuite))
}

func (s *HealthTestSuite) SetupTest() {
	tempDir, err := os.MkdirTemp("", "health-test")
	s.Require().NoError(err)
	s.tempDir = tempDir

	// Create test CLI directory
	s.testCLIDir = filepath.Join(s.tempDir, "cli")
	s.Require().NoError(os.MkdirAll(s.testCLIDir, 0o750))

	// Create mock CLI executable
	cliPath := filepath.Join(s.testCLIDir, "test-cli")
	cliScript := "#!/bin/bash\necho 'test-cli version 1.0.0'\n"
	s.Require().NoError(os.WriteFile(cliPath, []byte(cliScript), 0o750)) //nolint:gosec // executable script needs exec permissions

	s.logger = newMockLogger()
	s.config = &Config{
		CLI: CLIConfig{
			Path:       cliPath,
			WorkingDir: s.tempDir,
			MaxTimeout: 10 * time.Second,
		},
		Server: ServerConfig{
			Host:        "localhost",
			Port:        0,
			Timeout:     30 * time.Second,
			ReadTimeout: 5 * time.Second,
		},
		Security: SecurityConfig{
			AllowedCommands:       []string{"test-cli"},
			WorkingDir:            s.tempDir,
			SandboxEnabled:        true,
			FileAccessRestricted:  true,
			MaxCommandTimeout:     "10s",
			EnableInputValidation: true,
		},
		LogLevel: "info",
	}

	s.checker = NewHealthChecker(s.logger, s.config)
}

func (s *HealthTestSuite) TearDownTest() {
	if s.checker != nil {
		_ = s.checker.StopHealthMonitoring(context.Background())
	}
	if s.tempDir != "" {
		err := os.RemoveAll(s.tempDir)
		s.Require().NoError(err)
	}
}

func (s *HealthTestSuite) TestNewHealthChecker() {
	tests := []struct {
		name        string
		logger      Logger
		config      *Config
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
			expectPanic: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			if tt.expectPanic {
				s.Panics(func() {
					NewHealthChecker(tt.logger, tt.config)
				})
			} else {
				checker := NewHealthChecker(tt.logger, tt.config)
				s.NotNil(checker)
				s.Equal(tt.logger, checker.logger)
				s.Equal(tt.config, checker.config)
				s.NotNil(checker.customChecks)
				s.NotNil(checker.stats)
			}
		})
	}
}

func (s *HealthTestSuite) TestCheckHealthSuccess() {
	ctx := context.Background()

	status, err := s.checker.CheckHealth(ctx)
	s.Require().NoError(err)
	s.NotNil(status)

	// Verify basic status structure
	s.Equal("healthy", status.Status)
	s.Equal("1.0.0", status.Version)
	s.NotEmpty(status.Transport)
	s.Positive(status.Uptime)
	s.NotEmpty(status.Timestamp)
	s.NotNil(status.Metadata)
	s.NotNil(status.PerformanceStats)

	// Verify checks were performed
	s.Len(status.Checks, 2) // CLI and Storage checks

	checkNames := make(map[string]bool)
	for _, check := range status.Checks {
		checkNames[check.Name] = true
		s.NotEmpty(check.LastChecked)
		s.GreaterOrEqual(check.Duration, time.Duration(0))
	}
	s.True(checkNames["CLI Availability"])
	s.True(checkNames["Storage Health"])

	// Verify CLI status
	s.Equal("healthy", status.CLIStatus)
	s.Contains(status.CLIVersion, "1.0.0")

	// Verify storage status
	s.Equal("healthy", status.StorageStatus)
	s.Equal(s.tempDir, status.StoragePath)
}

func (s *HealthTestSuite) TestCheckHealthWithCustomChecks() {
	ctx := context.Background()

	// Register custom checks
	s.checker.RegisterCheck("Custom Success", func(ctx context.Context) error {
		return nil
	})
	s.checker.RegisterCheck("Custom Warning", func(ctx context.Context) error {
		return errCustomWarning
	})

	status, err := s.checker.CheckHealth(ctx)
	s.Require().NoError(err)
	s.NotNil(status)

	// Should be unhealthy due to custom check failure
	s.Equal("unhealthy", status.Status)
	s.Len(status.Checks, 4) // CLI, Storage, and 2 custom checks

	checkStatuses := make(map[string]string)
	for _, check := range status.Checks {
		checkStatuses[check.Name] = check.Status
	}

	s.Equal("healthy", checkStatuses["Custom Success"])
	s.Equal("unhealthy", checkStatuses["Custom Warning"])
}

func (s *HealthTestSuite) TestCheckHealthContextCancellation() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	status, err := s.checker.CheckHealth(ctx)
	s.Equal(context.Canceled, err)
	s.Nil(status)
}

func (s *HealthTestSuite) TestCheckHealthCLIFailure() {
	// Create config with non-existent CLI
	config := *s.config
	config.CLI.Path = "/non/existent/cli"
	checker := NewHealthChecker(s.logger, &config)

	ctx := context.Background()
	status, err := checker.CheckHealth(ctx)
	s.Require().NoError(err)
	s.NotNil(status)

	s.Equal("unhealthy", status.Status)
	s.Equal("unhealthy", status.CLIStatus)
	s.NotEmpty(status.LastError)
}

func (s *HealthTestSuite) TestCheckHealthStorageFailure() {
	// Create config with invalid storage path (file instead of directory)
	invalidPath := filepath.Join(s.tempDir, "not-a-dir")
	s.Require().NoError(os.WriteFile(invalidPath, []byte("test"), 0o600))

	config := *s.config
	config.CLI.WorkingDir = invalidPath
	checker := NewHealthChecker(s.logger, &config)

	ctx := context.Background()
	status, err := checker.CheckHealth(ctx)
	s.Require().NoError(err)
	s.NotNil(status)

	s.Equal("unhealthy", status.Status)
	s.Equal("unhealthy", status.StorageStatus)
	s.NotEmpty(status.LastError)
}

func (s *HealthTestSuite) TestStartStopHealthMonitoring() {
	ctx := context.Background()
	interval := 100 * time.Millisecond

	// Start monitoring
	err := s.checker.StartHealthMonitoring(ctx, interval)
	s.Require().NoError(err)

	// Verify monitoring is active
	s.checker.mu.RLock()
	monitoring := s.checker.monitoring
	s.checker.mu.RUnlock()
	s.True(monitoring)

	// Try to start again (should fail)
	err = s.checker.StartHealthMonitoring(ctx, interval)
	s.Equal(ErrMonitoringAlreadyActive, err)

	// Wait for at least one monitoring cycle
	time.Sleep(150 * time.Millisecond)

	// Stop monitoring
	err = s.checker.StopHealthMonitoring(ctx)
	s.Require().NoError(err)

	// Verify monitoring is stopped
	s.checker.mu.RLock()
	monitoring = s.checker.monitoring
	s.checker.mu.RUnlock()
	s.False(monitoring)

	// Try to stop again (should fail)
	err = s.checker.StopHealthMonitoring(ctx)
	s.Equal(ErrMonitoringNotActive, err)
}

func (s *HealthTestSuite) TestGetLastHealthStatus() {
	ctx := context.Background()

	// Initially should be nil
	status := s.checker.GetLastHealthStatus()
	s.Nil(status)

	// Perform health check
	_, err := s.checker.CheckHealth(ctx)
	s.Require().NoError(err)

	// Now should have status
	status = s.checker.GetLastHealthStatus()
	s.NotNil(status)
	s.Equal("healthy", status.Status)
}

func (s *HealthTestSuite) TestRegisterCheck() {
	testCheck := func(ctx context.Context) error {
		return nil
	}

	s.checker.RegisterCheck("Test Check", testCheck)

	s.checker.mu.RLock()
	check, exists := s.checker.customChecks["Test Check"]
	s.checker.mu.RUnlock()

	s.True(exists)
	s.NotNil(check)

	// Verify the check is called during health check
	ctx := context.Background()
	status, err := s.checker.CheckHealth(ctx)
	s.Require().NoError(err)

	found := false
	for _, healthCheck := range status.Checks {
		if healthCheck.Name == "Test Check" {
			found = true
			s.Equal("healthy", healthCheck.Status)
			break
		}
	}
	s.True(found, "Custom check should be included in health status")
}

func (s *HealthTestSuite) TestHealthErrors() {
	tests := []struct {
		name     string
		err      *HealthError
		expected string
	}{
		{
			name:     "HealthCheckFailed",
			err:      ErrHealthCheckFailed,
			expected: "health check: health check failed",
		},
		{
			name:     "CLINotAvailable",
			err:      ErrCLINotAvailable,
			expected: "health validate: CLI not available",
		},
		{
			name:     "StorageNotAccessible",
			err:      ErrStorageNotAccessible,
			expected: "health validate: storage not accessible",
		},
		{
			name:     "UnhealthyState",
			err:      ErrUnhealthyState,
			expected: "health status: service in unhealthy state",
		},
		{
			name:     "MonitoringAlreadyActive",
			err:      ErrMonitoringAlreadyActive,
			expected: "health start: health monitoring already active",
		},
		{
			name:     "MonitoringNotActive",
			err:      ErrMonitoringNotActive,
			expected: "health stop: health monitoring not active",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Equal(tt.expected, tt.err.Error())
		})
	}
}

func (s *HealthTestSuite) TestPerformanceTracker() {
	tracker := newPerformanceTracker()
	s.NotNil(tracker)
	s.Equal(uint64(0), tracker.totalRequests)
	s.Equal(uint64(0), tracker.errorCount)
	s.Empty(tracker.responseTimes)

	// Record some requests
	tracker.recordRequest(100*time.Millisecond, nil)
	tracker.recordRequest(200*time.Millisecond, errTestError)
	tracker.recordHealthCheck(50 * time.Millisecond)

	stats := tracker.getStats()
	s.Equal(uint64(3), stats.TotalRequests)
	s.Equal(uint64(1), stats.ErrorCount)
	s.Equal(uint64(3), stats.LastHourRequests)
	s.Positive(stats.AverageResponseTime)

	// Test response time limit (should keep only last 1000)
	for i := 0; i < 1001; i++ {
		tracker.recordRequest(time.Millisecond, nil)
	}
	s.Len(tracker.responseTimes, 1000)
}

func (s *HealthTestSuite) TestHealthHandler() {
	handler := NewHealthHandler(s.checker, s.logger)
	s.NotNil(handler)

	tests := []struct {
		name           string
		method         string
		expectedStatus int
		setupChecker   func(*DefaultHealthChecker)
	}{
		{
			name:           "HealthyResponse",
			method:         "GET",
			expectedStatus: http.StatusOK,
			setupChecker:   nil,
		},
		{
			name:           "MethodNotAllowed",
			method:         "POST",
			expectedStatus: http.StatusMethodNotAllowed,
			setupChecker:   nil,
		},
		{
			name:           "UnhealthyResponse",
			method:         "GET",
			expectedStatus: http.StatusServiceUnavailable,
			setupChecker: func(checker *DefaultHealthChecker) {
				checker.RegisterCheck("Failing Check", func(ctx context.Context) error {
					return errForcedFailure
				})
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			if tt.setupChecker != nil {
				tt.setupChecker(s.checker)
			}

			req := httptest.NewRequest(tt.method, "/health", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)
			s.Equal(tt.expectedStatus, w.Code)

			if tt.method == "GET" && tt.expectedStatus != http.StatusMethodNotAllowed {
				s.Equal("application/json", w.Header().Get("Content-Type"))

				var status HealthStatus
				err := json.NewDecoder(w.Body).Decode(&status)
				s.Require().NoError(err)
				s.NotEmpty(status.Status)
			}
		})
	}
}

func (s *HealthTestSuite) TestHealthHandlerPanics() {
	s.Panics(func() {
		NewHealthHandler(nil, s.logger)
	})

	s.Panics(func() {
		NewHealthHandler(s.checker, nil)
	})
}

func (s *HealthTestSuite) TestConcurrentHealthChecks() {
	ctx := context.Background()
	numGoroutines := 10
	var wg sync.WaitGroup
	results := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := s.checker.CheckHealth(ctx)
			results <- err
		}()
	}

	wg.Wait()
	close(results)

	for err := range results {
		s.NoError(err)
	}
}

func (s *HealthTestSuite) TestStorageHealthChecks() {
	tests := []struct {
		name        string
		setup       func() string
		expectMsg   string
		expectState string
	}{
		{
			name: "ExistingDirectory",
			setup: func() string {
				dir := filepath.Join(s.tempDir, "existing")
				s.Require().NoError(os.MkdirAll(dir, 0o750))
				return dir
			},
			expectMsg:   "Storage accessible and writable",
			expectState: "healthy",
		},
		{
			name: "NonExistentDirectory",
			setup: func() string {
				return filepath.Join(s.tempDir, "new-dir")
			},
			expectMsg:   "Storage directory created",
			expectState: "warning",
		},
		{
			name: "FileInsteadOfDirectory",
			setup: func() string {
				file := filepath.Join(s.tempDir, "not-a-dir")
				s.Require().NoError(os.WriteFile(file, []byte("test"), 0o600))
				return file
			},
			expectMsg:   "Storage path is not a directory",
			expectState: "unhealthy",
		},
		{
			name: "ReadOnlyDirectory",
			setup: func() string {
				dir := filepath.Join(s.tempDir, "readonly")
				s.Require().NoError(os.MkdirAll(dir, 0o555)) //nolint:gosec // intentional read-only for test
				return dir
			},
			expectMsg:   "Storage not writable",
			expectState: "unhealthy",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			storagePath := tt.setup()

			config := *s.config
			config.CLI.WorkingDir = storagePath
			checker := NewHealthChecker(s.logger, &config)

			check := checker.checkStorageHealth(context.Background())
			s.Equal("Storage Health", check.Name)
			s.Equal(tt.expectState, check.Status)
			s.Contains(check.Message, tt.expectMsg)
			s.GreaterOrEqual(check.Duration, time.Duration(0))
			s.False(check.LastChecked.IsZero())
		})
	}
}

func (s *HealthTestSuite) TestCLIHealthChecks() {
	tests := []struct {
		name        string
		cliPath     string
		expectState string
		expectMsg   string
	}{
		{
			name:        "ValidCLI",
			cliPath:     s.config.CLI.Path,
			expectState: "healthy",
			expectMsg:   "1.0.0",
		},
		{
			name:        "NonExistentCLI",
			cliPath:     "/non/existent/cli",
			expectState: "unhealthy",
			expectMsg:   "CLI check failed",
		},
		{
			name: "EmptyVersionOutput",
			cliPath: func() string {
				// Create CLI that returns empty output
				emptyCLI := filepath.Join(s.testCLIDir, "empty-cli")
				script := "#!/bin/bash\n"                                          // No echo command, just exit with 0
				s.Require().NoError(os.WriteFile(emptyCLI, []byte(script), 0o750)) //nolint:gosec // executable script needs exec permissions
				return emptyCLI
			}(),
			expectState: "warning",
			expectMsg:   "CLI returned empty version",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			config := *s.config
			config.CLI.Path = tt.cliPath
			checker := NewHealthChecker(s.logger, &config)

			check := checker.checkCLIAvailability(context.Background())
			s.Equal("CLI Availability", check.Name)
			s.Equal(tt.expectState, check.Status)
			s.Contains(check.Message, tt.expectMsg)
			s.GreaterOrEqual(check.Duration, time.Duration(0))
			s.False(check.LastChecked.IsZero())
		})
	}
}

func (s *HealthTestSuite) TestCustomCheckExecution() {
	ctx := context.Background()

	tests := []struct {
		name        string
		checkFunc   func(context.Context) error
		expectState string
		expectMsg   string
	}{
		{
			name: "SuccessfulCheck",
			checkFunc: func(ctx context.Context) error {
				return nil
			},
			expectState: "healthy",
			expectMsg:   "",
		},
		{
			name: "FailingCheck",
			checkFunc: func(ctx context.Context) error {
				return errCustomCheckFailed
			},
			expectState: "unhealthy",
			expectMsg:   "custom check failed",
		},
		{
			name: "ContextAwareCheck",
			checkFunc: func(ctx context.Context) error {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
					return nil
				}
			},
			expectState: "healthy",
			expectMsg:   "",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			check := s.checker.runCustomCheck(ctx, "Test Check", tt.checkFunc)
			s.Equal("Test Check", check.Name)
			s.Equal(tt.expectState, check.Status)
			if tt.expectMsg != "" {
				s.Contains(check.Message, tt.expectMsg)
			}
			s.GreaterOrEqual(check.Duration, time.Duration(0))
			s.False(check.LastChecked.IsZero())
		})
	}
}

// mockLogger implements the Logger interface for testing
type mockLogger struct {
	mu       sync.RWMutex
	debugMsg []string
	infoMsg  []string
	warnMsg  []string
	errorMsg []string
}

func newMockLogger() *mockLogger {
	return &mockLogger{
		debugMsg: make([]string, 0),
		infoMsg:  make([]string, 0),
		warnMsg:  make([]string, 0),
		errorMsg: make([]string, 0),
	}
}

func (m *mockLogger) Debug(msg string, keysAndValues ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.debugMsg = append(m.debugMsg, fmt.Sprintf("%s: %v", msg, keysAndValues))
}

func (m *mockLogger) Info(msg string, keysAndValues ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.infoMsg = append(m.infoMsg, fmt.Sprintf("%s: %v", msg, keysAndValues))
}

func (m *mockLogger) Warn(msg string, keysAndValues ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.warnMsg = append(m.warnMsg, fmt.Sprintf("%s: %v", msg, keysAndValues))
}

func (m *mockLogger) Error(msg string, keysAndValues ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errorMsg = append(m.errorMsg, fmt.Sprintf("%s: %v", msg, keysAndValues))
}

func (m *mockLogger) GetDebugMessages() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]string, len(m.debugMsg))
	copy(result, m.debugMsg)
	return result
}

func (m *mockLogger) GetInfoMessages() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]string, len(m.infoMsg))
	copy(result, m.infoMsg)
	return result
}

func (m *mockLogger) GetWarnMessages() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]string, len(m.warnMsg))
	copy(result, m.warnMsg)
	return result
}

func (m *mockLogger) GetErrorMessages() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]string, len(m.errorMsg))
	copy(result, m.errorMsg)
	return result
}

// Benchmark tests
func BenchmarkCheckHealth(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "health-bench")
	require.NoError(b, err)
	defer func() {
		removeErr := os.RemoveAll(tempDir)
		require.NoError(b, removeErr)
	}()

	// Create mock CLI
	cliPath := filepath.Join(tempDir, "bench-cli")
	cliScript := "#!/bin/bash\necho 'bench-cli version 1.0.0'\n"
	require.NoError(b, os.WriteFile(cliPath, []byte(cliScript), 0o750)) //nolint:gosec // executable script needs exec permissions

	logger := newMockLogger()
	config := &Config{
		CLI: CLIConfig{
			Path:       cliPath,
			WorkingDir: tempDir,
			MaxTimeout: 10 * time.Second,
		},
		Server:   ServerConfig{Host: "localhost", Port: 0},
		Security: SecurityConfig{AllowedCommands: []string{"bench-cli"}, WorkingDir: tempDir},
		LogLevel: "info",
	}

	checker := NewHealthChecker(logger, config)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := checker.CheckHealth(ctx)
		require.NoError(b, err)
	}
}

func BenchmarkPerformanceTracker(b *testing.B) {
	tracker := newPerformanceTracker()
	duration := 100 * time.Millisecond

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tracker.recordRequest(duration, nil)
	}
}
