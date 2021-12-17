package environment

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type command struct {
	*pcmd.AuthenticatedStateFlagCommand
	analyticsClient analytics.Client
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

	c.AddCommand(c.newCreateCommand())
	c.AddCommand(c.newDeleteCommand())
	c.AddCommand(c.newListCommand())
	c.AddCommand(c.newUpdateCommand())
	c.AddCommand(c.newUseCommand())

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
