package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v2types "github.com/cosmos/cosmos-sdk/x/staking/types/v2"
)

func (k Keeper) GetWhitelist(ctx sdk.Context, addr sdk.ValAddress) (whitelist v2types.WhitelistDelegator, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), v2types.KeyPrefix(v2types.WhiltelistKeyPrefix))

	value := store.Get(v2types.WhitelistKeyStore(addr))
	if value == nil {
		return whitelist, false
	}

	whitelist = v2types.MustUnmarshalWhitelist(k.cdc, value)
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
