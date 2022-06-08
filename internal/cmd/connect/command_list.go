package connect

import (
	"context"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *command) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "list",
		Short:       "List connectors.",
		Args:        cobra.NoArgs,
		RunE:        c.list,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List connectors in the current or specified Kafka cluster context.",
				Code: "confluent connect list",
			},
			examples.Example{
				Code: "confluent connect list --cluster lkc-123456",
			},
		),
	}

	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) list(cmd *cobra.Command, _ []string) error {
	kafkaCluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	connectors, err := c.Client.Connect.ListWithExpansions(context.Background(), &schedv1.Connector{AccountId: c.EnvironmentId(), KafkaClusterId: kafkaCluster.ID}, "status,info,id")
	if err != nil {
		return err
	}

	outputWriter, err := output.NewListOutputWriter(cmd, listFields, listFields, listStructuredLabels)
	if err != nil {
		return err
	}

	for name, connector := range connectors {
		connector := &connectorDescribeDisplay{
			Name:   name,
			ID:     connector.Id.Id,
			Status: connector.Status.Connector.State,
			Type:   connector.Info.Type,
			Trace:  connector.Status.Connector.Trace,
		}
		outputWriter.AddElement(connector)
	}

	return outputWriter.Out()
}
