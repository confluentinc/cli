package rtce

import (
	"github.com/spf13/cobra"

	rtcev1 "github.com/confluentinc/ccloud-sdk-go-v2/rtce/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/output"
)

type rtceTopicCommand struct {
	*pcmd.AuthenticatedCLICommand
}

type rtceTopicOut struct {
	TopicName    string `human:"Topic Name" serialized:"topic_name"`
	Cloud        string `human:"Cloud" serialized:"cloud"`
	Description  string `human:"Description" serialized:"description"`
	Environment  string `human:"Environment" serialized:"environment"`
	KafkaCluster string `human:"Kafka Cluster" serialized:"kafka_cluster"`
	Region       string `human:"Region" serialized:"region"`
	ErrorMessage string `human:"Error Message" serialized:"error_message"`
	Phase        string `human:"Phase" serialized:"phase"`
}

func newRtceTopicCommand(cfg *config.Config, prerunner pcmd.PreRunner) *cobra.Command { //nolint:unparam
	cmd := &cobra.Command{
		Use:         "rtce-topic",
		Short:       "Manage rtce topics.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &rtceTopicCommand{
		AuthenticatedCLICommand: pcmd.NewAuthenticatedCLICommand(cmd, prerunner),
	}

	cmd.AddCommand(
		c.newCreateCommand(),
		c.newDeleteCommand(),
		c.newDescribeCommand(),
		c.newListCommand(),
		c.newUpdateCommand(),
	)

	return cmd
}

func printRtceTopic(cmd *cobra.Command, rtceTopic rtcev1.RtceV1RtceTopic) error {
	table := output.NewTable(cmd)
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
	table.Add(out)
	return table.Print()
}

func (c *rtceTopicCommand) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	return c.validArgsMultiple(cmd, args)
}

func (c *rtceTopicCommand) validArgsMultiple(cmd *cobra.Command, args []string) []string {
	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return c.autocompleteRtceTopics()
}

func (c *rtceTopicCommand) autocompleteRtceTopics() []string {
	rtceTopics, err := c.V2Client.ListRtceTopics("", "", "", "")
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(rtceTopics))
	for i, rtceTopic := range rtceTopics {
		suggestions[i] = rtceTopic.Spec.GetTopicName()
	}
	return suggestions
}
