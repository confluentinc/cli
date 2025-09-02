package unifiedstreammanager

import (
	"strings"

	"github.com/spf13/cobra"

	usmv1 "github.com/confluentinc/ccloud-sdk-go-v2/usm/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/log"
)

func (c *command) newConnectRegisterCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register <confluent-platform-connect-cluster-id>",
		Short: "Register a Connect cluster.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.registerConnect,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Register a Confluent Platform Connect cluster with the ID connect-group-xyz123.",
				Code: "confluent unified-stream-manager connect register connect-group-xyz123 --confluent-platform-kafka-cluster 4k0R9d1GTS5tI9f4Y2xZ0Q --cloud aws --region us-east-1",
			},
		),
	}

	cmd.Flags().String("confluent-platform-kafka-cluster", "", "The ID of the metadata Kafka cluster for the Connect Cluster.")
	pcmd.AddCloudFlag(cmd)
	c.addRegionFlag(cmd)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("confluent-platform-kafka-cluster"))
	cmd.MarkFlagsRequiredTogether("cloud", "region")

	return cmd
}

func (c *command) registerConnect(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	onPremToCloudKafkaIdMap, err := c.getOnPremToCloudKafkaIdMap(environmentId)
	if err != nil {
		return err
	}

	kafkaClusterId, err := cmd.Flags().GetString("confluent-platform-kafka-cluster")
	if err != nil {
		return err
	}

	usmKafkaClusterId, ok := onPremToCloudKafkaIdMap[kafkaClusterId]
	if !ok {
		log.CliLogger.Errorf(kafkaClusterNotFoundErrorMsg, kafkaClusterId)
	}

	connectClusterRequest := usmv1.UsmV1ConnectCluster{
		ConfluentPlatformConnectClusterId: usmv1.PtrString(args[0]),
		KafkaClusterId:                    usmv1.PtrString(kafkaClusterId),
		Environment: &usmv1.EnvScopedObjectReference{
			Id: environmentId,
		},
	}

	if cmd.Flags().Changed("cloud") { // Cloud and region are marked as required together, so we only need to check for cloud
		cloud, err := cmd.Flags().GetString("cloud")
		if err != nil {
			return err
		}
		connectClusterRequest.SetCloud(strings.ToUpper(cloud))

		region, err := cmd.Flags().GetString("region")
		if err != nil {
			return err
		}
		connectClusterRequest.SetRegion(region)
	}

	cluster, err := c.V2Client.CreateUsmConnectCluster(connectClusterRequest)
	if err != nil {
		return err
	}

	return printConnectTable(cmd, cluster, usmKafkaClusterId)
}
