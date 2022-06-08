package iam

import (
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/spf13/cobra"
)

// May need to add more fields depending on OAUTH team
var poolListFields = []string{"Id", "DisplayName", "Description", "SubjectClaim"}

type identityPoolCommand struct {
	*pcmd.AuthenticatedCLICommand
}

// May need to add more fields depending on OAUTH team
type identityPool struct {
	Id           string
	DisplayName  string
	Description  string
	SubjectClaim string
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
	providerID, _ := cmd.Flags().GetString("provider")
	return pcmd.AutocompleteIdentityPools(c.V2Client, providerID)
}
