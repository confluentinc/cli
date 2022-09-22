package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type identityProviderCommand struct {
	*pcmd.AuthenticatedCLICommand
}

type identityProviderOut struct {
	Id          string `human:"ID" json:"id" yaml:"id"`
	Name        string `human:"Name" json:"name" yaml:"name"`
	Description string `human:"Description" json:"description" yaml:"description"`
	IssuerUri   string `human:"Issuer URI" json:"issuer_uri" yaml:"issuer_uri"`
	JwksUri     string `human:"JWKS URI" json:"jwks_uri" yaml:"jwks_uri"`
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

func (c *identityProviderCommand) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return pcmd.AutocompleteIdentityProviders(c.V2Client)
}
