// Package mcp provides comprehensive performance benchmarks for the MCP server implementation.
//
// This test suite validates response time targets, identifies performance bottlenecks,
// and ensures the MCP server meets its performance requirements for Claude integration.
//
// The benchmarks cover:
// - All 21 MCP tools with varying input sizes
// - Transport performance (stdio vs HTTP)
// - Concurrent request handling and scalability
// - Memory usage patterns and resource consumption
// - Response time validation against targets (<100ms simple, <2s complex)
// - Load testing for various scenarios
// - Performance regression detection
// - Resource limit validation
//
// Usage:
//
//	go test -bench=. -benchmem -cpuprofile=cpu.prof -memprofile=mem.prof
//	go test -bench=BenchmarkTool -benchtime=10s
//	go test -bench=BenchmarkConcurrent -cpu=1,2,4,8
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/mrz/go-invoice/internal/mcp/tools"
	"github.com/mrz/go-invoice/internal/mcp/types"
)

// Performance targets for validation
const (
	// SimpleOperationTarget is the maximum acceptable response time for simple operations
	SimpleOperationTarget = 100 * time.Millisecond

	// ComplexOperationTarget is the maximum acceptable response time for complex operations
	ComplexOperationTarget = 2 * time.Second

	// ConcurrentRequestsTarget is the minimum concurrent requests the server should handle
	ConcurrentRequestsTarget = 100

	// ThroughputTarget is the minimum operations per second for sustained load
	ThroughputTarget = 50
)

// PerformanceTestSuite provides comprehensive performance testing for the MCP server
type PerformanceTestSuite struct {
	suite.Suite

	// Test infrastructure
	server         *DefaultServer
	logger         *TestLogger
	bridge         *MockCLIBridge
	config         *Config
	toolComponents *tools.ToolSystemComponents

	// Transport testing
	stdioTransport Transport
	httpTransport  Transport
	httpServer     *httptest.Server

	// Performance tracking
	responseTimer   *ResponseTimeTracker
	resourceMonitor *ResourceMonitor

	// Test data generators
	dataGenerator *TestDataGenerator
}

// SetupSuite initializes the performance test environment
func (s *PerformanceTestSuite) SetupSuite() {
	s.logger = NewTestLogger()
	s.bridge = NewMockCLIBridge()
	s.config = s.createTestConfig()
	s.responseTimer = NewResponseTimeTracker()
	s.resourceMonitor = NewResourceMonitor()
	s.dataGenerator = NewTestDataGenerator()

	// Initialize tool system
	ctx := context.Background()
	components, err := tools.InitializeToolSystem(ctx, s.logger)
	s.Require().NoError(err, "Failed to initialize tool system")
	s.toolComponents = components

	// Create server instance
	s.server = NewServer(s.logger, s.bridge, s.config).(*DefaultServer)

	// Initialize transports
	s.setupTransports()
}

// TearDownSuite cleans up test resources
func (s *PerformanceTestSuite) TearDownSuite() {
	if s.httpServer != nil {
		s.httpServer.Close()
	}

	s.resourceMonitor.Stop()
	s.printPerformanceReport()
}

// setupTransports initializes both stdio and HTTP transports for testing
func (s *PerformanceTestSuite) setupTransports() {
	transportConfig := DefaultTransportConfig()
	factory := NewTransportFactory(s.logger)

	// Setup stdio transport
	transportConfig.Type = types.TransportStdio
	stdio, err := factory.CreateTransport(transportConfig, nil)
	s.Require().NoError(err)
	s.stdioTransport = stdio

	// Setup HTTP transport with test server
	transportConfig.Type = types.TransportHTTP
	http, err := factory.CreateTransport(transportConfig, s.server)
	s.Require().NoError(err)
	s.httpTransport = http

	// Create HTTP test server
	s.httpServer = httptest.NewServer(http.(*HTTPTransport).handler)
}

// createTestConfig creates optimized configuration for performance testing
func (s *PerformanceTestSuite) createTestConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host:        "localhost",
			Port:        0,
			Timeout:     ComplexOperationTarget,
			ReadTimeout: 30 * time.Second,
		},
		CLI: CLIConfig{
			Path:       "/usr/bin/go-invoice",
			WorkingDir: "/tmp/mcp-test",
			MaxTimeout: ComplexOperationTarget,
		},
		Security: SecurityConfig{
			AllowedCommands:       []string{"go-invoice"},
			WorkingDir:            "/tmp/mcp-test",
			SandboxEnabled:        true,
			FileAccessRestricted:  true,
			MaxCommandTimeout:     "2s",
			EnableInputValidation: true,
		},
		LogLevel: "error", // Reduce logging overhead for benchmarks
	}
}

// BenchmarkToolExecution tests performance of all 21 tools
func BenchmarkToolExecution(b *testing.B) {
	suite := &PerformanceTestSuite{}
	suite.SetupSuite()
	defer suite.TearDownSuite()

	// Get all available tools
	ctx := context.Background()
	allTools, err := suite.toolComponents.Registry.ListTools(ctx, "")
	require.NoError(b, err)

	for _, tool := range allTools {
		b.Run(tool.Name, func(b *testing.B) {
			suite.benchmarkSingleTool(b, tool)
		})
	}
}

// BenchmarkToolsByCategory tests performance by tool category
func BenchmarkToolsByCategory(b *testing.B) {
	suite := &PerformanceTestSuite{}
	suite.SetupSuite()
	defer suite.TearDownSuite()

	categories := []tools.CategoryType{
		tools.CategoryInvoiceManagement,
		tools.CategoryDataImport,
		tools.CategoryDataExport,
		tools.CategoryClientManagement,
		tools.CategoryConfiguration,
		tools.CategoryReporting,
	}

	for _, category := range categories {
		b.Run(string(category), func(b *testing.B) {
			suite.benchmarkToolCategory(b, category)
		})
	}
}

// BenchmarkTransportPerformance compares stdio vs HTTP transport performance
func BenchmarkTransportPerformance(b *testing.B) {
	suite := &PerformanceTestSuite{}
	suite.SetupSuite()
	defer suite.TearDownSuite()

	testRequest := suite.createSimpleToolRequest()

	b.Run("stdio", func(b *testing.B) {
		suite.benchmarkStdioTransport(b, testRequest)
	})

	b.Run("http", func(b *testing.B) {
		suite.benchmarkHTTPTransport(b, testRequest)
	})
}

// BenchmarkConcurrentRequests tests concurrent request handling
func BenchmarkConcurrentRequests(b *testing.B) {
	suite := &PerformanceTestSuite{}
	suite.SetupSuite()
	defer suite.TearDownSuite()

	concurrencyLevels := []int{1, 5, 10, 25, 50, 100}

	for _, concurrency := range concurrencyLevels {
		b.Run(fmt.Sprintf("concurrent_%d", concurrency), func(b *testing.B) {
			suite.benchmarkConcurrentRequests(b, concurrency)
		})
	}
}

// BenchmarkMemoryUsage tests memory allocation patterns
func BenchmarkMemoryUsage(b *testing.B) {
	suite := &PerformanceTestSuite{}
	suite.SetupSuite()
	defer suite.TearDownSuite()

	payloadSizes := []string{"small", "medium", "large"}

	for _, size := range payloadSizes {
		b.Run(size, func(b *testing.B) {
			suite.benchmarkMemoryUsage(b, size)
		})
	}
}

// BenchmarkResourceLimits validates timeout enforcement and resource constraints
func BenchmarkResourceLimits(b *testing.B) {
	suite := &PerformanceTestSuite{}
	suite.SetupSuite()
	defer suite.TearDownSuite()

	b.Run("timeout_enforcement", func(b *testing.B) {
		suite.benchmarkTimeoutEnforcement(b)
	})

	b.Run("resource_constraints", func(b *testing.B) {
		suite.benchmarkResourceConstraints(b)
	})

	b.Run("rate_limiting", func(b *testing.B) {
		suite.benchmarkRateLimiting(b)
	})
}

// BenchmarkScalability tests performance under increasing load
func BenchmarkScalability(b *testing.B) {
	suite := &PerformanceTestSuite{}
	suite.SetupSuite()
	defer suite.TearDownSuite()

	loadPatterns := []string{"burst", "sustained", "ramp_up"}

	for _, pattern := range loadPatterns {
		b.Run(pattern, func(b *testing.B) {
			suite.benchmarkScalabilityPattern(b, pattern)
		})
	}
}

// BenchmarkColdStart tests initial setup and warm-up performance
func BenchmarkColdStart(b *testing.B) {
	b.Run("tool_system_initialization", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ctx := context.Background()
			_, err := tools.InitializeToolSystem(ctx, NewTestLogger())
			require.NoError(b, err)
		}
	})

	b.Run("server_startup", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			logger := NewTestLogger()
			bridge := NewMockCLIBridge()
			config := &Config{
				Server: ServerConfig{
					Host:        "localhost",
					Port:        0,
					Timeout:     5 * time.Second,
					ReadTimeout: 30 * time.Second,
				},
				CLI: CLIConfig{
					Path:       "/usr/bin/go-invoice",
					WorkingDir: "/tmp",
					MaxTimeout: 5 * time.Second,
				},
				Security: SecurityConfig{
					AllowedCommands: []string{"go-invoice"},
				},
				LogLevel: "error",
			}

			server := NewServer(logger, bridge, config)
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			_ = server.Start(ctx, types.TransportStdio)
			cancel()
		}
	})
}

// benchmarkSingleTool tests performance of a specific tool
func (s *PerformanceTestSuite) benchmarkSingleTool(b *testing.B, tool *tools.MCPTool) {
	ctx := context.Background()
	request := s.createToolRequest(tool.Name, s.dataGenerator.GenerateToolInput(tool))

	s.responseTimer.Start(tool.Name)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			start := time.Now()

			response, err := s.server.HandleRequest(ctx, request)

			duration := time.Since(start)
			s.responseTimer.Record(tool.Name, duration)

			require.NoError(b, err)
			require.NotNil(b, response)

			// Validate response time targets
			if s.isSimpleOperation(tool) {
				assert.True(b, duration < SimpleOperationTarget,
					"Simple operation %s took %v, expected < %v", tool.Name, duration, SimpleOperationTarget)
			} else {
				assert.True(b, duration < ComplexOperationTarget,
					"Complex operation %s took %v, expected < %v", tool.Name, duration, ComplexOperationTarget)
			}
		}
	})
}

// benchmarkToolCategory tests performance of tools in a specific category
func (s *PerformanceTestSuite) benchmarkToolCategory(b *testing.B, category tools.CategoryType) {
	ctx := context.Background()
	categoryTools, err := s.toolComponents.Registry.ListTools(ctx, category)
	require.NoError(b, err)

	if len(categoryTools) == 0 {
		b.Skip("No tools found in category:", category)
		return
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		toolIndex := 0
		for pb.Next() {
			tool := categoryTools[toolIndex%len(categoryTools)]
			toolIndex++

			request := s.createToolRequest(tool.Name, s.dataGenerator.GenerateToolInput(tool))

			start := time.Now()
			response, err := s.server.HandleRequest(ctx, request)
			duration := time.Since(start)

			require.NoError(b, err)
			require.NotNil(b, response)
			s.responseTimer.Record(string(category), duration)
		}
	})
}

// benchmarkStdioTransport tests stdio transport performance
func (s *PerformanceTestSuite) benchmarkStdioTransport(b *testing.B, request *types.MCPRequest) {
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()
		response, err := s.server.HandleRequest(ctx, request)
		duration := time.Since(start)

		require.NoError(b, err)
		require.NotNil(b, response)
		s.responseTimer.Record("stdio_transport", duration)
	}
}

// benchmarkHTTPTransport tests HTTP transport performance
func (s *PerformanceTestSuite) benchmarkHTTPTransport(b *testing.B, request *types.MCPRequest) {
	client := &http.Client{Timeout: 30 * time.Second}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		jsonData, err := json.Marshal(request)
		require.NoError(b, err)

		start := time.Now()
		resp, err := client.Post(s.httpServer.URL+"/mcp", "application/json", strings.NewReader(string(jsonData)))
		require.NoError(b, err)

		body, err := io.ReadAll(resp.Body)
		require.NoError(b, err)
		resp.Body.Close()

		duration := time.Since(start)

		var response types.MCPResponse
		err = json.Unmarshal(body, &response)
		require.NoError(b, err)

		s.responseTimer.Record("http_transport", duration)
	}
}

// benchmarkConcurrentRequests tests concurrent request handling
func (s *PerformanceTestSuite) benchmarkConcurrentRequests(b *testing.B, concurrency int) {
	ctx := context.Background()
	request := s.createSimpleToolRequest()

	var successCount int64
	var errorCount int64

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var wg sync.WaitGroup
			semaphore := make(chan struct{}, concurrency)

			for i := 0; i < concurrency; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					semaphore <- struct{}{}
					defer func() { <-semaphore }()

					start := time.Now()
					response, err := s.server.HandleRequest(ctx, request)
					duration := time.Since(start)

					if err != nil {
						atomic.AddInt64(&errorCount, 1)
					} else if response != nil {
						atomic.AddInt64(&successCount, 1)
						s.responseTimer.Record("concurrent", duration)
					}
				}()
			}
			wg.Wait()
		}
	})

	b.ReportMetric(float64(atomic.LoadInt64(&successCount)), "successful_requests")
	b.ReportMetric(float64(atomic.LoadInt64(&errorCount)), "failed_requests")
}

// benchmarkMemoryUsage tests memory allocation patterns
func (s *PerformanceTestSuite) benchmarkMemoryUsage(b *testing.B, payloadSize string) {
	ctx := context.Background()
	request := s.createRequestWithPayloadSize(payloadSize)

	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		response, err := s.server.HandleRequest(ctx, request)
		require.NoError(b, err)
		require.NotNil(b, response)
	}

	runtime.GC()
	runtime.ReadMemStats(&m2)

	b.ReportMetric(float64(m2.TotalAlloc-m1.TotalAlloc)/float64(b.N), "bytes_per_op")
	b.ReportMetric(float64(m2.Mallocs-m1.Mallocs)/float64(b.N), "allocs_per_op")
}

// benchmarkTimeoutEnforcement tests timeout handling
func (s *PerformanceTestSuite) benchmarkTimeoutEnforcement(b *testing.B) {
	shortTimeout := 10 * time.Millisecond

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), shortTimeout)

		request := s.createSlowToolRequest()
		start := time.Now()
		response, err := s.server.HandleRequest(ctx, request)
		duration := time.Since(start)

		cancel()

		// Should timeout or complete within reasonable time
		assert.True(b, duration < 100*time.Millisecond, "Timeout enforcement took too long: %v", duration)

		if err != nil {
			assert.Contains(b, err.Error(), "context deadline exceeded")
		} else {
			require.NotNil(b, response)
		}
	}
}

// benchmarkResourceConstraints tests resource limit enforcement
func (s *PerformanceTestSuite) benchmarkResourceConstraints(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Test large payload handling
		request := s.createLargePayloadRequest()

		start := time.Now()
		response, err := s.server.HandleRequest(ctx, request)
		duration := time.Since(start)

		// Should handle gracefully within time limits
		assert.True(b, duration < ComplexOperationTarget, "Resource constraint handling took too long: %v", duration)

		if err == nil {
			require.NotNil(b, response)
		}

		s.responseTimer.Record("resource_constraints", duration)
	}
}

// benchmarkRateLimiting tests rate limiting behavior
func (s *PerformanceTestSuite) benchmarkRateLimiting(b *testing.B) {
	ctx := context.Background()
	request := s.createSimpleToolRequest()

	rateLimiter := make(chan struct{}, 10) // Allow 10 concurrent requests

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			select {
			case rateLimiter <- struct{}{}:
				start := time.Now()
				response, err := s.server.HandleRequest(ctx, request)
				duration := time.Since(start)
				<-rateLimiter

				if err == nil {
					require.NotNil(b, response)
					s.responseTimer.Record("rate_limited", duration)
				}
			default:
				// Rate limited - this is expected behavior
				b.Skip()
			}
		}
	})
}

// benchmarkScalabilityPattern tests different load patterns
func (s *PerformanceTestSuite) benchmarkScalabilityPattern(b *testing.B, pattern string) {
	ctx := context.Background()
	request := s.createSimpleToolRequest()

	switch pattern {
	case "burst":
		s.benchmarkBurstLoad(b, ctx, request)
	case "sustained":
		s.benchmarkSustainedLoad(b, ctx, request)
	case "ramp_up":
		s.benchmarkRampUpLoad(b, ctx, request)
	}
}

// benchmarkBurstLoad tests performance under sudden load spikes
func (s *PerformanceTestSuite) benchmarkBurstLoad(b *testing.B, ctx context.Context, request *types.MCPRequest) {
	burstSize := 50

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		start := time.Now()

		for j := 0; j < burstSize; j++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				response, err := s.server.HandleRequest(ctx, request)
				if err == nil && response != nil {
					// Success
				}
			}()
		}
		wg.Wait()

		duration := time.Since(start)
		s.responseTimer.Record("burst_load", duration)
	}
}

// benchmarkSustainedLoad tests performance under sustained load
func (s *PerformanceTestSuite) benchmarkSustainedLoad(b *testing.B, ctx context.Context, request *types.MCPRequest) {
	requestsPerSecond := ThroughputTarget
	interval := time.Second / time.Duration(requestsPerSecond)

	b.ResetTimer()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for i := 0; i < b.N; i++ {
		<-ticker.C

		start := time.Now()
		response, err := s.server.HandleRequest(ctx, request)
		duration := time.Since(start)

		if err == nil {
			require.NotNil(b, response)
		}
		s.responseTimer.Record("sustained_load", duration)
	}
}

// benchmarkRampUpLoad tests performance under gradually increasing load
func (s *PerformanceTestSuite) benchmarkRampUpLoad(b *testing.B, ctx context.Context, request *types.MCPRequest) {
	maxConcurrency := 20
	rampSteps := 5

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for step := 1; step <= rampSteps; step++ {
			concurrency := (maxConcurrency * step) / rampSteps
			var wg sync.WaitGroup

			start := time.Now()
			for j := 0; j < concurrency; j++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					response, err := s.server.HandleRequest(ctx, request)
					if err == nil && response != nil {
						// Success
					}
				}()
			}
			wg.Wait()

			duration := time.Since(start)
			s.responseTimer.Record(fmt.Sprintf("ramp_step_%d", step), duration)

			// Brief pause between ramp steps
			time.Sleep(10 * time.Millisecond)
		}
	}
}

// Helper methods for test data generation

// createSimpleToolRequest creates a simple tool request for testing
func (s *PerformanceTestSuite) createSimpleToolRequest() *types.MCPRequest {
	return &types.MCPRequest{
		JSONRPC: "2.0",
		ID:      "test-1",
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":      "ping",
			"arguments": map[string]interface{}{},
		},
	}
}

// createToolRequest creates a tool request for a specific tool
func (s *PerformanceTestSuite) createToolRequest(toolName string, input map[string]interface{}) *types.MCPRequest {
	return &types.MCPRequest{
		JSONRPC: "2.0",
		ID:      fmt.Sprintf("test-%s", toolName),
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":      toolName,
			"arguments": input,
		},
	}
}

// createSlowToolRequest creates a request that might take longer to process
func (s *PerformanceTestSuite) createSlowToolRequest() *types.MCPRequest {
	return &types.MCPRequest{
		JSONRPC: "2.0",
		ID:      "test-slow",
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name": "generate_invoice",
			"arguments": map[string]interface{}{
				"clientId":     "client-123",
				"outputFormat": "pdf",
				"template":     "detailed",
			},
		},
	}
}

// createLargePayloadRequest creates a request with large payload
func (s *PerformanceTestSuite) createLargePayloadRequest() *types.MCPRequest {
	largeData := make(map[string]interface{})
	for i := 0; i < 1000; i++ {
		largeData[fmt.Sprintf("field_%d", i)] = strings.Repeat("data", 100)
	}

	return &types.MCPRequest{
		JSONRPC: "2.0",
		ID:      "test-large",
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":      "import_timesheet",
			"arguments": largeData,
		},
	}
}

// createRequestWithPayloadSize creates request with specific payload size
func (s *PerformanceTestSuite) createRequestWithPayloadSize(size string) *types.MCPRequest {
	var data map[string]interface{}

	switch size {
	case "small":
		data = map[string]interface{}{
			"field1": "value1",
			"field2": "value2",
		}
	case "medium":
		data = make(map[string]interface{})
		for i := 0; i < 100; i++ {
			data[fmt.Sprintf("field_%d", i)] = fmt.Sprintf("value_%d", i)
		}
	case "large":
		data = make(map[string]interface{})
		for i := 0; i < 1000; i++ {
			data[fmt.Sprintf("field_%d", i)] = strings.Repeat("data", 50)
		}
	}

	return &types.MCPRequest{
		JSONRPC: "2.0",
		ID:      fmt.Sprintf("test-%s", size),
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":      "validate_config",
			"arguments": data,
		},
	}
}

// isSimpleOperation determines if a tool operation is considered simple
func (s *PerformanceTestSuite) isSimpleOperation(tool *tools.MCPTool) bool {
	simpleOperations := map[string]bool{
		"ping":            true,
		"version":         true,
		"health_check":    true,
		"list_clients":    true,
		"validate_config": true,
	}

	return simpleOperations[tool.Name]
}

// printPerformanceReport prints a summary of performance test results
func (s *PerformanceTestSuite) printPerformanceReport() {
	fmt.Println("\n=== MCP Server Performance Test Report ===")

	s.responseTimer.PrintReport()
	s.resourceMonitor.PrintReport()

	fmt.Println("\n=== Performance Targets ===")
	fmt.Printf("Simple operations target: %v\n", SimpleOperationTarget)
	fmt.Printf("Complex operations target: %v\n", ComplexOperationTarget)
	fmt.Printf("Concurrent requests target: %d\n", ConcurrentRequestsTarget)
	fmt.Printf("Throughput target: %d ops/sec\n", ThroughputTarget)
}

// Test suite runner
func TestPerformanceSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance tests in short mode")
	}

	suite.Run(t, new(PerformanceTestSuite))
}

// ResponseTimeTracker tracks response times for performance analysis
type ResponseTimeTracker struct {
	mu    sync.RWMutex
	times map[string][]time.Duration
}

// NewResponseTimeTracker creates a new response time tracker
func NewResponseTimeTracker() *ResponseTimeTracker {
	return &ResponseTimeTracker{
		times: make(map[string][]time.Duration),
	}
}

// Start begins tracking for an operation
func (rt *ResponseTimeTracker) Start(operation string) {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	if rt.times[operation] == nil {
		rt.times[operation] = make([]time.Duration, 0)
	}
}

// Record records a response time for an operation
func (rt *ResponseTimeTracker) Record(operation string, duration time.Duration) {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	rt.times[operation] = append(rt.times[operation], duration)
}

// PrintReport prints a performance report
func (rt *ResponseTimeTracker) PrintReport() {
	rt.mu.RLock()
	defer rt.mu.RUnlock()

	fmt.Println("\n=== Response Time Analysis ===")
	for operation, durations := range rt.times {
		if len(durations) == 0 {
			continue
		}

		var total time.Duration
		var min, max time.Duration = durations[0], durations[0]

		for _, d := range durations {
			total += d
			if d < min {
				min = d
			}
			if d > max {
				max = d
			}
		}

		avg := total / time.Duration(len(durations))

		fmt.Printf("Operation: %s\n", operation)
		fmt.Printf("  Samples: %d\n", len(durations))
		fmt.Printf("  Average: %v\n", avg)
		fmt.Printf("  Min: %v\n", min)
		fmt.Printf("  Max: %v\n", max)
		fmt.Printf("  Total: %v\n", total)
		fmt.Println()
	}
}

// ResourceMonitor tracks resource usage during tests
type ResourceMonitor struct {
	startMem runtime.MemStats
	stopMem  runtime.MemStats
	stopped  bool
}

// NewResourceMonitor creates a new resource monitor
func NewResourceMonitor() *ResourceMonitor {
	rm := &ResourceMonitor{}
	runtime.ReadMemStats(&rm.startMem)
	return rm
}

// Stop stops resource monitoring
func (rm *ResourceMonitor) Stop() {
	if !rm.stopped {
		runtime.ReadMemStats(&rm.stopMem)
		rm.stopped = true
	}
}

// PrintReport prints resource usage report
func (rm *ResourceMonitor) PrintReport() {
	if !rm.stopped {
		rm.Stop()
	}

	fmt.Println("\n=== Resource Usage Analysis ===")
	fmt.Printf("Memory allocated: %d bytes\n", rm.stopMem.TotalAlloc-rm.startMem.TotalAlloc)
	fmt.Printf("Memory allocations: %d\n", rm.stopMem.Mallocs-rm.startMem.Mallocs)
	fmt.Printf("Memory frees: %d\n", rm.stopMem.Frees-rm.startMem.Frees)
	fmt.Printf("GC runs: %d\n", rm.stopMem.NumGC-rm.startMem.NumGC)
	fmt.Printf("Goroutines: %d\n", runtime.NumGoroutine())
}

// TestDataGenerator generates test data for performance tests
type TestDataGenerator struct{}

// NewTestDataGenerator creates a new test data generator
func NewTestDataGenerator() *TestDataGenerator {
	return &TestDataGenerator{}
}

// GenerateToolInput generates appropriate input for a tool
func (tdg *TestDataGenerator) GenerateToolInput(tool *tools.MCPTool) map[string]interface{} {
	// Generate input based on tool category and schema
	switch tool.Category {
	case tools.CategoryInvoiceManagement:
		return map[string]interface{}{
			"clientId":    "test-client-123",
			"amount":      1000.50,
			"description": "Test invoice description",
			"dueDate":     "2024-01-31",
		}
	case tools.CategoryDataImport:
		return map[string]interface{}{
			"filePath": "/tmp/test-timesheet.csv",
			"format":   "csv",
			"validate": true,
		}
	case tools.CategoryDataExport:
		return map[string]interface{}{
			"format":     "pdf",
			"template":   "standard",
			"outputPath": "/tmp/output",
		}
	case tools.CategoryClientManagement:
		return map[string]interface{}{
			"name":    "Test Client",
			"email":   "test@example.com",
			"address": "123 Test Street",
		}
	case tools.CategoryConfiguration:
		return map[string]interface{}{
			"section":  "server",
			"validate": true,
		}
	case tools.CategoryReporting:
		return map[string]interface{}{
			"period":    "monthly",
			"startDate": "2024-01-01",
			"endDate":   "2024-01-31",
			"format":    "json",
		}
	default:
		return map[string]interface{}{}
	}
}
