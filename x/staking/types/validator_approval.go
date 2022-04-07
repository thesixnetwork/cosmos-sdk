package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// return the redelegation
func MustMarshalValidatorApproval(cdc codec.BinaryCodec, validatorApproval *ValidatorApproval) []byte {
	return cdc.MustMarshal(validatorApproval)
}

// unmarshal a redelegation from a store value
func MustUnmarshalValidatorApproval(cdc codec.BinaryCodec, value []byte) ValidatorApproval {
	validatorApproval, err := UnmarshalValidatorApproval(cdc, value)
	if err != nil {
		panic(err)
	}

	return validatorApproval
}

// unmarshal a redelegation from a store value
func UnmarshalValidatorApproval(cdc codec.BinaryCodec, value []byte) (v ValidatorApproval, err error) {
	err = cdc.Unmarshal(value, &v)
	return v, err
}
