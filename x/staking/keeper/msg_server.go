package keeper

import (
	"bytes"
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
		if !errorsmod.IsOf(err, types.ErrNoValidatorFound) {
			return nil, err
		}
		// the record is only written by InitGenesis, so a chain that upgraded
		// in-place starts without one; governance can bootstrap it below
		validatorApproval = types.ValidatorApproval{}
	}

	// an absent or empty approver can only be claimed through governance;
	// otherwise the first caller would appoint themselves approver
	if validatorApproval.ApproverAddress == "" {
		if msg.ApproverAddress != k.authority {
			return nil, errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "no approver is set; only the authority %s can bootstrap the approval state", k.authority)
		}
	} else if validatorApproval.ApproverAddress != msg.ApproverAddress {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "Msg sender is not current approver")
	}

	// an empty new approver keeps the current one
	newApproverAddress := validatorApproval.ApproverAddress
	if msg.NewApproverAddress != "" {
		if _, err := sdk.AccAddressFromBech32(msg.NewApproverAddress); err != nil {
			return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid new approver address")
		}
		newApproverAddress = msg.NewApproverAddress
	}
	if newApproverAddress == "" {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "new approver address must be provided")
	}

	approved := validatorApproval.ApprovedValidators
	for _, addr := range msg.RevokeValidators {
		for i, existing := range approved {
			if existing == addr {
				approved = append(approved[:i], approved[i+1:]...)
				break
			}
		}
	}
	for _, addr := range msg.ApproveValidators {
		if _, err := k.validatorAddressCodec.StringToBytes(addr); err != nil {
			return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid operator address to approve: %s", err)
		}
		duplicate := false
		for _, existing := range approved {
			if existing == addr {
				duplicate = true
				break
			}
		}
		if !duplicate {
			approved = append(approved, addr)
		}
	}

	newValidatorApprovalState := types.ValidatorApproval{
		ApproverAddress:    newApproverAddress,
		Enabled:            msg.Enabled,
		ApprovedValidators: approved,
	}

	if err := k.SetNewValidatorApprovalState(ctx, newValidatorApprovalState); err != nil {
		return nil, err
	}

	return &types.MsgSetValidatorApprovalResponse{}, nil
}

// CreateValidator defines a method for creating a new validator
func (k msgServer) CreateValidator(ctx context.Context, msg *types.MsgCreateValidator) (*types.MsgCreateValidatorResponse, error) {
	approval, err := k.GetValidatorApproval(ctx)
	if err != nil {
		if !errorsmod.IsOf(err, types.ErrNoValidatorFound) {
			return nil, err
		}
		// the approval state is only written by InitGenesis, so a chain that
		// upgraded in-place may not have it; treat that as approval disabled
		// instead of blocking validator creation forever
		approval = types.ValidatorApproval{}
	}

	valAddr, err := k.validatorAddressCodec.StringToBytes(msg.ValidatorAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid validator address: %s", err)
	}

	// msg.ApproverAddress is not a tx signer, so it proves nothing; the gate
	// instead requires an approval the approver granted on-chain (via the
	// signed MsgSetValidatorApproval) for this exact operator, consumed on use
	if approval.Enabled {
		approvedIdx := -1
		for i, addr := range approval.ApprovedValidators {
			approvedBz, err := k.validatorAddressCodec.StringToBytes(addr)
			if err != nil {
				continue
			}
			if bytes.Equal(approvedBz, valAddr) {
				approvedIdx = i
				break
			}
		}
		if approvedIdx == -1 {
			return nil, errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "validator %s has not been approved for creation", msg.ValidatorAddress)
		}
		approval.ApprovedValidators = append(approval.ApprovedValidators[:approvedIdx], approval.ApprovedValidators[approvedIdx+1:]...)
		if err := k.SetNewValidatorApprovalState(ctx, approval); err != nil {
			return nil, err
		}
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
	// CustomValidator
	validator.MinDelegation = msg.MinDelegation
	validator.DelegationIncrement = msg.DelegationIncrement
	// when Min Delegation is not defined, default as DelegationIncrement.
	// zero counts as undefined: the wire cannot carry "unset" for the
	// non-nullable Int fields, so a nil Int arrives here as zero
	if msg.MinDelegation.IsNil() || msg.MinDelegation.IsZero() {
		validator.MinDelegation = validator.DelegationIncrement
	}

	switch msg.Mode {
	case types.ValidatorMode_MODE_LICENSE:
		validator.Mode = types.ValidatorMode_MODE_LICENSE

		licenseCount, err := validator.LicenseCountCreationCalculate(msg.DelegationIncrement, msg.MinDelegation, msg.MaxLicense, msg.Value)
		if err != nil {
			return nil, err
		}
		validator.LicenseCount = licenseCount
	case types.ValidatorMode_MODE_FAST:
		validator.Mode = types.ValidatorMode_MODE_FAST
	case types.ValidatorMode_MODE_NORMAL:
		validator.Mode = types.ValidatorMode_MODE_NORMAL
	default:
		// the enum comes straight off the wire, so an unknown value must be
		// rejected instead of silently creating a normal validator
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "Invalid validator mode")
	}
	// redelegation eligibility is decided by validator mode at redelegate time
	// (LICENSE/NORMAL allowed, FAST excluded); the enable_redelegation wire
	// field is no longer used for gating
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

	requestedMode := validator.Mode
	if msg.Mode != "" {
		requestedMode, err = types.ParseValidatorMode(msg.Mode)
		if err != nil {
			return nil, err
		}
	}

	switch requestedMode {
	case types.ValidatorMode_MODE_LICENSE:
		// increment and min-delegation updates are applied before the license
		// recount below, so the edit only goes through when the new values
		// "make sense" for every delegation that already exists
		if msg.DelegationIncrement != nil && !msg.DelegationIncrement.IsNil() {
			if !msg.DelegationIncrement.IsPositive() {
				return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "delegation increment must be a positive integer")
			}
			// a new increment must divide every existing delegation exactly, so
			// the redistribution cannot silently split a stake; this only gates
			// CHANGING the increment — a slashed validator (whose delegations
			// are no longer whole increments) can still edit everything else
			if !msg.DelegationIncrement.Equal(validator.DelegationIncrement) {
				delegations, err := k.GetValidatorDelegations(ctx, valAddr)
				if err != nil {
					return nil, err
				}
				for _, delegation := range delegations {
					tokens := validator.TokensFromShares(delegation.Shares).TruncateInt()
					if tokens.Mod(*msg.DelegationIncrement).GT(math.ZeroInt()) {
						return nil, types.ErrInvalidIncrementDelegation
					}
				}
			}
			validator.DelegationIncrement = *msg.DelegationIncrement
		}

		if msg.MinDelegation != nil && !msg.MinDelegation.IsNil() {
			if !msg.MinDelegation.IsPositive() {
				return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "minimum delegation must be a positive integer")
			}

			delegations, err := k.GetValidatorDelegations(ctx, valAddr)
			if err != nil {
				return nil, err
			}

			for _, delegation := range delegations {
				tokens := validator.TokensFromShares(delegation.Shares).TruncateInt()
				if tokens.LT(*msg.MinDelegation) {
					return nil, types.ErrNewMinimumDelegationInvalid
				}
			}

			validator.MinDelegation = *msg.MinDelegation
		}

		// license accounting needs both knobs: MinDelegation is the entry
		// threshold, DelegationIncrement is the license unit; a validator
		// created in another mode may have them nil or zero
		if validator.DelegationIncrement.IsNil() || !validator.DelegationIncrement.IsPositive() ||
			validator.MinDelegation.IsNil() || !validator.MinDelegation.IsPositive() {
			return nil, types.ErrLicenseIncrement
		}

		// validate max license: the cap may be raised OR lowered. It is checked
		// against the freshly recomputed usage below (ErrNotEnoughLicense), not
		// the stored LicenseCount (which may be stale, e.g. after a slash), so a
		// valid reduction is never wrongly blocked while the count-<=-cap
		// invariant still holds.
		if msg.MaxLicense != nil && !msg.MaxLicense.IsNil() {
			if !msg.MaxLicense.IsPositive() {
				return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "max license must be a positive integer")
			}
			validator.MaxLicense = *msg.MaxLicense
		}

		if validator.MaxLicense.IsNil() {
			return nil, types.ErrMaxLicenseMustBeDefined
		}

		// redistribute license usage across the current state and re-check it
		// against the (possibly new) cap
		totalLicenses, err := k.calculateEditLicenseCount(ctx, sdk.ValAddress(valAddr), validator)
		if err != nil {
			return nil, err
		}

		if totalLicenses.GT(validator.MaxLicense) {
			return nil, types.ErrNotEnoughLicense
		}

		validator.Mode = types.ValidatorMode_MODE_LICENSE
		validator.LicenseCount = totalLicenses
	case types.ValidatorMode_MODE_FAST:
		// fast mode grants near-instant unbonding, so it can only be chosen at
		// create-validator (which is gated by the approver); switching an
		// existing validator into fast mode would bypass the unbonding period
		if validator.Mode != types.ValidatorMode_MODE_FAST {
			return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "cannot switch an existing validator into fast mode")
		}
	case types.ValidatorMode_MODE_NORMAL:
		// reached on an explicit "normal" request (an owner decision) or when
		// the mode was omitted and the validator already is normal
		validator.Mode = types.ValidatorMode_MODE_NORMAL
	default:
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "Invalid validator mode")
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
		if validator.MaxLicense.IsNil() {
			return nil, types.ErrMaxLicenseMustBeDefined
		}
		if validator.LicenseCount.IsNil() {
			validator.LicenseCount = math.ZeroInt()
		}
		delegateLicenseCount, err := k.calculateDelegateLicenseCount(ctx, msg.Amount, validator, delegatorAddress)
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
		k.SetValidator(ctx, validator)
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
	sourceVal, err := k.GetValidator(ctx, valSrcAddr)
	if err != nil {
		return nil, err
	}
	destVal, err := k.GetValidator(ctx, valDstAddr)
	if err != nil {
		return nil, err
	}

	if bytes.Equal(valSrcAddr, valDstAddr) {
		return nil, types.ErrSelfRedelegation
	}

	// Fast-mode validators have special instant-exit semantics and are excluded
	// from redelegation entirely — they cannot be a source or a destination.
	if sourceVal.Mode == types.ValidatorMode_MODE_FAST || destVal.Mode == types.ValidatorMode_MODE_FAST {
		return nil, errorsmod.Wrap(types.ErrRedelegationDisable, "fast-mode validators cannot redelegate")
	}

	// License accounting: the redelegated amount must be a whole multiple of
	// BOTH validators' delegation_increment. On the source it behaves like an
	// undelegation (frees licenses); on the destination like a delegation
	// (consumes licenses, respecting MinDelegation and MaxLicense). Non-license
	// validators skip these checks. Everything is validated before any state
	// change so an invalid redelegation fails cleanly.
	srcIsLicense := sourceVal.Mode == types.ValidatorMode_MODE_LICENSE
	dstIsLicense := destVal.Mode == types.ValidatorMode_MODE_LICENSE

	srcLicenseFreed := math.ZeroInt()
	if srcIsLicense {
		srcLicenseFreed, err = k.calculateUndelegateLicenseCount(ctx, msg.Amount, sourceVal, delegatorAddress)
		if err != nil {
			return nil, err
		}
	}

	dstLicenseUsed := math.ZeroInt()
	if dstIsLicense {
		if destVal.MaxLicense.IsNil() {
			return nil, types.ErrMaxLicenseMustBeDefined
		}
		if destVal.LicenseCount.IsNil() {
			destVal.LicenseCount = math.ZeroInt()
		}
		dstLicenseUsed, err = k.calculateDelegateLicenseCount(ctx, msg.Amount, destVal, delegatorAddress)
		if err != nil {
			return nil, err
		}
		if destVal.LicenseCount.GTE(destVal.MaxLicense) {
			return nil, types.ErrLicenseLimit
		}
		if dstLicenseUsed.Add(destVal.LicenseCount).GT(destVal.MaxLicense) {
			return nil, types.ErrNotEnoughLicense
		}
	}

	completionTime, err := k.BeginRedelegation(
		ctx, delegatorAddress, valSrcAddr, valDstAddr, shares,
	)
	if err != nil {
		return nil, err
	}

	// The stake has moved; update license counts on both validators. Re-fetch
	// so the token/share changes BeginRedelegation made are preserved and only
	// LicenseCount is adjusted.
	if srcIsLicense {
		sourceVal, err = k.GetValidator(ctx, valSrcAddr)
		if err != nil {
			// a full redelegation can drain and remove the source validator;
			// there is then nothing left to adjust
			if !errorsmod.IsOf(err, types.ErrNoValidatorFound) {
				return nil, err
			}
		} else if !sourceVal.LicenseCount.IsNil() {
			sourceVal.LicenseCount = sourceVal.LicenseCount.Sub(srcLicenseFreed)
			if sourceVal.LicenseCount.IsNegative() {
				sourceVal.LicenseCount = math.ZeroInt()
			}
			if err := k.SetValidator(ctx, sourceVal); err != nil {
				return nil, err
			}
		}
	}

	if dstIsLicense {
		destVal, err = k.GetValidator(ctx, valDstAddr)
		if err != nil {
			return nil, err
		}
		if destVal.LicenseCount.IsNil() {
			destVal.LicenseCount = math.ZeroInt()
		}
		destVal.LicenseCount = destVal.LicenseCount.Add(dstLicenseUsed)
		if err := k.SetValidator(ctx, destVal); err != nil {
			return nil, err
		}
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
		// validate the undelegation shape (whole increments, entry minimum,
		// full exit) but do NOT release the licenses yet: they stay occupied
		// until the unbonding period completes (see CompleteUnbonding), so a
		// freed slot cannot be taken while a cancel-unbonding is still possible
		if _, err := k.calculateUndelegateLicenseCount(ctx, msg.Amount, validator, delegatorAddress); err != nil {
			return nil, err
		}

		completionTime, undelegatedAmt, err = k.Keeper.Undelegate(ctx, delegatorAddress, addr, shares)
		if err != nil {
			return nil, err
		}
		undelegatedCoin = sdk.NewCoin(msg.Amount.Denom, undelegatedAmt)
	case types.ValidatorMode_MODE_FAST:
		// the instant exit is a whitelist privilege for delegators only: the
		// operator's self-bond backs consensus, so pulling it without the
		// unbonding period would leave no stake at risk for slashing — the
		// operator always waits the full unbonding time. A delegator that was
		// removed from the whitelist can still leave, but through the normal
		// unbonding period rather than being blocked (or instantly exiting)
		if !bytes.Equal(delegatorAddress, addr) && k.IsSpecialDelegator(ctx, sdk.ValAddress(addr), delegatorAddress) {
			completionTime, undelegatedAmt, err = k.UndelegateSpecial(ctx, delegatorAddress, addr, shares)
		} else {
			completionTime, undelegatedAmt, err = k.Keeper.Undelegate(ctx, delegatorAddress, addr, shares)
		}
		if err != nil {
			return nil, err
		}
		undelegatedCoin = sdk.NewCoin(msg.Amount.Denom, undelegatedAmt)
	default:
		completionTime, undelegatedAmt, err = k.Keeper.Undelegate(ctx, delegatorAddress, addr, shares)
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

	// licenses stay occupied during the whole unbonding period (they are only
	// released by CompleteUnbonding), so canceling is license-neutral: the
	// count is untouched and the cap cannot be exceeded by a
	// delegate -> undelegate -> cancel cycle. Only a partial cancel has its
	// delegation shape re-validated (whole increments, or the entry minimum
	// when the cancel re-creates a fully-exited delegation); canceling the
	// whole entry restores stake that was already accepted and is always
	// allowed — a slash may have left the balance at a non-increment amount
	if validator.Mode == types.ValidatorMode_MODE_LICENSE && !msg.Amount.Amount.Equal(unbondEntry.Balance) {
		if _, err := k.calculateDelegateLicenseCount(ctx, msg.Amount, validator, delegatorAddress); err != nil {
			return nil, err
		}
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

func (k msgServer) calculateDelegateLicenseCount(ctx context.Context, amount sdk.Coin, validator types.Validator, delegatorAddress sdk.AccAddress) (math.Int, error) {
	valAddr, err := k.validatorAddressCodec.StringToBytes(validator.GetOperator())
	if err != nil {
		return math.Int{}, sdkerrors.ErrInvalidAddress.Wrapf("invalid validator address: %s", err)
	}
	_, existed := k.GetExistingDelegation(ctx, delegatorAddress, valAddr)

	// MinDelegation is the entry threshold: it applies to NEW delegations
	// only, while every delegation (entry or top-up) must be a whole number
	// of DelegationIncrement, the license unit
	if !validator.MinDelegation.IsNil() && !existed && amount.Amount.LT(validator.MinDelegation) {
		return math.Int{}, types.ErrDelegationBelowMinimum
	}

	increment := math.OneInt()
	if !validator.DelegationIncrement.IsNil() {
		increment = validator.DelegationIncrement
	}

	if amount.Amount.Mod(increment).GT(math.ZeroInt()) {
		return math.NewInt(0), types.ErrInvalidIncrementDelegation
	}
	return amount.Amount.Quo(increment), nil
}

// calculateEditLicenseCount recomputes a license validator's total license
// usage from state for MsgEditValidator: each delegation occupies
// ceil(tokens/increment) licenses, and stakes still locked in the unbonding
// queue keep holding ceil(balance/increment) until CompleteUnbonding releases
// them with the same rule. The ceiling makes the recount slash-tolerant: a
// slash leaves token values at non-increment amounts, and requiring exact
// multiples here would permanently block every subsequent edit (even a
// description-only one). It also makes the recount the recovery path for any
// LicenseCount drift slashing causes.
func (k msgServer) calculateEditLicenseCount(ctx context.Context, valAddr sdk.ValAddress, validator types.Validator) (math.Int, error) {
	delegations, err := k.GetValidatorDelegations(ctx, valAddr)
	if err != nil {
		return math.Int{}, err
	}
	totalLicenses := math.ZeroInt()
	for _, delegation := range delegations {
		tokens := validator.TokensFromShares(delegation.Shares).TruncateInt()
		totalLicenses = totalLicenses.Add(licenseUnits(tokens, validator.DelegationIncrement))
	}

	ubds, err := k.GetUnbondingDelegationsFromValidator(ctx, valAddr)
	if err != nil {
		return math.Int{}, err
	}
	for _, ubd := range ubds {
		for _, entry := range ubd.Entries {
			totalLicenses = totalLicenses.Add(licenseUnits(entry.Balance, validator.DelegationIncrement))
		}
	}
	return totalLicenses, nil
}

// calculateUndelegateLicenseCount returns how many licenses an undelegation of
// amount releases. Removing the whole delegation releases all its licenses and
// is never blocked by increment rounding (e.g. after a slash) so a delegator
// can always fully exit. A partial undelegation must be an exact multiple of
// the increment and keep at least MinDelegation bonded.
func (k msgServer) calculateUndelegateLicenseCount(ctx context.Context, amount sdk.Coin, validator types.Validator, delegatorAddress sdk.AccAddress) (math.Int, error) {
	valAddr, err := k.validatorAddressCodec.StringToBytes(validator.GetOperator())
	if err != nil {
		return math.Int{}, sdkerrors.ErrInvalidAddress.Wrapf("invalid validator address: %s", err)
	}
	delegation, existed := k.GetExistingDelegation(ctx, delegatorAddress, valAddr)
	if !existed {
		return math.Int{}, types.ErrNoDelegation
	}

	increment := math.OneInt()
	if !validator.DelegationIncrement.IsNil() {
		increment = validator.DelegationIncrement
	}

	tokens := validator.TokensFromShares(delegation.Shares).TruncateInt()

	// full exit: release everything this delegation holds
	if amount.Amount.GTE(tokens) {
		return tokens.Quo(increment), nil
	}

	// partial: whole increments only, and the remaining delegation must not
	// fall below the entry minimum
	if amount.Amount.Mod(increment).GT(math.ZeroInt()) {
		return math.Int{}, types.ErrInvalidIncrementDelegation
	}
	if !validator.MinDelegation.IsNil() && tokens.Sub(amount.Amount).LT(validator.MinDelegation) {
		return math.Int{}, types.ErrDelegationBelowMinimum
	}
	return amount.Amount.Quo(increment), nil
}
