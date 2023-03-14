package connect

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/confluentinc/cli/internal/pkg/types"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *clusterCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id-1> [id-2] ... [id-n]",
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

	connectorIdToName, err := c.checkExistence(cmd, kafkaCluster.ID, args)
	if err != nil {
		return err
	}

	if _, err := form.ConfirmDeletionType(cmd, resource.Connector, connectorIdToName[args[0]], args); err != nil {
		return err
	}

	var errs error
	for _, connectorId := range args {
		if _, err := c.V2Client.DeleteConnector(connectorIdToName[connectorId], c.EnvironmentId(), kafkaCluster.ID); err != nil {
			errs = errors.Join(errs, err)
		} else {
			output.Printf(errors.DeletedResourceMsg, resource.Connector, connectorId)
		}
	}

	return errs
}

func (c *clusterCommand) checkExistence(cmd *cobra.Command, kafkaClusterId string, args []string) (map[string]string, error) {
	// Single
	if len(args) == 1 {
		if connector, err := c.V2Client.GetConnectorExpansionById(args[0], c.EnvironmentId(), kafkaClusterId); err != nil {
			return nil, errors.NewErrorWithSuggestions(fmt.Sprintf(errors.NotFoundErrorMsg, resource.Connector, args[0]), fmt.Sprintf(errors.DeleteNotFoundSuggestions, resource.Connector))
		} else {
			return map[string]string{args[0]: connector.Info.GetName()}, nil
		}
	}

	// Multiple
	connectors, err := c.V2Client.ListConnectorsWithExpansions(c.EnvironmentId(), kafkaClusterId, "id,status")
	if err != nil {
		return nil, err
	}

	set := types.NewSet()
	connectorIdToName := make(map[string]string)
	for _, connector := range connectors {
		set.Add(connector.Id.GetId())
		connectorIdToName[connector.Id.GetId()] = connector.Info.GetName()
	}

	validArgs, invalidArgs := set.IntersectionAndDifference(args)
	if force, err := cmd.Flags().GetBool("force"); err != nil {
		return nil, err
	} else if force && len(invalidArgs) > 0 {
		args = validArgs
		return connectorIdToName, nil
	}

	invalidArgsStr := utils.ArrayToCommaDelimitedStringWithAnd(invalidArgs)
	if len(invalidArgs) == 1 {
		return nil, errors.NewErrorWithSuggestions(fmt.Sprintf(errors.NotFoundErrorMsg, resource.Connector, invalidArgsStr), fmt.Sprintf(errors.DeleteNotFoundSuggestions, resource.Connector))
	} else if len(invalidArgs) > 1 {
		return nil, errors.NewErrorWithSuggestions(fmt.Sprintf(errors.NotFoundErrorMsg, resource.Plural(resource.Connector), invalidArgsStr), fmt.Sprintf(errors.DeleteNotFoundSuggestions, resource.Connector))
	}

	return connectorIdToName, nil
}
