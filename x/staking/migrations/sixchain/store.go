package sixchain

import (
	"bytes"

	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// OldWhitelistDelegatorKey reproduces the pre-fix key layout used by earlier
// six-network binaries: the raw validator-address bytes followed by "/", with
// no module-owned prefix. It exists so an upgrade (or a test) can locate the
// stale entries this migration re-keys.
func OldWhitelistDelegatorKey(validator sdk.ValAddress) []byte {
	key := make([]byte, 0, len(validator)+1)
	key = append(key, validator...)
	key = append(key, '/')
	return key
}

// MigrateStore re-keys whitelist delegator entries written by earlier
// six-network binaries under OldWhitelistDelegatorKey into the
// WhitelistDelegatorKeyPrefix namespace, where every reader (GetWhitelistDelegator,
// IsSpecialDelegator, GetAllWhitelistDelegator, the whitelist query and genesis
// export) now expects them. Without it, previously whitelisted delegators of
// fast-mode validators silently lose their whitelist status on upgrade.
//
// This is intentionally NOT tied to the staking ConsensusVersion /
// RegisterMigration machinery — trigger it once, explicitly, from the app
// upgrade handler. Either call the convenience keeper method:
//
//	err := app.StakingKeeper.MigrateWhitelistDelegatorKeys(ctx)
//
// or invoke this function directly with the staking store:
//
//	store := ctx.KVStore(app.GetKey(stakingtypes.StoreKey))
//	err := sixchain.MigrateStore(ctx, store, app.appCodec)
//
// It is idempotent: entries already living under the new prefix are longer than
// the old 21-byte key and are left untouched, so a second run is a no-op.
func MigrateStore(ctx sdk.Context, store storetypes.KVStore, cdc codec.BinaryCodec) error {
	type rekeyEntry struct {
		oldKey []byte
		newKey []byte
		value  []byte
	}
	var entries []rekeyEntry

	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		// old key shape: 20 raw validator-address bytes followed by "/"
		if len(key) != 21 || key[20] != '/' {
			continue
		}

		// only treat it as a whitelist entry when the value decodes to a
		// WhitelistDelegator whose validator address matches the key bytes —
		// this rules out collisions with any other 21-byte staking key that
		// happens to end in '/' (its value would not round-trip this way)
		var whitelist types.WhitelistDelegator
		if err := cdc.Unmarshal(iterator.Value(), &whitelist); err != nil {
			continue
		}
		valAddr, err := sdk.ValAddressFromBech32(whitelist.ValidatorAddress)
		if err != nil || !bytes.Equal(valAddr, key[:20]) {
			continue
		}

		entries = append(entries, rekeyEntry{
			oldKey: append([]byte(nil), key...),
			newKey: types.WhitelistDelegatorKey(valAddr),
			value:  append([]byte(nil), iterator.Value()...),
		})
	}

	for _, e := range entries {
		store.Delete(e.oldKey)
		store.Set(e.newKey, e.value)
	}

	if len(entries) > 0 {
		ctx.Logger().Info("re-keyed whitelist delegator entries", "count", len(entries), "module", types.ModuleName)
	}

	return nil
}
