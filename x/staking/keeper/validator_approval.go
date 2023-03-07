package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// get a single validator
func (k Keeper) GetValidatorApproval(ctx sdk.Context) (validatorApproval types.ValidatorApproval, found bool) {
	store := ctx.KVStore(k.storeKey)

	value := store.Get(types.ValidatorApprovalKey)
	if value == nil {
		return validatorApproval, false
	}

	validatorApproval = types.MustUnmarshalValidatorApproval(k.cdc, value)
	return validatorApproval, true
}

// set the main record holding validator details
func (k Keeper) SetNewValidatorApprovalState(ctx sdk.Context, validatorApproval types.ValidatorApproval) {
	store := ctx.KVStore(k.storeKey)
	bz := types.MustMarshalValidatorApproval(k.cdc, &validatorApproval)
	store.Set(types.ValidatorApprovalKey, bz)
}
