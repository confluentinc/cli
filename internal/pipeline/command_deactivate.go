package pipeline

import (
	"github.com/spf13/cobra"

	streamdesignerv1 "github.com/confluentinc/ccloud-sdk-go-v2/stream-designer/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/kafka"
)

func (c *command) newDeactivateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "deactivate <pipeline-id>",
		Short:             "Request to deactivate a pipeline.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.deactivate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Request to deactivate Stream Designer pipeline "pipe-12345" with 3 retained topics.`,
				Code: `confluent pipeline deactivate pipe-12345 --retained-topics topic1,topic2,topic3`,
			},
		),
	}

	cmd.Flags().StringSlice("retained-topics", []string{}, "A comma-separated list of topics to be retained after deactivation.")
	pcmd.AddOutputFlag(cmd)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	return cmd
}

func (c *command) deactivate(cmd *cobra.Command, args []string) error {
	retainedTopics, err := cmd.Flags().GetStringSlice("retained-topics")
	if err != nil {
		return err
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	cluster, err := kafka.GetClusterForCommand(c.V2Client, c.Context)
	if err != nil {
		return err
	}

	pipeline := streamdesignerv1.SdV1Pipeline{Spec: &streamdesignerv1.SdV1PipelineSpec{
		RetainedTopicNames: &retainedTopics,
		Activated:          streamdesignerv1.PtrBool(false),
		Environment:        &streamdesignerv1.ObjectReference{Id: environmentId},
		KafkaCluster:       &streamdesignerv1.ObjectReference{Id: cluster.ID},
	}}

	pipeline, err = c.V2Client.UpdateSdPipeline(args[0], pipeline)
	if err != nil {
		return err
	}

	return printTable(cmd, pipeline)
}
