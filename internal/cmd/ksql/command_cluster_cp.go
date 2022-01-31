package ksql

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

func newClusterCommandOnPrem(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "cluster",
		Short:       "Manage ksqlDB clusters in Confluent Platform.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireOnPremLogin},
	}

	c := &ksqlCommand{
		AuthenticatedStateFlagCommand: pcmd.NewAuthenticatedWithMDSStateFlagCommand(cmd, prerunner),
	}
	
	c.AddCommand(c.newListCommandOnPrem())

	return c.Command
}
