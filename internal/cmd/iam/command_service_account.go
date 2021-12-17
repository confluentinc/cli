package iam

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type serviceAccountCommand struct {
	*pcmd.AuthenticatedCLICommand
	completableChildren []*cobra.Command
}

func NewServiceAccountCommand(prerunner pcmd.PreRunner) *serviceAccountCommand {
	cmd := &cobra.Command{
		Use:         "service-account",
		Aliases:     []string{"sa"},
		Short:       "Manage service accounts.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &serviceAccountCommand{AuthenticatedCLICommand: pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	deleteCmd := c.newDeleteCommand()
	updateCmd := c.newUpdateCommand()

	c.AddCommand(c.newCreateCommand())
	c.AddCommand(deleteCmd)
	c.AddCommand(c.newListCommand())
	c.AddCommand(updateCmd)

	c.completableChildren = []*cobra.Command{updateCmd, deleteCmd}

	return c
}

func (c *serviceAccountCommand) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return pcmd.AutocompleteServiceAccounts(c.Client)
}

func requireLen(val string, maxLen int, field string) error {
	if len(val) > maxLen {
		return fmt.Errorf(field+" length should not exceed %d characters.", maxLen)
	}

	return nil
}
