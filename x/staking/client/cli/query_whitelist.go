package cli

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/spf13/cobra"
)

func CmdListWhitelistDelegator() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-whitelist",
		Short: "list all Whitelist Delegator",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryAllWhitelistDelegatorRequest{
				Pagination: pageReq,
			}

			res, err := queryClient.WhitelistdelegatorAll(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddPaginationFlagsToCmd(cmd, cmd.Use)
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdShowWhitelistDelegator() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-whitelist [validator]",
		Short: "shows a Whitelist Delegator",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			argValidator := args[0]

			params := &types.QueryGetWhitelistDelegatorRequest{
				Validator: argValidator,
			}

			res, err := queryClient.Whitelistdelegator(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
