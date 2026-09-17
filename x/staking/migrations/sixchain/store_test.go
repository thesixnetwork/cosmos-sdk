package sixchain_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/sims"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/staking/migrations/sixchain"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestMigrateWhitelistKeys(t *testing.T) {
	cdc := moduletestutil.MakeTestEncodingConfig(staking.AppModuleBasic{}).Codec
	storeKey := storetypes.NewKVStoreKey(stakingtypes.StoreKey)
	tKey := storetypes.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)
	store := ctx.KVStore(storeKey)

	accAddrs := sims.CreateIncrementalAccounts(4)
	valAddrs := sims.ConvertAddrsToValAddrs(accAddrs)

	// two whitelist entries written in the old raw-key format
	wl0 := stakingtypes.WhitelistDelegator{
		ValidatorAddress: valAddrs[0].String(),
		DelegatorAddress: []string{accAddrs[1].String(), accAddrs[2].String()},
	}
	wl1 := stakingtypes.WhitelistDelegator{
		ValidatorAddress: valAddrs[1].String(),
		DelegatorAddress: []string{accAddrs[3].String()},
	}
	store.Set(sixchain.OldWhitelistDelegatorKey(valAddrs[0]), cdc.MustMarshal(&wl0))
	store.Set(sixchain.OldWhitelistDelegatorKey(valAddrs[1]), cdc.MustMarshal(&wl1))

	// a decoy 21-byte key ending in '/' whose value does not round-trip as a
	// whitelist entry for that key must be left untouched (guard rejects it):
	// here the stored entry names a different validator than the key bytes
	decoyKey := sixchain.OldWhitelistDelegatorKey(valAddrs[2])
	require.Len(t, decoyKey, 21)
	decoyVal := cdc.MustMarshal(&stakingtypes.WhitelistDelegator{
		ValidatorAddress: valAddrs[3].String(), // != key bytes (valAddrs[2])
		DelegatorAddress: []string{accAddrs[0].String()},
	})
	store.Set(decoyKey, decoyVal)

	require.NoError(t, sixchain.MigrateStore(ctx, store, cdc))

	// old whitelist keys are gone, new prefixed keys hold the same values
	for i, wl := range []stakingtypes.WhitelistDelegator{wl0, wl1} {
		require.Nil(t, store.Get(sixchain.OldWhitelistDelegatorKey(valAddrs[i])), "old key %d should be deleted", i)

		newKey := stakingtypes.WhitelistDelegatorKey(valAddrs[i])
		bz := store.Get(newKey)
		require.NotNil(t, bz, "new key %d should exist", i)
		var got stakingtypes.WhitelistDelegator
		cdc.MustUnmarshal(bz, &got)
		require.Equal(t, wl, got)
	}

	// the decoy key (mismatched validator) is undisturbed
	require.Equal(t, decoyVal, store.Get(decoyKey))

	// re-running is a no-op (idempotent)
	require.NoError(t, sixchain.MigrateStore(ctx, store, cdc))
	require.NotNil(t, store.Get(stakingtypes.WhitelistDelegatorKey(valAddrs[0])))
	require.Equal(t, decoyVal, store.Get(decoyKey))
}
