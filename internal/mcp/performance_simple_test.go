package mcp

import (
	"context"
	"testing"
	"time"

	"github.com/mrz/go-invoice/internal/mcp/types"
	"github.com/stretchr/testify/require"
)

// BenchmarkSimpleServerOperations tests basic server operations without complex tool initialization
func BenchmarkSimpleServerOperations(b *testing.B) {
	// Create minimal configuration for performance testing
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

	logger := NewTestLogger()
	bridge := NewMockCLIBridge()

	// Set up mock bridge to return fast responses
	bridge.SetResponse(&CommandResponse{
		ExitCode: 0,
		Stdout:   `{"status": "ok"}`,
		Stderr:   "",
		Duration: 1 * time.Millisecond,
	}, nil)

	server := NewServer(logger, bridge, config)
	ctx := context.Background()

	// Test simple request handling
	b.Run("ping_request", func(b *testing.B) {
		request := &types.MCPRequest{
			JSONRPC: "2.0",
			ID:      "test-ping",
			Method:  "ping",
			Params:  nil,
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			response, err := server.HandleRequest(ctx, request)
			require.NoError(b, err)
			require.NotNil(b, response)
		}
	})

	b.Run("initialize_request", func(b *testing.B) {
		request := &types.MCPRequest{
			JSONRPC: "2.0",
			ID:      "test-init",
			Method:  "initialize",
			Params: map[string]interface{}{
				"protocolVersion": "2024-11-05",
				"capabilities":    map[string]interface{}{},
				"clientInfo": map[string]interface{}{
					"name":    "test-client",
					"version": "1.0.0",
				},
			},
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			response, err := server.HandleRequest(ctx, request)
			require.NoError(b, err)
			require.NotNil(b, response)
		}
	})
}

// BenchmarkBasicTransportOperations tests transport layer performance
func BenchmarkBasicTransportOperations(b *testing.B) {
	logger := NewTestLogger()
	config := DefaultTransportConfig()
	config.Type = types.TransportStdio

	factory := NewTransportFactory(logger)
	transport, err := factory.CreateTransport(config, nil)
	require.NoError(b, err)

	b.Run("transport_health_check", func(b *testing.B) {
		ctx := context.Background()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			healthy := transport.IsHealthy(ctx)
			require.True(b, healthy)
		}
	})

	b.Run("transport_type_check", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			transportType := transport.Type()
			require.Equal(b, types.TransportStdio, transportType)
		}
	})
}

// BenchmarkConfigOperations tests configuration loading and validation performance
func BenchmarkConfigOperations(b *testing.B) {
	b.Run("default_config_creation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			config := DefaultTransportConfig()
			require.NotNil(b, config)
			require.Equal(b, types.TransportStdio, config.Type)
		}
	})

	b.Run("transport_factory_creation", func(b *testing.B) {
		logger := NewTestLogger()
		for i := 0; i < b.N; i++ {
			factory := NewTransportFactory(logger)
			require.NotNil(b, factory)
		}
	})
}

// BenchmarkMockOperations tests mock component performance for baseline comparison
func BenchmarkMockOperations(b *testing.B) {
	b.Run("mock_cli_bridge_operations", func(b *testing.B) {
		bridge := NewMockCLIBridge()
		ctx := context.Background()

		bridge.SetResponse(&CommandResponse{
			ExitCode: 0,
			Stdout:   "test output",
			Duration: 1 * time.Millisecond,
		}, nil)

		request := &CommandRequest{
			Command: "test-command",
			Args:    []string{"arg1", "arg2"},
			Timeout: 5 * time.Second,
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			response, err := bridge.ExecuteCommand(ctx, request)
			require.NoError(b, err)
			require.NotNil(b, response)
		}
	})

	b.Run("test_logger_operations", func(b *testing.B) {
		logger := NewTestLogger()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("test message", "key", "value", "iteration", i)
		}
	})
}

// BenchmarkResponseTimeValidation validates that operations meet performance targets
func BenchmarkResponseTimeValidation(b *testing.B) {
	const SimpleOperationTarget = 100 * time.Millisecond

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

	logger := NewTestLogger()
	bridge := NewMockCLIBridge()
	bridge.SetResponse(&CommandResponse{
		ExitCode: 0,
		Stdout:   `{"status": "ok"}`,
		Duration: 1 * time.Millisecond,
	}, nil)

	server := NewServer(logger, bridge, config)
	ctx := context.Background()

	request := &types.MCPRequest{
		JSONRPC: "2.0",
		ID:      "test",
		Method:  "ping",
	}

	b.Run("response_time_under_target", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			start := time.Now()
			response, err := server.HandleRequest(ctx, request)
			duration := time.Since(start)

			require.NoError(b, err)
			require.NotNil(b, response)

			// Validate that simple operations complete within target time
			if duration > SimpleOperationTarget {
				b.Errorf("Operation took %v, expected < %v", duration, SimpleOperationTarget)
			}
		}
	})
}
