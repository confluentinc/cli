package pipeline

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *command) newActivateCommand(prerunner pcmd.PreRunner) *cobra.Command {
	return &cobra.Command{
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
}

func (c *command) activate(cmd *cobra.Command, args []string) error {
	// get kafka cluster
	cluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	// call api - will be replaced with activate API when minispec is updated
	pipeline, _, err := c.V2Client.GetSdPipeline(c.EnvironmentId(), cluster.ID, args[0])
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
