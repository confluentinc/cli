package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type identityPoolCommand struct {
	*pcmd.AuthenticatedCLICommand
}

type identityPool struct {
	Id            string
	DisplayName   string
	Description   string
	IdentityClaim string
	Filter        string
}

func newPoolCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "pool",
		Short:       "Manage identity pools.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
	}

	c := &identityPoolCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newCreateCommand())
	cmd.AddCommand(c.newDeleteCommand())
	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(c.newListCommand())
	cmd.AddCommand(c.newUpdateCommand())

	return cmd
}

func (c *identityPoolCommand) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}
	
	provider, _ := cmd.Flags().GetString("provider")
	return pcmd.AutocompleteIdentityPools(c.V2Client, provider)
}
