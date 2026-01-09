package testutil

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// NewValidator is a testing helper method to create validators in tests
func NewValidator(tb testing.TB, operator sdk.ValAddress, pubKey cryptotypes.PubKey) types.Validator {
	tb.Helper()
	v, err := types.NewValidatorNormalMode(operator.String(), pubKey, types.Description{})
	require.NoError(tb, err)
	return v
}

func NewValidatorLicenseMode(tb testing.TB, operator sdk.ValAddress, pubKey cryptotypes.PubKey, minDelegation, delegationIncrement, maxLicense math.Int) types.Validator {
	tb.Helper()
	v, err := types.NewValidatorLicenseMode(operator.String(), pubKey, types.Description{}, minDelegation, delegationIncrement, maxLicense)
	require.NoError(tb, err)
	return v
}
