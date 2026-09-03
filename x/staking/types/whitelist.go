package types

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ binary.ByteOrder

const (
	WhitelistDelegatorKeyPrefix = "WhitelistDelegator/value/"
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

func WhitelistDelegatorKey(
	validator sdk.ValAddress,
) []byte {
	// the key must live under WhitelistDelegatorKeyPrefix: GetAllWhitelistDelegator,
	// the whitelist query, and genesis export all iterate that prefix, and a raw
	// address key could collide with the module's binary-prefixed key space
	key := KeyPrefix(WhitelistDelegatorKeyPrefix)
	key = append(key, []byte(validator)...)
	key = append(key, []byte("/")...)

	return key
}
