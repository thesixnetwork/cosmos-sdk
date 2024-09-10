package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (k msgServer) CreateWhitelistdelegator(goCtx context.Context, msg *types.MsgCreateWhitelistDelegator) (*types.MsgWhitelistDelegatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// validate basic
	valAddr, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid validator address (%s)", err)
	}
	_, err = sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid delegator address (%s)", err)
	}

	whitelist, found := k.GetWhitelistDelegator(ctx, valAddr)
	if !found {
		whitelist = types.WhitelistDelegator{
			ValidatorAddress: msg.ValidatorAddress,
			DelegatorAddress: []string{},
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

	return &types.MsgWhitelistDelegatorResponse{WhitelistDelegator: &whitelist}, nil
}


// DeleteWhitelistdelegator implements types.MsgServer.
func (k msgServer) DeleteWhitelistdelegator(goCtx context.Context, msg *types.MsgDeleteWhitelistDelegator) (*types.MsgWhitelistDelegatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	
	// validate basic
	validatorAddr, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid validator address (%s)", err)
	}

	valdatorOp := k.Validator(ctx, validatorAddr)
	if valdatorOp == nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrorInvalidSigner, "Validator is not operate (%s)", err)
	}

	_, err = sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid delegator address (%s)", err)
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
