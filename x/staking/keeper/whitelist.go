package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
 	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (k Keeper) GetWhitelist(ctx sdk.Context, addr sdk.ValAddress) (whitelist types.WhitelistDelegator, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.WhiltelistKeyPrefix))

	value := store.Get(types.WhitelistKeyStore(addr))
	if value == nil {
		return whitelist, false
	}

	whitelist = types.MustUnmarshalWhitelist(k.cdc, value)
	return whitelist, true
}

func (k Keeper) IsSpecialDelegator(ctx sdk.Context, val sdk.ValAddress, delegator sdk.AccAddress) (found bool) {
	specialList, found := k.GetWhitelist(ctx, val)
	if !found {
		return false
	}

	for _, delegatorList := range specialList.Delegators {
		if delegatorList.DelegatorAddress == delegator.String() {
			return true
		}
	}

	return false
}
