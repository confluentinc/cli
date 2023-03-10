package connect

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	perrors "github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/confluentinc/cli/internal/pkg/set"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *clusterCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id-1> [id-2] ... [id-N]",
		Short:             "Delete connectors.",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
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
	kafkaCluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	connectorNames, err := c.checkExistence(cmd, kafkaCluster.ID, args)
	if err != nil {
		return err
	}

	if len(args) > 1 {
		promptMsg := fmt.Sprintf(perrors.DeleteResourcesConfirmYesNoMsg, resource.Connector, utils.ArrayToCommaDelimitedStringWithAnd(args))
		if ok, err := form.ConfirmDeletion(cmd, promptMsg, ""); err != nil || !ok {
			return err
		}
	} else {
		promptMsg := fmt.Sprintf(perrors.DeleteResourceConfirmMsg, resource.Connector, args[0], connectorNames[0])
		if _, err := form.ConfirmDeletion(cmd, promptMsg, connectorNames[0]); err != nil {
			return err
		}
	}

	var errs error
	for i, connectorName := range connectorNames {
		if _, err := c.V2Client.DeleteConnector(connectorName, c.EnvironmentId(), kafkaCluster.ID); err != nil {
			errs = errors.Join(errs, err)
		} else {
			output.Printf(perrors.DeletedResourceMsg, resource.Connector, args[i])
		}
	}

	return errs
}

func (c *clusterCommand) checkExistence(cmd *cobra.Command, kafkaClusterId string, args []string) ([]string, error) {
	// Single
	if len(args) == 1 {
		if connector, err := c.V2Client.GetConnectorExpansionById(args[0], c.EnvironmentId(), kafkaClusterId); err != nil {
			return nil, err
		} else {
			return []string{connector.Info.GetName()}, nil
		}
	}

	// Multiple
	connectors, err := c.V2Client.ListConnectorsWithExpansions(c.EnvironmentId(), kafkaClusterId, "id,status")
	if err != nil {
		return nil, err
	}

	connectorSet := set.New()
	connectorNames := make([]string, len(connectors))
	i := 0
	for _, connector := range connectors {
		connectorSet.Add(connector.Id.GetId())
		connectorNames[i] = connector.Info.GetName()
		i++
	}

	invalidConnectors := connectorSet.Difference(args)
	if len(invalidConnectors) > 0 {
		return nil, perrors.New("unknown connector ID(s): " + utils.ArrayToCommaDelimitedStringWithAnd(invalidConnectors))
	}

	return connectorNames, nil
}
