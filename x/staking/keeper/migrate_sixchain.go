package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/migrations/sixchain"
)

// MigrateWhitelistDelegatorKeys re-keys whitelist delegator entries written by
// earlier six-network binaries under the pre-fix raw keys into the current
// prefixed namespace. It is deliberately NOT registered against the module
// ConsensusVersion; call it once from the app upgrade handler:
//
//	app.UpgradeKeeper.SetUpgradeHandler("<name>", func(ctx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
//		if err := app.StakingKeeper.MigrateWhitelistDelegatorKeys(ctx); err != nil {
//			return nil, err
//		}
//		return app.ModuleManager.RunMigrations(ctx, app.configurator, vm)
//	})
func (k Keeper) MigrateWhitelistDelegatorKeys(ctx context.Context) error {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	return sixchain.MigrateStore(sdk.UnwrapSDKContext(ctx), store, k.cdc)
}
