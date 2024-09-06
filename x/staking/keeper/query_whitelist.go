package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)


// DelegatorWhitelistdelegatorAll implements types.QueryServer.
func (k Querier) WhitelistdelegatorAll(c context.Context, req *types.QueryAllWhitelistDelegatorRequest) (*types.QueryAllWhitelistdelegatorResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var delegatorWhitelist []types.WhitelistDelegator
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
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

	return &types.QueryAllWhitelistdelegatorResponse{WhitelistDelegator: delegatorWhitelist, Pagination: pageRes}, nil
}

// DelegatorWhitelistdelegator implements types.QueryServer.
func (k Querier) Whitelistdelegator(c context.Context, req *types.QueryGetWhitelistDelegatorRequest) (*types.QueryGetWhitelistDelegatorResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	val, found := k.GetWhitelistDelegator(
		ctx,
		req.Validator,
	)

	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetWhitelistDelegatorResponse{WhitelistDelegator: val}, nil
}
