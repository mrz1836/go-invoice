package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// Transport errors
var (
	ErrTransportNotInitialized = errors.New("transport not initialized")
	ErrInvalidTransportType    = errors.New("invalid transport type")
	ErrTransportClosed         = errors.New("transport closed")
	ErrMessageTooLarge         = errors.New("message exceeds size limit")
)

// TransportType represents the type of transport
type TransportType string

const (
	// TransportStdio uses standard input/output
	TransportStdio TransportType = "stdio"
	// TransportHTTP uses HTTP server
	TransportHTTP TransportType = "http"
)

// Transport defines the interface for MCP communication
type Transport interface {
	// Type returns the transport type
	Type() TransportType

	// Start initializes the transport
	Start(ctx context.Context) error

	// Stop gracefully shuts down the transport
	Stop(ctx context.Context) error

	// Receive reads the next message
	Receive(ctx context.Context) (*MCPRequest, error)

	// Send writes a response
	Send(ctx context.Context, response *MCPResponse) error

	// IsHealthy checks transport health
	IsHealthy(ctx context.Context) bool
}

// TransportConfig holds configuration for transports
type TransportConfig struct {
	Type           TransportType `json:"type"`
	Host           string        `json:"host,omitempty"`
	Port           int           `json:"port,omitempty"`
	ReadTimeout    time.Duration `json:"readTimeout"`
	WriteTimeout   time.Duration `json:"writeTimeout"`
	MaxMessageSize int64         `json:"maxMessageSize"`
	EnableLogging  bool          `json:"enableLogging"`
	LogLevel       string        `json:"logLevel"`
	MetricsEnabled bool          `json:"metricsEnabled"`
}

// DefaultTransportConfig returns default transport configuration
func DefaultTransportConfig() *TransportConfig {
	return &TransportConfig{
		Type:           TransportStdio,
		Host:           "localhost",
		Port:           0, // Auto-assign
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxMessageSize: 10 * 1024 * 1024, // 10MB
		EnableLogging:  true,
		LogLevel:       "info",
	}
}

// StdioTransport implements stdio-based transport for Claude Code
type StdioTransport struct {
	logger  Logger
	config  *TransportConfig
	reader  *bufio.Reader
	writer  *bufio.Writer
	encoder *json.Encoder
	decoder *json.Decoder
	mu      sync.Mutex
	closed  bool
	metrics *TransportMetrics
}

// NewStdioTransport creates a new stdio transport
func NewStdioTransport(logger Logger, config *TransportConfig) *StdioTransport {
	if logger == nil {
		panic("logger is required")
	}
	if config == nil {
		config = DefaultTransportConfig()
	}

	return &StdioTransport{
		logger:  logger,
		config:  config,
		reader:  bufio.NewReader(os.Stdin),
		writer:  bufio.NewWriter(os.Stdout),
		metrics: NewTransportMetrics(),
	}
}

// Type returns the transport type
func (t *StdioTransport) Type() TransportType {
	return TransportStdio
}

// Start initializes the stdio transport
func (t *StdioTransport) Start(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return ErrTransportClosed
	}

	// Set up JSON encoder/decoder
	t.encoder = json.NewEncoder(t.writer)
	t.decoder = json.NewDecoder(t.reader)

	// Configure decoder for streaming
	t.decoder.UseNumber()

	t.logger.Info("stdio transport started",
		"maxMessageSize", t.config.MaxMessageSize,
	)

	return nil
}

// Stop gracefully shuts down the stdio transport
func (t *StdioTransport) Stop(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return nil
	}

	t.closed = true

	// Flush any buffered output
	if err := t.writer.Flush(); err != nil {
		t.logger.Warn("failed to flush output buffer", "error", err)
	}

	t.logger.Info("stdio transport stopped",
		"messagesReceived", t.metrics.MessagesReceived(),
		"messagesSent", t.metrics.MessagesSent(),
	)

	return nil
}

// Receive reads the next message from stdin
func (t *StdioTransport) Receive(ctx context.Context) (*MCPRequest, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	t.mu.Lock()
	if t.closed {
		t.mu.Unlock()
		return nil, ErrTransportClosed
	}
	t.mu.Unlock()

	// Read with timeout
	type result struct {
		req *MCPRequest
		err error
	}

	resultCh := make(chan result, 1)
	go func() {
		var req MCPRequest
		err := t.decoder.Decode(&req)
		resultCh <- result{&req, err}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case r := <-resultCh:
		if r.err != nil {
			if r.err == io.EOF {
				return nil, ErrTransportClosed
			}
			return nil, fmt.Errorf("failed to decode request: %w", r.err)
		}

		t.metrics.RecordReceive(1)
		t.logger.Debug("received stdio message",
			"method", r.req.Method,
			"id", r.req.ID,
		)

		return r.req, nil
	}
}

// Send writes a response to stdout
func (t *StdioTransport) Send(ctx context.Context, response *MCPResponse) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return ErrTransportClosed
	}

	// Encode response
	if err := t.encoder.Encode(response); err != nil {
		return fmt.Errorf("failed to encode response: %w", err)
	}

	// Flush to ensure immediate delivery
	if err := t.writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush response: %w", err)
	}

	t.metrics.RecordSend(1)
	t.logger.Debug("sent stdio message",
		"id", response.ID,
		"hasError", response.Error != nil,
	)

	return nil
}

// IsHealthy checks if the stdio transport is healthy
func (t *StdioTransport) IsHealthy(ctx context.Context) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	return !t.closed
}

// HTTPTransport implements HTTP-based transport for Claude Desktop
type HTTPTransport struct {
	logger   Logger
	config   *TransportConfig
	server   *http.Server
	handler  http.Handler
	mu       sync.RWMutex
	closed   bool
	ready    bool
	metrics  *TransportMetrics
	requests chan *httpRequest
}

type httpRequest struct {
	req      *MCPRequest
	respChan chan *MCPResponse
	errChan  chan error
}

// NewHTTPTransport creates a new HTTP transport
func NewHTTPTransport(logger Logger, config *TransportConfig, handler MCPHandler) *HTTPTransport {
	if logger == nil {
		panic("logger is required")
	}
	if config == nil {
		config = DefaultTransportConfig()
		config.Type = TransportHTTP
	}
	if handler == nil {
		panic("handler is required for HTTP transport")
	}

	transport := &HTTPTransport{
		logger:   logger,
		config:   config,
		metrics:  NewTransportMetrics(),
		requests: make(chan *httpRequest, 100),
	}

	// Create HTTP handler
	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", transport.handleMCPRequest)
	mux.HandleFunc("/health", transport.handleHealthCheck)

	transport.handler = mux

	return transport
}

// Type returns the transport type
func (t *HTTPTransport) Type() TransportType {
	return TransportHTTP
}

// Start initializes the HTTP transport
func (t *HTTPTransport) Start(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return ErrTransportClosed
	}

	addr := fmt.Sprintf("%s:%d", t.config.Host, t.config.Port)
	t.server = &http.Server{
		Addr:         addr,
		Handler:      t.handler,
		ReadTimeout:  t.config.ReadTimeout,
		WriteTimeout: t.config.WriteTimeout,
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}

	// Start server in background
	go func() {
		t.logger.Info("starting HTTP transport", "addr", addr)
		if err := t.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.logger.Error("HTTP server error", "error", err)
		}
	}()

	// Wait for server to be ready
	time.Sleep(100 * time.Millisecond)
	t.ready = true

	t.logger.Info("HTTP transport started",
		"host", t.config.Host,
		"port", t.config.Port,
	)

	return nil
}

// Stop gracefully shuts down the HTTP transport
func (t *HTTPTransport) Stop(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return nil
	}

	t.closed = true
	t.ready = false

	// Shutdown server
	if t.server != nil {
		shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		if err := t.server.Shutdown(shutdownCtx); err != nil {
			t.logger.Warn("HTTP server shutdown error", "error", err)
			return err
		}
	}

	// Close request channel
	close(t.requests)

	t.logger.Info("HTTP transport stopped",
		"messagesReceived", t.metrics.MessagesReceived(),
		"messagesSent", t.metrics.MessagesSent(),
	)

	return nil
}

// Receive waits for the next HTTP request
func (t *HTTPTransport) Receive(ctx context.Context) (*MCPRequest, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case req, ok := <-t.requests:
		if !ok {
			return nil, ErrTransportClosed
		}
		return req.req, nil
	}
}

// Send sends a response for the current HTTP request
func (t *HTTPTransport) Send(ctx context.Context, response *MCPResponse) error {
	// For HTTP transport, responses are sent directly in the handler
	// This method is here for interface compliance
	return nil
}

// IsHealthy checks if the HTTP transport is healthy
func (t *HTTPTransport) IsHealthy(ctx context.Context) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return !t.closed && t.ready
}

// handleMCPRequest handles incoming MCP requests
func (t *HTTPTransport) handleMCPRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check content type
	contentType := r.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/json") {
		http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	// Limit request size
	r.Body = http.MaxBytesReader(w, r.Body, t.config.MaxMessageSize)

	// Decode request
	var req MCPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		t.logger.Error("failed to decode request", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	t.metrics.RecordReceive(1)
	t.logger.Debug("received HTTP request",
		"method", req.Method,
		"id", req.ID,
	)

	// Process request through channel
	httpReq := &httpRequest{
		req:      &req,
		respChan: make(chan *MCPResponse, 1),
		errChan:  make(chan error, 1),
	}

	select {
	case t.requests <- httpReq:
		// Wait for response
		select {
		case resp := <-httpReq.respChan:
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(resp); err != nil {
				t.logger.Error("failed to encode response", "error", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			t.metrics.RecordSend(1)
		case err := <-httpReq.errChan:
			t.logger.Error("request processing error", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		case <-r.Context().Done():
			http.Error(w, "Request timeout", http.StatusRequestTimeout)
		}
	default:
		http.Error(w, "Server busy", http.StatusServiceUnavailable)
	}
}

// handleHealthCheck handles health check requests
func (t *HTTPTransport) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	t.mu.RLock()
	healthy := !t.closed && t.ready
	t.mu.RUnlock()

	status := map[string]interface{}{
		"status":           "ok",
		"transport":        "http",
		"ready":            healthy,
		"messagesReceived": t.metrics.MessagesReceived(),
		"messagesSent":     t.metrics.MessagesSent(),
		"uptime":           t.metrics.Uptime(),
	}

	w.Header().Set("Content-Type", "application/json")
	if !healthy {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	json.NewEncoder(w).Encode(status)
}

// TransportFactory creates transports based on configuration
type TransportFactory struct {
	logger Logger
}

// NewTransportFactory creates a new transport factory
func NewTransportFactory(logger Logger) *TransportFactory {
	if logger == nil {
		panic("logger is required")
	}
	return &TransportFactory{logger: logger}
}

// CreateTransport creates a transport based on configuration
func (f *TransportFactory) CreateTransport(config *TransportConfig, handler MCPHandler) (Transport, error) {
	switch config.Type {
	case TransportStdio:
		return NewStdioTransport(f.logger, config), nil
	case TransportHTTP:
		if handler == nil {
			return nil, fmt.Errorf("handler required for HTTP transport")
		}
		return NewHTTPTransport(f.logger, config, handler), nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrInvalidTransportType, config.Type)
	}
}

// DetectTransport detects the appropriate transport type
func DetectTransport() TransportType {
	// Check command line arguments
	for _, arg := range os.Args[1:] {
		switch arg {
		case "--stdio":
			return TransportStdio
		case "--http":
			return TransportHTTP
		}
	}

	// Check environment variable
	if transport := os.Getenv("MCP_TRANSPORT"); transport != "" {
		return TransportType(strings.ToLower(transport))
	}

	// Check if running in a terminal
	if fileInfo, _ := os.Stdin.Stat(); (fileInfo.Mode() & os.ModeCharDevice) == 0 {
		// Input is from pipe, likely stdio
		return TransportStdio
	}

	// Default to stdio for Claude Code compatibility
	return TransportStdio
}

// TransportMetrics tracks transport performance
type TransportMetrics struct {
	startTime        time.Time
	messagesReceived uint64
	messagesSent     uint64
	mu               sync.RWMutex
}

// NewTransportMetrics creates new transport metrics
func NewTransportMetrics() *TransportMetrics {
	return &TransportMetrics{
		startTime: time.Now(),
	}
}

// RecordReceive records a received message
func (m *TransportMetrics) RecordReceive(count uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messagesReceived += count
}

// RecordSend records a sent message
func (m *TransportMetrics) RecordSend(count uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messagesSent += count
}

// MessagesReceived returns total messages received
func (m *TransportMetrics) MessagesReceived() uint64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.messagesReceived
}

// MessagesSent returns total messages sent
func (m *TransportMetrics) MessagesSent() uint64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.messagesSent
}

// Uptime returns transport uptime
func (m *TransportMetrics) Uptime() time.Duration {
	return time.Since(m.startTime)
}
