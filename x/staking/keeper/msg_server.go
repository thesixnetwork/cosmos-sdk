package keeper

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/go-metrics"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

type msgServer struct {
	*Keeper
}

// NewMsgServerImpl returns an implementation of the staking MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper *Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (k msgServer) SetValidatorApproval(ctx context.Context, msg *types.MsgSetValidatorApproval) (*types.MsgSetValidatorApprovalResponse, error) {
	validatorApproval, err := k.GetValidatorApproval(ctx)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrNotFound, "Validator approval is somehow does not existed")
	}

	if validatorApproval.ApproverAddress != msg.ApproverAddress {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "Msg sender is not current approver")
	}

	var newApproverAddress string
	if _, err := sdk.AccAddressFromBech32(msg.NewApproverAddress); err == nil {
		newApproverAddress = msg.NewApproverAddress
	} else {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid new approver address")
	}

	newValidatorApprovalState := types.ValidatorApproval{
		ApproverAddress: newApproverAddress,
		Enabled:         msg.Enabled,
	}

	k.SetNewValidatorApprovalState(ctx, newValidatorApprovalState)

	return &types.MsgSetValidatorApprovalResponse{}, nil
}

// CreateValidator defines a method for creating a new validator
func (k msgServer) CreateValidator(ctx context.Context, msg *types.MsgCreateValidator) (*types.MsgCreateValidatorResponse, error) {
	approval, err := k.GetValidatorApproval(ctx)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrNotFound, "Validator approval is somehow does not existed")
	}

	if approval.Enabled && msg.ApproverAddress != approval.ApproverAddress {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "Wrong approver for create validator")
	}

	valAddr, err := k.validatorAddressCodec.StringToBytes(msg.ValidatorAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid validator address: %s", err)
	}

	if err := msg.Validate(k.validatorAddressCodec); err != nil {
		return nil, err
	}

	minCommRate, err := k.MinCommissionRate(ctx)
	if err != nil {
		return nil, err
	}

	if msg.Commission.Rate.LT(minCommRate) {
		return nil, errorsmod.Wrapf(types.ErrCommissionLTMinRate, "cannot set validator commission to less than minimum rate of %s", minCommRate)
	}

	// check to see if the pubkey or sender has been registered before
	if _, err := k.GetValidator(ctx, valAddr); err == nil {
		return nil, types.ErrValidatorOwnerExists
	}

	pk, ok := msg.Pubkey.GetCachedValue().(cryptotypes.PubKey)
	if !ok {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidType, "Expecting cryptotypes.PubKey, got %T", pk)
	}

	if _, err := k.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(pk)); err == nil {
		return nil, types.ErrValidatorPubKeyExists
	}

	bondDenom, err := k.BondDenom(ctx)
	if err != nil {
		return nil, err
	}

	if msg.Value.Denom != bondDenom {
		return nil, errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest, "invalid coin denomination: got %s, expected %s", msg.Value.Denom, bondDenom,
		)
	}

	if _, err := msg.Description.EnsureLength(); err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	cp := sdkCtx.ConsensusParams()
	if cp.Validator != nil {
		pkType := pk.Type()
		hasKeyType := false
		for _, keyType := range cp.Validator.PubKeyTypes {
			if pkType == keyType {
				hasKeyType = true
				break
			}
		}
		if !hasKeyType {
			return nil, errorsmod.Wrapf(
				types.ErrValidatorPubKeyTypeNotSupported,
				"got: %s, expected: %s", pk.Type(), cp.Validator.PubKeyTypes,
			)
		}
	}

	validator, err := types.NewValidator(msg.ValidatorAddress, pk, msg.Description)
	if err != nil {
		return nil, err
	}

	commission := types.NewCommissionWithTime(
		msg.Commission.Rate, msg.Commission.MaxRate,
		msg.Commission.MaxChangeRate, sdkCtx.BlockHeader().Time,
	)

	validator, err = validator.SetInitialCommission(commission)
	if err != nil {
		return nil, err
	}

	validator.MinSelfDelegation = msg.MinSelfDelegation

	validator.MinSelfDelegation = msg.MinSelfDelegation
	// CustomValidator
	validator.MinDelegation = msg.MinDelegation
	validator.DelegationIncrement = msg.DelegationIncrement
	// when Min Delegation is not defined, default as DelegationIncrement
	if msg.MinDelegation.IsNil() {
		validator.MinDelegation = validator.DelegationIncrement
	}

	switch msg.Mode {
	case types.ValidatorMode_MODE_LICENSE:
		validator.Mode = types.ValidatorMode_MODE_LICENSE
		// Verify that MinDelegation and DelegationIncrement is  defined and contains the same value
		if msg.DelegationIncrement.IsNil() || !validator.MinDelegation.Equal(validator.DelegationIncrement) {
			return nil, types.ErrLicenseIncrement
		}

		if validator.MaxLicense = msg.MaxLicense; msg.MaxLicense.IsNil() {
			return nil, types.ErrMaxLicenseMustBeDefined
		} // bug is nill genesis

		// Count licesene amount for validator
		divAmount := msg.Value.Amount.Quo(validator.DelegationIncrement)
		modAmount := msg.Value.Amount.Mod(validator.DelegationIncrement)
		if modAmount.GT(math.ZeroInt()) {
			return nil, types.ErrInvalidIncrementDelegation
		}
		if divAmount.GT(validator.MaxLicense) {
			return nil, types.ErrNotEnoughLicense
		}
		validator.LicenseCount = divAmount
		// Force disable redelegation when
		validator.EnableRedelegation = false
	case types.ValidatorMode_MODE_FAST:
		validator.Mode = types.ValidatorMode_MODE_FAST
		validator.EnableRedelegation = msg.EnableRedelegation
	default:
		validator.Mode = types.ValidatorMode_MODE_NORMAL
		validator.EnableRedelegation = msg.EnableRedelegation
	}
	err = k.SetValidator(ctx, validator)
	if err != nil {
		return nil, err
	}

	err = k.SetValidatorByConsAddr(ctx, validator)
	if err != nil {
		return nil, err
	}

	err = k.SetNewValidatorByPowerIndex(ctx, validator)
	if err != nil {
		return nil, err
	}

	// call the after-creation hook
	if err := k.Hooks().AfterValidatorCreated(ctx, valAddr); err != nil {
		return nil, err
	}

	// move coins from the msg.Address account to a (self-delegation) delegator account
	// the validator account and global shares are updated within here
	// NOTE source will always be from a wallet which are unbonded
	_, err = k.Keeper.Delegate(ctx, sdk.AccAddress(valAddr), msg.Value.Amount, types.Unbonded, validator, true)
	if err != nil {
		return nil, err
	}

	sdkCtx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeCreateValidator,
			sdk.NewAttribute(types.AttributeKeyValidator, msg.ValidatorAddress),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Value.String()),
		),
	})

	return &types.MsgCreateValidatorResponse{}, nil
}

// EditValidator defines a method for editing an existing validator
func (k msgServer) EditValidator(ctx context.Context, msg *types.MsgEditValidator) (*types.MsgEditValidatorResponse, error) {
	valAddr, err := k.validatorAddressCodec.StringToBytes(msg.ValidatorAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid validator address: %s", err)
	}

	if msg.Description == (types.Description{}) {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "empty description")
	}

	if msg.MinSelfDelegation != nil && !msg.MinSelfDelegation.IsPositive() {
		return nil, errorsmod.Wrap(
			sdkerrors.ErrInvalidRequest,
			"minimum self delegation must be a positive integer",
		)
	}

	if msg.CommissionRate != nil {
		if msg.CommissionRate.GT(math.LegacyOneDec()) || msg.CommissionRate.IsNegative() {
			return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "commission rate must be between 0 and 1 (inclusive)")
		}

		minCommissionRate, err := k.MinCommissionRate(ctx)
		if err != nil {
			return nil, errorsmod.Wrap(sdkerrors.ErrLogic, err.Error())
		}

		if msg.CommissionRate.LT(minCommissionRate) {
			return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "commission rate cannot be less than the min commission rate %s", minCommissionRate.String())
		}
	}

	// validator must already be registered
	validator, err := k.GetValidator(ctx, valAddr)
	if err != nil {
		return nil, err
	}

	// replace all editable fields (clients should autofill existing values)
	description, err := validator.Description.UpdateDescription(msg.Description)
	if err != nil {
		return nil, err
	}

	validator.Description = description

	switch msg.Mode {
	case types.ValidatorMode_MODE_LICENSE:
		validator.Mode = types.ValidatorMode_MODE_LICENSE
		// validate max license
		if !msg.MaxLicense.IsNil() && msg.MaxLicense.LT(validator.MaxLicense) {
			return nil, types.ErrMaxLicenseMustBeGeater
		}

		if !msg.MaxLicense.IsNil() {
			validator.MaxLicense = *msg.MaxLicense
		}

		amount := validator.GetDelegatorShares().Ceil().TruncateInt()
		divAmount := amount.Quo(validator.DelegationIncrement)
		modAmount := amount.Mod(validator.DelegationIncrement)
		if modAmount.GT(math.ZeroInt()) {
			return nil, types.ErrInvalidIncrementDelegation
		}
		if divAmount.GT(validator.MaxLicense) {
			return nil, types.ErrNotEnoughLicense
		}

		validator.LicenseCount = divAmount
		validator.EnableRedelegation = false
	case types.ValidatorMode_MODE_FAST:
		validator.Mode = types.ValidatorMode_MODE_FAST
	default:
		validator.Mode = types.ValidatorMode_MODE_NORMAL
	}

	if msg.CommissionRate != nil {
		commission, err := k.UpdateValidatorCommission(ctx, validator, *msg.CommissionRate)
		if err != nil {
			return nil, err
		}

		// call the before-modification hook since we're about to update the commission
		if err := k.Hooks().BeforeValidatorModified(ctx, valAddr); err != nil {
			return nil, err
		}

		validator.Commission = commission
	}

	if msg.MinSelfDelegation != nil {
		if !msg.MinSelfDelegation.GT(validator.MinSelfDelegation) {
			return nil, types.ErrMinSelfDelegationDecreased
		}

		if msg.MinSelfDelegation.GT(validator.Tokens) {
			return nil, types.ErrSelfDelegationBelowMinimum
		}

		validator.MinSelfDelegation = *msg.MinSelfDelegation
	}

	err = k.SetValidator(ctx, validator)
	if err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeEditValidator,
			sdk.NewAttribute(types.AttributeKeyCommissionRate, validator.Commission.String()),
			sdk.NewAttribute(types.AttributeKeyMinSelfDelegation, validator.MinSelfDelegation.String()),
		),
	})

	return &types.MsgEditValidatorResponse{}, nil
}

// Delegate defines a method for performing a delegation of coins from a delegator to a validator
func (k msgServer) Delegate(ctx context.Context, msg *types.MsgDelegate) (*types.MsgDelegateResponse, error) {
	valAddr, valErr := k.validatorAddressCodec.StringToBytes(msg.ValidatorAddress)
	if valErr != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid validator address: %s", valErr)
	}

	delegatorAddress, err := k.authKeeper.AddressCodec().StringToBytes(msg.DelegatorAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid delegator address: %s", err)
	}

	if !msg.Amount.IsValid() || !msg.Amount.Amount.IsPositive() {
		return nil, errorsmod.Wrap(
			sdkerrors.ErrInvalidRequest,
			"invalid delegation amount",
		)
	}

	validator, err := k.GetValidator(ctx, valAddr)
	if err != nil {
		return nil, err
	}

	bondDenom, err := k.BondDenom(ctx)
	if err != nil {
		return nil, err
	}

	if msg.Amount.Denom != bondDenom {
		return nil, errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest, "invalid coin denomination: got %s, expected %s", msg.Amount.Denom, bondDenom,
		)
	}

	// CustomValidator
	// New Delegation or Update

	switch validator.Mode {
	case types.ValidatorMode_MODE_LICENSE:
		delegateLicenseCount, err := k.calculateDelegateLicenseCount(ctx, msg.Amount, validator, sdk.AccAddress(msg.DelegatorAddress), math.LegacyDec{})
		if err != nil {
			return nil, err
		}
		// Validate current license count with MaxLicense
		if validator.LicenseCount.GTE(validator.MaxLicense) {
			return nil, types.ErrLicenseLimit
		}
		// Validate delegatio with license count and max license
		if delegateLicenseCount.Add(validator.LicenseCount).GT(validator.MaxLicense) {
			return nil, types.ErrNotEnoughLicense
		}
		// increase license count in validator
		validator.LicenseCount = delegateLicenseCount.Add(validator.LicenseCount)
		// Update Validator
		k.Keeper.SetValidator(ctx, validator)
	case types.ValidatorMode_MODE_FAST:
		isSpecial := k.IsSpecialDelegator(ctx, valAddr, delegatorAddress)
		if !isSpecial {
			return nil, types.ErrDelegatorIsNotSpecial
		}
	}

	// NOTE: source funds are always unbonded
	newShares, err := k.Keeper.Delegate(ctx, delegatorAddress, msg.Amount.Amount, types.Unbonded, validator, true)
	if err != nil {
		return nil, err
	}

	if msg.Amount.Amount.IsInt64() {
		defer func() {
			telemetry.IncrCounter(1, types.ModuleName, "delegate")
			telemetry.SetGaugeWithLabels(
				[]string{"tx", "msg", sdk.MsgTypeURL(msg)},
				float32(msg.Amount.Amount.Int64()),
				[]metrics.Label{telemetry.NewLabel("denom", msg.Amount.Denom)},
			)
		}()
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeDelegate,
			sdk.NewAttribute(types.AttributeKeyValidator, msg.ValidatorAddress),
			sdk.NewAttribute(types.AttributeKeyDelegator, msg.DelegatorAddress),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyNewShares, newShares.String()),
		),
	})

	return &types.MsgDelegateResponse{}, nil
}

// BeginRedelegate defines a method for performing a redelegation of coins from a source validator to a destination validator of given delegator
func (k msgServer) BeginRedelegate(ctx context.Context, msg *types.MsgBeginRedelegate) (*types.MsgBeginRedelegateResponse, error) {
	valSrcAddr, err := k.validatorAddressCodec.StringToBytes(msg.ValidatorSrcAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid source validator address: %s", err)
	}

	valDstAddr, err := k.validatorAddressCodec.StringToBytes(msg.ValidatorDstAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid destination validator address: %s", err)
	}

	delegatorAddress, err := k.authKeeper.AddressCodec().StringToBytes(msg.DelegatorAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid delegator address: %s", err)
	}

	if !msg.Amount.IsValid() || !msg.Amount.Amount.IsPositive() {
		return nil, errorsmod.Wrap(
			sdkerrors.ErrInvalidRequest,
			"invalid shares amount",
		)
	}

	shares, err := k.ValidateUnbondAmount(
		ctx, delegatorAddress, valSrcAddr, msg.Amount.Amount,
	)
	if err != nil {
		return nil, err
	}

	bondDenom, err := k.BondDenom(ctx)
	if err != nil {
		return nil, err
	}

	if msg.Amount.Denom != bondDenom {
		return nil, errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest, "invalid coin denomination: got %s, expected %s", msg.Amount.Denom, bondDenom,
		)
	}

	// Validate source validator
	// Get Validator
	sourceVal, err := k.GetValidator(ctx, valSrcAddr) // TODO: to check
	if err != nil {
		return nil, err
	}
	destVal, err := k.GetValidator(ctx, valDstAddr) // TODO: to check
	if err != nil {
		return nil, err
	}
	if !sourceVal.EnableRedelegation || !destVal.EnableRedelegation {
		return nil, types.ErrRedelegationDisable
	}

	// Get Current Delegation
	currentSourceDelegation, err := k.Keeper.GetDelegation(ctx, delegatorAddress, sdk.ValAddress(sourceVal.GetOperator()))
	if err != nil {
		return nil, err
	}
	// Validate minimum amount , currentDelegation - unbond >= min delegation
	if !currentSourceDelegation.Shares.Equal(shares) {
		// NOT remove entire shares , only unbond some of it.
		if !sourceVal.MinDelegation.IsNil() && currentSourceDelegation.Shares.Sub(shares).LT(sourceVal.MinDelegation.ToLegacyDec()) {
			return nil, types.ErrDelegationBelowMinimum
		}
	}
	// Deduct minimum from value to validate increment
	amountToValidateIncrement := math.NewIntFromBigInt(msg.Amount.Amount.BigInt())
	if !sourceVal.MinDelegation.IsNil() {
		amountToValidateIncrement = amountToValidateIncrement.Sub(sourceVal.MinDelegation)
	}

	increment := math.OneInt()
	if !sourceVal.DelegationIncrement.IsNil() {
		increment = sourceVal.DelegationIncrement
	}
	// Validate DelegationIncrement
	if amountToValidateIncrement.GT(math.ZeroInt()) {
		// not remove
		modAmount := amountToValidateIncrement.Mod(increment)
		if modAmount.GT(math.ZeroInt()) {
			return nil, types.ErrInvalidIncrementDelegation
		}
	}

	// Validate destination
	// New Delegation or Update
	_, err = k.Keeper.GetDelegation(ctx, delegatorAddress, sdk.ValAddress(destVal.GetOperator()))
	if err != nil {
		return nil, err
	}
	// Validate Minimum and Increment
	if !destVal.MinDelegation.IsNil() && msg.Amount.Amount.LT(destVal.MinDelegation) {
		return nil, types.ErrDelegationBelowMinimum
	}
	// Deduct minimum from value to validate increment
	amountToValidateIncrement = math.NewIntFromBigInt(msg.Amount.Amount.BigInt())
	if !destVal.MinDelegation.IsNil() {
		amountToValidateIncrement = amountToValidateIncrement.Sub(destVal.MinDelegation)
	}
	// Validate DelegationIncrement
	increment = math.OneInt()
	if !destVal.DelegationIncrement.IsNil() {
		increment = destVal.DelegationIncrement
	}
	if amountToValidateIncrement.GT(math.ZeroInt()) {
		modAmount := amountToValidateIncrement.Mod(increment)
		if modAmount.GT(math.ZeroInt()) {
			return nil, types.ErrInvalidIncrementDelegation
		}
	}

	completionTime, err := k.BeginRedelegation(
		ctx, delegatorAddress, valSrcAddr, valDstAddr, shares,
	)
	if err != nil {
		return nil, err
	}

	if msg.Amount.Amount.IsInt64() {
		defer func() {
			telemetry.IncrCounter(1, types.ModuleName, "redelegate")
			telemetry.SetGaugeWithLabels(
				[]string{"tx", "msg", sdk.MsgTypeURL(msg)},
				float32(msg.Amount.Amount.Int64()),
				[]metrics.Label{telemetry.NewLabel("denom", msg.Amount.Denom)},
			)
		}()
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeRedelegate,
			sdk.NewAttribute(types.AttributeKeySrcValidator, msg.ValidatorSrcAddress),
			sdk.NewAttribute(types.AttributeKeyDstValidator, msg.ValidatorDstAddress),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyCompletionTime, completionTime.Format(time.RFC3339)),
		),
	})

	return &types.MsgBeginRedelegateResponse{
		CompletionTime: completionTime,
	}, nil
}

// Undelegate defines a method for performing an undelegation from a delegate and a validator
func (k msgServer) Undelegate(ctx context.Context, msg *types.MsgUndelegate) (*types.MsgUndelegateResponse, error) {
	addr, err := k.validatorAddressCodec.StringToBytes(msg.ValidatorAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid validator address: %s", err)
	}

	delegatorAddress, err := k.authKeeper.AddressCodec().StringToBytes(msg.DelegatorAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid delegator address: %s", err)
	}

	if !msg.Amount.IsValid() || !msg.Amount.Amount.IsPositive() {
		return nil, errorsmod.Wrap(
			sdkerrors.ErrInvalidRequest,
			"invalid shares amount",
		)
	}
	shares, err := k.ValidateUnbondAmount(
		ctx, delegatorAddress, addr, msg.Amount.Amount,
	)
	if err != nil {
		return nil, err
	}

	bondDenom, err := k.BondDenom(ctx)
	if err != nil {
		return nil, err
	}

	if msg.Amount.Denom != bondDenom {
		return nil, errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest, "invalid coin denomination: got %s, expected %s", msg.Amount.Denom, bondDenom,
		)
	}

	var completionTime time.Time
	var undelegatedCoin sdk.Coin
	var undelegatedAmt math.Int
	/// Custom Validator
	// Get Current Validator
	validator, err := k.GetValidator(ctx, addr)
	if err != nil {
		return nil, err
	}

	switch validator.Mode {
	case types.ValidatorMode_MODE_LICENSE:
		delegateLicenseCount, err := k.calculateDelegateLicenseCount(ctx, msg.Amount, validator, delegatorAddress, shares)
		if err != nil {
			return nil, err
		}
		// Validate License
		// decrease license count in validator
		validator.LicenseCount = validator.LicenseCount.Sub(delegateLicenseCount)
		// Update Validator
		k.Keeper.SetValidator(ctx, validator)

		completionTime, undelegatedAmt, err = k.Keeper.Undelegate(ctx, delegatorAddress, addr, shares)
		if err != nil {
			return nil, err
		}
		undelegatedCoin = sdk.NewCoin(msg.Amount.Denom, undelegatedAmt)
	case types.ValidatorMode_MODE_FAST:
		completionTime, _, err = k.Keeper.UndelegateSpecial(ctx, delegatorAddress, addr, shares)
		if err != nil {
			return nil, err
		}
		undelegatedCoin = sdk.NewCoin(msg.Amount.Denom, undelegatedAmt)
	default:
		completionTime, _, err = k.Keeper.Undelegate(ctx, delegatorAddress, addr, shares)
		if err != nil {
			return nil, err
		}
		undelegatedCoin = sdk.NewCoin(msg.Amount.Denom, undelegatedAmt)
	}

	if msg.Amount.Amount.IsInt64() {
		defer func() {
			telemetry.IncrCounter(1, types.ModuleName, "undelegate")
			telemetry.SetGaugeWithLabels(
				[]string{"tx", "msg", sdk.MsgTypeURL(msg)},
				float32(msg.Amount.Amount.Int64()),
				[]metrics.Label{telemetry.NewLabel("denom", msg.Amount.Denom)},
			)
		}()
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeUnbond,
			sdk.NewAttribute(types.AttributeKeyValidator, msg.ValidatorAddress),
			sdk.NewAttribute(types.AttributeKeyDelegator, msg.DelegatorAddress),
			sdk.NewAttribute(sdk.AttributeKeyAmount, undelegatedCoin.String()),
			sdk.NewAttribute(types.AttributeKeyCompletionTime, completionTime.Format(time.RFC3339)),
		),
	})

	return &types.MsgUndelegateResponse{
		CompletionTime: completionTime,
		Amount:         undelegatedCoin,
	}, nil
}

// CancelUnbondingDelegation defines a method for canceling the unbonding delegation
// and delegate back to the validator.
func (k msgServer) CancelUnbondingDelegation(ctx context.Context, msg *types.MsgCancelUnbondingDelegation) (*types.MsgCancelUnbondingDelegationResponse, error) {
	valAddr, err := k.validatorAddressCodec.StringToBytes(msg.ValidatorAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid validator address: %s", err)
	}

	delegatorAddress, err := k.authKeeper.AddressCodec().StringToBytes(msg.DelegatorAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid delegator address: %s", err)
	}

	if !msg.Amount.IsValid() || !msg.Amount.Amount.IsPositive() {
		return nil, errorsmod.Wrap(
			sdkerrors.ErrInvalidRequest,
			"invalid amount",
		)
	}

	if msg.CreationHeight <= 0 {
		return nil, errorsmod.Wrap(
			sdkerrors.ErrInvalidRequest,
			"invalid height",
		)
	}

	bondDenom, err := k.BondDenom(ctx)
	if err != nil {
		return nil, err
	}

	if msg.Amount.Denom != bondDenom {
		return nil, errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest, "invalid coin denomination: got %s, expected %s", msg.Amount.Denom, bondDenom,
		)
	}

	validator, err := k.GetValidator(ctx, valAddr)
	if err != nil {
		return nil, err
	}

	// In some situations, the exchange rate becomes invalid, e.g. if
	// Validator loses all tokens due to slashing. In this case,
	// make all future delegations invalid.
	if validator.InvalidExRate() {
		return nil, types.ErrDelegatorShareExRateInvalid
	}

	if validator.IsJailed() {
		return nil, types.ErrValidatorJailed
	}

	ubd, err := k.GetUnbondingDelegation(ctx, delegatorAddress, valAddr)
	if err != nil {
		return nil, status.Errorf(
			codes.NotFound,
			"unbonding delegation with delegator %s not found for validator %s",
			msg.DelegatorAddress, msg.ValidatorAddress,
		)
	}

	var (
		unbondEntry      types.UnbondingDelegationEntry
		unbondEntryIndex int64 = -1
	)

	for i, entry := range ubd.Entries {
		if entry.CreationHeight == msg.CreationHeight {
			unbondEntry = entry
			unbondEntryIndex = int64(i)
			break
		}
	}
	if unbondEntryIndex == -1 {
		return nil, sdkerrors.ErrNotFound.Wrapf("unbonding delegation entry is not found at block height %d", msg.CreationHeight)
	}

	if unbondEntry.Balance.LT(msg.Amount.Amount) {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("amount is greater than the unbonding delegation entry balance")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if unbondEntry.CompletionTime.Before(sdkCtx.BlockTime()) {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("unbonding delegation is already processed")
	}

	// delegate back the unbonding delegation amount to the validator
	_, err = k.Keeper.Delegate(ctx, delegatorAddress, msg.Amount.Amount, types.Unbonding, validator, false)
	if err != nil {
		return nil, err
	}

	amount := unbondEntry.Balance.Sub(msg.Amount.Amount)
	if amount.IsZero() {
		ubd.RemoveEntry(unbondEntryIndex)
	} else {
		// update the unbondingDelegationEntryBalance and InitialBalance for ubd entry
		unbondEntry.Balance = amount
		unbondEntry.InitialBalance = unbondEntry.InitialBalance.Sub(msg.Amount.Amount)
		ubd.Entries[unbondEntryIndex] = unbondEntry
	}

	// set the unbonding delegation or remove it if there are no more entries
	if len(ubd.Entries) == 0 {
		err = k.RemoveUnbondingDelegation(ctx, ubd)
	} else {
		err = k.SetUnbondingDelegation(ctx, ubd)
	}

	if err != nil {
		return nil, err
	}

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCancelUnbondingDelegation,
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyValidator, msg.ValidatorAddress),
			sdk.NewAttribute(types.AttributeKeyDelegator, msg.DelegatorAddress),
			sdk.NewAttribute(types.AttributeKeyCreationHeight, strconv.FormatInt(msg.CreationHeight, 10)),
		),
	)

	return &types.MsgCancelUnbondingDelegationResponse{}, nil
}

// UpdateParams defines a method to perform updation of params exist in x/staking module.
func (k msgServer) UpdateParams(ctx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if k.authority != msg.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.authority, msg.Authority)
	}

	if err := msg.Params.Validate(); err != nil {
		return nil, err
	}

	// store params
	if err := k.SetParams(ctx, msg.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

func (k msgServer) calculateDelegateLicenseCount(ctx context.Context, amount sdk.Coin, validator types.Validator, delegatorAddress sdk.AccAddress, shares math.LegacyDec) (math.Int, error) {
	currentDelegation, err := k.Keeper.GetDelegation(ctx, delegatorAddress, sdk.ValAddress(validator.GetOperator()))
	if err != nil {
		return math.Int{}, err
	}

	if !shares.IsNil() {
		// Validate minimum amount , currentDelegation - unbond >= min delegation
		if !currentDelegation.Shares.Equal(shares) {
			// NOT remove entire shares , only unbond some of it.
			if !validator.MinDelegation.IsNil() && currentDelegation.Shares.Sub(shares).LT(validator.MinDelegation.ToLegacyDec()) {
				return math.NewInt(0), types.ErrDelegationBelowMinimum
			}
		}
	}

	// Validate Minimum and Increment
	if !validator.MinDelegation.IsNil() && amount.Amount.LT(validator.MinDelegation) {
		return math.NewInt(0), types.ErrDelegationBelowMinimum
	}

	delegateLicenseCount := math.ZeroInt()
	// Deduct minimum from value to validate increment
	amountToValidateIncrement := math.NewIntFromBigInt(amount.Amount.BigInt())
	if !validator.MinDelegation.IsNil() {
		amountToValidateIncrement = amountToValidateIncrement.Sub(validator.MinDelegation)
		delegateLicenseCount = delegateLicenseCount.Add(math.OneInt())
	}

	// Validate DelegationIncrement
	increment := math.OneInt()
	if !validator.DelegationIncrement.IsNil() {
		increment = validator.DelegationIncrement
	}

	if amountToValidateIncrement.GT(math.ZeroInt()) {
		divAmount := amountToValidateIncrement.Quo(increment)
		modAmount := amountToValidateIncrement.Mod(increment)
		if modAmount.GT(math.ZeroInt()) {
			return math.NewInt(0), types.ErrInvalidIncrementDelegation
		}
		delegateLicenseCount = delegateLicenseCount.Add(divAmount)
	}
	return delegateLicenseCount, nil
}
