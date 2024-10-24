package v1

import "cosmossdk.io/math"

type DelegationI interface {
	GetDelegatorAddr() string  // delegator string for the bond
	GetValidatorAddr() string  // validator operator address
	GetShares() math.LegacyDec // amount of validator's shares held in this delegation
}
