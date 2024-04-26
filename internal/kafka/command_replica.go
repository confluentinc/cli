package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
)

type replicaCommand struct {
	*pcmd.AuthenticatedCLICommand
}

func newReplicaCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "replica",
		Short:       "Manage Kafka replicas.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireOnPremLogin},
	}

	c := &replicaCommand{pcmd.NewAuthenticatedWithMDSCLICommand(cmd, prerunner)}
	c.PersistentPreRunE = prerunner.InitializeOnPremKafkaRest(c.AuthenticatedCLICommand)

	cmd.AddCommand(c.newListCommand())
	cmd.AddCommand(c.newStatusCommand())

	return cmd
}
