package pipeline

import (
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
	// get kafka cluster
	cluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	updateBody := sdv1.NewSdV1PipelineUpdate()
	updateBody.SetActivated(false)
	
	// call api
	pipeline, err := c.V2Client.UpdateSdPipeline(c.EnvironmentId(), cluster.ID, args[0], *updateBody)
	if err != nil {
		return err
	}

	outputWriter, err := output.NewListOutputWriter(cmd, pipelineListFields, pipelineListHumanLabels, pipelineListStructuredLabels)
	if err != nil {
		return err
	}

	// *pipeline.state will be deactivating
	element := &Pipeline{Id: *pipeline.Id, Name: *pipeline.Name, State: *pipeline.State}
	outputWriter.AddElement(element)

	return outputWriter.Out()
}
