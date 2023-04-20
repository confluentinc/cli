package connect

import (
	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/deletion"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func (c *clusterCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id-1> [id-2] ... [id-n]",
		Short:             "Delete connectors.",
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

	errs := &multierror.Error{ErrorFormat: errors.CustomMultierrorList}
	var deleted []string
	for _, id := range args {
		if _, err := c.V2Client.DeleteConnector(connectorIdToName[id], environmentId, kafkaCluster.ID); err != nil {
			errs = multierror.Append(errs, err)
		} else {
			deleted = append(deleted, id)
		}
	}
	deletion.PrintSuccessMsg(deleted, resource.Connector)

	return errs.ErrorOrNil()
}

func (c *clusterCommand) confirmDeletion(cmd *cobra.Command, environmentId, kafkaClusterId string, args []string, connectorIdToName map[string]string) error {
	describeFunc := func(id string) error {
		connector, err := c.V2Client.GetConnectorExpansionById(id, environmentId, kafkaClusterId)
		if err == nil {
			connectorIdToName[id] = connector.Info.GetName()
		}
		return err
	}

	if err := deletion.ValidateArgs(cmd, args, resource.Connector, describeFunc); err != nil {
		return err
	}

	if len(args) == 1 {
		if err := form.ConfirmDeletionWithString(cmd, resource.Connector, args[0], connectorIdToName[args[0]]); err != nil {
			return err
		}
	} else {
		if ok, err := form.ConfirmDeletionYesNo(cmd, resource.Connector, args); err != nil || !ok {
			return err
		}
	}

	return nil
}
