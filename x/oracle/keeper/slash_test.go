package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
)

func TestSlashAndResetMissCounters(t *testing.T) {
	// initial setup
	input := CreateTestInput(t)
	bankKeeper := input.BankKeeper
	stakingKeeper := input.StakingKeeper
	oracleKeeper := input.OracleKeeper

	addr1, val1 := ValAddrs[0], ValPubKeys[0]
	addr2, val2 := ValAddrs[1], ValPubKeys[1]
	amount := sdk.TokensFromConsensusPower(100, sdk.DefaultPowerReduction)
	sh := staking.NewHandler(stakingKeeper)
	ctx := input.Ctx

	// Validators created
	_, err := sh(ctx, NewTestMsgCreateValidator(addr1, val1, amount))
	require.NoError(t, err)
	_, err = sh(ctx, NewTestMsgCreateValidator(addr2, val2, amount))
	require.NoError(t, err)
	staking.EndBlocker(ctx, stakingKeeper)

	balance1 := bankKeeper.GetAllBalances(ctx, sdk.AccAddress(addr1))
	expectedBalance := sdk.NewCoins(sdk.NewCoin(stakingKeeper.GetParams(ctx).BondDenom, InitTokens.Sub(amount)))
	bondedTokens1 := stakingKeeper.Validator(ctx, addr1).GetBondedTokens()
	require.Equal(t, balance1, expectedBalance)
	require.Equal(t, amount, bondedTokens1)

	balance2 := bankKeeper.GetAllBalances(ctx, sdk.AccAddress(addr2))
	bondedTokens2 := stakingKeeper.Validator(ctx, addr2).GetBondedTokens()
	require.Equal(t, balance2, expectedBalance)
	require.Equal(t, amount, bondedTokens2)

	//Define slash fraction
	votePeriodsPerWindow := sdk.NewDec(int64(oracleKeeper.SlashWindow(input.Ctx))).QuoInt64(int64(oracleKeeper.VotePeriod(input.Ctx))).TruncateInt64()
	slashFraction := oracleKeeper.SlashFraction(input.Ctx)
	minValidVotes := oracleKeeper.MinValidPerWindow(input.Ctx).MulInt64(votePeriodsPerWindow).TruncateInt64()

	t.Run("no slash", func(t *testing.T) {
		oracleKeeper.SetVotePenaltyCounter(input.Ctx, ValAddrs[0], uint64(votePeriodsPerWindow-minValidVotes), 0, uint64(minValidVotes))
		oracleKeeper.SlashAndResetCounters(input.Ctx)
		staking.EndBlocker(input.Ctx, stakingKeeper)

		validator, _ := stakingKeeper.GetValidator(input.Ctx, ValAddrs[0])
		require.Equal(t, amount, validator.GetBondedTokens())
	})

	t.Run("no slash - total votes is greater than votes per window", func(t *testing.T) {
		oracleKeeper.SetVotePenaltyCounter(input.Ctx, ValAddrs[0], uint64(votePeriodsPerWindow), 0, uint64(votePeriodsPerWindow))
		oracleKeeper.SlashAndResetCounters(input.Ctx)
		staking.EndBlocker(input.Ctx, stakingKeeper)

		validator, _ := stakingKeeper.GetValidator(input.Ctx, ValAddrs[0])
		require.Equal(t, amount, validator.GetBondedTokens())
	})

	t.Run("successfully slash", func(t *testing.T) {
		oracleKeeper.SetVotePenaltyCounter(input.Ctx, ValAddrs[0], uint64(votePeriodsPerWindow-minValidVotes+1), 0, uint64(minValidVotes-1))
		oracleKeeper.SlashAndResetCounters(input.Ctx)
		validator, _ := stakingKeeper.GetValidator(input.Ctx, ValAddrs[0])
		require.Equal(t, amount.Sub(slashFraction.MulInt(amount).TruncateInt()), validator.GetBondedTokens())
		require.True(t, validator.IsJailed())
	})

	t.Run("slash and jail for abstaining too much along with misses", func(t *testing.T) {
		validator, _ := stakingKeeper.GetValidator(input.Ctx, ValAddrs[0])
		validator.Jailed = false
		validator.Tokens = amount
		stakingKeeper.SetValidator(input.Ctx, validator)
		require.Equal(t, amount, validator.GetBondedTokens())
		oracleKeeper.SetVotePenaltyCounter(input.Ctx, ValAddrs[0], 0, uint64(votePeriodsPerWindow-minValidVotes+1), 0)
		oracleKeeper.SlashAndResetCounters(input.Ctx)
		validator, _ = stakingKeeper.GetValidator(input.Ctx, ValAddrs[0])

		// slashing for not voting validly sufficiently
		require.Equal(t, amount.Sub(slashFraction.MulInt(amount).TruncateInt()), validator.GetBondedTokens())
		require.True(t, validator.IsJailed())
	})

	t.Run("slash unbonded validator", func(t *testing.T) {
		validator, _ := stakingKeeper.GetValidator(input.Ctx, ValAddrs[0])
		validator.Status = stakingtypes.Unbonded
		validator.Jailed = false
		validator.Tokens = amount
		stakingKeeper.SetValidator(input.Ctx, validator)

		oracleKeeper.SetVotePenaltyCounter(input.Ctx, ValAddrs[0], uint64(votePeriodsPerWindow-minValidVotes+1), 0, 0)
		oracleKeeper.SlashAndResetCounters(input.Ctx)
		validator, _ = stakingKeeper.GetValidator(input.Ctx, ValAddrs[0])
		require.Equal(t, amount, validator.Tokens)
		require.False(t, validator.IsJailed())
	})

	t.Run("slash jailed validator", func(t *testing.T) {
		validator, _ := stakingKeeper.GetValidator(input.Ctx, ValAddrs[0])
		validator.Status = stakingtypes.Bonded
		validator.Jailed = true
		validator.Tokens = amount
		stakingKeeper.SetValidator(input.Ctx, validator)

		oracleKeeper.SetVotePenaltyCounter(input.Ctx, ValAddrs[0], uint64(votePeriodsPerWindow-minValidVotes+1), 0, 0)
		oracleKeeper.SlashAndResetCounters(input.Ctx)
		validator, _ = stakingKeeper.GetValidator(input.Ctx, ValAddrs[0])
		require.Equal(t, amount, validator.Tokens)
	})
}
