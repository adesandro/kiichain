package cmd_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"

	app "github.com/kiichain/kiichain/v2/app"
	"github.com/kiichain/kiichain/v2/cmd/kiichaind/cmd"
)

func TestRootCmdConfig(t *testing.T) {
	rootCmd := cmd.NewRootCmd()
	rootCmd.SetArgs([]string{
		"config", // Test the config cmd
		"get app pruning",
		"keyring-backend", // key
		"test",            // value
	})

	require.NoError(t, svrcmd.Execute(rootCmd, "", app.DefaultNodeHome))
}
