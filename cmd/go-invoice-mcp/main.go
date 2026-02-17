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

	"github.com/mrz1836/go-invoice/internal/mcp"
)

func main() {
	// Configure standard log to not include timestamps (MCP logger handles this)
	log.SetFlags(0)

	// Check for version flag first
	for _, arg := range os.Args {
		if arg == "--version" || arg == "-v" {
			log.Println("go-invoice-mcp version 2.0.0")
			os.Exit(0)
		}
		if arg == "--help" || arg == "-h" {
			printHelp()
			os.Exit(0)
		}
	}

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

	// Create production handler with all tools registered
	handler, err := mcp.CreateProductionHandler(config)
	if err != nil {
		log.Fatalf("Failed to create production handler: %v", err)
	}

	// Create logger for server
	logger := mcp.NewLogger(config.LogLevel)

	// Create MCP server with production handler
	server := mcp.NewServerWithHandler(logger, handler, config)

	logger.Info("Starting MCP server", "transport", transport, "version", "2.0.0")
	logger.Info("Configuration loaded", "logLevel", config.LogLevel, "cliPath", config.CLI.Path)
	logger.Info("Environment", "GO_INVOICE_HOME", os.Getenv("GO_INVOICE_HOME"), "MCP_LOG_LEVEL", os.Getenv("MCP_LOG_LEVEL"), "MCP_LOG_FILE", os.Getenv("MCP_LOG_FILE"))

	// Start server based on transport type
	logger.Info("Starting server with transport", "transport", transport)
	if err := server.Start(ctx, transport); err != nil {
		logger.Error("Failed to start MCP server", "error", err)
		os.Exit(1)
	}
	logger.Info("Server started successfully")

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

// printHelp displays usage information
func printHelp() {
	log.Println("go-invoice-mcp - MCP server for go-invoice CLI integration")
	log.Println()
	log.Println("Usage: go-invoice-mcp [options]")
	log.Println()
	log.Println("Options:")
	log.Println("  --stdio      Use stdio transport (default)")
	log.Println("  --http       Use HTTP transport")
	log.Println("  --config     Path to MCP configuration file")
	log.Println("  --version    Show version information")
	log.Println("  --help       Show this help message")
	log.Println()
	log.Println("Environment variables:")
	log.Println("  MCP_TRANSPORT     Transport type (stdio or http)")
	log.Println("  MCP_LOG_LEVEL     Log level (debug, info, warn, error)")
	log.Println("  MCP_LOG_FILE      Path to log file")
	log.Println("  GO_INVOICE_HOME   Path to go-invoice home directory")
}
