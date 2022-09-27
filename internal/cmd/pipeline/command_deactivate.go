package pipeline

import (
	sdv1 "github.com/confluentinc/ccloud-sdk-go-v2/stream-designer/v1"
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
				Text: "Request to deactivate a pipeline in Stream Designer",
				Code: `confluent pipeline deactivate pipe-12345`,
			},
		),
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) deactivate(cmd *cobra.Command, args []string) error {
	cluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	updatePipeline := sdv1.SdV1PipelineUpdate{
		Spec: &sdv1.SdV1PipelineSpecUpdate{},
	}
	updatePipeline.Spec.SetActivated(false)

	// call api (Current update API does not support this yet)
	pipeline, err := c.V2Client.UpdateSdPipeline(c.EnvironmentId(), cluster.ID, args[0], updatePipeline)
	if err != nil {
		return err
	}

	outputWriter, err := output.NewListOutputWriter(cmd, pipelineListFields, pipelineListHumanLabels, pipelineListStructuredLabels)
	if err != nil {
		return err
	}

	// *pipeline.state will be deactivating
	element := &Pipeline{Id: *pipeline.Id, Name: *pipeline.Spec.DisplayName, State: *pipeline.Status.State}
	outputWriter.AddElement(element)

	return outputWriter.Out()
}
