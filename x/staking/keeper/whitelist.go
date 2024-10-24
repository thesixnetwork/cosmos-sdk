package keeper

import (
	"context"

	"cosmossdk.io/store/prefix"

	errorsmod "cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// SetWhitelistDelegator set a specific whitelistDelegator in the store from its index
func (k Keeper) SetWhitelistDelegator(ctx context.Context, whitelistDelegator types.WhitelistDelegator) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.WhitelistDelegatorKeyPrefix))
	b := k.cdc.MustMarshal(&whitelistDelegator)
	valAddr, _ := sdk.ValAddressFromBech32(whitelistDelegator.ValidatorAddress)
	store.Set(types.WhitelistDelegatorKey(
		valAddr,
	), b)
}

// GetWhitelistDelegator returns a whitelistDelegator from its index
func (k Keeper) GetWhitelistDelegator(
	ctx context.Context,
	validator sdk.ValAddress,
) (val types.WhitelistDelegator, found bool) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.WhitelistDelegatorKeyPrefix))
	b := store.Get(types.WhitelistDelegatorKey(
		validator,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveWhitelistDelegator removes a whitelistDelegator from the store
func (k Keeper) _(ctx sdk.Context, validator sdk.ValAddress) {
	store := k.storeService.OpenKVStore(ctx)
	store.Delete(types.WhitelistDelegatorKey(
		validator,
	))
}

// GetAllWhitelistDelegator returns all whitelistDelegator
func (k Keeper) GetAllWhitelistDelegator(ctx context.Context) (list []types.WhitelistDelegator) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.WhitelistDelegatorKeyPrefix))
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})

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
	if val.Equals(delegator) {
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

func (k Keeper) DelDelegatorFromWhitelist(ctx sdk.Context, validator sdk.ValAddress, delegator string) (*types.MsgDeleteWhitelistdelegatorResponse, error) {
	specialList, found := k.GetWhitelistDelegator(ctx, validator)
	if !found {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "validator whitelist delegator doesn't exist")
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

	return &types.MsgDeleteWhitelistdelegatorResponse{WhitelistDelegator: &specialList}, nil
}
