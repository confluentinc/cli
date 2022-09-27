package pipeline

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *command) newListCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Display pipelines in the current environment and cluster.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

// todo: this will return error for now, because our API server is not returning the conventional response
//       object yet, hence you will see error message as below. Official API minispec is coming soon.
//       'Error: json: cannot unmarshal array into Go value of type v1.SdV1PipelineList'
func (c *command) list(cmd *cobra.Command, _ []string) error {
	cluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	// call api
	pipelines, err := c.V2Client.ListPipelines(c.EnvironmentId(), cluster.ID)
	if err != nil {
		return err
	}

	outputWriter, err := output.NewListOutputWriter(cmd, pipelineListFields, pipelineListHumanLabels, pipelineListStructuredLabels)
	if err != nil {
		return err
	}
	for _, pipeline := range pipelines {
		element := &Pipeline{Id: *pipeline.Id, Name: *pipeline.Spec.DisplayName, State: *pipeline.Status.State}
		outputWriter.AddElement(element)
	}
	return outputWriter.Out()
}
