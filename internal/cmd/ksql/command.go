package ksql

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type command struct {
	*pcmd.CLICommand
	analyticsClient analytics.Client
}

func New(prerunner pcmd.PreRunner, analyticsClient analytics.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "ksql",
		Short:       "Manage ksqlDB applications.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLoginOrOnPremLogin},
	}

	c := &command{
		CLICommand:      pcmd.NewCLICommand(cmd, prerunner),
		analyticsClient: analyticsClient,
	}

	appCmd := newAppCommand(prerunner, c.analyticsClient)

	c.AddCommand(appCmd.Command)
	c.AddCommand(newClusterCommand(prerunner))

	return c.Command
}
