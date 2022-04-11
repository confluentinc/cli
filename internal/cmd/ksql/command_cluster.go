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
		c := &ksqlCommand{pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner)}
		cmd.AddCommand(c.newConfigureAclsCommand(false))
		cmd.AddCommand(c.newCreateCommand(false))
		cmd.AddCommand(c.newDeleteCommand(false))
		cmd.AddCommand(c.newDescribeCommand(false))
		cmd.AddCommand(c.newListCommand(false))
	} else {
		c := &ksqlCommand{pcmd.NewAuthenticatedWithMDSStateFlagCommand(cmd, prerunner)}
		cmd.AddCommand(c.newListCommandOnPrem())
	}

	return cmd
}
