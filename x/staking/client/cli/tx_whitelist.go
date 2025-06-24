package cli

import (
	"cosmossdk.io/core/address"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/spf13/cobra"
)

func CmdCreateWhitelistDelegator(valAddrCodec, ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-whiltelist [validator] [delegator]",
		Short: "Create a new Whitelist Delegator",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			// Get indexes
			indexValidator := args[0]

			// Get value arguments
			argDelegator := args[1]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgCreateWhitelistDelegator(
				clientCtx.GetFromAddress().String(),
				indexValidator,
				argDelegator,
			)

			if err := msg.Validate(valAddrCodec); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func CmdDeleteWhitelistDelegator(valAddrCodec, ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete-whiltelist [validator]",
		Short: "Delete a Whitelist Delegator",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			indexValidator := args[0]

			// Get value arguments
			argDelegator := args[1]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgDeleteWhitelistDelegator(
				clientCtx.GetFromAddress().String(),
				indexValidator,
				argDelegator,
			)

			if err := msg.Validate(valAddrCodec); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
