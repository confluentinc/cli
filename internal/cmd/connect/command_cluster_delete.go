package connect

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *clusterCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id>",
		Short:             "Delete a connector.",
		Args:              cobra.ExactArgs(1),
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
	clusterId := args[0]
	kafkaCluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	connector, err := c.V2Client.GetConnectorExpansionById(clusterId, c.EnvironmentId(), kafkaCluster.ID)
	if err != nil {
		return err
	}

	promptMsg := fmt.Sprintf(errors.DeleteResourceConfirmMsg, resource.Connector, clusterId, connector.Info.GetName())
	if _, err := form.ConfirmDeletion(cmd, promptMsg, connector.Info.GetName()); err != nil {
		return err
	}

	if _, err := c.V2Client.DeleteConnector(connector.Info.GetName(), c.EnvironmentId(), kafkaCluster.ID); err != nil {
		return err
	}

	utils.Printf(errors.DeletedResourceMsg, resource.Connector, clusterId)
	return nil
}
