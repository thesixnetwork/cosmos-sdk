package keeper

import (
	"context"
	"time"

	metrics "github.com/armon/go-metrics"
	tmstrings "github.com/tendermint/tendermint/libs/strings"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the bank MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (k msgServer) SetValidatorApproval(goCtx context.Context, msg *types.MsgSetValidatorApproval) (*types.MsgSetValidatorApprovalResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	validatorApproval, found := k.GetValidatorApproval(ctx)
	if !found {
		panic("Validator approval not found")
	}

	if validatorApproval.ApproverAddress != msg.ApproverAddress {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "Msg sender is not current approver")
	}

	var newApproverAddress string
	if _, err := sdk.AccAddressFromBech32(msg.NewApproverAddress); err == nil {
		newApproverAddress = msg.NewApproverAddress
	} else {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid new approver address")
	}

	newValidatorApprovalState := types.ValidatorApproval{
		ApproverAddress: newApproverAddress,
		Enabled:         msg.Enabled,
	}

	k.SetNewValidatorApprovalState(ctx, newValidatorApprovalState)

	return &types.MsgSetValidatorApprovalResponse{}, nil
}

// CreateValidator defines a method for creating a new validator
func (k msgServer) CreateValidator(goCtx context.Context, msg *types.MsgCreateValidator) (*types.MsgCreateValidatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	approval, foundApproval := k.GetValidatorApproval(ctx)
	if !foundApproval {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrNotFound, "Validator approval is somehow does not existed")
	}

	if approval.Enabled && msg.ApproverAddress != approval.ApproverAddress {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "Wrong approver for create validator")
	}

	valAddr, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return nil, err
	}

	// check to see if the pubkey or sender has been registered before
	if _, found := k.GetValidator(ctx, valAddr); found {
		return nil, types.ErrValidatorOwnerExists
	}

	pk, ok := msg.Pubkey.GetCachedValue().(cryptotypes.PubKey)
	if !ok {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "Expecting cryptotypes.PubKey, got %T", pk)
	}

	if _, found := k.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(pk)); found {
		return nil, types.ErrValidatorPubKeyExists
	}

	bondDenom := k.BondDenom(ctx)
	if msg.Value.Denom != bondDenom {
		return nil, sdkerrors.Wrapf(
			sdkerrors.ErrInvalidRequest, "invalid coin denomination: got %s, expected %s", msg.Value.Denom, bondDenom,
		)
	}

	if _, err := msg.Description.EnsureLength(); err != nil {
		return nil, err
	}

	cp := ctx.ConsensusParams()
	if cp != nil && cp.Validator != nil {
		if !tmstrings.StringInSlice(pk.Type(), cp.Validator.PubKeyTypes) {
			return nil, sdkerrors.Wrapf(
				types.ErrValidatorPubKeyTypeNotSupported,
				"got: %s, expected: %s", pk.Type(), cp.Validator.PubKeyTypes,
			)
		}
	}

	validator, err := types.NewValidator(valAddr, pk, msg.Description)
	if err != nil {
		return nil, err
	}
	commission := types.NewCommissionWithTime(
		msg.Commission.Rate, msg.Commission.MaxRate,
		msg.Commission.MaxChangeRate, ctx.BlockHeader().Time,
	)

	validator, err = validator.SetInitialCommission(commission)
	if err != nil {
		return nil, err
	}

	delegatorAddress, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return nil, err
	}

	validator.MinSelfDelegation = msg.MinSelfDelegation
	// CustomValidator
	validator.MinDelegation = msg.MinDelegation
	validator.DelegationIncrement = msg.DelegationIncrement
	// when Min Delegation is not defined, default as DelegationIncrement
	if msg.MinDelegation.IsNil() {
		validator.MinDelegation = validator.DelegationIncrement
	}

	switch {
	case msg.LicenseMode:
		// Verify that MinDelegation and DelegationIncrement is  defined and contains the same value
		if msg.DelegationIncrement.IsNil() || !validator.MinDelegation.Equal(validator.DelegationIncrement) {
			return nil, types.ErrLicenseIncrement
		}

		validator.LicenseMode = true
		if validator.MaxLicense = msg.MaxLicense; msg.MaxLicense.IsNil() {
			return nil, types.ErrMaxLicenseMustBeDefined
		} // bug is nill genersis

		// Count licesene amount for validator
		divAmount := msg.Value.Amount.Quo(validator.DelegationIncrement)
		modAmount := msg.Value.Amount.Mod(validator.DelegationIncrement)
		if modAmount.GT(sdk.ZeroInt()) {
			return nil, types.ErrInvalidIncrementDelegation
		}
		if divAmount.GT(validator.MaxLicense) {
			return nil, types.ErrNotEnoughLicense
		}
		validator.LicenseCount = divAmount
		// Force disable redelegation when
		validator.EnableRedelegation = false
		validator.SpecialMode = false
	case msg.SpecialMode:
		validator.LicenseMode = false
		validator.SpecialMode = true
		validator.EnableRedelegation = msg.EnableRedelegation
	default:
		validator.EnableRedelegation = msg.EnableRedelegation
	}

	k.SetValidator(ctx, validator)
	k.SetValidatorByConsAddr(ctx, validator)
	k.SetNewValidatorByPowerIndex(ctx, validator)

	// call the after-creation hook
	k.AfterValidatorCreated(ctx, validator.GetOperator())

	// move coins from the msg.Address account to a (self-delegation) delegator account
	// the validator account and global shares are updated within here
	// NOTE source will always be from a wallet which are unbonded
	_, err = k.Keeper.Delegate(ctx, delegatorAddress, msg.Value.Amount, types.Unbonded, validator, true)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeCreateValidator,
			sdk.NewAttribute(types.AttributeKeyValidator, msg.ValidatorAddress),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Value.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.DelegatorAddress),
		),
	})

	return &types.MsgCreateValidatorResponse{}, nil
}

// EditValidator defines a method for editing an existing validator
func (k msgServer) EditValidator(goCtx context.Context, msg *types.MsgEditValidator) (*types.MsgEditValidatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	valAddr, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return nil, err
	}
	// validator must already be registered
	validator, found := k.GetValidator(ctx, valAddr)
	if !found {
		return nil, types.ErrNoValidatorFound
	}

	// replace all editable fields (clients should autofill existing values)
	description, err := validator.Description.UpdateDescription(msg.Description)
	if err != nil {
		return nil, err
	}

	validator.Description = description

	switch {
	case msg.LicenseMode:
		// validate max license
		if validator.LicenseMode && !msg.MaxLicense.IsNil() && msg.MaxLicense.LT(validator.MaxLicense) {
			return nil, types.ErrMaxLicenseMustBeGeater
		}

		if validator.LicenseMode && !msg.MaxLicense.IsNil() {
			validator.MaxLicense = msg.MaxLicense
		}
		validator.SpecialMode = false
		validator.LicenseMode = true
	case msg.SpecialMode:
		validator.SpecialMode = true
		validator.LicenseMode = false
	default:
		validator.SpecialMode = false
		validator.LicenseMode = false
	}

	if msg.CommissionRate != nil {
		commission, err := k.UpdateValidatorCommission(ctx, validator, *msg.CommissionRate)
		if err != nil {
			return nil, err
		}

		// call the before-modification hook since we're about to update the commission
		k.BeforeValidatorModified(ctx, valAddr)

		validator.Commission = commission
	}

	if msg.MinSelfDelegation != nil {
		if !msg.MinSelfDelegation.GT(validator.MinSelfDelegation) {
			return nil, types.ErrMinSelfDelegationDecreased
		}

		if msg.MinSelfDelegation.GT(validator.Tokens) {
			return nil, types.ErrSelfDelegationBelowMinimum
		}

		validator.MinSelfDelegation = (*msg.MinSelfDelegation)
	}

	k.SetValidator(ctx, validator)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeEditValidator,
			sdk.NewAttribute(types.AttributeKeyCommissionRate, validator.Commission.String()),
			sdk.NewAttribute(types.AttributeKeyMinSelfDelegation, validator.MinSelfDelegation.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.ValidatorAddress),
		),
	})

	return &types.MsgEditValidatorResponse{}, nil
}

// Delegate defines a method for performing a delegation of coins from a delegator to a validator
func (k msgServer) Delegate(goCtx context.Context, msg *types.MsgDelegate) (*types.MsgDelegateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	valAddr, valErr := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if valErr != nil {
		return nil, valErr
	}

	validator, found := k.GetValidator(ctx, valAddr)
	if !found {
		return nil, types.ErrNoValidatorFound
	}

	delegatorAddress, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return nil, err
	}

	bondDenom := k.BondDenom(ctx)
	if msg.Amount.Denom != bondDenom {
		return nil, sdkerrors.Wrapf(
			sdkerrors.ErrInvalidRequest, "invalid coin denomination: got %s, expected %s", msg.Amount.Denom, bondDenom,
		)
	}

	// CustomValidator
	// New Delegation or Update
	_, existsDelegation := k.Keeper.GetDelegation(ctx, delegatorAddress, validator.GetOperator())

	if validator.LicenseMode {
		// Validate Minimum and Increment
		if !validator.MinDelegation.IsNil() && !existsDelegation && msg.Amount.Amount.LT(validator.MinDelegation) {
			return nil, types.ErrDelegationBelowMinimum
		}
		delegateLicenseCount := sdk.ZeroInt()
		// Deduct minimum from value to validate increment
		amountToValidateIncrement := sdk.NewIntFromBigInt(msg.Amount.Amount.BigInt())
		if !validator.MinDelegation.IsNil() && !existsDelegation {
			amountToValidateIncrement = amountToValidateIncrement.Sub(validator.MinDelegation)
			delegateLicenseCount = delegateLicenseCount.Add(sdk.OneInt())
		}
		// Validate DelegationIncrement
		increment := sdk.OneInt()
		if !validator.DelegationIncrement.IsNil() {
			increment = validator.DelegationIncrement
		}
		// TODO: recheck div amount or mod
		if amountToValidateIncrement.GT(sdk.ZeroInt()) {
			divAmount := amountToValidateIncrement.Quo(increment) // TODO: recheck
			modAmount := amountToValidateIncrement.Mod(increment)
			if modAmount.GT(sdk.ZeroInt()) {
				return nil, types.ErrInvalidIncrementDelegation
			}
			delegateLicenseCount = delegateLicenseCount.Add(divAmount)
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
				[]string{"tx", "msg", msg.Type()},
				float32(msg.Amount.Amount.Int64()),
				[]metrics.Label{telemetry.NewLabel("denom", msg.Amount.Denom)},
			)
		}()
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeDelegate,
			sdk.NewAttribute(types.AttributeKeyValidator, msg.ValidatorAddress),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyNewShares, newShares.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.DelegatorAddress),
		),
	})

	return &types.MsgDelegateResponse{}, nil
}

// BeginRedelegate defines a method for performing a redelegation of coins from a delegator and source validator to a destination validator
func (k msgServer) BeginRedelegate(goCtx context.Context, msg *types.MsgBeginRedelegate) (*types.MsgBeginRedelegateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	valSrcAddr, err := sdk.ValAddressFromBech32(msg.ValidatorSrcAddress)
	if err != nil {
		return nil, err
	}
	valDestAddr, err := sdk.ValAddressFromBech32(msg.ValidatorDstAddress)
	if err != nil {
		return nil, err
	}
	delegatorAddress, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return nil, err
	}
	shares, err := k.ValidateUnbondAmount(
		ctx, delegatorAddress, valSrcAddr, msg.Amount.Amount,
	)
	if err != nil {
		return nil, err
	}

	bondDenom := k.BondDenom(ctx)
	if msg.Amount.Denom != bondDenom {
		return nil, sdkerrors.Wrapf(
			sdkerrors.ErrInvalidRequest, "invalid coin denomination: got %s, expected %s", msg.Amount.Denom, bondDenom,
		)
	}

	valDstAddr, err := sdk.ValAddressFromBech32(msg.ValidatorDstAddress)
	if err != nil {
		return nil, err
	}

	// Validate source validator
	// Get Validator
	sourceVal, sourceValFound := k.GetValidator(ctx, valSrcAddr) // TODO: to check
	if !sourceValFound {
		return nil, types.ErrNoValidatorFound
	}
	destVal, destValFound := k.GetValidator(ctx, valDestAddr) // TODO: to check
	if !destValFound {
		return nil, types.ErrNoValidatorFound
	}
	if !sourceVal.EnableRedelegation || !destVal.EnableRedelegation {
		return nil, types.ErrRedelegationDisable
	}

	// Get Current Delegation
	currentSourceDelegation, existsSourceDelegation := k.Keeper.GetDelegation(ctx, delegatorAddress, sourceVal.GetOperator())
	// Validate minimum amount , currentDelegation - unbond >= min delegation
	if !currentSourceDelegation.Shares.Equal(shares) {
		// NOT remove entire shares , only unbond some of it.
		if !sourceVal.MinDelegation.IsNil() && currentSourceDelegation.Shares.Sub(shares).LT(sourceVal.MinDelegation.ToDec()) {
			return nil, types.ErrDelegationBelowMinimum
		}
	}
	// Deduct minimum from value to validate increment
	amountToValidateIncrement := sdk.NewIntFromBigInt(msg.Amount.Amount.BigInt())
	if !sourceVal.MinDelegation.IsNil() && !existsSourceDelegation {
		amountToValidateIncrement = amountToValidateIncrement.Sub(sourceVal.MinDelegation)
	}

	increment := sdk.OneInt()
	if !sourceVal.DelegationIncrement.IsNil() {
		increment = sourceVal.DelegationIncrement
	}
	// Validate DelegationIncrement
	if amountToValidateIncrement.GT(sdk.ZeroInt()) {
		// not remove
		modAmount := amountToValidateIncrement.Mod(increment)
		if modAmount.GT(sdk.ZeroInt()) {
			return nil, types.ErrInvalidIncrementDelegation
		}
	}

	// Validate destination
	// New Delegation or Update
	_, existsDestinationDelegation := k.Keeper.GetDelegation(ctx, delegatorAddress, destVal.GetOperator())
	// Validate Minimum and Increment
	if !destVal.MinDelegation.IsNil() && !existsDestinationDelegation && msg.Amount.Amount.LT(destVal.MinDelegation) {
		return nil, types.ErrDelegationBelowMinimum
	}
	// Deduct minimum from value to validate increment
	amountToValidateIncrement = sdk.NewIntFromBigInt(msg.Amount.Amount.BigInt())
	if !destVal.MinDelegation.IsNil() && !existsDestinationDelegation {
		amountToValidateIncrement = amountToValidateIncrement.Sub(destVal.MinDelegation)
	}
	// Validate DelegationIncrement
	increment = sdk.OneInt()
	if !destVal.DelegationIncrement.IsNil() {
		increment = destVal.DelegationIncrement
	}
	if amountToValidateIncrement.GT(sdk.ZeroInt()) {
		modAmount := amountToValidateIncrement.Mod(increment)
		if modAmount.GT(sdk.ZeroInt()) {
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
				[]string{"tx", "msg", msg.Type()},
				float32(msg.Amount.Amount.Int64()),
				[]metrics.Label{telemetry.NewLabel("denom", msg.Amount.Denom)},
			)
		}()
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeRedelegate,
			sdk.NewAttribute(types.AttributeKeySrcValidator, msg.ValidatorSrcAddress),
			sdk.NewAttribute(types.AttributeKeyDstValidator, msg.ValidatorDstAddress),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyCompletionTime, completionTime.Format(time.RFC3339)),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.DelegatorAddress),
		),
	})

	return &types.MsgBeginRedelegateResponse{
		CompletionTime: completionTime,
	}, nil
}

// Undelegate defines a method for performing an undelegation from a delegate and a validator
func (k msgServer) Undelegate(goCtx context.Context, msg *types.MsgUndelegate) (*types.MsgUndelegateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	addr, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return nil, err
	}
	delegatorAddress, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return nil, err
	}
	shares, err := k.ValidateUnbondAmount(
		ctx, delegatorAddress, addr, msg.Amount.Amount,
	)
	if err != nil {
		return nil, err
	}

	bondDenom := k.BondDenom(ctx)
	if msg.Amount.Denom != bondDenom {
		return nil, sdkerrors.Wrapf(
			sdkerrors.ErrInvalidRequest, "invalid coin denomination: got %s, expected %s", msg.Amount.Denom, bondDenom,
		)
	}
	/// Custom Validator
	// Get Current Validator
	validator, found := k.GetValidator(ctx, addr)
	if !found {
		return nil, types.ErrNoValidatorFound
	}

	var completionTime time.Time

	switch {
	case validator.LicenseMode:
		// Get Current Delegation
		currentDelegation, existsDelegation := k.Keeper.GetDelegation(ctx, delegatorAddress, validator.GetOperator())
		// Validate minimum amount , currentDelegation - unbond >= min delegation
		if !currentDelegation.Shares.Equal(shares) {
			// NOT remove entire shares , only unbond some of it.
			if !validator.MinDelegation.IsNil() && currentDelegation.Shares.Sub(shares).LT(validator.MinDelegation.ToDec()) {
				return nil, types.ErrDelegationBelowMinimum
			}
		}
		delegateLicenseCount := sdk.ZeroInt()
		// Deduct minimum from value to validate increment
		amountToValidateIncrement := sdk.NewIntFromBigInt(msg.Amount.Amount.BigInt())
		if !validator.MinDelegation.IsNil() && !existsDelegation {
			amountToValidateIncrement = amountToValidateIncrement.Sub(validator.MinDelegation)
			delegateLicenseCount = delegateLicenseCount.Add(sdk.OneInt())
		}

		increment := sdk.OneInt()
		if !validator.DelegationIncrement.IsNil() {
			increment = validator.DelegationIncrement
		}
		// Validate DelegationIncrement
		if amountToValidateIncrement.GT(sdk.ZeroInt()) {
			// not remove
			divAmount := amountToValidateIncrement.Quo(increment)
			modAmount := amountToValidateIncrement.Mod(increment)
			if modAmount.GT(sdk.ZeroInt()) {
				return nil, types.ErrInvalidIncrementDelegation
			}
			delegateLicenseCount = delegateLicenseCount.Add(divAmount)
		}
		// Validate License
		// decrease license count in validator
		validator.LicenseCount = validator.LicenseCount.Sub(delegateLicenseCount)
		// Update Validator
		k.Keeper.SetValidator(ctx, validator)

		completionTime, err = k.Keeper.Undelegate(ctx, delegatorAddress, addr, shares)
		if err != nil {
			return nil, err
		}
	case validator.SpecialMode:
		completionTime, err = k.Keeper.UndelegatSpecial(ctx, delegatorAddress, addr, shares)
		if err != nil {
			return nil, err
		}
	default:
		completionTime, err = k.Keeper.Undelegate(ctx, delegatorAddress, addr, shares)
		if err != nil {
			return nil, err
		}
	}

	if msg.Amount.Amount.IsInt64() {
		defer func() {
			telemetry.IncrCounter(1, types.ModuleName, "undelegate")
			telemetry.SetGaugeWithLabels(
				[]string{"tx", "msg", msg.Type()},
				float32(msg.Amount.Amount.Int64()),
				[]metrics.Label{telemetry.NewLabel("denom", msg.Amount.Denom)},
			)
		}()
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeUnbond,
			sdk.NewAttribute(types.AttributeKeyValidator, msg.ValidatorAddress),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyCompletionTime, completionTime.Format(time.RFC3339)),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.DelegatorAddress),
		),
	})

	return &types.MsgUndelegateResponse{
		CompletionTime: completionTime,
	}, nil
}
