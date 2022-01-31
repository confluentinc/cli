package ksql

import (
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/spf13/cobra"
)

func newClusterCommand(cfg *v1.Config, prerunner pcmd.PreRunner) *ksqlCommand {
	cmd := &cobra.Command{
		Use:         "cluster",
		Short:       "Manage ksqlDB clusters.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLoginOrOnPremLogin},
	}

	var c *ksqlCommand
	var listCommand *cobra.Command
	if cfg.IsCloudLogin() {
		c = &ksqlCommand{
			AuthenticatedStateFlagCommand: pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner),
		}
		listCommand = c.newListCommand(false)
	} else {
		c = &ksqlCommand{
			AuthenticatedStateFlagCommand: pcmd.NewAuthenticatedWithMDSStateFlagCommand(cmd, prerunner),
		}
		listCommand = c.newListCommandOnPrem()
	}

	if cfg.IsCloudLogin() {
		c.AddCommand(c.newConfigureAclsCommand(false))
		c.AddCommand(c.newCreateCommand(false))
		c.AddCommand(c.newDeleteCommand(false))
		c.AddCommand(c.newDescribeCommand(false))
	}
	c.AddCommand(listCommand)

	return c
}
