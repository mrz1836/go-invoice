// Package main provides the MCP server executable for go-invoice.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mrz1836/go-invoice/internal/mcp"
)

var (
	version = "1.0.0"
	commit  = "dev"     //nolint:gochecknoglobals // build-time variable set via ldflags
	date    = "unknown" //nolint:gochecknoglobals // build-time variable set via ldflags

	// ErrServerUnhealthy represents a server health check failure
	ErrServerUnhealthy = errors.New("server is unhealthy")
)

func main() {
	var (
		stdio       = flag.Bool("stdio", false, "Use stdio transport")
		httpFlag    = flag.Bool("http", false, "Use HTTP transport")
		logLevel    = flag.String("log-level", "", "Log level (debug, info, warn, error)")
		versionFlag = flag.Bool("version", false, "Show version information")
		healthCheck = flag.Bool("health", false, "Perform health check and exit")
	)
	flag.Parse()

	// Show version information
	if *versionFlag {
		log.Printf("go-invoice-mcp %s", version)
		log.Printf("Commit: %s", commit)
		log.Printf("Built: %s", date)
		return
	}

	// Load configuration
	ctx := context.Background()
	config, err := mcp.LoadConfig(ctx)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Override log level if specified
	if *logLevel != "" {
		config.LogLevel = *logLevel
	}

	// Create logger
	logger := mcp.NewLogger(config.LogLevel)

	// Determine transport type
	transportType := determineTransportType(*stdio, *httpFlag)

	// Perform health check if requested
	if *healthCheck {
		if err := performHealthCheck(config, logger); err != nil {
			fmt.Fprintf(os.Stderr, "Health check failed: %v\n", err)
			os.Exit(1)
		}
		log.Println("Health check passed")
		return
	}

	// Create CLI bridge components
	validator := mcp.NewCommandValidator(config.Security.AllowedCommands)
	fileHandler := mcp.NewFileHandler(config.Security.WorkingDir)
	bridge := mcp.NewCLIBridge(logger, validator, fileHandler, config.CLI)

	// Create server
	server := mcp.NewServer(logger, bridge, config)

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		logger.Info("Received shutdown signal, stopping server...")
		cancel()
	}()

	logger.Info("MCP server starting",
		"transport", transportType,
		"version", version,
	)

	// Start server with detected transport
	if err := server.Start(ctx, transportType); err != nil {
		log.Fatalf("Server failed: %v", err)
	}

	logger.Info("Server stopped")
}

func determineTransportType(stdio, http bool) mcp.TransportType {
	// Explicit flags take precedence
	if stdio {
		return mcp.TransportStdio
	}
	if http {
		return mcp.TransportHTTP
	}

	// Use detection
	return mcp.DetectTransport()
}

func performHealthCheck(config *mcp.Config, logger mcp.Logger) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create a simple health checker
	healthChecker := mcp.NewHealthChecker(logger, config)

	status, err := healthChecker.CheckHealth(ctx)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	if status.Status != "healthy" {
		return fmt.Errorf("%w: %s", ErrServerUnhealthy, status.LastError)
	}

	return nil
}
