package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *identityProviderCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List identity providers.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *identityProviderCommand) list(cmd *cobra.Command, _ []string) error {
	identityProviders, err := c.V2Client.ListIdentityProviders()
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, provider := range identityProviders {
		list.Add(&identityProviderOut{
			Id:          provider.GetId(),
			Name:        provider.GetDisplayName(),
			Description: provider.GetDescription(),
			IssuerUri:   provider.GetIssuer(),
			JwksUri:     provider.GetJwksUri(),
		})
	}
	return list.Print()
}
