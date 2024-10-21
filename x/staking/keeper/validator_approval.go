package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// get a single validator
func (k Keeper) GetValidatorApproval(ctx context.Context) (validatorApproval types.ValidatorApproval, found bool) {
	store := k.storeService.OpenKVStore(ctx)

	value, err := store.Get(types.ValidatorApprovalKey)
	if err != nil {
		return validatorApproval, false
	}

	if value == nil {
		return validatorApproval, false
	}

	validatorApproval = types.MustUnmarshalValidatorApproval(k.cdc, value)
	return validatorApproval, true
}

// set the main record holding validator details
func (k Keeper) SetNewValidatorApprovalState(ctx context.Context, validatorApproval types.ValidatorApproval) {
	store := k.storeService.OpenKVStore(ctx)
	bz := types.MustMarshalValidatorApproval(k.cdc, &validatorApproval)
	store.Set(types.ValidatorApprovalKey, bz)
}
