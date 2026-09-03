package keeper

import (
	"context"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (k msgServer) CreateWhitelistdelegator(goCtx context.Context, msg *types.MsgCreateWhitelistDelegator) (*types.MsgCreateWhitelistdelegatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// validate basic
	valAddr, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid validator address (%s)", err)
	}
	_, err = sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid delegator address (%s)", err)
	}

	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	if !creator.Equals(sdk.AccAddress(valAddr)) {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "only the validator operator can modify its delegator whitelist")
	}

	whitelist, err := k.GetWhitelistDelegator(ctx, valAddr)
	if errors.Is(err, types.ErrNoWhiltelistFound) {
		whitelist = types.WhitelistDelegator{
			ValidatorAddress: msg.ValidatorAddress,
			DelegatorAddress: []string{},
		}
	}

	// check duplicate 
	for _, whitelist := range whitelist.DelegatorAddress {
		if whitelist == msg.DelegatorAddress{
			return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "duplicate delegator address (%s)", err)
		}
	}

	// append value to key store
	whitelist.DelegatorAddress = append(whitelist.DelegatorAddress, msg.DelegatorAddress)

	k.SetWhitelistDelegator(ctx, types.WhitelistDelegator{
		ValidatorAddress: whitelist.ValidatorAddress,
		DelegatorAddress: whitelist.DelegatorAddress,
	})

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeMsgCreateWhitelistDelegator,
			sdk.NewAttribute(types.AttributeKeyValidator, msg.ValidatorAddress),
			sdk.NewAttribute(types.AttributeKeyDelegator, msg.DelegatorAddress),
		),
	})

	return &types.MsgCreateWhitelistdelegatorResponse{WhitelistDelegator: &whitelist}, nil
}


// DeleteWhitelistdelegator implements types.MsgServer.
func (k msgServer) DeleteWhitelistdelegator(goCtx context.Context, msg *types.MsgDeleteWhitelistDelegator) (*types.MsgDeleteWhitelistdelegatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// validate basic
	validatorAddr, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid validator address (%s)", err)
	}

	valdatorOp, err := k.Validator(ctx, validatorAddr)
	if err != nil {
		return nil, err
	}
	if valdatorOp == nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrorInvalidSigner, "Validator is not operate (%s)", err)
	}

	_, err = sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid delegator address (%s)", err)
	}

	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	if !creator.Equals(sdk.AccAddress(validatorAddr)) {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "only the validator operator can modify its delegator whitelist")
	}

	whitelist, err := k.DelDelegatorFromWhitelist(ctx, validatorAddr, msg.DelegatorAddress)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeMsgDeleteWhitelistDelegator,
			sdk.NewAttribute(types.AttributeKeyValidator, msg.ValidatorAddress),
			sdk.NewAttribute(types.AttributeKeyDelegator, msg.DelegatorAddress),
		),
	})

	return whitelist, nil
}