package types

import (
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
		Creator: creator,
		ValidatorAddress: validator,
		DelegatorAddress: delegator,
	}
}

func (msg *MsgCreateWhitelistDelegator) Route() string {
	return RouterKey
}

func (msg *MsgCreateWhitelistDelegator) Type() string {
	return TypeMsgCreateWhitelistDelegator
}

func (msg *MsgCreateWhitelistDelegator) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}

	valAddr, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		panic(err)
	}

	valAccAdd := sdk.AccAddress(valAddr.Bytes())

	if !creator.Equals(valAccAdd) {
		panic("Signer and validator must be the same person")
	}

	return []sdk.AccAddress{creator}
}

func (msg *MsgCreateWhitelistDelegator) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgCreateWhitelistDelegator) ValidateBasic() error {
	_, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid validator address (%s)", err)
	}

	_, err = sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid delegator address (%s)", err)
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
		Creator: creator,
		ValidatorAddress: validator,
		DelegatorAddress: delegator,
	}
}
func (msg *MsgDeleteWhitelistDelegator) Route() string {
	return RouterKey
}

func (msg *MsgDeleteWhitelistDelegator) Type() string {
	return TypeMsgDeleteWhitelistDelegator
}

func (msg *MsgDeleteWhitelistDelegator) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}

	valAddr, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		panic(err)
	}

	valAccAdd := sdk.AccAddress(valAddr.Bytes())

	if !creator.Equals(valAccAdd) {
		panic("Signer and validator must be the same person")
	}

	return []sdk.AccAddress{creator}
}

func (msg *MsgDeleteWhitelistDelegator) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgDeleteWhitelistDelegator) ValidateBasic() error {
	_, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid validator address (%s)", err)
	}

	_, err = sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid delegator address (%s)", err)
	}
	return nil
}