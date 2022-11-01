package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type replicaCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
}

func newReplicaCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "replica",
		Short:       "Manage Kafka replicas.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireOnPremLogin},
	}

	c := &replicaCommand{pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner)}
	c.PersistentPreRunE = prerunner.InitializeOnPremKafkaRest(c.AuthenticatedCLICommand)

	cmd.AddCommand(c.newListCommand())

	return cmd
}
