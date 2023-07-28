package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

func (c *identityProviderCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <id>",
		Short:             "Describe an identity provider.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.describe,
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *identityProviderCommand) describe(cmd *cobra.Command, args []string) error {
	provider, err := c.V2Client.GetIdentityProvider(args[0])
	if err != nil {
		return err
	}

	return printIdentityProvider(cmd, provider)
}
