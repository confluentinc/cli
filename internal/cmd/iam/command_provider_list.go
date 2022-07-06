package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var (
	identityProviderListFields           = []string{"Id", "DisplayName", "Description", "Issuer", "JwksUri"}
	identityProviderListHumanLabels      = []string{"ID", "Display Name", "Description", "Issuer", "JWKS URI"}
	identityProviderListStructuredLabels = []string{"id", "display_name", "description", "issuer", "jwks_uri"}
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

	outputWriter, err := output.NewListOutputWriter(cmd, identityProviderListFields, identityProviderListHumanLabels, identityProviderListStructuredLabels)
	if err != nil {
		return err
	}
	for _, op := range identityProviders {
		element := &identityProvider{Id: *op.Id, DisplayName: *op.DisplayName, Description: *op.Description, Issuer: *op.Issuer, JwksUri: *op.JwksUri}
		outputWriter.AddElement(element)
	}
	return outputWriter.Out()
}
