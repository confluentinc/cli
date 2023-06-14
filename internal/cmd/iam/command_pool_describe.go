package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *identityPoolCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <id>",
		Short:             "Describe an identity pool.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.describe,
	}

	pcmd.AddProviderFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("provider"))

	return cmd
}

func (c *identityPoolCommand) describe(cmd *cobra.Command, args []string) error {
	provider, err := cmd.Flags().GetString("provider")
	if err != nil {
		return err
	}

	identityPoolProfile, err := c.V2Client.GetIdentityPool(args[0], provider)
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&identityPoolOut{
		Id:            identityPoolProfile.GetId(),
		DisplayName:   identityPoolProfile.GetDisplayName(),
		Description:   identityPoolProfile.GetDescription(),
		IdentityClaim: identityPoolProfile.GetIdentityClaim(),
		Filter:        identityPoolProfile.GetFilter(),
	})
	return table.Print()
}
