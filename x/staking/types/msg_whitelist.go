package types

import (
	"cosmossdk.io/core/address"
	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	TypeMsgCreateWhitelistDelegator = "create_whitelist_delegator"
	TypeMsgUpdateWhitelistDelegator = "update_whitelist_delegator"
	TypeMsgDeleteWhitelistDelegator = "delete_whitelist_delegator"
)

var _ sdk.Msg = &MsgCreateWhitelistDelegator{}

func NewMsgCreateWhitelistDelegator(
	creator string,
	validator string,
	delegator string,
) *MsgCreateWhitelistDelegator {
	return &MsgCreateWhitelistDelegator{
		Creator:          creator,
		ValidatorAddress: validator,
		DelegatorAddress: delegator,
	}
}

// Validate validates the MsgCreateWhitelistDelegator sdk msg.
func (msg MsgCreateWhitelistDelegator) Validate(ac address.Codec) error {
	// note that unmarshaling from bech32 ensures both non-empty and valid
	_, err := ac.StringToBytes(msg.ValidatorAddress)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid validator address: %s", err)
	}

	_, err = sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid delegator address (%s)", err)
	}

	return nil
}

var _ sdk.Msg = &MsgDeleteWhitelistDelegator{}

func NewMsgDeleteWhitelistDelegator(
	creator string,
	validator string,
	delegator string,
) *MsgDeleteWhitelistDelegator {
	return &MsgDeleteWhitelistDelegator{
		Creator:          creator,
		ValidatorAddress: validator,
		DelegatorAddress: delegator,
	}
}

func (msg *MsgDeleteWhitelistDelegator) Validate(ac address.Codec) error {
	// note that unmarshaling from bech32 ensures both non-empty and valid
	_, err := ac.StringToBytes(msg.ValidatorAddress)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid validator address: %s", err)
	}

	_, err = sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid delegator address (%s)", err)
	}

	return nil
}
