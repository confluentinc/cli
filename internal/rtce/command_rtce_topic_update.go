package rtce

import (
	"github.com/spf13/cobra"

	rtcev1 "github.com/confluentinc/ccloud-sdk-go-v2/rtce/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/kafka"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *rtceTopicCommand) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "update <topic-name>",
		Short:             "Update a rtce topic.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.update,
	}

	// Optional: flags for updatable attributes only
	cmd.Flags().String("description", "", "A model-readable description of the RTCE topic.")
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *rtceTopicCommand) update(cmd *cobra.Command, args []string) error {
	topicName := args[0]

	update := rtcev1.RtceV1RtceTopicUpdate{}
	specUpdate := rtcev1.RtceV1RtceTopicSpecUpdate{}
	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return err
	}
	if description != "" {
		specUpdate.Description = rtcev1.PtrString(description)
	}
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}
	if environmentId != "" {
		specUpdate.Environment = &rtcev1.EnvScopedObjectReference{Id: environmentId}
	}
	kafkaClusterConfig, err := kafka.GetClusterForCommand(c.V2Client, c.Context)
	if err != nil {
		return err
	}
	kafkaClusterId := kafkaClusterConfig.GetId()
	if kafkaClusterId != "" {
		specUpdate.KafkaCluster = &rtcev1.EnvScopedObjectReference{Id: kafkaClusterId}
	}
	update.Spec = &specUpdate
	rtceTopic, httpResp, err := c.V2Client.UpdateRtceTopic(topicName, update)
	if err != nil {
		return errors.CatchCCloudV2Error(err, httpResp)
	}

	output.Printf(c.Config.EnableColor, "Updated rtce topic \"%s\".\n", topicName)
	return printRtceTopic(cmd, rtceTopic)
}
