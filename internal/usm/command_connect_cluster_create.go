package usm

import (
	"strings"

	"github.com/spf13/cobra"

	usmv1 "github.com/confluentinc/ccloud-sdk-go-v2/usm/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/examples"
)

func (c *connectClusterCommand) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create <confluent-platform-connect-cluster-id>",
		Short:   "Create a USM Connect cluster.",
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"register"},
		RunE:    c.create,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Register a Confluent Platform Connect cluster with the ID connect-group-xyz123.",
				Code: "confluent usm connect-cluster create connect-group-xyz123 --confluent-platform-kafka-cluster 4k0R9d1GTS5tI9f4Y2xZ0Q --cloud aws --region us-east-1",
			},
		),
	}

	// Kafka cluster flags (exactly one required, mutually exclusive)
	cmd.Flags().String("confluent-platform-kafka-cluster", "", "The unique identifier of the metadata Kafka cluster for the Connect Cluster.")
	cmd.Flags().String("kafka-cluster", "", "The unique identifier of the metadata Kafka cluster for the Connect Cluster.")

	// Optional flags
	pcmd.AddCloudFlag(cmd)
	cmd.Flags().String("region", "", "The home region of the Confluent Platform Connect cluster where the metadata should be stored. This field is optional. If provided, 'cloud' must also be provided. If neither 'cloud' nor 'region' are provided, the home region of the associated metadata Kafka cluster (identified by 'kafka_cluster_id') will be used as a fallback.")
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cmd.MarkFlagsOneRequired("confluent-platform-kafka-cluster", "kafka-cluster")
	cmd.MarkFlagsMutuallyExclusive("confluent-platform-kafka-cluster", "kafka-cluster")

	return cmd
}

func (c *connectClusterCommand) create(cmd *cobra.Command, args []string) error {
	arg := args[0]

	createReq := usmv1.UsmV1ConnectCluster{}

	createReq.ConfluentPlatformConnectClusterId = usmv1.PtrString(arg)
	kafkaClusterId, err := cmd.Flags().GetString("confluent-platform-kafka-cluster")
	if err != nil {
		return err
	}
	if cmd.Flags().Changed("kafka-cluster") {
		kafkaClusterId, err = cmd.Flags().GetString("kafka-cluster")
		if err != nil {
			return err
		}
	}
	createReq.KafkaClusterId = usmv1.PtrString(kafkaClusterId)

	cloud, err := cmd.Flags().GetString("cloud")
	if err != nil {
		return err
	}
	cloud = strings.ToUpper(cloud)
	if cloud != "" {
		createReq.Cloud = usmv1.PtrString(cloud)
	}

	region, err := cmd.Flags().GetString("region")
	if err != nil {
		return err
	}
	if region != "" {
		createReq.Region = usmv1.PtrString(region)
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}
	createReq.Environment = &usmv1.EnvScopedObjectReference{Id: environmentId}

	connectCluster, httpResp, err := c.V2Client.CreateUsmConnectCluster(createReq)
	if err != nil {
		return errors.CatchCCloudV2Error(err, httpResp)
	}

	return printConnectCluster(cmd, connectCluster)
}
