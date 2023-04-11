package ksql

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
)

func newClusterCommand(cfg *v1.Config, prerunner pcmd.PreRunner) *cobra.Command {
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
