package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	nameconversions "github.com/confluentinc/cli/internal/pkg/name-conversions"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c identityProviderCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <id|name>",
		Short:             "Describe an identity provider.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.describe,
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c identityProviderCommand) describe(cmd *cobra.Command, args []string) error {
	identityProvider, err := c.V2Client.GetIdentityProvider(args[0])
	if err != nil {
		providerId, err := nameconversions.ConvertIamProviderNameToId(args[0], c.V2Client)
		if err != nil {
			return err
		}
		if identityProvider, err = c.V2Client.GetIdentityProvider(providerId); err != nil {
			return err
		}
	}

	table := output.NewTable(cmd)
	table.Add(&identityProviderOut{
		Id:          identityProvider.GetId(),
		Name:        identityProvider.GetDisplayName(),
		Description: identityProvider.GetDescription(),
		IssuerUri:   identityProvider.GetIssuer(),
		JwksUri:     identityProvider.GetJwksUri(),
	})
	return table.Print()
}
