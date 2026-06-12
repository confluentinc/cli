package switchover

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

type command struct {
	*pcmd.AuthenticatedCLICommand
}

// New returns the parent `confluent switchover` command, which groups the
// Kafka Disaster Recovery resources: switchover pairs and switchover endpoints.
func New(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "switchover",
		Short:       "Manage Kafka Disaster Recovery switchover resources.",
		Long:        "Manage Kafka Disaster Recovery switchover pairs and switchover endpoints in Confluent Cloud.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
	}

	c := &command{AuthenticatedCLICommand: pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newPairCommand())
	cmd.AddCommand(c.newEndpointCommand())

	return cmd
}
