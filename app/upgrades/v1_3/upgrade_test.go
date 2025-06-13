package v130_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	tmtypes "github.com/cometbft/cometbft/proto/tendermint/types"

	"github.com/kiichain/kiichain/v2/app/helpers"
	v130 "github.com/kiichain/kiichain/v2/app/upgrades/v1_3"
	"github.com/kiichain/kiichain/v2/precompiles/wasmd"
)

// TestUpgrade tests the upgrade handler for v1.3.0
func TestUpgrade(t *testing.T) {
	// Create the app and the context
	app := helpers.Setup(t)
	ctx := app.BaseApp.NewUncachedContext(true, tmtypes.Header{Height: 1, ChainID: "test_1010-1", Time: time.Now().UTC()})

	// Create a pre-populated list of pre-compiles
	precompiles := []string{
		"0x0000000000000000000000000000000000000001",
		"0x0000000000000000000000000000000000000002",
	}

	// Install the precompiles
	evmParams := app.EVMKeeper.GetParams(ctx)
	evmParams.ActiveStaticPrecompiles = precompiles
	err := app.EVMKeeper.SetParams(ctx, evmParams)
	require.NoError(t, err)

	// Now run add wasmd upgrade
	err = v130.InstallNewPrecompile(ctx, &app.AppKeepers)
	require.NoError(t, err)

	// Get the params again
	evmParams = app.EVMKeeper.GetParams(ctx)

	// Check that the precompiles was added
	require.Len(t, evmParams.ActiveStaticPrecompiles, 3)
	require.Contains(t, evmParams.ActiveStaticPrecompiles, "0x0000000000000000000000000000000000000001")
	require.Contains(t, evmParams.ActiveStaticPrecompiles, "0x0000000000000000000000000000000000000002")
	require.Contains(t, evmParams.ActiveStaticPrecompiles, wasmd.WasmdPrecompileAddress)
}
