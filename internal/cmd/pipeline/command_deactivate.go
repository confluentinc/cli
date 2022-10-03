package pipeline

import (
	streamdesignerv1 "github.com/confluentinc/ccloud-sdk-go-v2/stream-designer/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *command) newDeactivateCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deactivate <pipeline-id>",
		Short: "Request to deactivate a pipeline.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.deactivate,
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
	retainedTopics, _ := cmd.Flags().GetStringSlice("retained-topics")

	cluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	updatePipeline := streamdesignerv1.SdV1PipelineUpdate{
		Spec: &streamdesignerv1.SdV1PipelineSpecUpdate{
			Activated:          streamdesignerv1.PtrBool(false),
			RetainedTopicNames: &retainedTopics,
		},
	}

	pipeline, err := c.V2Client.UpdateSdPipeline(c.EnvironmentId(), cluster.ID, args[0], updatePipeline)
	if err != nil {
		return err
	}

	// *pipeline.state will be deactivating
	element := &Pipeline{
		Id:          *pipeline.Id,
		Name:        *pipeline.Spec.DisplayName,
		Description: *pipeline.Spec.Description,
		KsqlCluster: pipeline.Spec.KsqlCluster.Id,
		State:       *pipeline.Status.State,
		CreatedAt:   *pipeline.Metadata.CreatedAt,
		UpdatedAt:   *pipeline.Metadata.UpdatedAt,
	}

	return output.DescribeObject(cmd, element, pipelineDescribeFields, pipelineDescribeHumanLabels, pipelineDescribeStructuredLabels)
}
