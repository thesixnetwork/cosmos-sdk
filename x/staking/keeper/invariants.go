package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// Upstream removed the standard staking invariants (and x/crisis) in v0.53;
// only the six-network license invariant is kept here. The module manager no
// longer calls RegisterInvariants, so wire this from the app if periodic
// checking is wanted.

// RegisterInvariants registers the six-network staking invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k *Keeper) {
	ir.RegisterRoute(types.ModuleName, "license-count",
		LicenseCountInvariant(k))
}

// AllInvariants runs all invariants of the staking module.
func AllInvariants(k *Keeper) sdk.Invariant {
	return LicenseCountInvariant(k)
}

// LicenseCountInvariant checks that no license-mode validator uses more
// licenses than its owner-set MaxLicense. MaxLicense is only ever written from
// MsgCreateValidator/MsgEditValidator, so a break here means some delegation
// path skipped the cap check.
func LicenseCountInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var (
			msg    string
			broken bool
		)

		validators, err := k.GetAllValidators(ctx)
		if err != nil {
			panic(err)
		}

		for _, validator := range validators {
			if validator.Mode != types.ValidatorMode_MODE_LICENSE {
				continue
			}
			if validator.MaxLicense.IsNil() || validator.LicenseCount.IsNil() {
				continue
			}
			if validator.LicenseCount.GT(validator.MaxLicense) {
				broken = true
				msg += fmt.Sprintf("\tvalidator %s license count %s exceeds max license %s\n",
					validator.GetOperator(), validator.LicenseCount, validator.MaxLicense)
			}
		}

		return sdk.FormatInvariant(types.ModuleName, "license count",
			fmt.Sprintf("license-mode validators over their max license\n%s", msg)), broken
	}
}
