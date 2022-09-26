package pipeline

import (
	"github.com/spf13/cobra"

	sdv1 "github.com/confluentinc/ccloud-sdk-go-v2/stream-designer/v1"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *command) newActivateCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "activate <pipeline-id>",
		Short: "Request to activate a pipeline.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.activate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Request to activate a pipeline in Stream Designer",
				Code: `confluent pipeline activate pipe-12345`,
			},
		),
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) activate(cmd *cobra.Command, args []string) error {
	// get kafka cluster
	cluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	updateBody := sdv1.NewSdV1PipelineUpdate()
	updateBody.SetActivated(true)
	
	// call api (Current update API does not support this yet)
	pipeline, err := c.V2Client.UpdateSdPipeline(c.EnvironmentId(), cluster.ID, args[0], *updateBody)
	if err != nil {
		return err
	}

	outputWriter, err := output.NewListOutputWriter(cmd, pipelineListFields, pipelineListHumanLabels, pipelineListStructuredLabels)
	if err != nil {
		return err
	}

	// *pipeline.state will be activating
	element := &Pipeline{Id: *pipeline.Id, Name: *pipeline.Name, State: *pipeline.State}
	outputWriter.AddElement(element)

	return outputWriter.Out()
}
