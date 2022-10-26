package ksql

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

const app = "app"

func newAppCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         app,
		Short:       "Manage ksqlDB apps.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
	}

	c := &ksqlCommand{pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner)}

	cmd.AddCommand(c.newConfigureAclsCommand(app))
	cmd.AddCommand(c.newCreateCommand(app))
	cmd.AddCommand(c.newDeleteCommand(app))
	cmd.AddCommand(c.newDescribeCommand(app))
	cmd.AddCommand(c.newListCommand(app))

	return cmd
}
