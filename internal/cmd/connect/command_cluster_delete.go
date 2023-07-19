package connect

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/resource"
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

	connectorIdToName := make(map[string]string)
	if err := c.confirmDeletion(cmd, environmentId, kafkaCluster.ID, args, connectorIdToName); err != nil {
		return err
	}

	deleteFunc := func(id string) error {
		if _, err := c.V2Client.DeleteConnector(connectorIdToName[id], environmentId, kafkaCluster.ID); err != nil {
			return err
		}
		return nil
	}

	deleted, err := resource.Delete(args, deleteFunc, nil)
	resource.PrintDeleteSuccessMsg(deleted, resource.Connector)

	return err
}

func (c *clusterCommand) confirmDeletion(cmd *cobra.Command, environmentId, kafkaClusterId string, args []string, connectorIdToName map[string]string) error {
	describeFunc := func(id string) error {
		connector, err := c.V2Client.GetConnectorExpansionById(id, environmentId, kafkaClusterId)
		if err == nil {
			connectorIdToName[id] = connector.Info.GetName()
		}
		return err
	}

	if err := resource.ValidateArgs(pcmd.FullParentName(cmd), args, resource.Connector, describeFunc); err != nil {
		return err
	}

	if len(args) == 1 {
		displayName := connectorIdToName[args[0]]
		if err := form.ConfirmDeletionWithString(cmd, form.DefaultPromptString(resource.Connector, args[0], displayName), displayName); err != nil {
			return err
		}
	} else {
		if ok, err := form.ConfirmDeletionYesNo(cmd, form.DefaultYesNoPromptString(resource.Connector, args)); err != nil || !ok {
			return err
		}
	}

	return nil
}
