package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mrz/go-invoice/internal/mcp"
)

var (
	version = "1.0.0"
	commit  = "dev"
	date    = "unknown"
)

func main() {
	var (
		configPath  = flag.String("config", "~/.go-invoice/mcp-config.json", "Path to configuration file")
		stdio       = flag.Bool("stdio", false, "Use stdio transport")
		httpFlag    = flag.Bool("http", false, "Use HTTP transport")
		port        = flag.Int("port", 0, "HTTP port (0 for auto-assign)")
		host        = flag.String("host", "127.0.0.1", "HTTP host")
		logLevel    = flag.String("log-level", "", "Log level (debug, info, warn, error)")
		versionFlag = flag.Bool("version", false, "Show version information")
		healthCheck = flag.Bool("health", false, "Perform health check and exit")
	)
	flag.Parse()

	// Show version information
	if *versionFlag {
		fmt.Printf("go-invoice-mcp %s\n", version)
		fmt.Printf("Commit: %s\n", commit)
		fmt.Printf("Built: %s\n", date)
		return
	}

	// Load configuration
	config, err := mcp.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Override log level if specified
	if *logLevel != "" {
		config.Server.LogLevel = *logLevel
	}

	// Create logger
	logger := mcp.NewDefaultLogger(config.Server.LogLevel)

	// Determine transport type
	transportType := determineTransportType(*stdio, *httpFlag)

	// Perform health check if requested
	if *healthCheck {
		if err := performHealthCheck(config, logger); err != nil {
			fmt.Fprintf(os.Stderr, "Health check failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Health check passed")
		return
	}

	// Create CLI bridge
	bridge := mcp.NewDefaultCLIBridge(logger, config)

	// Create MCP handler
	handler := mcp.NewMCPHandler(logger, bridge, config)

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
	healthChecker := mcp.NewHealthChecker(&mcp.HealthConfig{
		Enabled:  true,
		Interval: "30s",
		Timeout:  "5s",
		Checks: mcp.HealthChecks{
			CLI:     true,
			Storage: true,
			Memory:  true,
		},
		CLIPath:    config.CLI.Path,
		WorkingDir: config.CLI.WorkingDir,
	})

	status, err := healthChecker.CheckHealth(ctx)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	if !status.Healthy {
		return fmt.Errorf("server is unhealthy: %s", status.Message)
	}

	return nil
}
