package keeper_test

import (
	"github.com/golang/mock/gomock"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (s *KeeperTestSuite) TestMsgWhitelistDelegator() {
	ctx, keeper, msgServer := s.ctx, s.stakingKeeper, s.msgServer
	require := s.Require()
	s.execExpectCalls()

	pk := ed25519.GenPrivKey().PubKey()
	comm := stakingtypes.NewCommissionRates(math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0))
	require.NoError(keeper.SetNewValidatorApprovalState(ctx, stakingtypes.ValidatorApproval{ApproverAddress: pk.Address().String(), Enabled: false}))

	// fast validator whose whitelist is managed below
	msg, err := stakingtypes.NewMsgCreateValidator(ValAddr.String(), pk.Address().String(), pk, sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(10000)), stakingtypes.Description{Moniker: "FastVal"}, comm, math.OneInt())
	require.NoError(err)
	msg.Mode = stakingtypes.ValidatorMode_MODE_FAST
	_, err = msgServer.CreateValidator(ctx, msg)
	require.NoError(err)

	delegator := sdk.AccAddress(PKS[1].Address())
	delegator2 := sdk.AccAddress(PKS[2].Address())

	// only the operator may modify the whitelist
	_, err = msgServer.CreateWhitelistdelegator(ctx, &stakingtypes.MsgCreateWhitelistDelegator{
		Creator:          delegator.String(),
		ValidatorAddress: ValAddr.String(),
		DelegatorAddress: delegator.String(),
	})
	require.Error(err)
	require.Contains(err.Error(), "only the validator operator can modify its delegator whitelist")

	// operator adds a delegator
	_, err = msgServer.CreateWhitelistdelegator(ctx, &stakingtypes.MsgCreateWhitelistDelegator{
		Creator:          Addr.String(),
		ValidatorAddress: ValAddr.String(),
		DelegatorAddress: delegator.String(),
	})
	require.NoError(err)
	require.True(keeper.IsSpecialDelegator(ctx, ValAddr, delegator))

	// the same delegator cannot be added twice
	_, err = msgServer.CreateWhitelistdelegator(ctx, &stakingtypes.MsgCreateWhitelistDelegator{
		Creator:          Addr.String(),
		ValidatorAddress: ValAddr.String(),
		DelegatorAddress: delegator.String(),
	})
	require.Error(err)
	require.Contains(err.Error(), "duplicate delegator address")

	// a second delegator lands in the same list
	_, err = msgServer.CreateWhitelistdelegator(ctx, &stakingtypes.MsgCreateWhitelistDelegator{
		Creator:          Addr.String(),
		ValidatorAddress: ValAddr.String(),
		DelegatorAddress: delegator2.String(),
	})
	require.NoError(err)
	whitelist, err := keeper.GetWhitelistDelegator(ctx, ValAddr)
	require.NoError(err)
	require.Len(whitelist.DelegatorAddress, 2)

	// only the operator may delete from the whitelist
	_, err = msgServer.DeleteWhitelistdelegator(ctx, &stakingtypes.MsgDeleteWhitelistDelegator{
		Creator:          delegator.String(),
		ValidatorAddress: ValAddr.String(),
		DelegatorAddress: delegator.String(),
	})
	require.Error(err)
	require.Contains(err.Error(), "only the validator operator can modify its delegator whitelist")

	// deleting for a validator that does not exist fails
	_, err = msgServer.DeleteWhitelistdelegator(ctx, &stakingtypes.MsgDeleteWhitelistDelegator{
		Creator:          delegator.String(),
		ValidatorAddress: sdk.ValAddress(delegator).String(),
		DelegatorAddress: delegator.String(),
	})
	require.Error(err)
	require.Contains(err.Error(), "validator does not exist")

	// operator removes the first delegator; the second one stays whitelisted
	_, err = msgServer.DeleteWhitelistdelegator(ctx, &stakingtypes.MsgDeleteWhitelistDelegator{
		Creator:          Addr.String(),
		ValidatorAddress: ValAddr.String(),
		DelegatorAddress: delegator.String(),
	})
	require.NoError(err)
	require.False(keeper.IsSpecialDelegator(ctx, ValAddr, delegator))
	require.True(keeper.IsSpecialDelegator(ctx, ValAddr, delegator2))

	// deleting from a validator that never created a whitelist fails
	otherAccAddr := delegator
	otherValAddr := sdk.ValAddress(otherAccAddr)
	s.bankKeeper.EXPECT().DelegateCoinsFromAccountToModule(gomock.Any(), otherAccAddr, stakingtypes.NotBondedPoolName, gomock.Any()).AnyTimes()
	otherPk := ed25519.GenPrivKey().PubKey()
	msg, err = stakingtypes.NewMsgCreateValidator(otherValAddr.String(), pk.Address().String(), otherPk, sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(10000)), stakingtypes.Description{Moniker: "OtherVal"}, comm, math.OneInt())
	require.NoError(err)
	_, err = msgServer.CreateValidator(ctx, msg)
	require.NoError(err)
	_, err = msgServer.DeleteWhitelistdelegator(ctx, &stakingtypes.MsgDeleteWhitelistDelegator{
		Creator:          otherAccAddr.String(),
		ValidatorAddress: otherValAddr.String(),
		DelegatorAddress: delegator2.String(),
	})
	require.Error(err)
	require.Contains(err.Error(), "validator whitelist delegator doesn't exist")

	// the operator is always special, whitelist or not
	require.True(keeper.IsSpecialDelegator(ctx, ValAddr, Addr))
	require.True(keeper.IsSpecialDelegator(ctx, otherValAddr, otherAccAddr))
}

func (s *KeeperTestSuite) TestWhitelistDelegatorStore() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	valAddr := sdk.ValAddress(PKs[0].Address())
	delegator := sdk.AccAddress(PKs[1].Address())

	// unknown validator has no whitelist
	_, err := keeper.GetWhitelistDelegator(ctx, valAddr)
	require.ErrorIs(err, stakingtypes.ErrNoWhiltelistFound)

	whitelist := stakingtypes.WhitelistDelegator{
		ValidatorAddress: valAddr.String(),
		DelegatorAddress: []string{delegator.String()},
	}
	keeper.SetWhitelistDelegator(ctx, whitelist)

	stored, err := keeper.GetWhitelistDelegator(ctx, valAddr)
	require.NoError(err)
	require.Equal(whitelist, stored)

	all, err := keeper.GetAllWhitelistDelegator(ctx)
	require.NoError(err)
	require.Len(all, 1)

	keeper.RemoveWhitelistDelegator(ctx, valAddr)
	_, err = keeper.GetWhitelistDelegator(ctx, valAddr)
	require.ErrorIs(err, stakingtypes.ErrNoWhiltelistFound)
}
