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
	var key []byte

	validatorBytes := []byte(validator)
	key = append(key, validatorBytes...)
	key = append(key, []byte("/")...)

	return key
}
