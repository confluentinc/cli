package ksql

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type clusterCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
}

func newClusterCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "cluster",
		Short:       "Manage ksqlDB clusters.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireOnPremLogin},
	}

	c := &clusterCommand{pcmd.NewAuthenticatedWithMDSStateFlagCommand(cmd, prerunner)}

	c.AddCommand(c.newListCommand())

	return c.Command
}
