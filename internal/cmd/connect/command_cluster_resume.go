package connect

import (
	connectv1 "github.com/confluentinc/ccloud-sdk-go-v2/connect/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *clusterCommand) newResumeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "resume <id-1> [id-2] ... [id-N]",
		Short:             "Resume connectors.",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.resume,
		Annotations:       map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Resume connectors "lcc-000001" and "lcc-000002":`,
				Code: "confluent connect cluster resume lcc-000001 lcc-000002",
			},
		),
	}

	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	return cmd
}

func (c *clusterCommand) resume(cmd *cobra.Command, args []string) error {
	kafkaCluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	connectorsByName, err := c.V2Client.ListConnectorsWithExpansions(c.EnvironmentId(), kafkaCluster.ID, "id,info")
	if err != nil {
		return err
	}

	connectorsById := make(map[string]connectv1.ConnectV1ConnectorExpansion)
	for _, connector := range connectorsByName {
		connectorsById[connector.Id.GetId()] = connector
	}

	for _, id := range args {
		connector, ok := connectorsById[id]
		if !ok {
			return errors.Errorf(errors.UnknownConnectorIdErrorMsg, id)
		}

		if err := c.V2Client.ResumeConnector(connector.Info.GetName(), c.EnvironmentId(), kafkaCluster.ID); err != nil {
			return err
		}

		utils.Printf(cmd, errors.ResumedConnectorMsg, id)
	}

	return nil
}
