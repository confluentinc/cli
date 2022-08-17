package ksql

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
)

const Cluster = "cluster"

func newClusterCommand(cfg *v1.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         Cluster,
		Short:       "Manage ksqlDB clusters.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLoginOrOnPremLogin},
	}

	if cfg.IsCloudLogin() {
		c := &ksqlCommand{pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner)}
		cmd.AddCommand(c.newConfigureAclsCommand(Cluster))
		cmd.AddCommand(c.newCreateCommand(Cluster))
		cmd.AddCommand(c.newDeleteCommand(Cluster))
		cmd.AddCommand(c.newDescribeCommand(Cluster))
		cmd.AddCommand(c.newListCommand(Cluster))
	} else {
		c := &ksqlCommand{pcmd.NewAuthenticatedWithMDSStateFlagCommand(cmd, prerunner)}
		cmd.AddCommand(c.newListCommandOnPrem())
	}

	return cmd
}
