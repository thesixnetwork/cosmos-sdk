package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// SetWhitelistDelegator set a specific whitelistDelegator in the store from its index
func (k Keeper) SetWhitelistDelegator(ctx sdk.Context, whitelistDelegator types.WhitelistDelegator) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.WhitelistDelegatorKeyPrefix))
	b := k.cdc.MustMarshal(&whitelistDelegator)
	store.Set(types.WhitelistDelegatorKey(
		sdk.ValAddress(whitelistDelegator.ValidatorAddress),
	), b)
}

// GetWhitelistDelegator returns a whitelistDelegator from its index
func (k Keeper) GetWhitelistDelegator(
	ctx sdk.Context,
	validator sdk.ValAddress,
) (val types.WhitelistDelegator, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.WhitelistDelegatorKeyPrefix))

	b := store.Get(types.WhitelistDelegatorKey(
		sdk.ValAddress(validator),
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveWhitelistDelegator removes a whitelistDelegator from the store
func (k Keeper) RemoveWhitelistDelegator(
	ctx sdk.Context,
	validator sdk.ValAddress,
) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.WhitelistDelegatorKeyPrefix))
	store.Delete(types.WhitelistDelegatorKey(
		validator,
	))
}

// GetAllWhitelistDelegator returns all whitelistDelegator
func (k Keeper) GetAllWhitelistDelegator(ctx sdk.Context) (list []types.WhitelistDelegator) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.WhitelistDelegatorKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.WhitelistDelegator
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

func (k Keeper) IsSpecialDelegator(ctx sdk.Context, val sdk.ValAddress, delegator sdk.AccAddress) (found bool) {
	// chekc if delegator is validator itself then return true
	// if not then validator must add specific delegator to whitelist
	if val.Equals(delegator){
		return true
	}

	specialList, found := k.GetWhitelistDelegator(ctx, val)
	if !found {
		return false
	}

	for _, whiltelistAddress := range specialList.DelegatorAddress {
		whiltelistAddressBech32, _ := sdk.AccAddressFromBech32(whiltelistAddress)

		if whiltelistAddressBech32.Equals(delegator) {
			return true
		}
	}

	return false
}

func (k Keeper) DelDelegatorFromWhitelist(ctx sdk.Context, validator sdk.ValAddress, delegator string) (*types.MsgWhitelistDelegatorResponse, error) {

	specialList, found := k.GetWhitelistDelegator(ctx, validator)
	if !found {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "validator whitelist delegator doesn't exist")
	}

	for i, spDelegator := range specialList.DelegatorAddress {
		if spDelegator == delegator {
			specialList.DelegatorAddress = append(specialList.DelegatorAddress[:i], specialList.DelegatorAddress[i+1:]...)
			break
		}
	}

	k.SetWhitelistDelegator(ctx, types.WhitelistDelegator{
		ValidatorAddress: specialList.ValidatorAddress,
		DelegatorAddress: specialList.DelegatorAddress,
	})

	return &types.MsgWhitelistDelegatorResponse{WhitelistDelegator: &specialList}, nil
}
