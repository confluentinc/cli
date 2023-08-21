package connect

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/deletion"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

func (c *clusterCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id-1> [id-2] ... [id-n]",
		Short:             "Delete one or more connectors.",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgsMultiple),
		RunE:              c.delete,
		Annotations:       map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Delete a connector in the current or specified Kafka cluster context.",
				Code: "confluent connect cluster delete",
			},
			examples.Example{
				Code: "confluent connect cluster delete --cluster lkc-123456",
			},
		),
	}

	pcmd.AddForceFlag(cmd)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	return cmd
}

func (c *clusterCommand) delete(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	kafkaCluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	connectorIdToName, err := c.mapConnectorIdToName(environmentId, kafkaCluster.ID)
	if err != nil {
		return err
	}

	existenceFunc := func(id string) bool {
		_, ok := connectorIdToName[id]
		return ok
	}

	if err := deletion.ValidateAndConfirmDeletion(cmd, args, existenceFunc, resource.Connector, connectorIdToName[args[0]]); err != nil {
		return err
	}

	deleteFunc := func(id string) error {
		_, err := c.V2Client.DeleteConnector(connectorIdToName[id], environmentId, kafkaCluster.ID)
		return err
	}

	_, err = deletion.Delete(args, deleteFunc, resource.Connector)
	return err
}

func (c *clusterCommand) mapConnectorIdToName(environmentId, kafkaClusterId string) (map[string]string, error) {
	// NOTE: Do NOT replace this with `V2Client.GetConnectorExpansionById` calls; that function itself calls `V2Client.ListConnectorsWithExpansions`
	connectors, err := c.V2Client.ListConnectorsWithExpansions(environmentId, kafkaClusterId, "id,info,status")
	if err != nil {
		return nil, err
	}

	connectorIdToName := make(map[string]string)
	for _, connector := range connectors {
		connectorIdToName[connector.Id.GetId()] = connector.Info.GetName()
	}

	return connectorIdToName, nil
}
