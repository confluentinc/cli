package iam

import (
	identityproviderv2 "github.com/confluentinc/ccloud-sdk-go-v2/identity-provider/v2"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type identityProviderCommand struct {
	*pcmd.AuthenticatedCLICommand
}

type identityProviderOut struct {
	Id          string `human:"ID" serialized:"id"`
	Name        string `human:"Name" serialized:"name"`
	Description string `human:"Description" serialized:"description"`
	IssuerUri   string `human:"Issuer URI" serialized:"issuer_uri"`
	JwksUri     string `human:"JWKS URI" serialized:"jwks_uri"`
}

func newProviderCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "provider",
		Short:       "Manage identity providers.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
	}

	c := &identityProviderCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newCreateCommand())
	cmd.AddCommand(c.newDeleteCommand())
	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(c.newListCommand())
	cmd.AddCommand(c.newUpdateCommand())

	return cmd
}

func printIdentityProvider(cmd *cobra.Command, provider identityproviderv2.IamV2IdentityProvider) error {
	table := output.NewTable(cmd)
	table.Add(&identityProviderOut{
		Id:          provider.GetId(),
		Name:        provider.GetDisplayName(),
		Description: provider.GetDescription(),
		IssuerUri:   provider.GetIssuer(),
		JwksUri:     provider.GetJwksUri(),
	})
	return table.Print()
}

func (c *identityProviderCommand) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return pcmd.AutocompleteIdentityProviders(c.V2Client)
}
