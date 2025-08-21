package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	"cosmossdk.io/core/address"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"

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
	defaultValidatorMode           = types.ValidatorMode_MODE_NORMAL
)

// NewTxCmd returns a root CLI command handler for all x/staking transaction commands.
func NewTxCmd(valAddrCodec, ac address.Codec) *cobra.Command {
	stakingTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Staking transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	stakingTxCmd.AddCommand(
		NewSetValidatorApprovalCmd(),
		NewCreateValidatorLegacyCmd(valAddrCodec),
		NewCreateValidatorCmd(valAddrCodec),
		NewEditValidatorCmd(valAddrCodec),
		NewDelegateCmd(valAddrCodec, ac),
		NewRedelegateCmd(valAddrCodec, ac),
		NewUnbondCmd(valAddrCodec, ac),
		NewCancelUnbondingDelegation(valAddrCodec, ac),
		CmdCreateWhitelistDelegator(valAddrCodec, ac),
		CmdDeleteWhitelistDelegator(valAddrCodec, ac),
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

func NewCreateValidatorLegacyCmd(ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-validator-legacy",
		Short: "create new validator initialized with a self-delegation to it (legacy version)",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			moniker, _ := cmd.Flags().GetString(FlagMoniker)
			identity, _ := cmd.Flags().GetString(FlagIdentity)
			website, _ := cmd.Flags().GetString(FlagWebsite)
			security, _ := cmd.Flags().GetString(FlagSecurityContact)
			details, _ := cmd.Flags().GetString(FlagDetails)

			pkStr, err := cmd.Flags().GetString(FlagPubKey)
			if err != nil {
				return err
			}

			var pk cryptotypes.PubKey
			if err := clientCtx.Codec.UnmarshalInterfaceJSON([]byte(pkStr), &pk); err != nil {
				return err
			}

			fAmount, _ := cmd.Flags().GetString(FlagAmount)
			amount, err := sdk.ParseCoinNormalized(fAmount)
			if err != nil {
				return err
			}

			// get the initial validator commission parameters
			rateStr, _ := cmd.Flags().GetString(FlagCommissionRate)
			maxRateStr, _ := cmd.Flags().GetString(FlagCommissionMaxRate)
			maxChangeRateStr, _ := cmd.Flags().GetString(FlagCommissionMaxChangeRate)

			commissionRates, err := buildCommissionRates(rateStr, maxRateStr, maxChangeRateStr)
			if err != nil {
				return err
			}

			var valMinSelfDelegation *math.Int
			minSelfDelegationString, _ := cmd.Flags().GetString(FlagMinSelfDelegation)
			if minSelfDelegationString != "" {
				msb, ok := math.NewIntFromString(minSelfDelegationString)
				if !ok {
					return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "minimum self delegation must be a positive integer")
				}

				valMinSelfDelegation = &msb
			}

			validator := validator{
				Amount:            amount,
				PubKey:            pk,
				Moniker:           moniker,
				Identity:          identity,
				Website:           website,
				Security:          security,
				Details:           details,
				CommissionRates:   commissionRates,
				MinSelfDelegation: *valMinSelfDelegation,
			}

			txf, err := tx.NewFactoryCLI(clientCtx, cmd.Flags())
			// .WithTxConfig(clientCtx.TxConfig)
			// .WithAccountRetriever(clientCtx.AccountRetriever)
			if err != nil {
				return err
			}
			txf, msg, err := newBuildCreateValidatorMsg(clientCtx, txf, cmd.Flags(), validator, ac)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	// add flag
	cmd.Flags().AddFlagSet(FlagSetPublicKey())
	cmd.Flags().AddFlagSet(FlagSetAmount())
	cmd.Flags().AddFlagSet(flagSetDescriptionCreate())
	cmd.Flags().AddFlagSet(FlagSetCommissionCreate())
	cmd.Flags().AddFlagSet(FlagSetMinSelfDelegation())
	cmd.Flags().AddFlagSet(FlagSetApprover())
	cmd.Flags().AddFlagSet(FlagMinDelegationCreate())
	cmd.Flags().AddFlagSet(FlagDelegationIncrementCreate())
	cmd.Flags().AddFlagSet(FlagValidatorModeCreate())
	cmd.Flags().AddFlagSet(FlagEnableRedelegationCreate())

	cmd.Flags().String(FlagIP, "", fmt.Sprintf("The node's public IP. It takes effect only when used in combination with --%s", flags.FlagGenerateOnly))
	cmd.Flags().String(FlagNodeID, "", "The node's ID")
	flags.AddTxFlagsToCmd(cmd)

	// mark flag that required of this message
	_ = cmd.MarkFlagRequired(flags.FlagFrom)
	_ = cmd.MarkFlagRequired(FlagAmount)
	_ = cmd.MarkFlagRequired(FlagPubKey)
	_ = cmd.MarkFlagRequired(FlagMinDelegation)
	_ = cmd.MarkFlagRequired(FlagDelegationIncrement)
	_ = cmd.MarkFlagRequired(FlagMaxLicense)
	_ = cmd.MarkFlagRequired(FlagValidatorMode)
	_ = cmd.MarkFlagRequired(FlagIdentity)
	_ = cmd.MarkFlagRequired(FlagWebsite)
	_ = cmd.MarkFlagRequired(FlagSecurityContact)
	_ = cmd.MarkFlagRequired(FlagDetails)

	return cmd
}

// NewCreateValidatorCmd returns a CLI command handler for creating a MsgCreateValidator transaction.
func NewCreateValidatorCmd(ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-validator [path/to/validator.json]",
		Short: "create new validator initialized with a self-delegation to it",
		Args:  cobra.ExactArgs(1),
		Long:  `Create a new validator initialized with a self-delegation by submitting a JSON file with the new validator details.`,
		Example: strings.TrimSpace(
			fmt.Sprintf(`
$ %s tx staking create-validator path/to/validator.json --from keyname

Where validator.json contains:

{
	"pubkey": {"@type":"/cosmos.crypto.ed25519.PubKey","key":"oWg2ISpLF405Jcm2vXV+2v4fnjodh6aafuIdeoW+rUw="},
	"amount": "1000000stake",
	"moniker": "myvalidator",
	"identity": "optional identity signature (ex. UPort or Keybase)",
	"website": "validator's (optional) website",
	"security": "validator's (optional) security contact email",
	"details": "validator's (optional) details",
	"commission-rate": "0.1",
	"commission-max-rate": "0.2",
	"commission-max-change-rate": "0.01",
	"min-self-delegation": "1"
}

where we can get the pubkey using "%s tendermint show-validator"
`, version.AppName, version.AppName)),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf, err := tx.NewFactoryCLI(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			validator, err := parseAndValidateValidatorJSON(clientCtx.Codec, args[0])
			if err != nil {
				return err
			}

			txf, msg, err := newBuildCreateValidatorMsg(clientCtx, txf, cmd.Flags(), validator, ac)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	cmd.Flags().AddFlagSet(FlagSetApprover())
	cmd.Flags().AddFlagSet(FlagMinDelegationCreate())
	cmd.Flags().AddFlagSet(FlagDelegationIncrementCreate())
	cmd.Flags().AddFlagSet(FlagEnableRedelegationCreate())
	cmd.Flags().AddFlagSet(FlagValidatorModeCreate())

	cmd.Flags().String(FlagIP, "", fmt.Sprintf("The node's public IP. It takes effect only when used in combination with --%s", flags.FlagGenerateOnly))
	cmd.Flags().String(FlagNodeID, "", "The node's ID")
	flags.AddTxFlagsToCmd(cmd)

	_ = cmd.MarkFlagRequired(flags.FlagFrom)

	return cmd
}

// NewEditValidatorCmd returns a CLI command handler for creating a MsgEditValidator transaction.
func NewEditValidatorCmd(ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit-validator",
		Short: "edit an existing validator account",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			moniker, _ := cmd.Flags().GetString(FlagEditMoniker)
			identity, _ := cmd.Flags().GetString(FlagIdentity)
			website, _ := cmd.Flags().GetString(FlagWebsite)
			security, _ := cmd.Flags().GetString(FlagSecurityContact)
			details, _ := cmd.Flags().GetString(FlagDetails)
			validatorMode, _ := cmd.Flags().GetString(FlagValidatorMode)
			maxLicense, _ := cmd.Flags().GetString(FlagMaxLicense)
			description := types.NewDescription(moniker, identity, website, security, details)

			var newRate *math.LegacyDec

			commissionRate, _ := cmd.Flags().GetString(FlagCommissionRate)
			if commissionRate != "" {
				rate, err := math.LegacyNewDecFromStr(commissionRate)
				if err != nil {
					return fmt.Errorf("invalid new commission rate: %v", err)
				}

				newRate = &rate
			}

			var newMinSelfDelegation *math.Int

			minSelfDelegationString, _ := cmd.Flags().GetString(FlagMinSelfDelegation)
			if minSelfDelegationString != "" {
				msb, ok := math.NewIntFromString(minSelfDelegationString)
				if !ok {
					return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "minimum self delegation must be a positive integer")
				}

				newMinSelfDelegation = &msb
			}

			var newMaxLicense *math.Int
			if maxLicense != "" {

				msb, ok := math.NewIntFromString(maxLicense)
				if !ok {
					return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "When license mode is used, max license is required and must be positive")
				}
				newMaxLicense = &msb
			}

			valAddr, err := ac.BytesToString(clientCtx.GetFromAddress())
			if err != nil {
				return err
			}

			mode := convertValidatorFlag(validatorMode)

			msg := types.NewMsgEditValidator(valAddr, description, newRate, newMinSelfDelegation, mode, newMaxLicense)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().AddFlagSet(flagSetDescriptionEdit())
	cmd.Flags().AddFlagSet(flagSetCommissionUpdate())
	cmd.Flags().AddFlagSet(FlagSetMinSelfDelegation())
	// cmd.Flags().AddFlagSet(FlagMaxLicenseEdit())
	cmd.Flags().AddFlagSet(FlagValidatorModeEdit())
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewDelegateCmd returns a CLI command handler for creating a MsgDelegate transaction.
func NewDelegateCmd(valAddrCodec, ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delegate [validator-addr] [amount]",
		Args:  cobra.ExactArgs(2),
		Short: "Delegate liquid tokens to a validator",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Delegate an amount of liquid coins to a validator from your wallet.

Example:
$ %s tx staking delegate cosmosvalopers1l2rsakp388kuv9k8qzq6lrm9taddae7fpx59wm 1000stake --from mykey
`,
				version.AppName,
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

			delAddr, err := ac.BytesToString(clientCtx.GetFromAddress())
			if err != nil {
				return err
			}

			_, err = valAddrCodec.StringToBytes(args[0])
			if err != nil {
				return err
			}

			msg := types.NewMsgDelegate(delAddr, args[0], amount)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewRedelegateCmd returns a CLI command handler for creating a MsgBeginRedelegate transaction.
func NewRedelegateCmd(valAddrCodec, ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "redelegate [src-validator-addr] [dst-validator-addr] [amount]",
		Short: "Redelegate illiquid tokens from one validator to another",
		Args:  cobra.ExactArgs(3),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Redelegate an amount of illiquid staking tokens from one validator to another.

Example:
$ %s tx staking redelegate cosmosvalopers1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj cosmosvalopers1l2rsakp388kuv9k8qzq6lrm9taddae7fpx59wm 100stake --from mykey
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			delAddr, err := ac.BytesToString(clientCtx.GetFromAddress())
			if err != nil {
				return err
			}

			_, err = valAddrCodec.StringToBytes(args[0])
			if err != nil {
				return err
			}

			_, err = valAddrCodec.StringToBytes(args[1])
			if err != nil {
				return err
			}

			amount, err := sdk.ParseCoinNormalized(args[2])
			if err != nil {
				return err
			}

			msg := types.NewMsgBeginRedelegate(delAddr, args[0], args[1], amount)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewUnbondCmd returns a CLI command handler for creating a MsgUndelegate transaction.
func NewUnbondCmd(valAddrCodec, ac address.Codec) *cobra.Command {
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

			delAddr, err := ac.BytesToString(clientCtx.GetFromAddress())
			if err != nil {
				return err
			}
			_, err = valAddrCodec.StringToBytes(args[0])
			if err != nil {
				return err
			}

			amount, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}

			msg := types.NewMsgUndelegate(delAddr, args[0], amount)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func newBuildCreateValidatorMsg(clientCtx client.Context, txf tx.Factory, fs *flag.FlagSet, val validator, valAc address.Codec) (tx.Factory, *types.MsgCreateValidator, error) {
	valAddr := clientCtx.GetFromAddress()

	description := types.NewDescription(
		val.Moniker,
		val.Identity,
		val.Website,
		val.Security,
		val.Details,
	)

	valStr, err := valAc.BytesToString(sdk.ValAddress(valAddr))
	if err != nil {
		return txf, nil, err
	}
	approver, _ := fs.GetString(FlagAddressApprover)
	msg, err := types.NewMsgCreateValidator(
		valStr, approver, val.PubKey, val.Amount, description, val.CommissionRates, val.MinSelfDelegation,
	)
	if err != nil {
		return txf, nil, err
	}

	// Custom Validator
	// Min delegation
	mdStr, _ := fs.GetString(FlagMinDelegation)
	if mdStr != "" {
		minDelegation, ok := math.NewIntFromString(mdStr)
		if !ok {
			return txf, nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "minimum delegation must be a positive integer")
		}
		msg.MinDelegation = minDelegation
	}
	// delegation increment
	dincStr, _ := fs.GetString(FlagDelegationIncrement)
	if dincStr != "" {
		delegationIncrement, ok := math.NewIntFromString(dincStr)
		if !ok {
			return txf, nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "delegation increment must be a positive integer")
		}
		msg.DelegationIncrement = delegationIncrement
		if mdStr == "" {
			// if min delegation is not defined and increment is. Assign min delegation = increment
			msg.MinDelegation = delegationIncrement
		}
	}

	enableRedelegation, _ := fs.GetBool(FlagEnableRedelegation)

	flagMode, err := fs.GetString(FlagValidatorMode)
	if err != nil {
		return txf, nil, err
	}

	validatorMode := convertValidatorFlag(flagMode)

	switch validatorMode {
	case types.ValidatorMode_MODE_LICENSE:
		msg.Mode = types.ValidatorMode_MODE_LICENSE
		mlcStr, _ := fs.GetString(FlagMaxLicense)
		maxLicense, ok := math.NewIntFromString(mlcStr)
		if !ok {
			return txf, nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "When license mode is used, max license is required and must be positive")
		}
		msg.MaxLicense = maxLicense

		if enableRedelegation {
			return txf, nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "When license mode is used, redelegation must be disabled")
		}

		// check Count licesene amount for validator
		divAmount := msg.Value.Amount.Quo(msg.DelegationIncrement)
		modAmount := msg.Value.Amount.Mod(msg.DelegationIncrement)
		if modAmount.GT(math.ZeroInt()) {
			return txf, nil, types.ErrInvalidIncrementDelegation
		}
		if divAmount.GT(msg.MaxLicense) {
			return txf, nil, types.ErrNotEnoughLicense
		}
	case types.ValidatorMode_MODE_FAST:
		msg.Mode = types.ValidatorMode_MODE_FAST
	default:
		msg.Mode = types.ValidatorMode_MODE_NORMAL
	}

	// Enable Redelegation
	msg.EnableRedelegation = enableRedelegation
	if err := msg.Validate(valAc); err != nil {
		return txf, nil, err
	}

	genOnly, _ := fs.GetBool(flags.FlagGenerateOnly)
	if genOnly {
		ip, _ := fs.GetString(FlagIP)
		p2pPort, _ := fs.GetUint(FlagP2PPort)
		nodeID, _ := fs.GetString(FlagNodeID)

		if nodeID != "" && ip != "" && p2pPort > 0 {
			txf = txf.WithMemo(fmt.Sprintf("%s@%s:%d", nodeID, ip, p2pPort))
		}
	}

	return txf, msg, nil
}

// NewCancelUnbondingDelegation returns a CLI command handler for creating a MsgCancelUnbondingDelegation transaction.
func NewCancelUnbondingDelegation(valAddrCodec, ac address.Codec) *cobra.Command {
	bech32PrefixValAddr := sdk.GetConfig().GetBech32ValidatorAddrPrefix()

	cmd := &cobra.Command{
		Use:   "cancel-unbond [validator-addr] [amount] [creation-height]",
		Short: "Cancel unbonding delegation and delegate back to the validator",
		Args:  cobra.ExactArgs(3),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Cancel Unbonding Delegation and delegate back to the validator.

Example:
$ %s tx staking cancel-unbond %s1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj 100stake 2 --from mykey
`,
				version.AppName, bech32PrefixValAddr,
			),
		),
		Example: fmt.Sprintf(`$ %s tx staking cancel-unbond %s1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj 100stake 2 --from mykey`,
			version.AppName, bech32PrefixValAddr),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			delAddr, err := ac.BytesToString(clientCtx.GetFromAddress())
			if err != nil {
				return err
			}

			_, err = valAddrCodec.StringToBytes(args[0])
			if err != nil {
				return err
			}

			amount, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}

			creationHeight, err := strconv.ParseInt(args[2], 10, 64)
			if err != nil {
				return errorsmod.Wrap(fmt.Errorf("invalid height: %d", creationHeight), "invalid height")
			}

			msg := types.NewMsgCancelUnbondingDelegation(delAddr, args[0], creationHeight, amount)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// Return the flagset, particular flags, and a description of defaults
// this is anticipated to be used with the gen-tx
func CreateValidatorMsgFlagSet(ipDefault string) (fs *flag.FlagSet, defaultsDesc string) {
	fsCreateValidator := flag.NewFlagSet("", flag.ContinueOnError)
	fsCreateValidator.String(FlagIP, ipDefault, "The node's public P2P IP")
	fsCreateValidator.Uint(FlagP2PPort, 26656, "The node's public P2P port")
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
	fsCreateValidator.AddFlagSet(FlagValidatorModeCreate())
	fsCreateValidator.AddFlagSet(FlagEnableRedelegationCreate())

	defaultsDesc = fmt.Sprintf(`
	delegation amount:           %s
	commission rate:             %s
	commission max rate:         %s
	commission max change rate:  %s
	minimum self delegation:     %s
	minimum delegation:          %s
	delegation increment:        %s
`, defaultAmount, defaultCommissionRate,
		defaultCommissionMaxRate, defaultCommissionMaxChangeRate,
		defaultMinSelfDelegation,
		defaultMinDelegation,
		defaultDelegationIncrement,
	)

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

	ValidatorMode      types.ValidatorMode
	MaxLicense         string
	EnableRedelegation bool

	PubKey cryptotypes.PubKey

	IP              string
	P2PPort         uint
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
		_, _ = fmt.Fprintf(os.Stderr, "failed to retrieve an external IP; the tx's memo field will be unset")
	}

	p2pPort, err := flagSet.GetUint(FlagP2PPort)
	if err != nil {
		return c, err
	}

	website, err := flagSet.GetString(FlagWebsite)
	if err != nil {
		return c, err
	}

	securityContact, err := flagSet.GetString(FlagSecurityContact)
	if err != nil {
		return c, err
	}

	details, err := flagSet.GetString(FlagDetails)
	if err != nil {
		return c, err
	}

	identity, err := flagSet.GetString(FlagIdentity)
	if err != nil {
		return c, err
	}

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

	c.MinDelegation, err = flagSet.GetString(FlagMinDelegation)
	if err != nil {
		return c, err
	}

	c.DelegationIncrement, err = flagSet.GetString(FlagDelegationIncrement)
	if err != nil {
		return c, err
	}

	c.IP = ip
	c.P2PPort = p2pPort
	c.Website = website
	c.SecurityContact = securityContact
	c.Identity = identity
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

	if c.MinDelegation == "" {
		c.MinDelegation = defaultMinDelegation
	}

	if c.DelegationIncrement == "" {
		c.DelegationIncrement = defaultDelegationIncrement
	}

	validatorMode, err := flagSet.GetString(FlagValidatorMode)
	if err != nil {
		return c, err
	}

	c.ValidatorMode = convertValidatorFlag(validatorMode)

	c.EnableRedelegation, err = flagSet.GetBool(FlagEnableRedelegation)
	if err != nil {
		return c, err
	}

	c.MaxLicense, err = flagSet.GetString(FlagMaxLicense)
	if err != nil {
		return c, err
	}

	return c, nil
}

// BuildCreateValidatorMsg makes a new MsgCreateValidator.
func BuildCreateValidatorMsg(clientCtx client.Context, config TxCreateValidatorConfig, txBldr tx.Factory, generateOnly bool, valCodec address.Codec) (tx.Factory, sdk.Msg, error) {
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
	minSelfDelegation, ok := math.NewIntFromString(msbStr)

	if !ok {
		return txBldr, nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "minimum self delegation must be a positive integer")
	}

	valStr, err := valCodec.BytesToString(sdk.ValAddress(valAddr))
	if err != nil {
		return txBldr, nil, err
	}

	msg, err := types.NewMsgCreateValidator(
		valStr,
		"",
		config.PubKey,
		amount,
		description,
		commissionRates,
		minSelfDelegation,
	)

	mdlStr := config.MinDelegation
	minDelegation, ok := math.NewIntFromString(mdlStr)
	if !ok {
		return txBldr, nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "minimum delegation must be a positive integer")
	}
	msg.MinDelegation = minDelegation

	dliStr := config.DelegationIncrement
	delegationIncrement, ok := math.NewIntFromString(dliStr)
	if !ok {
		return txBldr, nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "delegation increment must be a positive integer")
	}
	msg.DelegationIncrement = delegationIncrement

	enableRedelegation := config.EnableRedelegation

	switch config.ValidatorMode {
	case types.ValidatorMode_MODE_LICENSE:
		msg.Mode = types.ValidatorMode_MODE_LICENSE
		mlcStr := config.MaxLicense
		maxLicense, ok := math.NewIntFromString(mlcStr)
		if !ok {
			return txBldr, nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "When license mode is used, max license is required and must be positive")
		}
		msg.MaxLicense = maxLicense

		if enableRedelegation {
			return txBldr, nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "When license mode is used, redelegation must be disabled")
		}
	case types.ValidatorMode_MODE_FAST:
		msg.Mode = types.ValidatorMode_MODE_FAST
	default:
		msg.Mode = types.ValidatorMode_MODE_NORMAL
	}

	// Enable Redelegation
	msg.EnableRedelegation = enableRedelegation

	if err != nil {
		return txBldr, msg, err
	}
	if generateOnly {
		ip := config.IP
		p2pPort := config.P2PPort
		nodeID := config.NodeID

		if nodeID != "" && ip != "" && p2pPort > 0 {
			txBldr = txBldr.WithMemo(fmt.Sprintf("%s@%s:%d", nodeID, ip, p2pPort))
		}
	}

	return txBldr, msg, nil
}
