package rtce

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/kafka"
)

func (c *rtceTopicCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <topic-name>",
		Short:             "Describe a rtce topic.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.describe,
	}
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *rtceTopicCommand) describe(cmd *cobra.Command, args []string) error {
	topicName := args[0]
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}
	kafkaClusterConfig, err := kafka.GetClusterForCommand(c.V2Client, c.Context)
	if err != nil {
		return err
	}
	kafkaClusterId := kafkaClusterConfig.GetId()

	rtceTopic, httpResp, err := c.V2Client.GetRtceTopic(topicName, environmentId, kafkaClusterId)
	if err != nil {
		return errors.CatchCCloudV2Error(err, httpResp)
	}

	return printRtceTopic(cmd, rtceTopic)
}
