package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"sync"
)

// Static errors for err113 compliance
var (
	ErrUnsupportedTransport = errors.New("unsupported transport type")
)

// DefaultServer implements the Server interface with MCP protocol support
type DefaultServer struct {
	logger   Logger
	bridge   CLIBridge
	config   *Config
	handler  MCPHandler
	listener net.Listener
	server   *http.Server
	wg       sync.WaitGroup
	shutdown chan struct{}
}

// NewServer creates a new MCP server with dependency injection
func NewServer(logger Logger, bridge CLIBridge, config *Config) Server {
	handler := NewMCPHandler(logger, bridge, config)
	return &DefaultServer{
		logger:   logger,
		bridge:   bridge,
		config:   config,
		handler:  handler,
		shutdown: make(chan struct{}),
	}
}

// NewServerWithHandler creates a new MCP server with a custom handler
func NewServerWithHandler(logger Logger, handler MCPHandler, config *Config) Server {
	return &DefaultServer{
		logger:   logger,
		handler:  handler,
		config:   config,
		shutdown: make(chan struct{}),
	}
}

// Start starts the MCP server with the specified transport
func (s *DefaultServer) Start(ctx context.Context, transport TransportType) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	s.logger.Info("Starting MCP server", "transport", transport)

	switch transport {
	case TransportStdio:
		return s.startStdioTransport(ctx)
	case TransportHTTP:
		return s.startHTTPTransport(ctx)
	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedTransport, transport)
	}
}

// Shutdown gracefully shuts down the MCP server
func (s *DefaultServer) Shutdown(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	s.logger.Info("Shutting down MCP server")

	close(s.shutdown)

	if s.server != nil {
		if err := s.server.Shutdown(ctx); err != nil {
			return fmt.Errorf("HTTP server shutdown failed: %w", err)
		}
	}

	if s.listener != nil {
		if err := s.listener.Close(); err != nil {
			s.logger.Warn("Error closing listener", "error", err)
		}
	}

	// Wait for all goroutines to finish
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		s.logger.Info("MCP server shutdown complete")
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// startStdioTransport starts the server using stdio transport for Claude Code
func (s *DefaultServer) startStdioTransport(ctx context.Context) error {
	s.logger.Info("Starting stdio transport")

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.handleStdioRequests(ctx)
	}()

	return nil
}

// startHTTPTransport starts the server using HTTP transport for Claude Desktop
func (s *DefaultServer) startHTTPTransport(ctx context.Context) error {
	s.logger.Info("Starting HTTP transport", "host", s.config.Server.Host, "port", s.config.Server.Port)

	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", s.handleHTTPRequest)

	s.server = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port),
		Handler:      mux,
		ReadTimeout:  s.config.Server.ReadTimeout,
		WriteTimeout: s.config.Server.Timeout,
	}

	lc := &net.ListenConfig{}
	listener, err := lc.Listen(ctx, "tcp", s.server.Addr)
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}
	s.listener = listener

	actualAddr := listener.Addr().String()
	s.logger.Info("HTTP server listening", "address", actualAddr)

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := s.server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("HTTP server error", "error", err)
		}
	}()

	return nil
}

// handleStdioRequests handles MCP requests over stdio
func (s *DefaultServer) handleStdioRequests(ctx context.Context) {
	decoder := json.NewDecoder(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.shutdown:
			return
		default:
		}

		// Parse and handle request using JSON decoder
		var req MCPRequest
		if parseErr := decoder.Decode(&req); parseErr != nil {
			if errors.Is(parseErr, io.EOF) {
				s.logger.Debug("Stdin closed, shutting down")
				return
			}
			s.logger.Error("Failed to parse MCP request", "error", parseErr)
			continue
		}

		response, err := s.HandleRequest(ctx, &req)
		if err != nil {
			s.logger.Error("Failed to handle MCP request", "error", err, "method", req.Method)
			response = &MCPResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error: &MCPError{
					Code:    -32603,
					Message: "Internal error",
					Data:    err.Error(),
				},
			}
		}

		// Send response using JSON encoder (skip if nil for notifications)
		if response != nil {
			if err := encoder.Encode(response); err != nil {
				s.logger.Error("Failed to encode response", "error", err)
				return
			}
		}
	}
}

// handleHTTPRequest handles MCP requests over HTTP
func (s *DefaultServer) handleHTTPRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.logger.Error("Failed to read request body", "error", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	var req MCPRequest
	if parseErr := json.Unmarshal(body, &req); parseErr != nil {
		s.logger.Error("Failed to parse MCP request", "error", parseErr)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	response, err := s.HandleRequest(ctx, &req)
	if err != nil {
		s.logger.Error("Failed to handle MCP request", "error", err, "method", req.Method)
		response = &MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &MCPError{
				Code:    -32603,
				Message: "Internal error",
				Data:    err.Error(),
			},
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.logger.Error("Failed to write response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// HandleRequest processes an MCP request and returns a response
func (s *DefaultServer) HandleRequest(ctx context.Context, req *MCPRequest) (*MCPResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	s.logger.Debug("Handling MCP request", "method", req.Method, "id", req.ID)

	switch req.Method {
	case "initialize":
		return s.handler.HandleInitialize(ctx, req)
	case "notifications/initialized":
		// Handle initialized notification (no response needed for notifications)
		s.logger.Debug("Received initialized notification")
		return nil, nil
	case "ping":
		return s.handler.HandlePing(ctx, req)
	case "tools/list":
		return s.handler.HandleToolsList(ctx, req)
	case "tools/call":
		return s.handler.HandleToolCall(ctx, req)
	default:
		return &MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &MCPError{
				Code:    -32601,
				Message: "Method not found",
				Data:    fmt.Sprintf("Unknown method: %s", req.Method),
			},
		}, nil
	}
}
