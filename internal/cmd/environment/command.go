package environment

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type command struct {
	*pcmd.AuthenticatedStateFlagCommand
	completableChildren []*cobra.Command
	analyticsClient     analytics.Client
}

func New(prerunner pcmd.PreRunner, analyticsClient analytics.Client) *command {
	cmd := &cobra.Command{
		Use:         "environment",
		Aliases:     []string{"env"},
		Short:       "Manage and select Confluent Cloud environments.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &command{
		AuthenticatedStateFlagCommand: pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner),
		analyticsClient:               analyticsClient,
	}

	deleteCmd := c.newDeleteCommand()
	updateCmd := c.newUpdateCommand()
	useCmd := c.newUseCommand()

	c.AddCommand(c.newCreateCommand())
	c.AddCommand(deleteCmd)
	c.AddCommand(c.newListCommand())
	c.AddCommand(updateCmd)
	c.AddCommand(useCmd)

	c.completableChildren = []*cobra.Command{deleteCmd, updateCmd, useCmd}

	return c
}

func (c *command) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return pcmd.AutocompleteEnvironments(c.Client)
}
