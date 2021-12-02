package ksql

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

const clusterType = "ksql-cluster"

type clusterCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
}

func NewClusterCommandOnPrem(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "cluster",
		Short:       "Manage ksqlDB clusters.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireOnPremLogin},
	}

	c := &clusterCommand{AuthenticatedStateFlagCommand: pcmd.NewAuthenticatedWithMDSStateFlagCommand(cmd, prerunner, onPremClusterSubcommandFlags)}

	c.AddCommand(c.newListCommand())

	return c.Command
}
