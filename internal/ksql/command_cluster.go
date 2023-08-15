package ksql

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
)

func newClusterCommand(cfg *config.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "cluster",
		Short:       "Manage ksqlDB clusters.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLoginOrOnPremLogin},
	}

	if cfg.IsCloudLogin() {
		c := &ksqlCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}
		cmd.AddCommand(c.newConfigureAclsCommand())
		cmd.AddCommand(c.newCreateCommand())
		cmd.AddCommand(c.newDeleteCommand())
		cmd.AddCommand(c.newDescribeCommand())
		cmd.AddCommand(c.newListCommand())
	} else {
		c := &ksqlCommand{pcmd.NewAuthenticatedWithMDSCLICommand(cmd, prerunner)}
		cmd.AddCommand(c.newListCommandOnPrem())
	}

	return cmd
}
