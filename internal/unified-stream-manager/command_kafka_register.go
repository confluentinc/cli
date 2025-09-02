package unifiedstreammanager

import (
	"strings"

	"github.com/spf13/cobra"

	usmv1 "github.com/confluentinc/ccloud-sdk-go-v2/usm/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
)

func (c *command) newKafkaRegisterCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register <confluent-platform-kafka-cluster-id>",
		Short: "Register a Kafka cluster.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.registerKafka,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Register a Confluent Platform Kafka cluster with the ID 4k0R9d1GTS5tI9f4Y2xZ0Q.",
				Code: "confluent unified-stream-manager kafka register 4k0R9d1GTS5tI9f4Y2xZ0Q --name my-kafka-cluster --cloud aws --region us-east-1",
			},
		),
	}

	cmd.Flags().String("name", "", "Name of the USM Kafka cluster.")
	pcmd.AddCloudFlag(cmd)
	c.addRegionFlag(cmd)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("name"))
	cobra.CheckErr(cmd.MarkFlagRequired("cloud"))
	cobra.CheckErr(cmd.MarkFlagRequired("region"))

	return cmd
}

func (c *command) registerKafka(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}

	cloud, err := cmd.Flags().GetString("cloud")
	if err != nil {
		return err
	}

	region, err := cmd.Flags().GetString("region")
	if err != nil {
		return err
	}

	cluster, err := c.V2Client.CreateUsmKafkaCluster(usmv1.UsmV1KafkaCluster{
		ConfluentPlatformKafkaClusterId: usmv1.PtrString(args[0]),
		DisplayName:                     usmv1.PtrString(name),
		Cloud:                           usmv1.PtrString(strings.ToUpper(cloud)),
		Region:                          usmv1.PtrString(region),
		Environment: &usmv1.EnvScopedObjectReference{
			Id: environmentId,
		},
	})
	if err != nil {
		return err
	}

	return printKafkaTable(cmd, cluster)
}
