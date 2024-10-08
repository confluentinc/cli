package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *poolCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List identity pools.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	pcmd.AddProviderFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("provider"))

	return cmd
}

func (c *poolCommand) list(cmd *cobra.Command, _ []string) error {
	provider, err := cmd.Flags().GetString("provider")
	if err != nil {
		return err
	}

	identityPools, err := c.V2Client.ListIdentityPools(provider)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, pool := range identityPools {
		list.Add(&poolOut{
			Id:            pool.GetId(),
			DisplayName:   pool.GetDisplayName(),
			Description:   pool.GetDescription(),
			IdentityClaim: pool.GetIdentityClaim(),
			Filter:        pool.GetFilter(),
		})
	}
	return list.Print()
}
