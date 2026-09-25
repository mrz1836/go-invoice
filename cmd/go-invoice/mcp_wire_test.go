package main

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrz1836/go-invoice/internal/mcp/executor"
	"github.com/mrz1836/go-invoice/internal/mcp/tools"
	"github.com/mrz1836/go-invoice/internal/models"
)

var errUnexpectedBridgeCommand = errors.New("unexpected bridge command")

// quietMCPLogger discards MCP log output; it satisfies both the tools and executor loggers
type quietMCPLogger struct{}

func (quietMCPLogger) Debug(_ string, _ ...interface{}) {}
func (quietMCPLogger) Info(_ string, _ ...interface{})  {}
func (quietMCPLogger) Warn(_ string, _ ...interface{})  {}
func (quietMCPLogger) Error(_ string, _ ...interface{}) {}

// unusedFileHandler satisfies the bridge constructor; client tools never prepare or collect files
type unusedFileHandler struct{}

func (unusedFileHandler) PrepareWorkspace(_ context.Context, _ []executor.FileReference) (string, func(), error) {
	return "", func() {}, nil
}

func (unusedFileHandler) CollectOutputFiles(_ context.Context, _ string, _ []string) ([]executor.FileReference, error) {
	return []executor.FileReference{}, nil
}

func (unusedFileHandler) ValidateFile(_ context.Context, _ string) error { return nil }

func (unusedFileHandler) CreateTempFile(_ context.Context, _ string, _ []byte) (string, error) {
	return "", nil
}

// inProcessCLI runs the argument lists built by the MCP bridge against the real client
// commands instead of spawning a go-invoice subprocess
type inProcessCLI struct {
	t     *testing.T
	calls [][]string
}

func (c *inProcessCLI) Execute(_ context.Context, req *executor.ExecutionRequest) (*executor.ExecutionResponse, error) {
	c.calls = append(c.calls, req.Args)

	// The bridge places the global --config flag ahead of the "client <verb>" subcommands
	if len(req.Args) < 4 || req.Args[0] != "--config" || req.Args[2] != "client" {
		return nil, errUnexpectedBridgeCommand
	}

	// Like a subprocess, the command creates its own context, so none is passed in
	app := newWireTestApp()
	var cmd *cobra.Command
	switch req.Args[3] {
	case "create":
		cmd = app.buildClientCreateCommand() //nolint:contextcheck // the CLI command owns its context, as in a subprocess
	case "update":
		cmd = app.buildClientUpdateCommand() //nolint:contextcheck // the CLI command owns its context, as in a subprocess
	default:
		return nil, errUnexpectedBridgeCommand
	}

	// A failing command is reported through its exit status, as a subprocess would report it
	resp := &executor.ExecutionResponse{}
	if runErr := runClientCommand(c.t, cmd, req.Args[1], req.Args[4:]...); runErr != nil {
		resp.ExitCode = 1
		resp.Stderr = runErr.Error()
		resp.Error = runErr.Error()
	}
	return resp, nil
}

func (c *inProcessCLI) ValidateCommand(_ context.Context, _ string, _ []string) error { return nil }

func (c *inProcessCLI) GetAllowedCommands(_ context.Context) ([]string, error) {
	return []string{"go-invoice"}, nil
}

// TestMCPClientWireSettingsMatchCLI drives client create and update through the MCP tool path
// (schema validation, then the CLI bridge) and checks the stored client matches the CLI.
func TestMCPClientWireSettingsMatchCLI(t *testing.T) {
	// The bridge always points at the config file in the home directory, so isolate it
	home := t.TempDir()
	t.Setenv("HOME", home)

	dataDir := t.TempDir()
	configPath := setWireTestEnv(t, dataDir, nil)
	store := newInitializedStorage(t, dataDir)

	ctx := context.Background()
	logger := quietMCPLogger{}
	registry := tools.NewDefaultToolRegistry(tools.NewDefaultInputValidator(logger), logger)
	require.NoError(t, tools.RegisterClientManagementTools(ctx, registry))
	cli := &inProcessCLI{t: t}
	bridge := executor.NewCLIBridge(logger, cli, unusedFileHandler{}, "")

	callTool := func(t *testing.T, name string, input map[string]interface{}) error {
		t.Helper()
		if err := registry.ValidateToolInput(ctx, name, input); err != nil {
			return err
		}
		resp, err := bridge.ExecuteToolCommand(ctx, name, input)
		if err != nil {
			return err
		}
		require.Equal(t, 0, resp.ExitCode, resp.Stderr)
		return nil
	}

	findClient := func(t *testing.T, email string) *models.Client {
		t.Helper()
		client, err := store.FindClientByEmail(ctx, email)
		require.NoError(t, err)
		return client
	}

	assertSameWireSettings := func(t *testing.T, expected, actual *models.Client) {
		t.Helper()
		assert.Equal(t, expected.WireType, actual.WireType)
		assert.Equal(t, expected.WireFeeEnabled, actual.WireFeeEnabled)
		assert.InDelta(t, expected.WireFeeAmount, actual.WireFeeAmount, 0.001)
	}

	t.Run("BridgeUsesHomeConfig", func(t *testing.T) {
		require.NoError(t, callTool(t, "client_create", map[string]interface{}{
			"name":  "Example Probe Co",
			"email": "probe@example.com",
		}))
		require.NotEmpty(t, cli.calls)
		assert.Equal(t, filepath.Join(home, ".go-invoice", ".env.config"), cli.calls[0][1])
	})

	t.Run("CreateWithoutWireSettingsMatchesCLI", func(t *testing.T) {
		require.NoError(t, runClientCommand(t, newWireTestApp().buildClientCreateCommand(), configPath,
			"--name", "Example Plain CLI Co", "--email", "plaincli@example.com"))
		require.NoError(t, callTool(t, "client_create", map[string]interface{}{
			"name":  "Example Plain MCP Co",
			"email": "plainmcp@example.com",
		}))

		mcpClient := findClient(t, "plainmcp@example.com")
		assert.Equal(t, models.WireTypeNone, mcpClient.WireType)
		assert.False(t, mcpClient.WireFeeEnabled)
		assertSameWireSettings(t, findClient(t, "plaincli@example.com"), mcpClient)
	})

	t.Run("CreateWithWireSettingsMatchesCLI", func(t *testing.T) {
		require.NoError(t, runClientCommand(t, newWireTestApp().buildClientCreateCommand(), configPath,
			"--name", "Example Wire CLI Co", "--email", "wirecli@example.com",
			"--wire-type", "international", "--wire-fee", "--wire-fee-amount", "20"))
		require.NoError(t, callTool(t, "client_create", map[string]interface{}{
			"name":             "Acme Corp",
			"email":            "ap@acme.example",
			"wire_type":        "international",
			"wire_fee_enabled": true,
			"wire_fee_amount":  20.0,
		}))

		mcpClient := findClient(t, "ap@acme.example")
		assert.Equal(t, models.WireTypeInternational, mcpClient.WireType)
		assert.True(t, mcpClient.WireFeeEnabled)
		assert.InDelta(t, models.DefaultWireFeeAmount, mcpClient.WireFeeAmount, 0.001)
		assertSameWireSettings(t, findClient(t, "wirecli@example.com"), mcpClient)
	})

	t.Run("UpdateExplicitFalseDisablesOnlyTheFee", func(t *testing.T) {
		clientID := string(findClient(t, "ap@acme.example").ID)

		require.NoError(t, callTool(t, "client_update", map[string]interface{}{
			"client_id":        clientID,
			"wire_fee_enabled": false,
		}))

		client := findClient(t, "ap@acme.example")
		assert.False(t, client.WireFeeEnabled)
		assert.Equal(t, models.WireTypeInternational, client.WireType, "omitted wire_type is unchanged")
		assert.InDelta(t, models.DefaultWireFeeAmount, client.WireFeeAmount, 0.001, "omitted amount is unchanged")
	})

	t.Run("UpdateChangesWireTypeAndAmount", func(t *testing.T) {
		clientID := string(findClient(t, "ap@acme.example").ID)

		require.NoError(t, callTool(t, "client_update", map[string]interface{}{
			"client_id":       clientID,
			"wire_type":       "domestic",
			"wire_fee_amount": 35.0,
		}))

		client := findClient(t, "ap@acme.example")
		assert.Equal(t, models.WireTypeDomestic, client.WireType)
		assert.InDelta(t, 35.0, client.WireFeeAmount, 0.001)
		assert.False(t, client.WireFeeEnabled, "omitted wire_fee_enabled is unchanged")
	})

	t.Run("InvalidWireTypeNeverReachesCLI", func(t *testing.T) {
		clientID := string(findClient(t, "ap@acme.example").ID)
		callsBefore := len(cli.calls)

		err := callTool(t, "client_update", map[string]interface{}{
			"client_id": clientID,
			"wire_type": "carrier-pigeon",
		})
		require.ErrorIs(t, err, executor.ErrInvalidWireType)

		err = callTool(t, "client_create", map[string]interface{}{
			"name":      "Example Invalid Co",
			"email":     "invalid@example.com",
			"wire_type": "carrier-pigeon",
		})
		require.ErrorIs(t, err, executor.ErrInvalidWireType)

		assert.Len(t, cli.calls, callsBefore, "the CLI must not run for an invalid wire type")
		assert.Equal(t, models.WireTypeDomestic, findClient(t, "ap@acme.example").WireType)
		_, findErr := store.FindClientByEmail(ctx, "invalid@example.com")
		require.Error(t, findErr, "rejected client must not be stored")
	})

	t.Run("OutOfRangeWireFeeRejectedBySchema", func(t *testing.T) {
		clientID := string(findClient(t, "ap@acme.example").ID)
		callsBefore := len(cli.calls)

		err := callTool(t, "client_update", map[string]interface{}{
			"client_id":       clientID,
			"wire_fee_amount": -5.0,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "wire_fee_amount")

		assert.Len(t, cli.calls, callsBefore, "the CLI must not run for an out-of-range fee")
		assert.InDelta(t, 35.0, findClient(t, "ap@acme.example").WireFeeAmount, 0.001)
	})
}

// TestMCPSandboxForwardsEveryWireSetting checks the MCP executor passes every wire setting the
// configuration reads through to the CLI, so generation over MCP sees the same environment
func TestMCPSandboxForwardsEveryWireSetting(t *testing.T) {
	t.Parallel()

	whitelist := executor.DefaultSecurityConfig().Sandbox.EnvironmentWhitelist
	for _, key := range wireEnvKeys() {
		assert.Contains(t, whitelist, key)
	}
}
