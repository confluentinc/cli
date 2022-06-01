package connect

import (
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

	connectorExpansions, _, err := c.V2Client.ListConnectorsWithExpansions(c.EnvironmentId(), kafkaCluster.ID, "status,id")
	if err != nil {
		return err
	}

	outputWriter, err := output.NewListOutputWriter(cmd, listFields, listFields, listStructuredLabels)
	if err != nil {
		return err
	}

	for name, connector := range connectorExpansions {
		connector := &connectorDescribeDisplay{
			Name:   name,
			ID:     connector.Id.GetId(),
			Status: connector.Status.GetConnector().State,
			Type:   connector.Status.GetType(),
			Trace:  *connector.GetStatus().Connector.Trace,
		}
		outputWriter.AddElement(connector)
	}

	return outputWriter.Out()
}
