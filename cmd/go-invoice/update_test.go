package main

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUpdateCommandWiring locks the seam between the root command and the
// go-selfupdate cobracmd package: the self-update command is registered under
// the name "update" with the "upgrade" alias (preserving the old command name)
// and the check/force/verbose boolean flags. The command's behavior itself is
// covered by the library's own suites, so this asserts only the wiring.
func TestUpdateCommandWiring(t *testing.T) {
	t.Parallel()

	var updateCmd *cobra.Command
	for _, c := range NewApp().rootCmd.Commands() {
		if c.Name() == "update" {
			updateCmd = c
			break
		}
	}
	require.NotNil(t, updateCmd, "the root command registers an update command")
	assert.Contains(t, updateCmd.Aliases, "upgrade", "the update command keeps the upgrade alias")

	for _, name := range []string{"check", "force", "verbose"} {
		flag := updateCmd.Flags().Lookup(name)
		require.NotNilf(t, flag, "the update command registers --%s", name)
		assert.Equalf(t, "bool", flag.Value.Type(), "--%s is a boolean flag", name)
	}
}
