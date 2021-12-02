package ksql

import (
	"github.com/spf13/cobra"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/shell/completer"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type command struct {
	*pcmd.CLICommand
	serverCompleter completer.ServerSideCompleter
	analyticsClient analytics.Client
}

func New(cfg *v1.Config, prerunner pcmd.PreRunner, serverCompleter completer.ServerSideCompleter, analyticsClient analytics.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "ksql",
		Short:       "Manage ksqlDB applications.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLoginOrOnPremLogin},
	}

	c := &command{
		CLICommand:      pcmd.NewCLICommand(cmd, prerunner),
		serverCompleter: serverCompleter,
		analyticsClient: analyticsClient,
	}

	appCmd := newAppCommand(prerunner, c.analyticsClient)

	c.AddCommand(appCmd.Command)
	c.AddCommand(NewClusterCommandOnPrem(prerunner))

	if cfg.IsCloudLogin() {
		c.serverCompleter.AddCommand(appCmd)
	}

	return c.Command
}
