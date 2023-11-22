package connect

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	dynamicconfig "github.com/confluentinc/cli/v3/pkg/dynamic-config"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *clusterCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "list",
		Short:       "List connectors.",
		Args:        cobra.NoArgs,
		RunE:        c.list,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List connectors in the current or specified Kafka cluster context.",
				Code: "confluent connect cluster list",
			},
			examples.Example{
				Code: "confluent connect cluster list --cluster lkc-123456",
			},
		),
	}

	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *clusterCommand) list(cmd *cobra.Command, _ []string) error {
	kafkaCluster, err := dynamicconfig.GetKafkaClusterForCommand(c.V2Client, c.Context)
	if err != nil {
		return err
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	connectors, err := c.V2Client.ListConnectorsWithExpansions(environmentId, kafkaCluster.ID, "id,status")
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for name, connector := range connectors {
		list.Add(&connectOut{
			Name:   name,
			Id:     connector.Id.GetId(),
			Status: connector.Status.Connector.GetState(),
			Type:   connector.Status.GetType(),
			Trace:  connector.Status.Connector.GetTrace(),
		})
	}
	return list.Print()
}
