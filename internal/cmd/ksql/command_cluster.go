package ksql

import (
	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/spf13/cobra"
)

func newClusterCommand(prerunner pcmd.PreRunner, analyticsClient analytics.Client) *ksqlCommand {
	cmd := &cobra.Command{
		Use:         "cluster",
		Short:       "Manage ksqlDB clusters.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
	}

	c := &ksqlCommand{
		AuthenticatedStateFlagCommand: pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner),
	}

	c.AddCommand(c.newConfigureAclsCommand(false))
	c.AddCommand(c.newCreateCommand(false))
	c.AddCommand(c.newDeleteCommand(false))
	c.AddCommand(c.newDescribeCommand(false))
	c.AddCommand(c.newListCommand(false))

	return c
}
