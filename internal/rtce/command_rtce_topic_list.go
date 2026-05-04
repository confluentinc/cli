package rtce

import (
	"strings"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/kafka"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *rtceTopicCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List rtce topics.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}
	pcmd.AddCloudAwsFlag(cmd)
	cmd.Flags().String("region", "", "Filter the results by exact match for spec.region.")
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *rtceTopicCommand) list(cmd *cobra.Command, _ []string) error {
	cloud, err := cmd.Flags().GetString("cloud")
	if err != nil {
		return err
	}
	cloud = strings.ToUpper(cloud)
	region, err := cmd.Flags().GetString("region")
	if err != nil {
		return err
	}
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}
	kafkaClusterConfig, err := kafka.GetClusterForCommand(c.V2Client, c.Context)
	if err != nil {
		return err
	}
	kafkaClusterId := kafkaClusterConfig.GetId()

	rtceTopics, err := c.V2Client.ListRtceTopics(cloud, region, environmentId, kafkaClusterId)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, rtceTopic := range rtceTopics {
		out := &rtceTopicOut{
			TopicName:    rtceTopic.Spec.GetTopicName(),
			Cloud:        rtceTopic.Spec.GetCloud(),
			Description:  rtceTopic.Spec.GetDescription(),
			Environment:  rtceTopic.Spec.Environment.GetId(),
			KafkaCluster: rtceTopic.Spec.KafkaCluster.GetId(),
			Region:       rtceTopic.Spec.GetRegion(),
			ErrorMessage: rtceTopic.Status.GetErrorMessage(),
			Phase:        rtceTopic.Status.GetPhase(),
		}
		list.Add(out)
	}
	return list.Print()
}
