package unifiedstreammanager

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/deletion"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/resource"
)

func (c *command) newKafkaDeregisterCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "deregister <usm-cluster-id-1> [usm-cluster-id-2] [usm-cluster-id-3] ... [usm-cluster-id-n]",
		Short:             "Deregister Kafka clusters.",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validKafkaArgsMultiple),
		RunE:              c.deregisterKafka,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Deregister a Confluent Platform Kafka cluster.",
				Code: "confluent unified-stream-manager kafka deregister usmkc-abc123",
			},
		),
	}

	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddForceFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)

	return cmd
}

func (c *command) deregisterKafka(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	existenceFunc := func(id string) bool {
		_, err := c.V2Client.GetUsmKafkaCluster(id, environmentId)
		return err == nil
	}

	if err := deletion.ValidateAndConfirm(cmd, args, existenceFunc, resource.UsmKafkaCluster); err != nil {
		return err
	}

	deleteFunc := func(id string) error {
		return c.V2Client.DeleteUsmKafkaCluster(id, environmentId)
	}

	_, err = deletion.Delete(cmd, args, deleteFunc, resource.UsmKafkaCluster)
	return err
}
