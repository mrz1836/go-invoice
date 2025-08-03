// Package main provides the MCP (Model Context Protocol) server for go-invoice CLI integration.
// This server enables external tools to interact with the go-invoice CLI through a standardized protocol.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mrz/go-invoice/internal/mcp"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received shutdown signal, gracefully shutting down...")
		cancel()
	}()

	// Parse command line arguments for transport detection
	transport := detectTransport()

	// Load configuration
	config, err := mcp.LoadConfig(ctx)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create logger with dependency injection
	logger := mcp.NewLogger(config.LogLevel)

	// Create CLI bridge with dependency injection
	validator := mcp.NewCommandValidator(config.Security.AllowedCommands)
	fileHandler := mcp.NewFileHandler(config.Security.WorkingDir)
	bridge := mcp.NewCLIBridge(logger, validator, fileHandler, config.CLI)

	// Create MCP server with dependency injection
	server := mcp.NewServer(logger, bridge, config)

	logger.Info("Starting MCP server", "transport", transport, "version", "1.0.0")

	// Start server based on transport type
	if err := server.Start(ctx, transport); err != nil {
		logger.Error("Failed to start MCP server", "error", err)
		os.Exit(1)
	}

	// Wait for context cancellation
	<-ctx.Done()

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Error during shutdown", "error", err)
		os.Exit(1)
	}

	logger.Info("MCP server shutdown complete")
}

// detectTransport determines whether to use stdio or HTTP transport based on command line args and environment
func detectTransport() mcp.TransportType {
	for _, arg := range os.Args {
		switch arg {
		case "--stdio":
			return mcp.TransportStdio
		case "--http":
			return mcp.TransportHTTP
		}
	}

	// Check environment variable
	if transport := os.Getenv("MCP_TRANSPORT"); transport != "" {
		switch transport {
		case "stdio":
			return mcp.TransportStdio
		case "http":
			return mcp.TransportHTTP
		}
	}

	// Default to stdio for Claude Code compatibility
	return mcp.TransportStdio
}
