package v2

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	WhiltelistKeyPrefix = "Whitelist/value/"
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

func WhitelistKeyStore(
	valAddr sdk.ValAddress,
) []byte {
	var key []byte

	valAddrBytes := []byte(valAddr)
	key = append(key, valAddrBytes...)
	key = append(key, []byte("/")...)

	return key
}

// unmarshal a redelegation from a store value
func MustUnmarshalWhitelist(cdc codec.BinaryCodec, value []byte) WhitelistDelegator {
	whitelist, err := UnmarshalWhitelist(cdc, value)
	if err != nil {
		panic(err)
	}

	return whitelist
}


// unmarshal a redelegation from a store value
func UnmarshalWhitelist(cdc codec.BinaryCodec, value []byte) (v WhitelistDelegator, err error) {
	err = cdc.Unmarshal(value, &v)
	return v, err
}