package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// SetWhitelistDelegator set a specific whitelistDelegator in the store from its index
func (k Keeper) SetWhitelistDelegator(ctx sdk.Context, whitelistDelegator types.WhitelistDelegator) {
	store := k.storeService.OpenKVStore(ctx)
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
) (val types.WhitelistDelegator, err error) {
	store := k.storeService.OpenKVStore(ctx)
	value, err := store.Get(types.WhitelistDelegatorKey(
		validator,
	))
	if err != nil {
		return val, err
	}

	if value == nil {
		return val, types.ErrNoWhiltelistFound
	}

	k.cdc.MustUnmarshal(value, &val)
	return val, nil
}

// RemoveWhitelistDelegator removes a whitelistDelegator from the store
func (k Keeper) RemoveWhitelistDelegator(ctx sdk.Context, validator sdk.ValAddress) {
	store := k.storeService.OpenKVStore(ctx)
	store.Delete(types.WhitelistDelegatorKey(
		validator,
	))
}

// GetAllWhitelistDelegator returns all whitelistDelegator
func (k Keeper) GetAllWhitelistDelegator(ctx sdk.Context) (list []types.WhitelistDelegator, err error) {
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

func (k Keeper) IsSpecialDelegator(ctx context.Context, val sdk.ValAddress, delegator sdk.AccAddress) (found bool) {
	// chekc if delegator is validator itself then return true
	// if not then validator must add specific delegator to whitelist
	if val.Equals(delegator) {
		return true
	}

	specialList, err := k.GetWhitelistDelegator(ctx, val)
	if err != nil {
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

	specialList, err := k.GetWhitelistDelegator(ctx, validator)
	if err != nil {
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
