package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// default values
var (
	DefaultTokens                  = sdk.TokensFromConsensusPower(100, sdk.DefaultPowerReduction)
	defaultAmount                  = DefaultTokens.String() + sdk.DefaultBondDenom
	defaultCommissionRate          = "0.1"
	defaultCommissionMaxRate       = "0.2"
	defaultCommissionMaxChangeRate = "0.01"
	defaultMinSelfDelegation       = "1"
	defaultMinDelegation           = "1"
	defaultDelegationIncrement     = "1"
)

// NewTxCmd returns a root CLI command handler for all x/staking transaction commands.
func NewTxCmd() *cobra.Command {
	stakingTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Staking transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	stakingTxCmd.AddCommand(
		NewSetValidatorApprovalCmd(),
		NewCreateValidatorCmd(),
		NewEditValidatorCmd(),
		NewDelegateCmd(),
		NewRedelegateCmd(),
		NewUnbondCmd(),
		CmdCreateWhitelistDelegator(),
		CmdDeleteWhitelistDelegator(),
	)

	return stakingTxCmd
}

func NewSetValidatorApprovalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-validator-approval",
		Short: "set new validator approval state",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			approverAddr := clientCtx.GetFromAddress()
			newApproverAddr, _ := cmd.Flags().GetString(FlagAddressNewApprover)
			approvalEnabled, _ := cmd.Flags().GetBool(FlagApprovalEnabled)

			msg, err := types.NewMsgSetValidatorApproval(approverAddr.String(), newApproverAddr, approvalEnabled)
			if err != nil {
				return fmt.Errorf("error create message: %v", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().AddFlagSet(FlagSetNewApprover())
	cmd.Flags().AddFlagSet(FlagSetApprovalEnabled())
	flags.AddTxFlagsToCmd(cmd)

	_ = cmd.MarkFlagRequired(flags.FlagFrom)
	_ = cmd.MarkFlagRequired(FlagAddressNewApprover)
	_ = cmd.MarkFlagRequired(FlagApprovalEnabled)

	return cmd
}

func NewCreateValidatorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-validator",
		Short: "create new validator initialized with a self-delegation to it",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf := tx.NewFactoryCLI(clientCtx, cmd.Flags()).
				WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)
			txf, msg, err := newBuildCreateValidatorMsg(clientCtx, txf, cmd.Flags())
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	cmd.Flags().AddFlagSet(FlagSetPublicKey())
	cmd.Flags().AddFlagSet(FlagSetAmount())
	cmd.Flags().AddFlagSet(flagSetDescriptionCreate())
	cmd.Flags().AddFlagSet(FlagSetCommissionCreate())
	cmd.Flags().AddFlagSet(FlagSetMinSelfDelegation())
	cmd.Flags().AddFlagSet(FlagSetApprover())
	cmd.Flags().AddFlagSet(FlagMinDelegationCreate())
	cmd.Flags().AddFlagSet(FlagDelegationIncrementCreate())
	cmd.Flags().AddFlagSet(FlagLicenseModeCreate())
	cmd.Flags().AddFlagSet(FlagEnableRedelegationCreate())
	cmd.Flags().AddFlagSet(FlagSpecialModeCreate())

	cmd.Flags().String(FlagIP, "", fmt.Sprintf("The node's public IP. It takes effect only when used in combination with --%s", flags.FlagGenerateOnly))
	cmd.Flags().String(FlagNodeID, "", "The node's ID")
	flags.AddTxFlagsToCmd(cmd)

	_ = cmd.MarkFlagRequired(flags.FlagFrom)
	_ = cmd.MarkFlagRequired(FlagAmount)
	_ = cmd.MarkFlagRequired(FlagPubKey)
	_ = cmd.MarkFlagRequired(FlagMoniker)	

	return cmd
}

func NewEditValidatorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit-validator",
		Short: "edit an existing validator account",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			valAddr := clientCtx.GetFromAddress()
			moniker, _ := cmd.Flags().GetString(FlagEditMoniker)
			identity, _ := cmd.Flags().GetString(FlagIdentity)
			website, _ := cmd.Flags().GetString(FlagWebsite)
			security, _ := cmd.Flags().GetString(FlagSecurityContact)
			details, _ := cmd.Flags().GetString(FlagDetails)
			maxLicense, _ := cmd.Flags().GetString(FlagMaxLicense)
			licenceMode, _ := cmd.Flags().GetBool(FlagLicenseMode)
			specialMode, _ := cmd.Flags().GetBool(FlagSpecialMode)
			description := types.NewDescription(moniker, identity, website, security, details)

			var newRate *sdk.Dec

			commissionRate, _ := cmd.Flags().GetString(FlagCommissionRate)
			if commissionRate != "" {
				rate, err := sdk.NewDecFromStr(commissionRate)
				if err != nil {
					return fmt.Errorf("invalid new commission rate: %v", err)
				}

				newRate = &rate
			}

			var newMinSelfDelegation *sdk.Int

			minSelfDelegationString, _ := cmd.Flags().GetString(FlagMinSelfDelegation)
			if minSelfDelegationString != "" {
				msb, ok := sdk.NewIntFromString(minSelfDelegationString)
				if !ok {
					return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "minimum self delegation must be a positive integer")
				}

				newMinSelfDelegation = &msb
			}

			var newMaxLicense sdk.Int
			if maxLicense != "" {

				msb, ok := sdk.NewIntFromString(maxLicense)
				if !ok {
					return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "When license mode is used, max license is required and must be positive")
				}
				newMaxLicense = msb
			}

			msg := types.NewMsgEditValidator(sdk.ValAddress(valAddr), description, newRate, newMinSelfDelegation, newMaxLicense, licenceMode, specialMode)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().AddFlagSet(flagSetDescriptionEdit())
	cmd.Flags().AddFlagSet(flagSetCommissionUpdate())
	cmd.Flags().AddFlagSet(FlagSetMinSelfDelegation())
	// cmd.Flags().AddFlagSet(FlagMaxLicenseEdit())
	cmd.Flags().AddFlagSet(FlagLicenseModeEdit())
	cmd.Flags().AddFlagSet(FlagSpecialModeEdit())
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewDelegateCmd() *cobra.Command {
	bech32PrefixValAddr := sdk.GetConfig().GetBech32ValidatorAddrPrefix()

	cmd := &cobra.Command{
		Use:   "delegate [validator-addr] [amount]",
		Args:  cobra.ExactArgs(2),
		Short: "Delegate liquid tokens to a validator",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Delegate an amount of liquid coins to a validator from your wallet.

Example:
$ %s tx staking delegate %s1l2rsakp388kuv9k8qzq6lrm9taddae7fpx59wm 1000stake --from mykey
`,
				version.AppName, bech32PrefixValAddr,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			amount, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}

			delAddr := clientCtx.GetFromAddress()
			valAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			msg := types.NewMsgDelegate(delAddr, valAddr, amount)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewRedelegateCmd() *cobra.Command {
	bech32PrefixValAddr := sdk.GetConfig().GetBech32ValidatorAddrPrefix()

	cmd := &cobra.Command{
		Use:   "redelegate [src-validator-addr] [dst-validator-addr] [amount]",
		Short: "Redelegate illiquid tokens from one validator to another",
		Args:  cobra.ExactArgs(3),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Redelegate an amount of illiquid staking tokens from one validator to another.

Example:
$ %s tx staking redelegate %s1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj %s1l2rsakp388kuv9k8qzq6lrm9taddae7fpx59wm 100stake --from mykey
`,
				version.AppName, bech32PrefixValAddr, bech32PrefixValAddr,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			delAddr := clientCtx.GetFromAddress()
			valSrcAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			valDstAddr, err := sdk.ValAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			amount, err := sdk.ParseCoinNormalized(args[2])
			if err != nil {
				return err
			}

			msg := types.NewMsgBeginRedelegate(delAddr, valSrcAddr, valDstAddr, amount)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewUnbondCmd() *cobra.Command {
	bech32PrefixValAddr := sdk.GetConfig().GetBech32ValidatorAddrPrefix()

	cmd := &cobra.Command{
		Use:   "unbond [validator-addr] [amount]",
		Short: "Unbond shares from a validator",
		Args:  cobra.ExactArgs(2),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Unbond an amount of bonded shares from a validator.

Example:
$ %s tx staking unbond %s1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj 100stake --from mykey
`,
				version.AppName, bech32PrefixValAddr,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			delAddr := clientCtx.GetFromAddress()
			valAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			amount, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}

			msg := types.NewMsgUndelegate(delAddr, valAddr, amount)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func newBuildCreateValidatorMsg(clientCtx client.Context, txf tx.Factory, fs *flag.FlagSet) (tx.Factory, *types.MsgCreateValidator, error) {
	fAmount, _ := fs.GetString(FlagAmount)
	amount, err := sdk.ParseCoinNormalized(fAmount)
	if err != nil {
		return txf, nil, err
	}

	valAddr := clientCtx.GetFromAddress()
	pkStr, err := fs.GetString(FlagPubKey)
	if err != nil {
		return txf, nil, err
	}

	var pk cryptotypes.PubKey
	if err := clientCtx.Codec.UnmarshalInterfaceJSON([]byte(pkStr), &pk); err != nil {
		return txf, nil, err
	}

	approver, _ := fs.GetString(FlagAddressApprover)
	moniker, _ := fs.GetString(FlagMoniker)
	identity, _ := fs.GetString(FlagIdentity)
	website, _ := fs.GetString(FlagWebsite)
	security, _ := fs.GetString(FlagSecurityContact)
	details, _ := fs.GetString(FlagDetails)
	description := types.NewDescription(
		moniker,
		identity,
		website,
		security,
		details,
	)

	// get the initial validator commission parameters
	rateStr, _ := fs.GetString(FlagCommissionRate)
	maxRateStr, _ := fs.GetString(FlagCommissionMaxRate)
	maxChangeRateStr, _ := fs.GetString(FlagCommissionMaxChangeRate)

	commissionRates, err := buildCommissionRates(rateStr, maxRateStr, maxChangeRateStr)
	if err != nil {
		return txf, nil, err
	}

	// get the initial validator min self delegation
	msbStr, _ := fs.GetString(FlagMinSelfDelegation)

	minSelfDelegation, ok := sdk.NewIntFromString(msbStr)
	if !ok {
		return txf, nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "minimum self delegation must be a positive integer")
	}

	msg, err := types.NewMsgCreateValidator(
		sdk.ValAddress(valAddr), approver, pk, amount, description, commissionRates, minSelfDelegation,
	)
	if err != nil {
		return txf, nil, err
	}

	// Custom Validator
	// Min delegation
	mdStr, _ := fs.GetString(FlagMinDelegation)
	if mdStr != "" {
		minDelegation, ok := sdk.NewIntFromString(mdStr)
		if !ok {
			return txf, nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "minimum delegation must be a positive integer")
		}
		msg.MinDelegation = minDelegation
	}
	// delegation increment
	dincStr, _ := fs.GetString(FlagDelegationIncrement)
	if dincStr != "" {
		delegationIncrement, ok := sdk.NewIntFromString(dincStr)
		if !ok {
			return txf, nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "delegation increment must be a positive integer")
		}
		msg.DelegationIncrement = delegationIncrement
		if mdStr == "" {
			// if min delegation is not defined and increment is. Assign min delegation = increment
			msg.MinDelegation = delegationIncrement
		}
	}

	enableRedelegation, _ := fs.GetBool(FlagEnableRedelegation)

	// License
	licenseMode, _ := fs.GetBool(FlagLicenseMode)

	// Special mode
	specialMode, _ := fs.GetBool(FlagSpecialMode)

	switch{
	case licenseMode:
		msg.LicenseMode = true
		msg.SpecialMode = false
		mlcStr, _ := fs.GetString(FlagMaxLicense)
		maxLicense, ok := sdk.NewIntFromString(mlcStr)
		if !ok {
			return txf, nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "When license mode is used, max license is required and must be positive")
		}
		msg.MaxLicense = maxLicense

		if enableRedelegation {
			return txf, nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "When license mode is used, redelegation must be disabled")
		}

		if specialMode {
			return txf, nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "When license mode is used, special mode must be disabled")
		}

		// check Count licesene amount for validator
		divAmount := amount.Amount.Quo(msg.DelegationIncrement)
		modAmount := msg.Value.Amount.Mod(msg.DelegationIncrement)
		if modAmount.GT(sdk.ZeroInt()) {
			return txf, nil, types.ErrInvalidIncrementDelegation
		}
		if divAmount.GT(msg.MaxLicense) {
			return txf, nil, types.ErrNotEnoughLicense
		}
	case specialMode:
		msg.LicenseMode = false
		msg.SpecialMode = true
	default:
		msg.LicenseMode = false
		msg.SpecialMode = false
	}

	if(msg.LicenseMode && msg.SpecialMode){
		return txf, nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Cannot enable license mode along side with specialmode")
	}

	// Enable Redelegation
	msg.EnableRedelegation = enableRedelegation

	if err := msg.ValidateBasic(); err != nil {
		return txf, nil, err
	}

	genOnly, _ := fs.GetBool(flags.FlagGenerateOnly)
	if genOnly {
		ip, _ := fs.GetString(FlagIP)
		nodeID, _ := fs.GetString(FlagNodeID)

		if nodeID != "" && ip != "" {
			txf = txf.WithMemo(fmt.Sprintf("%s@%s:26656", nodeID, ip))
		}
	}

	return txf, msg, nil
}

// Return the flagset, particular flags, and a description of defaults
// this is anticipated to be used with the gen-tx
func CreateValidatorMsgFlagSet(ipDefault string) (fs *flag.FlagSet, defaultsDesc string) {
	fsCreateValidator := flag.NewFlagSet("", flag.ContinueOnError)
	fsCreateValidator.String(FlagIP, ipDefault, "The node's public IP")
	fsCreateValidator.String(FlagNodeID, "", "The node's NodeID")
	fsCreateValidator.String(FlagMoniker, "", "The validator's (optional) moniker")
	fsCreateValidator.String(FlagWebsite, "", "The validator's (optional) website")
	fsCreateValidator.String(FlagSecurityContact, "", "The validator's (optional) security contact email")
	fsCreateValidator.String(FlagDetails, "", "The validator's (optional) details")
	fsCreateValidator.String(FlagIdentity, "", "The (optional) identity signature (ex. UPort or Keybase)")
	fsCreateValidator.AddFlagSet(FlagSetCommissionCreate())
	fsCreateValidator.AddFlagSet(FlagSetMinSelfDelegation())
	fsCreateValidator.AddFlagSet(FlagSetAmount())
	fsCreateValidator.AddFlagSet(FlagSetPublicKey())
	fsCreateValidator.AddFlagSet(FlagMinDelegationCreate())
	fsCreateValidator.AddFlagSet(FlagDelegationIncrementCreate())
	fsCreateValidator.AddFlagSet(FlagLicenseModeCreate())
	fsCreateValidator.AddFlagSet(FlagEnableRedelegationCreate())
	fsCreateValidator.AddFlagSet(FlagSpecialModeCreate())

	defaultsDesc = fmt.Sprintf(`
	delegation amount:           %s
	commission rate:             %s
	commission max rate:         %s
	commission max change rate:  %s
	minimum self delegation:     %s
	minimum delegation:     %s
	delegation increment:     %s
`, defaultAmount, defaultCommissionRate,
		defaultCommissionMaxRate, defaultCommissionMaxChangeRate,
		defaultMinSelfDelegation,
		defaultMinDelegation,
		defaultDelegationIncrement)

	return fsCreateValidator, defaultsDesc
}

type TxCreateValidatorConfig struct {
	ChainID string
	NodeID  string
	Moniker string

	Amount string

	CommissionRate          string
	CommissionMaxRate       string
	CommissionMaxChangeRate string
	MinSelfDelegation       string
	MinDelegation           string
	DelegationIncrement     string

	LicenseMode        bool
	MaxLicense         string
	EnableRedelegation bool
	SpecialMode        bool

	PubKey cryptotypes.PubKey

	IP              string
	Website         string
	SecurityContact string
	Details         string
	Identity        string
}

func PrepareConfigForTxCreateValidator(flagSet *flag.FlagSet, moniker, nodeID, chainID string, valPubKey cryptotypes.PubKey) (TxCreateValidatorConfig, error) {
	c := TxCreateValidatorConfig{}
	ip, err := flagSet.GetString(FlagIP)
	if err != nil {
		return c, err
	}
	if ip == "" {
		_, _ = fmt.Fprintf(os.Stderr, "couldn't retrieve an external IP; "+
			"the tx's memo field will be unset")
	}
	c.IP = ip

	website, err := flagSet.GetString(FlagWebsite)
	if err != nil {
		return c, err
	}
	c.Website = website

	securityContact, err := flagSet.GetString(FlagSecurityContact)
	if err != nil {
		return c, err
	}
	c.SecurityContact = securityContact

	details, err := flagSet.GetString(FlagDetails)
	if err != nil {
		return c, err
	}
	c.SecurityContact = details

	identity, err := flagSet.GetString(FlagIdentity)
	if err != nil {
		return c, err
	}
	c.Identity = identity

	c.Amount, err = flagSet.GetString(FlagAmount)
	if err != nil {
		return c, err
	}

	c.CommissionRate, err = flagSet.GetString(FlagCommissionRate)
	if err != nil {
		return c, err
	}

	c.CommissionMaxRate, err = flagSet.GetString(FlagCommissionMaxRate)
	if err != nil {
		return c, err
	}

	c.CommissionMaxChangeRate, err = flagSet.GetString(FlagCommissionMaxChangeRate)
	if err != nil {
		return c, err
	}

	c.MinSelfDelegation, err = flagSet.GetString(FlagMinSelfDelegation)
	if err != nil {
		return c, err
	}

	c.NodeID = nodeID
	c.PubKey = valPubKey
	c.Website = website
	c.SecurityContact = securityContact
	c.Details = details
	c.Identity = identity
	c.ChainID = chainID
	c.Moniker = moniker

	if c.Amount == "" {
		c.Amount = defaultAmount
	}

	if c.CommissionRate == "" {
		c.CommissionRate = defaultCommissionRate
	}

	if c.CommissionMaxRate == "" {
		c.CommissionMaxRate = defaultCommissionMaxRate
	}

	if c.CommissionMaxChangeRate == "" {
		c.CommissionMaxChangeRate = defaultCommissionMaxChangeRate
	}

	if c.MinSelfDelegation == "" {
		c.MinSelfDelegation = defaultMinSelfDelegation
	}
	fmt.Println("c.MinDelegation ", c.MinDelegation)
	if c.MinDelegation == "" {
		c.MinDelegation = defaultMinDelegation
	}

	if c.DelegationIncrement == "" {
		c.DelegationIncrement = defaultDelegationIncrement
	}

	c.LicenseMode, err = flagSet.GetBool(FlagLicenseMode)
	if err != nil {
		return c, err
	}

	c.EnableRedelegation, err = flagSet.GetBool(FlagEnableRedelegation)
	if err != nil {
		return c, err
	}

	c.MaxLicense, err = flagSet.GetString(FlagMaxLicense)
	if err != nil {
		return c, err
	}

	c.SpecialMode, err = flagSet.GetBool(FlagSpecialMode)
	if err != nil {
		return c, err
	}

	return c, nil
}

// BuildCreateValidatorMsg makes a new MsgCreateValidator.
func BuildCreateValidatorMsg(clientCtx client.Context, config TxCreateValidatorConfig, txBldr tx.Factory, generateOnly bool) (tx.Factory, sdk.Msg, error) {
	amounstStr := config.Amount
	amount, err := sdk.ParseCoinNormalized(amounstStr)

	if err != nil {
		return txBldr, nil, err
	}

	valAddr := clientCtx.GetFromAddress()
	description := types.NewDescription(
		config.Moniker,
		config.Identity,
		config.Website,
		config.SecurityContact,
		config.Details,
	)

	// get the initial validator commission parameters
	rateStr := config.CommissionRate
	maxRateStr := config.CommissionMaxRate
	maxChangeRateStr := config.CommissionMaxChangeRate
	commissionRates, err := buildCommissionRates(rateStr, maxRateStr, maxChangeRateStr)

	if err != nil {
		return txBldr, nil, err
	}

	// get the initial validator min self delegation
	msbStr := config.MinSelfDelegation
	minSelfDelegation, ok := sdk.NewIntFromString(msbStr)

	if !ok {
		return txBldr, nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "minimum self delegation must be a positive integer")
	}

	msg, err := types.NewMsgCreateValidator(
		sdk.ValAddress(valAddr), "", config.PubKey, amount, description, commissionRates, minSelfDelegation,
	)

	mdlStr := config.MinDelegation
	minDelegation, ok := sdk.NewIntFromString(mdlStr)
	if !ok {
		return txBldr, nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "minimum delegation must be a positive integer")
	}
	msg.MinDelegation = minDelegation

	dliStr := config.DelegationIncrement
	delegationIncrement, ok := sdk.NewIntFromString(dliStr)
	if !ok {
		return txBldr, nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "delegation increment must be a positive integer")
	}
	msg.DelegationIncrement = delegationIncrement

	enableRedelegation := config.EnableRedelegation

	switch{
	case config.LicenseMode:
		msg.LicenseMode = true
		msg.SpecialMode = false
		mlcStr := config.MaxLicense
		maxLicense, ok := sdk.NewIntFromString(mlcStr)
		if !ok {
			return txBldr, nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "When license mode is used, max license is required and must be positive")
		}
		msg.MaxLicense = maxLicense

		if enableRedelegation {
			return txBldr, nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "When license mode is used, redelegation must be disabled")
		}
	case config.SpecialMode:
		msg.LicenseMode = false
		msg.SpecialMode = true
	default:
		msg.LicenseMode = false
		msg.SpecialMode = false
	}

	if(msg.LicenseMode && msg.SpecialMode){
		return txBldr, nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Cannot enable license mode along side with specialmode")
	}

	// Enable Redelegation
	msg.EnableRedelegation = enableRedelegation

	if err != nil {
		return txBldr, msg, err
	}
	if generateOnly {
		ip := config.IP
		nodeID := config.NodeID

		if nodeID != "" && ip != "" {
			txBldr = txBldr.WithMemo(fmt.Sprintf("%s@%s:26656", nodeID, ip))
		}
	}

	return txBldr, msg, nil
}
