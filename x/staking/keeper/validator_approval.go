package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// get a single validator
func (k Keeper) GetValidatorApproval(ctx context.Context) (validatorApproval types.ValidatorApproval, err error) {
	store := k.storeService.OpenKVStore(ctx)

	value, err := store.Get(types.ValidatorApprovalKey)
	if err != nil {
		return validatorApproval, err
	}

	if value == nil {
		return validatorApproval, types.ErrNoValidatorFound
	}

	validatorApproval = types.MustUnmarshalValidatorApproval(k.cdc, value)
	return validatorApproval, nil
}

// set the main record holding validator details
func (k Keeper) SetNewValidatorApprovalState(ctx context.Context, validatorApproval types.ValidatorApproval) error {
	store := k.storeService.OpenKVStore(ctx)
	bz := types.MustMarshalValidatorApproval(k.cdc, &validatorApproval)
	return store.Set(types.ValidatorApprovalKey, bz)
}
