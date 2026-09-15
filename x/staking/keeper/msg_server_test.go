package keeper_test

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var (
	PKS     = simtestutil.CreateTestPubKeys(3)
	Addr    = sdk.AccAddress(PKS[0].Address())
	ValAddr = sdk.ValAddress(Addr)
)

func (s *KeeperTestSuite) execExpectCalls() {
	s.accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec("cosmos")).AnyTimes()
	s.bankKeeper.EXPECT().DelegateCoinsFromAccountToModule(gomock.Any(), Addr, stakingtypes.NotBondedPoolName, gomock.Any()).AnyTimes()
}

func (s *KeeperTestSuite) TestMsgCreateValidator() {
	ctx, msgServer := s.ctx, s.msgServer
	require := s.Require()
	s.execExpectCalls()

	pk1 := ed25519.GenPrivKey().PubKey()
	require.NotNil(pk1)

	pubkey, err := codectypes.NewAnyWithValue(pk1)
	require.NoError(err)

	testCases := []struct {
		name      string
		input     *stakingtypes.MsgCreateValidator
		expErr    bool
		expErrMsg string
	}{
		{
			name: "empty description",
			input: &stakingtypes.MsgCreateValidator{
				Description: stakingtypes.Description{},
				Commission: stakingtypes.CommissionRates{
					Rate:          math.LegacyNewDecWithPrec(5, 1),
					MaxRate:       math.LegacyNewDecWithPrec(5, 1),
					MaxChangeRate: math.LegacyNewDec(0),
				},
				MinSelfDelegation: math.NewInt(1),
				DelegatorAddress:  Addr.String(),
				ValidatorAddress:  ValAddr.String(),
				Pubkey:            pubkey,
				Value:             sdk.NewInt64Coin("stake", 10000),
				ApproverAddress:   Addr.String(),
			},
			expErr:    true,
			expErrMsg: "empty description",
		},
		{
			name: "invalid validator address",
			input: &stakingtypes.MsgCreateValidator{
				Description: stakingtypes.Description{
					Moniker: "NewValidator",
				},
				Commission: stakingtypes.CommissionRates{
					Rate:          math.LegacyNewDecWithPrec(5, 1),
					MaxRate:       math.LegacyNewDecWithPrec(5, 1),
					MaxChangeRate: math.LegacyNewDec(0),
				},
				MinSelfDelegation: math.NewInt(1),
				DelegatorAddress:  Addr.String(),
				ValidatorAddress:  sdk.AccAddress([]byte("invalid")).String(),
				Pubkey:            pubkey,
				Value:             sdk.NewInt64Coin("stake", 10000),
				ApproverAddress:   Addr.String(),
			},
			expErr:    true,
			expErrMsg: "invalid validator address",
		},
		{
			name: "empty validator pubkey",
			input: &stakingtypes.MsgCreateValidator{
				Description: stakingtypes.Description{
					Moniker: "NewValidator",
				},
				Commission: stakingtypes.CommissionRates{
					Rate:          math.LegacyNewDecWithPrec(5, 1),
					MaxRate:       math.LegacyNewDecWithPrec(5, 1),
					MaxChangeRate: math.LegacyNewDec(0),
				},
				MinSelfDelegation: math.NewInt(1),
				DelegatorAddress:  Addr.String(),
				ValidatorAddress:  ValAddr.String(),
				Pubkey:            nil,
				Value:             sdk.NewInt64Coin("stake", 10000),
				ApproverAddress:   Addr.String(),
			},
			expErr:    true,
			expErrMsg: "empty validator public key",
		},
		{
			name: "empty delegation amount",
			input: &stakingtypes.MsgCreateValidator{
				Description: stakingtypes.Description{
					Moniker: "NewValidator",
				},
				Commission: stakingtypes.CommissionRates{
					Rate:          math.LegacyNewDecWithPrec(5, 1),
					MaxRate:       math.LegacyNewDecWithPrec(5, 1),
					MaxChangeRate: math.LegacyNewDec(0),
				},
				MinSelfDelegation: math.NewInt(1),
				DelegatorAddress:  Addr.String(),
				ValidatorAddress:  ValAddr.String(),
				Pubkey:            pubkey,
				Value:             sdk.NewInt64Coin("stake", 0),
				ApproverAddress:   Addr.String(),
			},
			expErr:    true,
			expErrMsg: "invalid delegation amount",
		},
		{
			name: "nil delegation amount",
			input: &stakingtypes.MsgCreateValidator{
				Description: stakingtypes.Description{
					Moniker: "NewValidator",
				},
				Commission: stakingtypes.CommissionRates{
					Rate:          math.LegacyNewDecWithPrec(5, 1),
					MaxRate:       math.LegacyNewDecWithPrec(5, 1),
					MaxChangeRate: math.LegacyNewDec(0),
				},
				MinSelfDelegation: math.NewInt(1),
				DelegatorAddress:  Addr.String(),
				ValidatorAddress:  ValAddr.String(),
				Pubkey:            pubkey,
				Value:             sdk.Coin{},
				ApproverAddress:   Addr.String(),
			},
			expErr:    true,
			expErrMsg: "invalid delegation amount",
		},
		{
			name: "zero minimum self delegation",
			input: &stakingtypes.MsgCreateValidator{
				Description: stakingtypes.Description{
					Moniker: "NewValidator",
				},
				Commission: stakingtypes.CommissionRates{
					Rate:          math.LegacyNewDecWithPrec(5, 1),
					MaxRate:       math.LegacyNewDecWithPrec(5, 1),
					MaxChangeRate: math.LegacyNewDec(0),
				},
				MinSelfDelegation: math.NewInt(0),
				DelegatorAddress:  Addr.String(),
				ValidatorAddress:  ValAddr.String(),
				Pubkey:            pubkey,
				Value:             sdk.NewInt64Coin("stake", 10000),
				ApproverAddress:   Addr.String(),
			},
			expErr:    true,
			expErrMsg: "minimum self delegation must be a positive integer",
		},
		{
			name: "negative minimum self delegation",
			input: &stakingtypes.MsgCreateValidator{
				Description: stakingtypes.Description{
					Moniker: "NewValidator",
				},
				Commission: stakingtypes.CommissionRates{
					Rate:          math.LegacyNewDecWithPrec(5, 1),
					MaxRate:       math.LegacyNewDecWithPrec(5, 1),
					MaxChangeRate: math.LegacyNewDec(0),
				},
				MinSelfDelegation: math.NewInt(-1),
				DelegatorAddress:  Addr.String(),
				ValidatorAddress:  ValAddr.String(),
				Pubkey:            pubkey,
				Value:             sdk.NewInt64Coin("stake", 10000),
				ApproverAddress:   Addr.String(),
			},
			expErr:    true,
			expErrMsg: "minimum self delegation must be a positive integer",
		},
		{
			name: "delegation less than minimum self delegation",
			input: &stakingtypes.MsgCreateValidator{
				Description: stakingtypes.Description{
					Moniker: "NewValidator",
				},
				Commission: stakingtypes.CommissionRates{
					Rate:          math.LegacyNewDecWithPrec(5, 1),
					MaxRate:       math.LegacyNewDecWithPrec(5, 1),
					MaxChangeRate: math.LegacyNewDec(0),
				},
				MinSelfDelegation: math.NewInt(100),
				DelegatorAddress:  Addr.String(),
				ValidatorAddress:  ValAddr.String(),
				Pubkey:            pubkey,
				Value:             sdk.NewInt64Coin("stake", 10),
				ApproverAddress:   Addr.String(),
			},
			expErr:    true,
			expErrMsg: "validator's self delegation must be greater than their minimum self delegation",
		},
		{
			name: "valid msg",
			input: &stakingtypes.MsgCreateValidator{
				Description: stakingtypes.Description{
					Moniker:         "NewValidator",
					Identity:        "xyz",
					Website:         "xyz.com",
					SecurityContact: "xyz@gmail.com",
					Details:         "details",
				},
				Commission: stakingtypes.CommissionRates{
					Rate:          math.LegacyNewDecWithPrec(5, 1),
					MaxRate:       math.LegacyNewDecWithPrec(5, 1),
					MaxChangeRate: math.LegacyNewDec(0),
				},
				MinSelfDelegation: math.NewInt(1),
				DelegatorAddress:  Addr.String(),
				ValidatorAddress:  ValAddr.String(),
				Pubkey:            pubkey,
				Value:             sdk.NewInt64Coin("stake", 10000),
				ApproverAddress:   Addr.String(),
			},
			expErr: false,
		},
	}
	for _, tc := range testCases {
		tc := tc
		s.T().Run(tc.name, func(t *testing.T) {
			s.stakingKeeper.SetNewValidatorApprovalState(ctx, stakingtypes.ValidatorApproval{ApproverAddress: tc.input.ApproverAddress, Enabled: false})
			_, err := msgServer.CreateValidator(ctx, tc.input)
			if tc.expErr {
				require.Error(err)
				require.Contains(err.Error(), tc.expErrMsg)
			} else {
				require.NoError(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestMsgEditValidator() {
	ctx, msgServer := s.ctx, s.msgServer
	require := s.Require()
	s.execExpectCalls()

	// create new context with updated block time
	newCtx := ctx.WithBlockTime(ctx.BlockTime().AddDate(0, 0, 1))

	pk := ed25519.GenPrivKey().PubKey()
	require.NotNil(pk)

	s.stakingKeeper.SetNewValidatorApprovalState(ctx, stakingtypes.ValidatorApproval{ApproverAddress: pk.Address().String(), Enabled: false})

	comm := stakingtypes.NewCommissionRates(math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0))
	msg, err := stakingtypes.NewMsgCreateValidator(ValAddr.String(), pk.Address().String(), pk, sdk.NewCoin("stake", math.NewInt(10)), stakingtypes.Description{Moniker: "NewVal"}, comm, math.OneInt())
	require.NoError(err)

	res, err := msgServer.CreateValidator(ctx, msg)
	require.NoError(err)
	require.NotNil(res)

	newRate := math.LegacyZeroDec()
	invalidRate := math.LegacyNewDec(2)

	lowSelfDel := math.OneInt()
	highSelfDel := math.NewInt(100)
	negSelfDel := math.NewInt(-1)
	newSelfDel := math.NewInt(3)

	testCases := []struct {
		name      string
		ctx       sdk.Context
		input     *stakingtypes.MsgEditValidator
		expErr    bool
		expErrMsg string
	}{
		{
			name: "invalid validator",
			ctx:  newCtx,
			input: &stakingtypes.MsgEditValidator{
				Description: stakingtypes.Description{
					Moniker: "TestValidator",
				},
				ValidatorAddress:  sdk.AccAddress([]byte("invalid")).String(),
				CommissionRate:    &newRate,
				MinSelfDelegation: &newSelfDel,
			},
			expErr:    true,
			expErrMsg: "invalid validator address",
		},
		{
			name: "empty description",
			ctx:  newCtx,
			input: &stakingtypes.MsgEditValidator{
				Description:       stakingtypes.Description{},
				ValidatorAddress:  ValAddr.String(),
				CommissionRate:    &newRate,
				MinSelfDelegation: &newSelfDel,
			},
			expErr:    true,
			expErrMsg: "empty description",
		},
		{
			name: "negative self delegation",
			ctx:  newCtx,
			input: &stakingtypes.MsgEditValidator{
				Description: stakingtypes.Description{
					Moniker: "TestValidator",
				},
				ValidatorAddress:  ValAddr.String(),
				CommissionRate:    &newRate,
				MinSelfDelegation: &negSelfDel,
			},
			expErr:    true,
			expErrMsg: "minimum self delegation must be a positive integer",
		},
		{
			name: "invalid commission rate",
			ctx:  newCtx,
			input: &stakingtypes.MsgEditValidator{
				Description: stakingtypes.Description{
					Moniker: "TestValidator",
				},
				ValidatorAddress:  ValAddr.String(),
				CommissionRate:    &invalidRate,
				MinSelfDelegation: &newSelfDel,
			},
			expErr:    true,
			expErrMsg: "commission rate must be between 0 and 1 (inclusive)",
		},
		{
			name: "validator does not exist",
			ctx:  newCtx,
			input: &stakingtypes.MsgEditValidator{
				Description: stakingtypes.Description{
					Moniker: "TestValidator",
				},
				ValidatorAddress:  sdk.ValAddress([]byte("val")).String(),
				CommissionRate:    &newRate,
				MinSelfDelegation: &newSelfDel,
			},
			expErr:    true,
			expErrMsg: "validator does not exist",
		},
		{
			name: "change commmission rate in <24hrs",
			ctx:  ctx,
			input: &stakingtypes.MsgEditValidator{
				Description: stakingtypes.Description{
					Moniker: "TestValidator",
				},
				ValidatorAddress:  ValAddr.String(),
				CommissionRate:    &newRate,
				MinSelfDelegation: &newSelfDel,
			},
			expErr:    true,
			expErrMsg: "commission cannot be changed more than once in 24h",
		},
		{
			name: "minimum self delegation cannot decrease",
			ctx:  newCtx,
			input: &stakingtypes.MsgEditValidator{
				Description: stakingtypes.Description{
					Moniker: "TestValidator",
				},
				ValidatorAddress:  ValAddr.String(),
				CommissionRate:    &newRate,
				MinSelfDelegation: &lowSelfDel,
			},
			expErr:    true,
			expErrMsg: "minimum self delegation cannot be decrease",
		},
		{
			name: "validator self-delegation must be greater than min self delegation",
			ctx:  newCtx,
			input: &stakingtypes.MsgEditValidator{
				Description: stakingtypes.Description{
					Moniker: "TestValidator",
				},
				ValidatorAddress:  ValAddr.String(),
				CommissionRate:    &newRate,
				MinSelfDelegation: &highSelfDel,
			},
			expErr:    true,
			expErrMsg: "validator's self delegation must be greater than their minimum self delegation",
		},
		{
			name: "valid msg",
			ctx:  newCtx,
			input: &stakingtypes.MsgEditValidator{
				Description: stakingtypes.Description{
					Moniker:         "TestValidator",
					Identity:        "abc",
					Website:         "abc.com",
					SecurityContact: "abc@gmail.com",
					Details:         "newDetails",
				},
				ValidatorAddress:  ValAddr.String(),
				CommissionRate:    &newRate,
				MinSelfDelegation: &newSelfDel,
			},
			expErr: false,
		},
	}
	for _, tc := range testCases {
		tc := tc
		s.T().Run(tc.name, func(t *testing.T) {
			_, err := msgServer.EditValidator(tc.ctx, tc.input)
			if tc.expErr {
				require.Error(err)
				require.Contains(err.Error(), tc.expErrMsg)
			} else {
				require.NoError(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestMsgDelegate() {
	ctx, keeper, msgServer := s.ctx, s.stakingKeeper, s.msgServer
	require := s.Require()
	s.execExpectCalls()

	pk := ed25519.GenPrivKey().PubKey()
	require.NotNil(pk)

	comm := stakingtypes.NewCommissionRates(math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0))
	s.stakingKeeper.SetNewValidatorApprovalState(ctx, stakingtypes.ValidatorApproval{ApproverAddress: pk.Address().String(), Enabled: false})

	msg, err := stakingtypes.NewMsgCreateValidator(ValAddr.String(), pk.Address().String(), pk, sdk.NewCoin("stake", math.NewInt(10)), stakingtypes.Description{Moniker: "NewVal"}, comm, math.OneInt())
	require.NoError(err)

	res, err := msgServer.CreateValidator(ctx, msg)
	require.NoError(err)
	require.NotNil(res)

	testCases := []struct {
		name      string
		input     *stakingtypes.MsgDelegate
		expErr    bool
		expErrMsg string
	}{
		{
			name: "invalid validator",
			input: &stakingtypes.MsgDelegate{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: sdk.AccAddress([]byte("invalid")).String(),
				Amount:           sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: keeper.TokensFromConsensusPower(s.ctx, int64(100))},
			},
			expErr:    true,
			expErrMsg: "invalid validator address",
		},
		{
			name: "empty delegator",
			input: &stakingtypes.MsgDelegate{
				DelegatorAddress: "",
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: keeper.TokensFromConsensusPower(s.ctx, int64(100))},
			},
			expErr:    true,
			expErrMsg: "invalid delegator address: empty address string is not allowed",
		},
		{
			name: "invalid delegator",
			input: &stakingtypes.MsgDelegate{
				DelegatorAddress: "invalid",
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: keeper.TokensFromConsensusPower(s.ctx, int64(100))},
			},
			expErr:    true,
			expErrMsg: "invalid delegator address: decoding bech32 failed",
		},
		{
			name: "validator does not exist",
			input: &stakingtypes.MsgDelegate{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: sdk.ValAddress([]byte("val")).String(),
				Amount:           sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: keeper.TokensFromConsensusPower(s.ctx, int64(100))},
			},
			expErr:    true,
			expErrMsg: "validator does not exist",
		},
		{
			name: "zero amount",
			input: &stakingtypes.MsgDelegate{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: keeper.TokensFromConsensusPower(s.ctx, int64(0))},
			},
			expErr:    true,
			expErrMsg: "invalid delegation amount",
		},
		{
			name: "negative amount",
			input: &stakingtypes.MsgDelegate{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: keeper.TokensFromConsensusPower(s.ctx, int64(-1))},
			},
			expErr:    true,
			expErrMsg: "invalid delegation amount",
		},
		{
			name: "invalid BondDenom",
			input: &stakingtypes.MsgDelegate{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.Coin{Denom: "test", Amount: keeper.TokensFromConsensusPower(s.ctx, int64(100))},
			},
			expErr:    true,
			expErrMsg: "invalid coin denomination",
		},
		{
			name: "valid msg",
			input: &stakingtypes.MsgDelegate{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: keeper.TokensFromConsensusPower(s.ctx, int64(100))},
			},
			expErr: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.T().Run(tc.name, func(t *testing.T) {
			_, err := msgServer.Delegate(ctx, tc.input)
			if tc.expErr {
				require.Error(err)
				require.Contains(err.Error(), tc.expErrMsg)
			} else {
				require.NoError(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestMsgDelegateLicenseMode() {
	ctx, keeper, msgServer := s.ctx, s.stakingKeeper, s.msgServer
	require := s.Require()
	s.execExpectCalls()

	pk := ed25519.GenPrivKey().PubKey()
	require.NotNil(pk)

	comm := stakingtypes.NewCommissionRates(math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0))
	s.stakingKeeper.SetNewValidatorApprovalState(ctx, stakingtypes.ValidatorApproval{ApproverAddress: pk.Address().String(), Enabled: false})

	msg, err := stakingtypes.NewMsgCreateValidator(ValAddr.String(), pk.Address().String(), pk, sdk.NewCoin("stake", math.NewInt(500000000000)), stakingtypes.Description{Moniker: "NewVal"}, comm, math.NewInt(10000000000))
	msg.Mode = stakingtypes.ValidatorMode_MODE_LICENSE
	msg.EnableRedelegation = false
	msg.DelegationIncrement = math.NewInt(10000000000)
	msg.MinDelegation = math.NewInt(10000000000)
	msg.MaxLicense = math.NewInt(150)
	require.NoError(err)

	res, err := msgServer.CreateValidator(ctx, msg)
	require.NoError(err)
	require.NotNil(res)

	testCases := []struct {
		name      string
		input     *stakingtypes.MsgDelegate
		expErr    bool
		expErrMsg string
	}{
		{
			name: "invalid validator",
			input: &stakingtypes.MsgDelegate{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: sdk.AccAddress([]byte("invalid")).String(),
				Amount:           sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: keeper.TokensFromConsensusPower(s.ctx, int64(100))},
			},
			expErr:    true,
			expErrMsg: "invalid validator address",
		},
		{
			name: "empty delegator",
			input: &stakingtypes.MsgDelegate{
				DelegatorAddress: "",
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: keeper.TokensFromConsensusPower(s.ctx, int64(100))},
			},
			expErr:    true,
			expErrMsg: "invalid delegator address: empty address string is not allowed",
		},
		{
			name: "invalid delegator",
			input: &stakingtypes.MsgDelegate{
				DelegatorAddress: "invalid",
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: keeper.TokensFromConsensusPower(s.ctx, int64(100))},
			},
			expErr:    true,
			expErrMsg: "invalid delegator address: decoding bech32 failed",
		},
		{
			name: "validator does not exist",
			input: &stakingtypes.MsgDelegate{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: sdk.ValAddress([]byte("val")).String(),
				Amount:           sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: keeper.TokensFromConsensusPower(s.ctx, int64(100))},
			},
			expErr:    true,
			expErrMsg: "validator does not exist",
		},
		{
			name: "zero amount",
			input: &stakingtypes.MsgDelegate{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: keeper.TokensFromConsensusPower(s.ctx, int64(0))},
			},
			expErr:    true,
			expErrMsg: "invalid delegation amount",
		},
		{
			name: "negative amount",
			input: &stakingtypes.MsgDelegate{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: keeper.TokensFromConsensusPower(s.ctx, int64(-1))},
			},
			expErr:    true,
			expErrMsg: "invalid delegation amount",
		},
		{
			name: "invalid BondDenom",
			input: &stakingtypes.MsgDelegate{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.Coin{Denom: "test", Amount: keeper.TokensFromConsensusPower(s.ctx, int64(100))},
			},
			expErr:    true,
			expErrMsg: "invalid coin denomination",
		},
		{
			// a delegator without an existing delegation must meet MinDelegation
			name: "invalid minimun increment amount msg",
			input: &stakingtypes.MsgDelegate{
				DelegatorAddress: sdk.AccAddress([]byte("fresh_delegator")).String(),
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: keeper.TokensFromConsensusPower(s.ctx, int64(101))},
			},
			expErr:    true,
			expErrMsg: "delegation amount less than minimum",
		},
		{
			// Addr already holds the self-delegation, so the minimum is skipped
			// and the amount must be a multiple of DelegationIncrement
			name: "invalid increment for existing delegation",
			input: &stakingtypes.MsgDelegate{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: keeper.TokensFromConsensusPower(s.ctx, int64(101))},
			},
			expErr:    true,
			expErrMsg: "delegation amount must meet increment condition",
		},
		{
			name: "valid msg - small delegation",
			input: &stakingtypes.MsgDelegate{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: math.NewInt(10000000000)}, // MinDelegation amount, adds 1 license
			},
			expErr: false,
		},
		{
			name: "valid msg - multiple increments",
			input: &stakingtypes.MsgDelegate{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: math.NewInt(50000000000)}, // MinDelegation + 4 increments, adds 5 more licenses (total 1+5=6)
			},
			expErr: false,
		},
		{
			name: "exceed max license with large delegation",
			input: &stakingtypes.MsgDelegate{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: math.NewInt(1000000000000)}, // MinDelegation + 99 increments = 100 licenses, would exceed max (50+6+100=156 > 150)
			},
			expErr:    true,
			expErrMsg: "There is no license enough for the delegation",
		},
		{
			name: "reach exactly max license",
			input: &stakingtypes.MsgDelegate{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: math.NewInt(940000000000)}, // Adds 94 licenses to reach exactly 150 (50+6+94=150)
			},
			expErr: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.T().Run(tc.name, func(t *testing.T) {
			_, err := msgServer.Delegate(ctx, tc.input)
			if tc.expErr {
				require.Error(err)
				require.Contains(err.Error(), tc.expErrMsg)
			} else {
				require.NoError(err)
				// validVal, _ := keeper.GetValidator(ctx, ValAddr)
				// t.Logf("Validator: %+v", validVal)
				// t.Logf("LicenseCount: %v", validVal.LicenseCount)
				// t.Logf("MaxLicense: %v", validVal.MaxLicense)
			}
		})
	}
}

func (s *KeeperTestSuite) TestMsgBeginRedelegate() {
	ctx, keeper, msgServer := s.ctx, s.stakingKeeper, s.msgServer
	require := s.Require()
	s.execExpectCalls()

	srcValAddr := ValAddr
	addr2 := sdk.AccAddress(PKS[1].Address())
	dstValAddr := sdk.ValAddress(addr2)

	pk := ed25519.GenPrivKey().PubKey()
	require.NotNil(pk)
	dstPk := ed25519.GenPrivKey().PubKey()
	require.NotNil(dstPk)

	keeper.SetNewValidatorApprovalState(ctx, stakingtypes.ValidatorApproval{ApproverAddress: pk.Address().String(), Enabled: false})

	comm := stakingtypes.NewCommissionRates(math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0))
	amt := sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: keeper.TokensFromConsensusPower(s.ctx, int64(100))}

	msg, err := stakingtypes.NewMsgCreateValidator(srcValAddr.String(), pk.Address().String(), pk, amt, stakingtypes.Description{Moniker: "NewVal"}, comm, math.OneInt())
	require.NoError(err)
	res, err := msgServer.CreateValidator(ctx, msg)
	require.NoError(err)
	require.NotNil(res)
	s.bankKeeper.EXPECT().DelegateCoinsFromAccountToModule(gomock.Any(), addr2, stakingtypes.NotBondedPoolName, gomock.Any()).AnyTimes()

	msg, err = stakingtypes.NewMsgCreateValidator(dstValAddr.String(), dstPk.Address().String(), dstPk, amt, stakingtypes.Description{Moniker: "NewVal"}, comm, math.OneInt())
	require.NoError(err)

	res, err = msgServer.CreateValidator(ctx, msg)
	require.NoError(err)
	require.NotNil(res)

	shares := math.LegacyNewDec(100)
	del := stakingtypes.NewDelegation(Addr.String(), srcValAddr.String(), shares)
	require.NoError(keeper.SetDelegation(ctx, del))
	_, err = keeper.GetDelegation(ctx, Addr, srcValAddr)
	require.NoError(err)

	testCases := []struct {
		name      string
		input     *stakingtypes.MsgBeginRedelegate
		expErr    bool
		expErrMsg string
	}{
		{
			name: "invalid source validator",
			input: &stakingtypes.MsgBeginRedelegate{
				DelegatorAddress:    Addr.String(),
				ValidatorSrcAddress: sdk.AccAddress([]byte("invalid")).String(),
				ValidatorDstAddress: dstValAddr.String(),
				Amount:              sdk.NewCoin(sdk.DefaultBondDenom, shares.RoundInt()),
			},
			expErr:    true,
			expErrMsg: "invalid source validator address",
		},
		{
			name: "empty delegator",
			input: &stakingtypes.MsgBeginRedelegate{
				DelegatorAddress:    "",
				ValidatorSrcAddress: srcValAddr.String(),
				ValidatorDstAddress: dstValAddr.String(),
				Amount:              sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: keeper.TokensFromConsensusPower(s.ctx, int64(100))},
			},
			expErr:    true,
			expErrMsg: "invalid delegator address: empty address string is not allowed",
		},
		{
			name: "invalid delegator",
			input: &stakingtypes.MsgBeginRedelegate{
				DelegatorAddress:    "invalid",
				ValidatorSrcAddress: srcValAddr.String(),
				ValidatorDstAddress: dstValAddr.String(),
				Amount:              sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: keeper.TokensFromConsensusPower(s.ctx, int64(100))},
			},
			expErr:    true,
			expErrMsg: "invalid delegator address: decoding bech32 failed: invalid bech32 string length 7",
		},
		{
			name: "invalid destination validator",
			input: &stakingtypes.MsgBeginRedelegate{
				DelegatorAddress:    Addr.String(),
				ValidatorSrcAddress: srcValAddr.String(),
				ValidatorDstAddress: sdk.AccAddress([]byte("invalid")).String(),
				Amount:              sdk.NewCoin(sdk.DefaultBondDenom, shares.RoundInt()),
			},
			expErr:    true,
			expErrMsg: "invalid destination validator address",
		},
		{
			name: "validator does not exist",
			input: &stakingtypes.MsgBeginRedelegate{
				DelegatorAddress:    Addr.String(),
				ValidatorSrcAddress: sdk.ValAddress([]byte("invalid")).String(),
				ValidatorDstAddress: dstValAddr.String(),
				Amount:              sdk.NewCoin(sdk.DefaultBondDenom, shares.RoundInt()),
			},
			expErr:    true,
			expErrMsg: "validator does not exist",
		},
		{
			name: "self redelegation",
			input: &stakingtypes.MsgBeginRedelegate{
				DelegatorAddress:    Addr.String(),
				ValidatorSrcAddress: srcValAddr.String(),
				ValidatorDstAddress: srcValAddr.String(),
				Amount:              sdk.NewCoin(sdk.DefaultBondDenom, shares.RoundInt()),
			},
			expErr:    true,
			expErrMsg: "cannot redelegate to the same validator",
		},
		{
			name: "amount greater than delegated shares amount",
			input: &stakingtypes.MsgBeginRedelegate{
				DelegatorAddress:    Addr.String(),
				ValidatorSrcAddress: srcValAddr.String(),
				ValidatorDstAddress: dstValAddr.String(),
				Amount:              sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(101)),
			},
			expErr:    true,
			expErrMsg: "invalid shares amount",
		},
		{
			name: "zero amount",
			input: &stakingtypes.MsgBeginRedelegate{
				DelegatorAddress:    Addr.String(),
				ValidatorSrcAddress: srcValAddr.String(),
				ValidatorDstAddress: dstValAddr.String(),
				Amount:              sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(0)),
			},
			expErr:    true,
			expErrMsg: "invalid shares amount",
		},
		{
			name: "invalid coin denom",
			input: &stakingtypes.MsgBeginRedelegate{
				DelegatorAddress:    Addr.String(),
				ValidatorSrcAddress: srcValAddr.String(),
				ValidatorDstAddress: dstValAddr.String(),
				Amount:              sdk.NewCoin("test", shares.RoundInt()),
			},
			expErr:    true,
			expErrMsg: "invalid coin denomination",
		},
		{
			name: "redelegation is force-disabled",
			input: &stakingtypes.MsgBeginRedelegate{
				DelegatorAddress:    Addr.String(),
				ValidatorSrcAddress: srcValAddr.String(),
				ValidatorDstAddress: dstValAddr.String(),
				Amount:              sdk.NewCoin(sdk.DefaultBondDenom, shares.RoundInt()),
			},
			expErr:    true,
			expErrMsg: "Redelegation is disable",
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.T().Run(tc.name, func(t *testing.T) {
			_, err := msgServer.BeginRedelegate(ctx, tc.input)
			if tc.expErr {
				require.Error(err)
				require.Contains(err.Error(), tc.expErrMsg)
			} else {
				require.NoError(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestRedelegationForceDisabled() {
	ctx, keeper, msgServer := s.ctx, s.stakingKeeper, s.msgServer
	require := s.Require()
	s.execExpectCalls()

	increment := math.NewInt(10000000000)

	srcValAddr := ValAddr
	dstAccAddr := sdk.AccAddress(PKS[1].Address())
	dstValAddr := sdk.ValAddress(dstAccAddr)

	srcPk := ed25519.GenPrivKey().PubKey()
	dstPk := ed25519.GenPrivKey().PubKey()

	keeper.SetNewValidatorApprovalState(ctx, stakingtypes.ValidatorApproval{ApproverAddress: srcPk.Address().String(), Enabled: false})
	s.bankKeeper.EXPECT().DelegateCoinsFromAccountToModule(gomock.Any(), dstAccAddr, stakingtypes.NotBondedPoolName, gomock.Any()).AnyTimes()

	comm := stakingtypes.NewCommissionRates(math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0))
	selfBond := sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(500000000000))

	// asking for redelegation at create-validator is ignored: the create
	// succeeds and the flag is force-stored as disabled
	msg, err := stakingtypes.NewMsgCreateValidator(srcValAddr.String(), srcPk.Address().String(), srcPk, selfBond, stakingtypes.Description{Moniker: "SrcVal"}, comm, increment)
	require.NoError(err)
	msg.Mode = stakingtypes.ValidatorMode_MODE_LICENSE
	msg.EnableRedelegation = true
	msg.DelegationIncrement = increment
	msg.MinDelegation = increment
	msg.MaxLicense = math.NewInt(150)
	_, err = msgServer.CreateValidator(ctx, msg)
	require.NoError(err)
	srcVal, err := keeper.GetValidator(ctx, srcValAddr)
	require.NoError(err)
	require.False(srcVal.EnableRedelegation)
	require.Equal(math.NewInt(50), srcVal.LicenseCount)

	msg, err = stakingtypes.NewMsgCreateValidator(dstValAddr.String(), dstPk.Address().String(), dstPk, selfBond, stakingtypes.Description{Moniker: "DstVal"}, comm, math.OneInt())
	require.NoError(err)
	_, err = msgServer.CreateValidator(ctx, msg)
	require.NoError(err)

	// edit-validator cannot turn redelegation on: the edit itself succeeds
	// but the flag is force-kept disabled
	_, err = msgServer.EditValidator(ctx, &stakingtypes.MsgEditValidator{
		ValidatorAddress:   srcValAddr.String(),
		Description:        stakingtypes.Description{Moniker: "SrcVal"},
		EnableRedelegation: stakingtypes.RedelegationUpdate_REDELEGATION_UPDATE_ENABLE,
	})
	require.NoError(err)
	srcVal, err = keeper.GetValidator(ctx, srcValAddr)
	require.NoError(err)
	require.False(srcVal.EnableRedelegation)

	// even a stale true in the store is cleaned up by any edit
	srcVal.EnableRedelegation = true
	require.NoError(keeper.SetValidator(ctx, srcVal))
	_, err = msgServer.EditValidator(ctx, &stakingtypes.MsgEditValidator{
		ValidatorAddress: srcValAddr.String(),
		Description:      stakingtypes.Description{Moniker: "SrcVal"},
	})
	require.NoError(err)
	srcVal, err = keeper.GetValidator(ctx, srcValAddr)
	require.NoError(err)
	require.False(srcVal.EnableRedelegation)

	// with no way to enable the flag, every redelegation fails the flag check
	_, err = msgServer.BeginRedelegate(ctx, &stakingtypes.MsgBeginRedelegate{
		DelegatorAddress:    Addr.String(),
		ValidatorSrcAddress: srcValAddr.String(),
		ValidatorDstAddress: dstValAddr.String(),
		Amount:              sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(10000000000)),
	})
	require.Error(err)
	require.Contains(err.Error(), "Redelegation is disable")

	// license accounting is untouched by the rejected attempts
	srcVal, err = keeper.GetValidator(ctx, srcValAddr)
	require.NoError(err)
	require.Equal(math.NewInt(50), srcVal.LicenseCount)
}

func (s *KeeperTestSuite) TestMsgUndelegate() {
	ctx, keeper, msgServer := s.ctx, s.stakingKeeper, s.msgServer
	require := s.Require()
	s.execExpectCalls()

	pk := ed25519.GenPrivKey().PubKey()
	require.NotNil(pk)

	keeper.SetNewValidatorApprovalState(ctx, stakingtypes.ValidatorApproval{ApproverAddress: pk.Address().String(), Enabled: false})

	comm := stakingtypes.NewCommissionRates(math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0))
	amt := sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: keeper.TokensFromConsensusPower(s.ctx, int64(100))}
	msg, err := stakingtypes.NewMsgCreateValidator(ValAddr.String(), pk.Address().String(), pk, amt, stakingtypes.Description{Moniker: "NewVal"}, comm, math.OneInt())
	require.NoError(err)
	res, err := msgServer.CreateValidator(ctx, msg)
	require.NoError(err)
	require.NotNil(res)

	shares := math.LegacyNewDec(100)
	del := stakingtypes.NewDelegation(Addr.String(), ValAddr.String(), shares)
	require.NoError(keeper.SetDelegation(ctx, del))
	_, err = keeper.GetDelegation(ctx, Addr, ValAddr)
	require.NoError(err)

	testCases := []struct {
		name      string
		input     *stakingtypes.MsgUndelegate
		expErr    bool
		expErrMsg string
	}{
		{
			name: "invalid validator",
			input: &stakingtypes.MsgUndelegate{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: sdk.AccAddress([]byte("invalid")).String(),
				Amount:           sdk.NewCoin(sdk.DefaultBondDenom, shares.RoundInt()),
			},
			expErr:    true,
			expErrMsg: "invalid validator address",
		},
		{
			name: "empty delegator",
			input: &stakingtypes.MsgUndelegate{
				DelegatorAddress: "",
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: shares.RoundInt()},
			},
			expErr:    true,
			expErrMsg: "invalid delegator address: empty address string is not allowed",
		},
		{
			name: "invalid delegator",
			input: &stakingtypes.MsgUndelegate{
				DelegatorAddress: "invalid",
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: shares.RoundInt()},
			},
			expErr:    true,
			expErrMsg: "invalid delegator address: decoding bech32 failed",
		},
		{
			name: "validator does not exist",
			input: &stakingtypes.MsgUndelegate{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: sdk.ValAddress([]byte("invalid")).String(),
				Amount:           sdk.NewCoin(sdk.DefaultBondDenom, shares.RoundInt()),
			},
			expErr:    true,
			expErrMsg: "validator does not exist",
		},
		{
			name: "amount greater than delegated shares amount",
			input: &stakingtypes.MsgUndelegate{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(101)),
			},
			expErr:    true,
			expErrMsg: "invalid shares amount",
		},
		{
			name: "zero amount",
			input: &stakingtypes.MsgUndelegate{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(0)),
			},
			expErr:    true,
			expErrMsg: "invalid shares amount",
		},
		{
			name: "invalid coin denom",
			input: &stakingtypes.MsgUndelegate{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.NewCoin("test", shares.RoundInt()),
			},
			expErr:    true,
			expErrMsg: "invalid coin denomination",
		},
		{
			name: "valid msg",
			input: &stakingtypes.MsgUndelegate{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.NewCoin(sdk.DefaultBondDenom, shares.RoundInt()),
			},
			expErr: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.T().Run(tc.name, func(t *testing.T) {
			_, err := msgServer.Undelegate(ctx, tc.input)
			if tc.expErr {
				require.Error(err)
				require.Contains(err.Error(), tc.expErrMsg)
			} else {
				require.NoError(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestMsgCancelUnbondingDelegation() {
	ctx, keeper, msgServer, ak := s.ctx, s.stakingKeeper, s.msgServer, s.accountKeeper
	require := s.Require()

	pk := ed25519.GenPrivKey().PubKey()
	require.NotNil(pk)

	comm := stakingtypes.NewCommissionRates(math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0))
	amt := sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: keeper.TokensFromConsensusPower(s.ctx, int64(100))}

	s.bankKeeper.EXPECT().DelegateCoinsFromAccountToModule(gomock.Any(), Addr, stakingtypes.NotBondedPoolName, gomock.Any()).AnyTimes()

	// set approval
	keeper.SetNewValidatorApprovalState(ctx, stakingtypes.ValidatorApproval{
		ApproverAddress: pk.Address().String(),
		Enabled:         true,
	})

	msg, err := stakingtypes.NewMsgCreateValidator(ValAddr.String(), pk.Address().String(), pk, amt, stakingtypes.Description{Moniker: "NewVal"}, comm, math.OneInt())
	require.NoError(err)
	res, err := msgServer.CreateValidator(ctx, msg)
	require.NoError(err)
	require.NotNil(res)

	shares := math.LegacyNewDec(100)
	del := stakingtypes.NewDelegation(Addr.String(), ValAddr.String(), shares)
	require.NoError(keeper.SetDelegation(ctx, del))
	resDel, err := keeper.GetDelegation(ctx, Addr, ValAddr)
	require.NoError(err)
	require.Equal(del, resDel)

	ubd := stakingtypes.NewUnbondingDelegation(Addr, ValAddr, 10, ctx.BlockTime().Add(time.Minute*10), shares.RoundInt(), 0, keeper.ValidatorAddressCodec(), ak.AddressCodec())
	require.NoError(keeper.SetUnbondingDelegation(ctx, ubd))
	resUnbond, err := keeper.GetUnbondingDelegation(ctx, Addr, ValAddr)
	require.NoError(err)
	require.Equal(ubd, resUnbond)

	testCases := []struct {
		name      string
		input     *stakingtypes.MsgCancelUnbondingDelegation
		expErr    bool
		expErrMsg string
	}{
		{
			name: "invalid validator",
			input: &stakingtypes.MsgCancelUnbondingDelegation{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: sdk.AccAddress([]byte("invalid")).String(),
				Amount:           sdk.NewCoin(sdk.DefaultBondDenom, shares.RoundInt()),
				CreationHeight:   10,
			},
			expErr:    true,
			expErrMsg: "invalid validator address",
		},
		{
			name: "empty delegator",
			input: &stakingtypes.MsgCancelUnbondingDelegation{
				DelegatorAddress: "",
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.NewCoin(sdk.DefaultBondDenom, shares.RoundInt()),
				CreationHeight:   10,
			},
			expErr:    true,
			expErrMsg: "invalid delegator address: empty address string is not allowed",
		},
		{
			name: "invalid delegator",
			input: &stakingtypes.MsgCancelUnbondingDelegation{
				DelegatorAddress: "invalid",
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.NewCoin(sdk.DefaultBondDenom, shares.RoundInt()),
				CreationHeight:   10,
			},
			expErr:    true,
			expErrMsg: "invalid delegator address: decoding bech32 failed",
		},
		{
			name: "entry not found at height",
			input: &stakingtypes.MsgCancelUnbondingDelegation{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.NewCoin(sdk.DefaultBondDenom, shares.RoundInt()),
				CreationHeight:   11,
			},
			expErr:    true,
			expErrMsg: "unbonding delegation entry is not found at block height",
		},
		{
			name: "invalid height",
			input: &stakingtypes.MsgCancelUnbondingDelegation{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.NewCoin(sdk.DefaultBondDenom, shares.RoundInt()),
				CreationHeight:   -1,
			},
			expErr:    true,
			expErrMsg: "invalid height",
		},
		{
			name: "invalid coin",
			input: &stakingtypes.MsgCancelUnbondingDelegation{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.NewCoin("test", shares.RoundInt()),
				CreationHeight:   10,
			},
			expErr:    true,
			expErrMsg: "invalid coin denomination",
		},
		{
			name: "validator does not exist",
			input: &stakingtypes.MsgCancelUnbondingDelegation{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: sdk.ValAddress([]byte("invalid")).String(),
				Amount:           sdk.NewCoin(sdk.DefaultBondDenom, shares.RoundInt()),
				CreationHeight:   10,
			},
			expErr:    true,
			expErrMsg: "validator does not exist",
		},
		{
			name: "amount is greater than balance",
			input: &stakingtypes.MsgCancelUnbondingDelegation{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(101)),
				CreationHeight:   10,
			},
			expErr:    true,
			expErrMsg: "amount is greater than the unbonding delegation entry balance",
		},
		{
			name: "zero amount",
			input: &stakingtypes.MsgCancelUnbondingDelegation{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(0)),
				CreationHeight:   10,
			},
			expErr:    true,
			expErrMsg: "invalid amount",
		},
		{
			name: "valid msg",
			input: &stakingtypes.MsgCancelUnbondingDelegation{
				DelegatorAddress: Addr.String(),
				ValidatorAddress: ValAddr.String(),
				Amount:           sdk.NewCoin(sdk.DefaultBondDenom, shares.RoundInt()),
				CreationHeight:   10,
			},
			expErr: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.T().Run(tc.name, func(t *testing.T) {
			_, err := msgServer.CancelUnbondingDelegation(ctx, tc.input)
			if tc.expErr {
				require.Error(err)
				require.Contains(err.Error(), tc.expErrMsg)
			} else {
				require.NoError(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestMsgUpdateParams() {
	ctx, keeper, msgServer := s.ctx, s.stakingKeeper, s.msgServer
	require := s.Require()

	testCases := []struct {
		name      string
		input     *stakingtypes.MsgUpdateParams
		expErr    bool
		expErrMsg string
	}{
		{
			name: "valid params",
			input: &stakingtypes.MsgUpdateParams{
				Authority: keeper.GetAuthority(),
				Params:    stakingtypes.DefaultParams(),
			},
			expErr: false,
		},
		{
			name: "invalid authority",
			input: &stakingtypes.MsgUpdateParams{
				Authority: "invalid",
				Params:    stakingtypes.DefaultParams(),
			},
			expErr:    true,
			expErrMsg: "invalid authority",
		},
		{
			name: "negative commission rate",
			input: &stakingtypes.MsgUpdateParams{
				Authority: keeper.GetAuthority(),
				Params: stakingtypes.Params{
					MinCommissionRate: math.LegacyNewDec(-10),
					UnbondingTime:     stakingtypes.DefaultUnbondingTime,
					MaxValidators:     stakingtypes.DefaultMaxValidators,
					MaxEntries:        stakingtypes.DefaultMaxEntries,
					HistoricalEntries: stakingtypes.DefaultHistoricalEntries,
					BondDenom:         stakingtypes.BondStatusBonded,
				},
			},
			expErr:    true,
			expErrMsg: "minimum commission rate cannot be negative",
		},
		{
			name: "commission rate cannot be bigger than 100",
			input: &stakingtypes.MsgUpdateParams{
				Authority: keeper.GetAuthority(),
				Params: stakingtypes.Params{
					MinCommissionRate: math.LegacyNewDec(2),
					UnbondingTime:     stakingtypes.DefaultUnbondingTime,
					MaxValidators:     stakingtypes.DefaultMaxValidators,
					MaxEntries:        stakingtypes.DefaultMaxEntries,
					HistoricalEntries: stakingtypes.DefaultHistoricalEntries,
					BondDenom:         stakingtypes.BondStatusBonded,
				},
			},
			expErr:    true,
			expErrMsg: "minimum commission rate cannot be greater than 100%",
		},
		{
			name: "invalid bond denom",
			input: &stakingtypes.MsgUpdateParams{
				Authority: keeper.GetAuthority(),
				Params: stakingtypes.Params{
					MinCommissionRate: stakingtypes.DefaultMinCommissionRate,
					UnbondingTime:     stakingtypes.DefaultUnbondingTime,
					MaxValidators:     stakingtypes.DefaultMaxValidators,
					MaxEntries:        stakingtypes.DefaultMaxEntries,
					HistoricalEntries: stakingtypes.DefaultHistoricalEntries,
					BondDenom:         "",
				},
			},
			expErr:    true,
			expErrMsg: "bond denom cannot be blank",
		},
		{
			name: "max validators must be positive",
			input: &stakingtypes.MsgUpdateParams{
				Authority: keeper.GetAuthority(),
				Params: stakingtypes.Params{
					MinCommissionRate: stakingtypes.DefaultMinCommissionRate,
					UnbondingTime:     stakingtypes.DefaultUnbondingTime,
					MaxValidators:     0,
					MaxEntries:        stakingtypes.DefaultMaxEntries,
					HistoricalEntries: stakingtypes.DefaultHistoricalEntries,
					BondDenom:         stakingtypes.BondStatusBonded,
				},
			},
			expErr:    true,
			expErrMsg: "max validators must be positive",
		},
		{
			name: "max entries most be positive",
			input: &stakingtypes.MsgUpdateParams{
				Authority: keeper.GetAuthority(),
				Params: stakingtypes.Params{
					MinCommissionRate: stakingtypes.DefaultMinCommissionRate,
					UnbondingTime:     stakingtypes.DefaultUnbondingTime,
					MaxValidators:     stakingtypes.DefaultMaxValidators,
					MaxEntries:        0,
					HistoricalEntries: stakingtypes.DefaultHistoricalEntries,
					BondDenom:         stakingtypes.BondStatusBonded,
				},
			},
			expErr:    true,
			expErrMsg: "max entries must be positive",
		},
		{
			name: "negative unbounding time",
			input: &stakingtypes.MsgUpdateParams{
				Authority: keeper.GetAuthority(),
				Params: stakingtypes.Params{
					UnbondingTime:     time.Hour * 24 * 7 * 3 * -1,
					MaxEntries:        stakingtypes.DefaultMaxEntries,
					MaxValidators:     stakingtypes.DefaultMaxValidators,
					HistoricalEntries: stakingtypes.DefaultHistoricalEntries,
					MinCommissionRate: stakingtypes.DefaultMinCommissionRate,
					BondDenom:         "denom",
				},
			},
			expErr:    true,
			expErrMsg: "unbonding time must be positive",
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.T().Run(tc.name, func(t *testing.T) {
			_, err := msgServer.UpdateParams(ctx, tc.input)
			if tc.expErr {
				require.Error(err)
				require.Contains(err.Error(), tc.expErrMsg)
			} else {
				require.NoError(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestMsgSetValidatorApproval() {
	ctx, keeper, msgServer := s.ctx, s.stakingKeeper, s.msgServer
	require := s.Require()

	approver := Addr
	newApprover := sdk.AccAddress(PKS[1].Address())

	// the approval state is only written by InitGenesis; before it exists the
	// message must fail instead of silently creating one
	_, err := keeper.GetValidatorApproval(ctx)
	require.ErrorIs(err, stakingtypes.ErrNoValidatorFound)
	_, err = msgServer.SetValidatorApproval(ctx, &stakingtypes.MsgSetValidatorApproval{
		ApproverAddress:    approver.String(),
		NewApproverAddress: newApprover.String(),
		Enabled:            true,
	})
	require.Error(err)
	require.Contains(err.Error(), "Validator approval is somehow does not existed")

	require.NoError(keeper.SetNewValidatorApprovalState(ctx, stakingtypes.ValidatorApproval{
		ApproverAddress: approver.String(),
		Enabled:         true,
	}))

	testCases := []struct {
		name      string
		input     *stakingtypes.MsgSetValidatorApproval
		expErr    bool
		expErrMsg string
	}{
		{
			name: "sender is not the current approver",
			input: &stakingtypes.MsgSetValidatorApproval{
				ApproverAddress:    newApprover.String(),
				NewApproverAddress: newApprover.String(),
				Enabled:            true,
			},
			expErr:    true,
			expErrMsg: "Msg sender is not current approver",
		},
		{
			name: "invalid new approver address",
			input: &stakingtypes.MsgSetValidatorApproval{
				ApproverAddress:    approver.String(),
				NewApproverAddress: "invalid",
				Enabled:            true,
			},
			expErr:    true,
			expErrMsg: "Invalid new approver address",
		},
		{
			name: "valid handover to new approver",
			input: &stakingtypes.MsgSetValidatorApproval{
				ApproverAddress:    approver.String(),
				NewApproverAddress: newApprover.String(),
				Enabled:            false,
			},
			expErr: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.T().Run(tc.name, func(t *testing.T) {
			_, err := msgServer.SetValidatorApproval(ctx, tc.input)
			if tc.expErr {
				require.Error(err)
				require.Contains(err.Error(), tc.expErrMsg)
			} else {
				require.NoError(err)
			}
		})
	}

	approval, err := keeper.GetValidatorApproval(ctx)
	require.NoError(err)
	require.Equal(newApprover.String(), approval.ApproverAddress)
	require.False(approval.Enabled)

	// the previous approver lost control with the handover
	_, err = msgServer.SetValidatorApproval(ctx, &stakingtypes.MsgSetValidatorApproval{
		ApproverAddress:    approver.String(),
		NewApproverAddress: approver.String(),
		Enabled:            true,
	})
	require.Error(err)
	require.Contains(err.Error(), "Msg sender is not current approver")
}

func (s *KeeperTestSuite) TestMsgCreateValidatorApproval() {
	ctx, keeper, msgServer := s.ctx, s.stakingKeeper, s.msgServer
	require := s.Require()
	s.execExpectCalls()

	approver := sdk.AccAddress(PKS[1].Address())
	comm := stakingtypes.NewCommissionRates(math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0))
	selfBond := sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(10000))

	require.NoError(keeper.SetNewValidatorApprovalState(ctx, stakingtypes.ValidatorApproval{
		ApproverAddress: approver.String(),
		Enabled:         true,
	}))

	// while approval is enabled, a message naming the wrong approver is rejected
	pk := ed25519.GenPrivKey().PubKey()
	msg, err := stakingtypes.NewMsgCreateValidator(ValAddr.String(), Addr.String(), pk, selfBond, stakingtypes.Description{Moniker: "NewVal"}, comm, math.OneInt())
	require.NoError(err)
	_, err = msgServer.CreateValidator(ctx, msg)
	require.Error(err)
	require.Contains(err.Error(), "Wrong approver for create validator")

	// the configured approver address passes the gate
	msg, err = stakingtypes.NewMsgCreateValidator(ValAddr.String(), approver.String(), pk, selfBond, stakingtypes.Description{Moniker: "NewVal"}, comm, math.OneInt())
	require.NoError(err)
	_, err = msgServer.CreateValidator(ctx, msg)
	require.NoError(err)

	// with approval disabled the approver field is not checked at all
	require.NoError(keeper.SetNewValidatorApprovalState(ctx, stakingtypes.ValidatorApproval{
		ApproverAddress: approver.String(),
		Enabled:         false,
	}))
	otherAddr := sdk.AccAddress(PKS[2].Address())
	s.bankKeeper.EXPECT().DelegateCoinsFromAccountToModule(gomock.Any(), otherAddr, stakingtypes.NotBondedPoolName, gomock.Any()).AnyTimes()
	otherPk := ed25519.GenPrivKey().PubKey()
	msg, err = stakingtypes.NewMsgCreateValidator(sdk.ValAddress(otherAddr).String(), Addr.String(), otherPk, selfBond, stakingtypes.Description{Moniker: "OtherVal"}, comm, math.OneInt())
	require.NoError(err)
	_, err = msgServer.CreateValidator(ctx, msg)
	require.NoError(err)
}

func (s *KeeperTestSuite) TestMsgCreateValidatorLicenseMode() {
	ctx, keeper, msgServer := s.ctx, s.stakingKeeper, s.msgServer
	require := s.Require()
	s.execExpectCalls()

	increment := math.NewInt(10000000000)
	pk := ed25519.GenPrivKey().PubKey()
	comm := stakingtypes.NewCommissionRates(math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0))
	require.NoError(keeper.SetNewValidatorApprovalState(ctx, stakingtypes.ValidatorApproval{ApproverAddress: pk.Address().String(), Enabled: false}))

	// baseline license-mode msg; each case mutates a copy of it
	newLicenseMsg := func() *stakingtypes.MsgCreateValidator {
		msg, err := stakingtypes.NewMsgCreateValidator(ValAddr.String(), pk.Address().String(), pk, sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(500000000000)), stakingtypes.Description{Moniker: "LicenseVal"}, comm, math.OneInt())
		require.NoError(err)
		msg.Mode = stakingtypes.ValidatorMode_MODE_LICENSE
		msg.DelegationIncrement = increment
		msg.MinDelegation = increment
		msg.MaxLicense = math.NewInt(150)
		return msg
	}

	testCases := []struct {
		name      string
		mutate    func(*stakingtypes.MsgCreateValidator)
		expErr    bool
		expErrMsg string
	}{
		{
			name:      "missing max license",
			mutate:    func(msg *stakingtypes.MsgCreateValidator) { msg.MaxLicense = math.Int{} },
			expErr:    true,
			expErrMsg: "max license is required when license mode is used",
		},
		{
			name:      "non-positive max license",
			mutate:    func(msg *stakingtypes.MsgCreateValidator) { msg.MaxLicense = math.NewInt(0) },
			expErr:    true,
			expErrMsg: "max license must be a positive integer",
		},
		{
			name: "missing delegation increment",
			mutate: func(msg *stakingtypes.MsgCreateValidator) {
				msg.DelegationIncrement = math.Int{}
				msg.MinDelegation = math.Int{}
			},
			expErr:    true,
			expErrMsg: "Min Delegation and DelegationIncrement must be defined and the same",
		},
		{
			name: "min delegation differs from increment",
			mutate: func(msg *stakingtypes.MsgCreateValidator) {
				msg.MinDelegation = math.NewInt(20000000000)
			},
			expErr:    true,
			expErrMsg: "Min Delegation and DelegationIncrement must be defined and the same",
		},
		{
			name: "self bond not a multiple of increment",
			mutate: func(msg *stakingtypes.MsgCreateValidator) {
				msg.Value = sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(500000000005))
			},
			expErr:    true,
			expErrMsg: "delegation amount must meet increment condition",
		},
		{
			name: "self bond exceeds max license",
			mutate: func(msg *stakingtypes.MsgCreateValidator) {
				msg.Value = sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1510000000000))
			},
			expErr:    true,
			expErrMsg: "There is no license enough for the delegation",
		},
		{
			name:      "unknown mode enum value",
			mutate:    func(msg *stakingtypes.MsgCreateValidator) { msg.Mode = stakingtypes.ValidatorMode(99) },
			expErr:    true,
			expErrMsg: "Invalid validator mode",
		},
		{
			name:   "valid license validator",
			mutate: func(msg *stakingtypes.MsgCreateValidator) {},
			expErr: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.T().Run(tc.name, func(t *testing.T) {
			msg := newLicenseMsg()
			tc.mutate(msg)
			_, err := msgServer.CreateValidator(ctx, msg)
			if tc.expErr {
				require.Error(err)
				require.Contains(err.Error(), tc.expErrMsg)
			} else {
				require.NoError(err)
			}
		})
	}

	// the created validator carries the license bookkeeping from the msg
	validator, err := keeper.GetValidator(ctx, ValAddr)
	require.NoError(err)
	require.Equal(stakingtypes.ValidatorMode_MODE_LICENSE, validator.Mode)
	require.Equal(math.NewInt(50), validator.LicenseCount)
	require.Equal(math.NewInt(150), validator.MaxLicense)
	require.Equal(increment, validator.MinDelegation)
	require.Equal(increment, validator.DelegationIncrement)
	require.False(validator.EnableRedelegation)

	// MinDelegation left empty defaults to DelegationIncrement, and a
	// requested enable_redelegation is ignored — the flag is always stored off
	otherAddr := sdk.AccAddress(PKS[1].Address())
	s.bankKeeper.EXPECT().DelegateCoinsFromAccountToModule(gomock.Any(), otherAddr, stakingtypes.NotBondedPoolName, gomock.Any()).AnyTimes()
	otherPk := ed25519.GenPrivKey().PubKey()
	msg, err := stakingtypes.NewMsgCreateValidator(sdk.ValAddress(otherAddr).String(), pk.Address().String(), otherPk, sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(500000000000)), stakingtypes.Description{Moniker: "LicenseVal2"}, comm, math.OneInt())
	require.NoError(err)
	msg.Mode = stakingtypes.ValidatorMode_MODE_LICENSE
	msg.DelegationIncrement = increment
	msg.MaxLicense = math.NewInt(150)
	msg.EnableRedelegation = true
	_, err = msgServer.CreateValidator(ctx, msg)
	require.NoError(err)
	validator, err = keeper.GetValidator(ctx, sdk.ValAddress(otherAddr))
	require.NoError(err)
	require.Equal(increment, validator.MinDelegation)
	require.False(validator.EnableRedelegation)
}

func (s *KeeperTestSuite) TestMsgEditValidatorLicenseMode() {
	ctx, keeper, msgServer := s.ctx, s.stakingKeeper, s.msgServer
	require := s.Require()
	s.execExpectCalls()

	increment := math.NewInt(10000000000)
	pk := ed25519.GenPrivKey().PubKey()
	comm := stakingtypes.NewCommissionRates(math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0))
	require.NoError(keeper.SetNewValidatorApprovalState(ctx, stakingtypes.ValidatorApproval{ApproverAddress: pk.Address().String(), Enabled: false}))

	// license validator with 50 licenses in use, cap 150
	msg, err := stakingtypes.NewMsgCreateValidator(ValAddr.String(), pk.Address().String(), pk, sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(500000000000)), stakingtypes.Description{Moniker: "LicenseVal"}, comm, math.OneInt())
	require.NoError(err)
	msg.Mode = stakingtypes.ValidatorMode_MODE_LICENSE
	msg.DelegationIncrement = increment
	msg.MinDelegation = increment
	msg.MaxLicense = math.NewInt(150)
	_, err = msgServer.CreateValidator(ctx, msg)
	require.NoError(err)

	// plain normal-mode validator without a delegation increment
	normalAccAddr := sdk.AccAddress(PKS[1].Address())
	normalValAddr := sdk.ValAddress(normalAccAddr)
	s.bankKeeper.EXPECT().DelegateCoinsFromAccountToModule(gomock.Any(), normalAccAddr, stakingtypes.NotBondedPoolName, gomock.Any()).AnyTimes()
	normalPk := ed25519.GenPrivKey().PubKey()
	msg, err = stakingtypes.NewMsgCreateValidator(normalValAddr.String(), pk.Address().String(), normalPk, sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(10000)), stakingtypes.Description{Moniker: "NormalVal"}, comm, math.OneInt())
	require.NoError(err)
	_, err = msgServer.CreateValidator(ctx, msg)
	require.NoError(err)

	desc := stakingtypes.Description{Moniker: "LicenseVal"}
	lowerMax := math.NewInt(100)
	higherMax := math.NewInt(200)

	// the owner cannot shrink MaxLicense below the current one
	_, err = msgServer.EditValidator(ctx, &stakingtypes.MsgEditValidator{
		ValidatorAddress: ValAddr.String(),
		Description:      desc,
		Mode:             "license",
		MaxLicense:       &lowerMax,
	})
	require.Error(err)
	require.Contains(err.Error(), "max license must be greater than or equal to the existing one")

	// raising MaxLicense is allowed and license usage is recomputed
	_, err = msgServer.EditValidator(ctx, &stakingtypes.MsgEditValidator{
		ValidatorAddress: ValAddr.String(),
		Description:      desc,
		Mode:             "license",
		MaxLicense:       &higherMax,
	})
	require.NoError(err)
	validator, err := keeper.GetValidator(ctx, ValAddr)
	require.NoError(err)
	require.Equal(higherMax, validator.MaxLicense)
	require.Equal(math.NewInt(50), validator.LicenseCount)

	// an edit that omits mode (proto zero = MODE_NORMAL) must not downgrade the validator
	_, err = msgServer.EditValidator(ctx, &stakingtypes.MsgEditValidator{
		ValidatorAddress: ValAddr.String(),
		Description:      desc,
	})
	require.NoError(err)
	validator, err = keeper.GetValidator(ctx, ValAddr)
	require.NoError(err)
	require.Equal(stakingtypes.ValidatorMode_MODE_LICENSE, validator.Mode)
	require.Equal(higherMax, validator.MaxLicense)

	// fast mode can only be chosen at create-validator
	_, err = msgServer.EditValidator(ctx, &stakingtypes.MsgEditValidator{
		ValidatorAddress: ValAddr.String(),
		Description:      desc,
		Mode:             "fast",
	})
	require.Error(err)
	require.Contains(err.Error(), "cannot switch an existing validator into fast mode")

	// redelegation is force-disabled: an enable request is ignored and the
	// flag stays off no matter what the edit asks for
	_, err = msgServer.EditValidator(ctx, &stakingtypes.MsgEditValidator{
		ValidatorAddress:   ValAddr.String(),
		Description:        desc,
		EnableRedelegation: stakingtypes.RedelegationUpdate_REDELEGATION_UPDATE_ENABLE,
	})
	require.NoError(err)
	validator, err = keeper.GetValidator(ctx, ValAddr)
	require.NoError(err)
	require.False(validator.EnableRedelegation)

	_, err = msgServer.EditValidator(ctx, &stakingtypes.MsgEditValidator{
		ValidatorAddress: ValAddr.String(),
		Description:      desc,
	})
	require.NoError(err)
	validator, err = keeper.GetValidator(ctx, ValAddr)
	require.NoError(err)
	require.False(validator.EnableRedelegation)

	_, err = msgServer.EditValidator(ctx, &stakingtypes.MsgEditValidator{
		ValidatorAddress:   ValAddr.String(),
		Description:        desc,
		EnableRedelegation: stakingtypes.RedelegationUpdate_REDELEGATION_UPDATE_DISABLE,
	})
	require.NoError(err)
	validator, err = keeper.GetValidator(ctx, ValAddr)
	require.NoError(err)
	require.False(validator.EnableRedelegation)

	// a validator created without an increment cannot switch into license mode
	_, err = msgServer.EditValidator(ctx, &stakingtypes.MsgEditValidator{
		ValidatorAddress: normalValAddr.String(),
		Description:      stakingtypes.Description{Moniker: "NormalVal"},
		Mode:             "license",
		MaxLicense:       &higherMax,
	})
	require.Error(err)
	require.Contains(err.Error(), "Min Delegation and DelegationIncrement must be defined and the same")

	// an unrecognized mode string is rejected instead of being coerced
	_, err = msgServer.EditValidator(ctx, &stakingtypes.MsgEditValidator{
		ValidatorAddress: ValAddr.String(),
		Description:      desc,
		Mode:             "turbo",
	})
	require.Error(err)
	require.Contains(err.Error(), "Invalid validator mode input")

	// the legacy enum field (pre-upgrade txs) is still honored when the
	// string field is empty
	_, err = msgServer.EditValidator(ctx, &stakingtypes.MsgEditValidator{
		ValidatorAddress: ValAddr.String(),
		Description:      desc,
		LegacyMode:       stakingtypes.ValidatorMode_MODE_LICENSE,
	})
	require.NoError(err)
	validator, err = keeper.GetValidator(ctx, ValAddr)
	require.NoError(err)
	require.Equal(stakingtypes.ValidatorMode_MODE_LICENSE, validator.Mode)

	// an explicit "normal" is unambiguous now and converts the validator
	_, err = msgServer.EditValidator(ctx, &stakingtypes.MsgEditValidator{
		ValidatorAddress: ValAddr.String(),
		Description:      desc,
		Mode:             "normal",
	})
	require.NoError(err)
	validator, err = keeper.GetValidator(ctx, ValAddr)
	require.NoError(err)
	require.Equal(stakingtypes.ValidatorMode_MODE_NORMAL, validator.Mode)

	// old tx bytes (enum mode on field 6) still unmarshal under the new schema
	oldTx := &stakingtypes.MsgEditValidator{
		ValidatorAddress: ValAddr.String(),
		Description:      desc,
		LegacyMode:       stakingtypes.ValidatorMode_MODE_FAST,
	}
	bz, err := oldTx.Marshal()
	require.NoError(err)
	var decoded stakingtypes.MsgEditValidator
	require.NoError(decoded.Unmarshal(bz))
	require.Equal(stakingtypes.ValidatorMode_MODE_FAST, decoded.LegacyMode)
	require.Empty(decoded.Mode)
}

func (s *KeeperTestSuite) TestMsgEditValidatorDelegationIncrement() {
	ctx, keeper, msgServer := s.ctx, s.stakingKeeper, s.msgServer
	require := s.Require()
	s.execExpectCalls()

	increment := math.NewInt(10000000000)
	pk := ed25519.GenPrivKey().PubKey()
	comm := stakingtypes.NewCommissionRates(math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0))
	require.NoError(keeper.SetNewValidatorApprovalState(ctx, stakingtypes.ValidatorApproval{ApproverAddress: pk.Address().String(), Enabled: false}))

	// license validator: self-bond 50 increments = 50 licenses, cap 150
	msg, err := stakingtypes.NewMsgCreateValidator(ValAddr.String(), pk.Address().String(), pk, sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(500000000000)), stakingtypes.Description{Moniker: "LicenseVal"}, comm, math.OneInt())
	require.NoError(err)
	msg.Mode = stakingtypes.ValidatorMode_MODE_LICENSE
	msg.DelegationIncrement = increment
	msg.MinDelegation = increment
	msg.MaxLicense = math.NewInt(150)
	_, err = msgServer.CreateValidator(ctx, msg)
	require.NoError(err)

	// second delegation of 3 increments -> 53 licenses in use
	delegator := sdk.AccAddress(PKS[1].Address())
	s.bankKeeper.EXPECT().DelegateCoinsFromAccountToModule(gomock.Any(), delegator, stakingtypes.NotBondedPoolName, gomock.Any()).AnyTimes()
	_, err = msgServer.Delegate(ctx, stakingtypes.NewMsgDelegate(delegator.String(), ValAddr.String(), sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(30000000000))))
	require.NoError(err)
	validator, err := keeper.GetValidator(ctx, ValAddr)
	require.NoError(err)
	require.Equal(math.NewInt(53), validator.LicenseCount)

	desc := stakingtypes.Description{Moniker: "LicenseVal"}
	editWithIncrement := func(newIncrement math.Int) (*stakingtypes.MsgEditValidatorResponse, error) {
		return msgServer.EditValidator(ctx, &stakingtypes.MsgEditValidator{
			ValidatorAddress:    ValAddr.String(),
			Description:         desc,
			Mode:                "license",
			DelegationIncrement: &newIncrement,
		})
	}

	// non-positive increment is rejected up front
	_, err = editWithIncrement(math.NewInt(0))
	require.Error(err)
	require.Contains(err.Error(), "delegation increment must be a positive integer")

	// 15G does not divide the 500G self-bond
	_, err = editWithIncrement(math.NewInt(15000000000))
	require.Error(err)
	require.Contains(err.Error(), "delegation amount must meet increment condition")

	// 53G divides the 530G total but not each delegation: the check is
	// per-delegation, so the redistribution cannot silently split a stake
	_, err = editWithIncrement(math.NewInt(53000000000))
	require.Error(err)
	require.Contains(err.Error(), "delegation amount must meet increment condition")

	// 2G divides everything but needs 265 licenses > cap 150
	_, err = editWithIncrement(math.NewInt(2000000000))
	require.Error(err)
	require.Contains(err.Error(), "There is no license enough for the delegation")

	// rejected edits leave the stored increment and count untouched
	validator, err = keeper.GetValidator(ctx, ValAddr)
	require.NoError(err)
	require.Equal(increment, validator.DelegationIncrement)
	require.Equal(math.NewInt(53), validator.LicenseCount)

	// 5G divides both delegations: 100 + 6 = 106 licenses under the cap
	newIncrement := math.NewInt(5000000000)
	_, err = editWithIncrement(newIncrement)
	require.NoError(err)
	validator, err = keeper.GetValidator(ctx, ValAddr)
	require.NoError(err)
	require.Equal(newIncrement, validator.DelegationIncrement)
	require.Equal(newIncrement, validator.MinDelegation)
	require.Equal(math.NewInt(106), validator.LicenseCount)

	// follow-up delegations must respect the new, finer increment
	_, err = msgServer.Delegate(ctx, stakingtypes.NewMsgDelegate(delegator.String(), ValAddr.String(), sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(5000000000))))
	require.NoError(err)
	validator, err = keeper.GetValidator(ctx, ValAddr)
	require.NoError(err)
	require.Equal(math.NewInt(107), validator.LicenseCount)
}

func (s *KeeperTestSuite) TestMsgDelegateFastMode() {
	ctx, keeper, msgServer := s.ctx, s.stakingKeeper, s.msgServer
	require := s.Require()
	s.execExpectCalls()

	pk := ed25519.GenPrivKey().PubKey()
	comm := stakingtypes.NewCommissionRates(math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0))
	require.NoError(keeper.SetNewValidatorApprovalState(ctx, stakingtypes.ValidatorApproval{ApproverAddress: pk.Address().String(), Enabled: false}))

	msg, err := stakingtypes.NewMsgCreateValidator(ValAddr.String(), pk.Address().String(), pk, sdk.NewCoin(sdk.DefaultBondDenom, keeper.TokensFromConsensusPower(ctx, 100)), stakingtypes.Description{Moniker: "FastVal"}, comm, math.OneInt())
	require.NoError(err)
	msg.Mode = stakingtypes.ValidatorMode_MODE_FAST
	_, err = msgServer.CreateValidator(ctx, msg)
	require.NoError(err)

	delegator := sdk.AccAddress(PKS[1].Address())
	amount := sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1000000))

	// a delegator outside the whitelist cannot enter a fast validator
	_, err = msgServer.Delegate(ctx, &stakingtypes.MsgDelegate{
		DelegatorAddress: delegator.String(),
		ValidatorAddress: ValAddr.String(),
		Amount:           amount,
	})
	require.Error(err)
	require.Contains(err.Error(), "Cannot find delegator in whitelist")

	// the operator itself is always allowed
	_, err = msgServer.Delegate(ctx, &stakingtypes.MsgDelegate{
		DelegatorAddress: Addr.String(),
		ValidatorAddress: ValAddr.String(),
		Amount:           amount,
	})
	require.NoError(err)

	// whitelisting through the msg server opens the gate
	_, err = msgServer.CreateWhitelistdelegator(ctx, &stakingtypes.MsgCreateWhitelistDelegator{
		Creator:          Addr.String(),
		ValidatorAddress: ValAddr.String(),
		DelegatorAddress: delegator.String(),
	})
	require.NoError(err)

	s.bankKeeper.EXPECT().DelegateCoinsFromAccountToModule(gomock.Any(), delegator, stakingtypes.NotBondedPoolName, gomock.Any()).AnyTimes()
	_, err = msgServer.Delegate(ctx, &stakingtypes.MsgDelegate{
		DelegatorAddress: delegator.String(),
		ValidatorAddress: ValAddr.String(),
		Amount:           amount,
	})
	require.NoError(err)
}

func (s *KeeperTestSuite) TestMsgUndelegateLicenseMode() {
	ctx, keeper, msgServer := s.ctx, s.stakingKeeper, s.msgServer
	require := s.Require()
	s.execExpectCalls()

	increment := math.NewInt(10000000000)
	pk := ed25519.GenPrivKey().PubKey()
	comm := stakingtypes.NewCommissionRates(math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0))
	require.NoError(keeper.SetNewValidatorApprovalState(ctx, stakingtypes.ValidatorApproval{ApproverAddress: pk.Address().String(), Enabled: false}))

	msg, err := stakingtypes.NewMsgCreateValidator(ValAddr.String(), pk.Address().String(), pk, sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(500000000000)), stakingtypes.Description{Moniker: "LicenseVal"}, comm, math.OneInt())
	require.NoError(err)
	msg.Mode = stakingtypes.ValidatorMode_MODE_LICENSE
	msg.DelegationIncrement = increment
	msg.MinDelegation = increment
	msg.MaxLicense = math.NewInt(150)
	_, err = msgServer.CreateValidator(ctx, msg)
	require.NoError(err)

	// withdrawals must respect the increment so remaining stake stays a
	// whole number of licenses
	_, err = msgServer.Undelegate(ctx, &stakingtypes.MsgUndelegate{
		DelegatorAddress: Addr.String(),
		ValidatorAddress: ValAddr.String(),
		Amount:           sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(15000000000)),
	})
	require.Error(err)
	require.Contains(err.Error(), "delegation amount must meet increment condition")

	// undelegating two increments releases two licenses
	res, err := msgServer.Undelegate(ctx, &stakingtypes.MsgUndelegate{
		DelegatorAddress: Addr.String(),
		ValidatorAddress: ValAddr.String(),
		Amount:           sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(20000000000)),
	})
	require.NoError(err)
	require.Equal(math.NewInt(20000000000), res.Amount.Amount)

	validator, err := keeper.GetValidator(ctx, ValAddr)
	require.NoError(err)
	require.Equal(math.NewInt(48), validator.LicenseCount)
}

func (s *KeeperTestSuite) TestMsgUndelegateFastMode() {
	ctx, keeper, msgServer := s.ctx, s.stakingKeeper, s.msgServer
	require := s.Require()
	s.execExpectCalls()

	pk := ed25519.GenPrivKey().PubKey()
	comm := stakingtypes.NewCommissionRates(math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0))
	require.NoError(keeper.SetNewValidatorApprovalState(ctx, stakingtypes.ValidatorApproval{ApproverAddress: pk.Address().String(), Enabled: false}))

	msg, err := stakingtypes.NewMsgCreateValidator(ValAddr.String(), pk.Address().String(), pk, sdk.NewCoin(sdk.DefaultBondDenom, keeper.TokensFromConsensusPower(ctx, 100)), stakingtypes.Description{Moniker: "FastVal"}, comm, math.OneInt())
	require.NoError(err)
	msg.Mode = stakingtypes.ValidatorMode_MODE_FAST
	_, err = msgServer.CreateValidator(ctx, msg)
	require.NoError(err)

	unbondingTime, err := keeper.UnbondingTime(ctx)
	require.NoError(err)

	// fast-mode undelegation completes next block instead of waiting the
	// full unbonding period
	res, err := msgServer.Undelegate(ctx, &stakingtypes.MsgUndelegate{
		DelegatorAddress: Addr.String(),
		ValidatorAddress: ValAddr.String(),
		Amount:           sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1000000)),
	})
	require.NoError(err)
	require.True(res.CompletionTime.Before(ctx.BlockTime().Add(unbondingTime)))
	require.True(res.CompletionTime.Equal(ctx.BlockTime()))

	// the whitelist gates entry only; a non-whitelisted delegator can still exit
	delegator := sdk.AccAddress(PKS[1].Address())
	del := stakingtypes.NewDelegation(delegator.String(), ValAddr.String(), math.LegacyNewDec(100))
	require.NoError(keeper.SetDelegation(ctx, del))
	res, err = msgServer.Undelegate(ctx, &stakingtypes.MsgUndelegate{
		DelegatorAddress: delegator.String(),
		ValidatorAddress: ValAddr.String(),
		Amount:           sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(100)),
	})
	require.NoError(err)
	require.True(res.CompletionTime.Equal(ctx.BlockTime()))
}

func (s *KeeperTestSuite) TestMsgCancelUnbondingDelegationLicenseMode() {
	keeper, msgServer := s.stakingKeeper, s.msgServer
	require := s.Require()
	s.execExpectCalls()

	// cancel-unbonding requires a positive creation height
	ctx := s.ctx.WithBlockHeight(5)

	increment := math.NewInt(10000000000)
	pk := ed25519.GenPrivKey().PubKey()
	comm := stakingtypes.NewCommissionRates(math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0))
	require.NoError(keeper.SetNewValidatorApprovalState(ctx, stakingtypes.ValidatorApproval{ApproverAddress: pk.Address().String(), Enabled: false}))

	// the cap is exactly the self-bond, so every license is in use
	msg, err := stakingtypes.NewMsgCreateValidator(ValAddr.String(), pk.Address().String(), pk, sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(500000000000)), stakingtypes.Description{Moniker: "LicenseVal"}, comm, math.OneInt())
	require.NoError(err)
	msg.Mode = stakingtypes.ValidatorMode_MODE_LICENSE
	msg.DelegationIncrement = increment
	msg.MinDelegation = increment
	msg.MaxLicense = math.NewInt(50)
	_, err = msgServer.CreateValidator(ctx, msg)
	require.NoError(err)

	// release two licenses into an unbonding entry
	_, err = msgServer.Undelegate(ctx, &stakingtypes.MsgUndelegate{
		DelegatorAddress: Addr.String(),
		ValidatorAddress: ValAddr.String(),
		Amount:           sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(20000000000)),
	})
	require.NoError(err)
	validator, err := keeper.GetValidator(ctx, ValAddr)
	require.NoError(err)
	require.Equal(math.NewInt(48), validator.LicenseCount)

	// cancelling must also respect the increment
	_, err = msgServer.CancelUnbondingDelegation(ctx, &stakingtypes.MsgCancelUnbondingDelegation{
		DelegatorAddress: Addr.String(),
		ValidatorAddress: ValAddr.String(),
		Amount:           sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(5000000000)),
		CreationHeight:   5,
	})
	require.Error(err)
	require.Contains(err.Error(), "delegation amount must meet increment condition")

	// cancelling one increment claims its license back
	_, err = msgServer.CancelUnbondingDelegation(ctx, &stakingtypes.MsgCancelUnbondingDelegation{
		DelegatorAddress: Addr.String(),
		ValidatorAddress: ValAddr.String(),
		Amount:           sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(10000000000)),
		CreationHeight:   5,
	})
	require.NoError(err)
	validator, err = keeper.GetValidator(ctx, ValAddr)
	require.NoError(err)
	require.Equal(math.NewInt(49), validator.LicenseCount)

	// refill the freed license with a fresh delegation...
	_, err = msgServer.Delegate(ctx, &stakingtypes.MsgDelegate{
		DelegatorAddress: Addr.String(),
		ValidatorAddress: ValAddr.String(),
		Amount:           sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(10000000000)),
	})
	require.NoError(err)
	validator, err = keeper.GetValidator(ctx, ValAddr)
	require.NoError(err)
	require.Equal(math.NewInt(50), validator.LicenseCount)

	// ...so cancelling the remaining entry would exceed MaxLicense and must fail
	// BEFORE any tokens are re-bonded
	_, err = msgServer.CancelUnbondingDelegation(ctx, &stakingtypes.MsgCancelUnbondingDelegation{
		DelegatorAddress: Addr.String(),
		ValidatorAddress: ValAddr.String(),
		Amount:           sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(10000000000)),
		CreationHeight:   5,
	})
	require.Error(err)
	require.Contains(err.Error(), "There is no license enough for the delegation")

	validator, err = keeper.GetValidator(ctx, ValAddr)
	require.NoError(err)
	require.Equal(math.NewInt(50), validator.LicenseCount)
}
