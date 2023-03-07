package cli

import (
	flag "github.com/spf13/pflag"

	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

const (
	FlagAddressApprover    = "approver"
	FlagAddressNewApprover = "new-approver"
	FlagApprovalEnabled    = "approval-enabled"

	FlagAddressValidator    = "validator"
	FlagAddressValidatorSrc = "addr-validator-source"
	FlagAddressValidatorDst = "addr-validator-dest"
	FlagPubKey              = "pubkey"
	FlagAmount              = "amount"
	FlagSharesAmount        = "shares-amount"
	FlagSharesFraction      = "shares-fraction"

	FlagMoniker         = "moniker"
	FlagEditMoniker     = "new-moniker"
	FlagIdentity        = "identity"
	FlagWebsite         = "website"
	FlagSecurityContact = "security-contact"
	FlagDetails         = "details"

	FlagCommissionRate          = "commission-rate"
	FlagCommissionMaxRate       = "commission-max-rate"
	FlagCommissionMaxChangeRate = "commission-max-change-rate"

	FlagMinSelfDelegation = "min-self-delegation"

	FlagGenesisFormat = "genesis-format"
	FlagNodeID        = "node-id"
	FlagIP            = "ip"

	// Flag for custom validator
	FlagMinDelegation       = "min-delegation"
	FlagDelegationIncrement = "delegation-increment"
	FlagLicenseMode         = "license-mode"
	FlagMaxLicense          = "max-license"
	FlagEnableRedelegation  = "enable-redelegation"
)

// common flagsets to add to various functions
var (
	fsShares       = flag.NewFlagSet("", flag.ContinueOnError)
	fsValidator    = flag.NewFlagSet("", flag.ContinueOnError)
	fsRedelegation = flag.NewFlagSet("", flag.ContinueOnError)
)

func init() {
	fsShares.String(FlagSharesAmount, "", "Amount of source-shares to either unbond or redelegate as a positive integer or decimal")
	fsShares.String(FlagSharesFraction, "", "Fraction of source-shares to either unbond or redelegate as a positive integer or decimal >0 and <=1")
	fsValidator.String(FlagAddressValidator, "", "The Bech32 address of the validator")
	fsRedelegation.String(FlagAddressValidatorSrc, "", "The Bech32 address of the source validator")
	fsRedelegation.String(FlagAddressValidatorDst, "", "The Bech32 address of the destination validator")
}

func FlagSetApprover() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.String(FlagAddressApprover, "", "Approver for create validator")
	return fs
}

func FlagSetNewApprover() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.String(FlagAddressNewApprover, "", "New approver for create validator")
	return fs
}

func FlagSetApprovalEnabled() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.Bool(FlagApprovalEnabled, true, "Enable approval for create validator")
	return fs
}

// FlagMinDelegation       = "min-delegation"
func FlagMinDelegationCreate() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.String(FlagMinDelegation, "", "The minimum delegation")
	return fs
}

// FlagDelegationIncrement = "delegation-increment"
func FlagDelegationIncrementCreate() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.String(FlagDelegationIncrement, "", "The delegation imcrement")
	return fs
}

// FlagLicenseMode         = "license-mode"
// FlagMaxLicense          = "max-license"
func FlagLicenseModeCreate() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.Bool(FlagLicenseMode, false, "License mode or not")
	fs.String(FlagMaxLicense, "", "The maximum license when license mode is on")

	return fs
}
func FlagMaxLicenseEdit() *flag.FlagSet {

	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.String(FlagMaxLicense, "", "The max license should set to current or gather current")
	return fs
}

// FlagEnableRedelegation  = "enable-redelegation"
func FlagEnableRedelegationCreate() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.Bool(FlagEnableRedelegation, true, "To enable redelegation for this validator (default true)")

	return fs
}

// FlagSetCommissionCreate Returns the FlagSet used for commission create.
func FlagSetCommissionCreate() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagCommissionRate, "", "The initial commission rate percentage")
	fs.String(FlagCommissionMaxRate, "", "The maximum commission rate percentage")
	fs.String(FlagCommissionMaxChangeRate, "", "The maximum commission change rate percentage (per day)")

	return fs
}

// FlagSetMinSelfDelegation Returns the FlagSet used for minimum set delegation.
func FlagSetMinSelfDelegation() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.String(FlagMinSelfDelegation, "", "The minimum self delegation required on the validator")
	return fs
}

// FlagSetAmount Returns the FlagSet for amount related operations.
func FlagSetAmount() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.String(FlagAmount, "", "Amount of coins to bond")
	return fs
}

// FlagSetPublicKey Returns the flagset for Public Key related operations.
func FlagSetPublicKey() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.String(FlagPubKey, "", "The validator's Protobuf JSON encoded public key")
	return fs
}

func flagSetDescriptionEdit() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagEditMoniker, types.DoNotModifyDesc, "The validator's name")
	fs.String(FlagIdentity, types.DoNotModifyDesc, "The (optional) identity signature (ex. UPort or Keybase)")
	fs.String(FlagWebsite, types.DoNotModifyDesc, "The validator's (optional) website")
	fs.String(FlagSecurityContact, types.DoNotModifyDesc, "The validator's (optional) security contact email")
	fs.String(FlagDetails, types.DoNotModifyDesc, "The validator's (optional) details")

	return fs
}

func flagSetCommissionUpdate() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagCommissionRate, "", "The new commission rate percentage")

	return fs
}

func flagSetDescriptionCreate() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagMoniker, "", "The validator's name")
	fs.String(FlagIdentity, "", "The optional identity signature (ex. UPort or Keybase)")
	fs.String(FlagWebsite, "", "The validator's (optional) website")
	fs.String(FlagSecurityContact, "", "The validator's (optional) security contact email")
	fs.String(FlagDetails, "", "The validator's (optional) details")

	return fs
}
