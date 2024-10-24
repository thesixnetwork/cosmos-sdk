package keeper

import (
	"context"

	"cosmossdk.io/store/prefix"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// DelegatorWhitelistdelegatorAll implements types.QueryServer.
func (k Querier) WhitelistdelegatorAll(c context.Context, req *types.QueryAllWhitelistDelegatorRequest) (*types.QueryWhitelistdelegatorAllResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var delegatorWhitelist []types.WhitelistDelegator
	ctx := sdk.UnwrapSDKContext(c)

	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	whitelistStore := prefix.NewStore(store, types.KeyPrefix(types.WhitelistDelegatorKeyPrefix))

	pageRes, err := query.Paginate(whitelistStore, req.Pagination, func(key []byte, value []byte) error {
		var wlDelegator types.WhitelistDelegator
		if err := k.cdc.Unmarshal(value, &wlDelegator); err != nil {
			return err
		}

		delegatorWhitelist = append(delegatorWhitelist, wlDelegator)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryWhitelistdelegatorAllResponse{WhitelistDelegator: delegatorWhitelist, Pagination: pageRes}, nil
}

// DelegatorWhitelistdelegator implements types.QueryServer.
func (k Querier) Whitelistdelegator(c context.Context, req *types.QueryGetWhitelistDelegatorRequest) (*types.QueryWhitelistdelegatorResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	valAddr, err := sdk.ValAddressFromBech32(req.Validator)
	if err != nil {
		return nil, err
	}

	val, found := k.GetWhitelistDelegator(
		ctx,
		valAddr,
	)

	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryWhitelistdelegatorResponse{WhitelistDelegator: val}, nil
}
