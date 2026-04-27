package rtce

import (
	"strings"

	"github.com/spf13/cobra"

	rtcev1 "github.com/confluentinc/ccloud-sdk-go-v2/rtce/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/kafka"
)

func (c *rtceTopicCommand) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a rtce topic.",
		Args:  cobra.NoArgs,
		RunE:  c.create,
	}

	// Flags derived from OpenAPI attributes:
	// - optional fields become flags
	// - required fields are marked required after context/output flags
	pcmd.AddCloudAwsFlag(cmd)
	cmd.Flags().String("description", "", "A model-readable description of the RTCE topic.")
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().String("region", "", "The cloud region where the RTCE topic is deployed.")
	cmd.Flags().String("topic-name", "", "The Kafka topic name containing the data for the RTCE topic.")

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)
	cobra.CheckErr(cmd.MarkFlagRequired("cloud"))
	cobra.CheckErr(cmd.MarkFlagRequired("description"))
	cobra.CheckErr(cmd.MarkFlagRequired("region"))
	cobra.CheckErr(cmd.MarkFlagRequired("topic-name"))

	return cmd
}

func (c *rtceTopicCommand) create(cmd *cobra.Command, args []string) error {
	// Build request model from flags (v0).
	spec := rtcev1.RtceV1RtceTopicSpec{}
	req := rtcev1.RtceV1RtceTopic{}
	cloud, err := cmd.Flags().GetString("cloud")
	if err != nil {
		return err
	}
	cloud = strings.ToUpper(cloud)
	spec.Cloud = rtcev1.PtrString(cloud)
	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return err
	}
	spec.Description = rtcev1.PtrString(description)
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}
	spec.Environment = &rtcev1.EnvScopedObjectReference{Id: environmentId}
	kafkaClusterConfig, err := kafka.GetClusterForCommand(c.V2Client, c.Context)
	if err != nil {
		return err
	}
	kafkaClusterId := kafkaClusterConfig.GetId()
	spec.KafkaCluster = &rtcev1.EnvScopedObjectReference{Id: kafkaClusterId}
	region, err := cmd.Flags().GetString("region")
	if err != nil {
		return err
	}
	spec.Region = rtcev1.PtrString(region)
	topicName, err := cmd.Flags().GetString("topic-name")
	if err != nil {
		return err
	}
	spec.TopicName = rtcev1.PtrString(topicName)
	req.Spec = &spec
	rtceTopic, httpResp, err := c.V2Client.CreateRtceTopic(req)
	if err != nil {
		return errors.CatchCCloudV2Error(err, httpResp)
	}

	return printRtceTopic(cmd, rtceTopic)
}
