package iam

import (
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/spf13/cobra"
)

var providerListFields = []string{"Id", "DisplayName", "Description", "Issuer", "JwksUri"}

type identityProviderCommand struct {
	*pcmd.AuthenticatedCLICommand
}

type identityProvider struct {
	Id          string
	DisplayName string
	Description string
	Issuer      string
	JwksUri     string
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
