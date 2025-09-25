package unifiedstreammanager

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/log"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newConnectListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Connect clusters.",
		Args:  cobra.NoArgs,
		RunE:  c.listConnect,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List Connect clusters.",
				Code: "confluent unified-stream-manager connect list",
			},
		),
	}

	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) listConnect(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	onPremToCloudKafkaIdMap, err := c.getOnPremToCloudKafkaIdMap(environmentId)
	if err != nil {
		return err
	}

	clusters, err := c.V2Client.ListUsmConnectClusters(environmentId)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, cluster := range clusters {
		usmKafkaClusterId, ok := onPremToCloudKafkaIdMap[cluster.GetKafkaClusterId()]
		if !ok {
			log.CliLogger.Errorf(kafkaClusterNotFoundErrorMsg, cluster.GetKafkaClusterId())
		}

		out := &connectOut{
			Id:                              cluster.GetId(),
			ConfluentPlatformConnectCluster: cluster.GetConfluentPlatformConnectClusterId(),
			USMKafkaClusterId:               usmKafkaClusterId,
			ConfluentPlatformKafkaClusterId: cluster.GetKafkaClusterId(),
			Cloud:                           cluster.GetCloud(),
			Region:                          cluster.GetRegion(),
			Environment:                     cluster.Environment.GetId(),
		}

		list.Add(out)
	}

	return list.Print()
}
