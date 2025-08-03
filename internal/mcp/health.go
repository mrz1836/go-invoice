package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

// Health check errors
var (
	ErrHealthCheckFailed    = errors.New("health check failed")
	ErrCLINotAvailable      = errors.New("CLI not available")
	ErrStorageNotAccessible = errors.New("storage not accessible")
	ErrUnhealthyState       = errors.New("service in unhealthy state")
)

// HealthStatus represents the overall health status
type HealthStatus struct {
	Status           string            `json:"status"` // "healthy", "degraded", "unhealthy"
	Version          string            `json:"version"`
	Transport        string            `json:"transport"`
	Uptime           time.Duration     `json:"uptime"`
	CLIStatus        string            `json:"cliStatus"`
	CLIVersion       string            `json:"cliVersion,omitempty"`
	StorageStatus    string            `json:"storageStatus"`
	StoragePath      string            `json:"storagePath"`
	Checks           []HealthCheck     `json:"checks"`
	LastError        string            `json:"lastError,omitempty"`
	Timestamp        time.Time         `json:"timestamp"`
	Metadata         map[string]string `json:"metadata"`
	PerformanceStats *PerformanceStats `json:"performanceStats,omitempty"`
}

// HealthCheck represents an individual health check
type HealthCheck struct {
	Name        string        `json:"name"`
	Status      string        `json:"status"` // "healthy", "warning", "unhealthy"
	Duration    time.Duration `json:"duration"`
	Message     string        `json:"message,omitempty"`
	LastChecked time.Time     `json:"lastChecked"`
}

// PerformanceStats tracks performance metrics
type PerformanceStats struct {
	AverageResponseTime time.Duration `json:"averageResponseTime"`
	TotalRequests       uint64        `json:"totalRequests"`
	ErrorCount          uint64        `json:"errorCount"`
	LastHourRequests    uint64        `json:"lastHourRequests"`
	MemoryUsage         uint64        `json:"memoryUsageBytes,omitempty"`
}

// HealthChecker defines the interface for health checking
type HealthChecker interface {
	// CheckHealth performs a comprehensive health check
	CheckHealth(ctx context.Context) (*HealthStatus, error)

	// StartHealthMonitoring starts continuous health monitoring
	StartHealthMonitoring(ctx context.Context, interval time.Duration) error

	// StopHealthMonitoring stops the monitoring loop
	StopHealthMonitoring(ctx context.Context) error

	// GetLastHealthStatus returns the last known health status
	GetLastHealthStatus() *HealthStatus

	// RegisterCheck adds a custom health check
	RegisterCheck(name string, check func(context.Context) error)
}

// DefaultHealthChecker implements comprehensive health checking
type DefaultHealthChecker struct {
	logger        Logger
	config        *Config
	cliPath       string
	storagePath   string
	startTime     time.Time
	lastStatus    *HealthStatus
	customChecks  map[string]func(context.Context) error
	mu            sync.RWMutex
	monitoring    bool
	monitorCancel context.CancelFunc
	stats         *performanceTracker
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(logger Logger, config *Config) *DefaultHealthChecker {
	if logger == nil {
		panic("logger is required")
	}
	if config == nil {
		panic("config is required")
	}

	storagePath := filepath.Join(os.Getenv("HOME"), ".go-invoice")
	if config.CLI.WorkingDir != "" {
		storagePath = config.CLI.WorkingDir
	}

	return &DefaultHealthChecker{
		logger:       logger,
		config:       config,
		cliPath:      config.CLI.Path,
		storagePath:  storagePath,
		startTime:    time.Now(),
		customChecks: make(map[string]func(context.Context) error),
		stats:        newPerformanceTracker(),
	}
}

// CheckHealth performs a comprehensive health check
func (h *DefaultHealthChecker) CheckHealth(ctx context.Context) (*HealthStatus, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	h.logger.Debug("performing health check")
	startTime := time.Now()

	status := &HealthStatus{
		Status:    "unknown",
		Version:   "1.0.0", // TODO: Get from build info
		Transport: string(DetectTransport()),
		Uptime:    time.Since(h.startTime),
		Checks:    []HealthCheck{},
		Timestamp: time.Now(),
		Metadata:  make(map[string]string),
	}

	// Add performance stats
	status.PerformanceStats = h.stats.getStats()

	// Perform individual checks
	var wg sync.WaitGroup
	checkResults := make([]HealthCheck, 0)
	checkChan := make(chan HealthCheck, 10)

	// CLI availability check
	wg.Add(1)
	go func() {
		defer wg.Done()
		check := h.checkCLIAvailability(ctx)
		checkChan <- check
	}()

	// Storage accessibility check
	wg.Add(1)
	go func() {
		defer wg.Done()
		check := h.checkStorageHealth(ctx)
		checkChan <- check
	}()

	// Custom checks
	h.mu.RLock()
	for name, checkFunc := range h.customChecks {
		wg.Add(1)
		go func(n string, f func(context.Context) error) {
			defer wg.Done()
			check := h.runCustomCheck(ctx, n, f)
			checkChan <- check
		}(name, checkFunc)
	}
	h.mu.RUnlock()

	// Close channel when all checks complete
	go func() {
		wg.Wait()
		close(checkChan)
	}()

	// Collect results
	for check := range checkChan {
		checkResults = append(checkResults, check)
	}

	status.Checks = checkResults

	// Determine overall status
	overallHealthy := true
	hasWarnings := false

	for _, check := range checkResults {
		switch check.Status {
		case "unhealthy":
			overallHealthy = false
			if status.LastError == "" {
				status.LastError = check.Message
			}
		case "warning":
			hasWarnings = true
		}

		// Set specific statuses
		switch check.Name {
		case "CLI Availability":
			status.CLIStatus = check.Status
			if check.Status == "healthy" && check.Message != "" {
				status.CLIVersion = check.Message
			}
		case "Storage Health":
			status.StorageStatus = check.Status
			status.StoragePath = h.storagePath
		}
	}

	if overallHealthy {
		if hasWarnings {
			status.Status = "degraded"
		} else {
			status.Status = "healthy"
		}
	} else {
		status.Status = "unhealthy"
	}

	// Update last status
	h.mu.Lock()
	h.lastStatus = status
	h.mu.Unlock()

	// Record health check completion
	h.stats.recordHealthCheck(time.Since(startTime))

	h.logger.Info("health check completed",
		"status", status.Status,
		"duration", time.Since(startTime),
		"checks", len(status.Checks),
	)

	return status, nil
}

// checkCLIAvailability checks if the CLI is available and functional
func (h *DefaultHealthChecker) checkCLIAvailability(ctx context.Context) HealthCheck {
	check := HealthCheck{
		Name:        "CLI Availability",
		Status:      "unhealthy",
		LastChecked: time.Now(),
	}

	startTime := time.Now()
	defer func() {
		check.Duration = time.Since(startTime)
	}()

	// Create command with timeout
	cmdCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// #nosec G204 -- CLI path comes from validated configuration
	cmd := exec.CommandContext(cmdCtx, h.cliPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		check.Message = fmt.Sprintf("CLI check failed: %v", err)
		h.logger.Error("CLI availability check failed", "error", err)
		return check
	}

	// Parse version from output
	version := string(output)
	if version != "" {
		check.Status = "healthy"
		check.Message = version
		h.logger.Debug("CLI check passed", "version", version)
	} else {
		check.Status = "warning"
		check.Message = "CLI returned empty version"
	}

	return check
}

// checkStorageHealth checks if storage is accessible
func (h *DefaultHealthChecker) checkStorageHealth(ctx context.Context) HealthCheck {
	check := HealthCheck{
		Name:        "Storage Health",
		Status:      "unhealthy",
		LastChecked: time.Now(),
	}

	startTime := time.Now()
	defer func() {
		check.Duration = time.Since(startTime)
	}()

	// Check if storage directory exists
	info, err := os.Stat(h.storagePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Try to create it
			if err := os.MkdirAll(h.storagePath, 0o755); err != nil {
				check.Message = fmt.Sprintf("Cannot create storage directory: %v", err)
				return check
			}
			check.Status = "warning"
			check.Message = "Storage directory created"
			return check
		}
		check.Message = fmt.Sprintf("Cannot access storage: %v", err)
		return check
	}

	// Check if it's a directory
	if !info.IsDir() {
		check.Message = "Storage path is not a directory"
		return check
	}

	// Check write permissions by creating a test file
	testFile := filepath.Join(h.storagePath, ".health-check")
	if err := os.WriteFile(testFile, []byte("test"), 0o644); err != nil {
		check.Message = fmt.Sprintf("Storage not writable: %v", err)
		return check
	}

	// Clean up test file
	os.Remove(testFile)

	// Check available space (simplified check)
	check.Status = "healthy"
	check.Message = "Storage accessible and writable"

	return check
}

// runCustomCheck runs a custom health check
func (h *DefaultHealthChecker) runCustomCheck(ctx context.Context, name string, checkFunc func(context.Context) error) HealthCheck {
	check := HealthCheck{
		Name:        name,
		Status:      "unhealthy",
		LastChecked: time.Now(),
	}

	startTime := time.Now()
	defer func() {
		check.Duration = time.Since(startTime)
	}()

	if err := checkFunc(ctx); err != nil {
		check.Message = err.Error()
		return check
	}

	check.Status = "healthy"
	return check
}

// StartHealthMonitoring starts continuous health monitoring
func (h *DefaultHealthChecker) StartHealthMonitoring(ctx context.Context, interval time.Duration) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.monitoring {
		return fmt.Errorf("health monitoring already active")
	}

	// Create cancellable context
	monitorCtx, cancel := context.WithCancel(ctx)
	h.monitorCancel = cancel
	h.monitoring = true

	// Start monitoring goroutine
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		h.logger.Info("health monitoring started", "interval", interval)

		// Initial check
		h.CheckHealth(monitorCtx)

		for {
			select {
			case <-monitorCtx.Done():
				h.logger.Info("health monitoring stopped")
				return
			case <-ticker.C:
				h.CheckHealth(monitorCtx)
			}
		}
	}()

	return nil
}

// StopHealthMonitoring stops the monitoring loop
func (h *DefaultHealthChecker) StopHealthMonitoring(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.monitoring {
		return fmt.Errorf("health monitoring not active")
	}

	if h.monitorCancel != nil {
		h.monitorCancel()
	}
	h.monitoring = false

	return nil
}

// GetLastHealthStatus returns the last known health status
func (h *DefaultHealthChecker) GetLastHealthStatus() *HealthStatus {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.lastStatus
}

// RegisterCheck adds a custom health check
func (h *DefaultHealthChecker) RegisterCheck(name string, check func(context.Context) error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.customChecks[name] = check
	h.logger.Debug("registered custom health check", "name", name)
}

// performanceTracker tracks performance metrics
type performanceTracker struct {
	totalRequests    uint64
	errorCount       uint64
	responseTimes    []time.Duration
	lastHourRequests uint64
	lastHourReset    time.Time
	mu               sync.RWMutex
}

func newPerformanceTracker() *performanceTracker {
	return &performanceTracker{
		responseTimes: make([]time.Duration, 0, 1000),
		lastHourReset: time.Now(),
	}
}

func (p *performanceTracker) recordRequest(duration time.Duration, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.totalRequests++
	p.lastHourRequests++

	if err != nil {
		p.errorCount++
	}

	// Keep last 1000 response times for averaging
	if len(p.responseTimes) >= 1000 {
		p.responseTimes = p.responseTimes[1:]
	}
	p.responseTimes = append(p.responseTimes, duration)

	// Reset hourly counter if needed
	if time.Since(p.lastHourReset) > time.Hour {
		p.lastHourRequests = 0
		p.lastHourReset = time.Now()
	}
}

func (p *performanceTracker) recordHealthCheck(duration time.Duration) {
	p.recordRequest(duration, nil)
}

func (p *performanceTracker) getStats() *PerformanceStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	stats := &PerformanceStats{
		TotalRequests:    p.totalRequests,
		ErrorCount:       p.errorCount,
		LastHourRequests: p.lastHourRequests,
	}

	// Calculate average response time
	if len(p.responseTimes) > 0 {
		var total time.Duration
		for _, d := range p.responseTimes {
			total += d
		}
		stats.AverageResponseTime = total / time.Duration(len(p.responseTimes))
	}

	return stats
}

// HealthHandler provides HTTP endpoints for health checks
type HealthHandler struct {
	checker HealthChecker
	logger  Logger
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(checker HealthChecker, logger Logger) *HealthHandler {
	if checker == nil {
		panic("checker is required")
	}
	if logger == nil {
		panic("logger is required")
	}
	return &HealthHandler{
		checker: checker,
		logger:  logger,
	}
}

// ServeHTTP handles health check requests
func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	status, err := h.checker.CheckHealth(ctx)
	if err != nil {
		h.logger.Error("health check failed", "error", err)
		http.Error(w, "Health check failed", http.StatusInternalServerError)
		return
	}

	// Set appropriate status code
	statusCode := http.StatusOK
	switch status.Status {
	case "unhealthy":
		statusCode = http.StatusServiceUnavailable
	case "degraded":
		statusCode = http.StatusOK // Still return 200 for degraded
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(status); err != nil {
		h.logger.Error("failed to encode health status", "error", err)
	}
}
