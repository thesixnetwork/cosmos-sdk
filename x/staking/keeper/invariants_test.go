package keeper_test

import (
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (s *KeeperTestSuite) TestLicenseCountInvariant() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	invariant := stakingkeeper.LicenseCountInvariant(keeper)

	// empty state is healthy
	_, broken := invariant(ctx)
	require.False(broken)

	// normal validators are ignored by the invariant
	normalVal, err := stakingtypes.NewValidator(sdk.ValAddress(PKs[10].Address()).String(), PKs[10], stakingtypes.Description{Moniker: "normal"})
	require.NoError(err)
	require.NoError(keeper.SetValidator(ctx, normalVal))

	// license validator with nil bookkeeping is skipped, not reported
	nilVal, err := stakingtypes.NewValidator(sdk.ValAddress(PKs[11].Address()).String(), PKs[11], stakingtypes.Description{Moniker: "nil-license"})
	require.NoError(err)
	nilVal.Mode = stakingtypes.ValidatorMode_MODE_LICENSE
	require.NoError(keeper.SetValidator(ctx, nilVal))

	// license validator within its cap
	licenseVal, err := stakingtypes.NewValidator(sdk.ValAddress(PKs[12].Address()).String(), PKs[12], stakingtypes.Description{Moniker: "license"})
	require.NoError(err)
	licenseVal.Mode = stakingtypes.ValidatorMode_MODE_LICENSE
	licenseVal.LicenseCount = math.NewInt(50)
	licenseVal.MaxLicense = math.NewInt(150)
	require.NoError(keeper.SetValidator(ctx, licenseVal))

	_, broken = invariant(ctx)
	require.False(broken)

	// usage above the owner-set cap breaks the invariant and names the validator
	licenseVal.LicenseCount = math.NewInt(151)
	licenseVal.MaxLicense = math.NewInt(150)
	require.NoError(keeper.SetValidator(ctx, licenseVal))

	msg, broken := invariant(ctx)
	require.True(broken)
	require.Contains(msg, licenseVal.GetOperator())

	// restoring the cap heals it
	licenseVal.LicenseCount = math.NewInt(150)
	require.NoError(keeper.SetValidator(ctx, licenseVal))
	_, broken = invariant(ctx)
	require.False(broken)
}
