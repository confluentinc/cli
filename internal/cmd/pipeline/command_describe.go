package pipeline

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *command) newDescribeCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <pipeline-id>",
		Short: "Describe a Stream Designer pipeline.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.describe,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Describe Stream Designer pipeline "pipe-12345".`,
				Code: `confluent pipeline describe pipe-12345`,
			},
		),
	}

	pcmd.AddOutputFlag(cmd)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	return cmd
}

func (c *command) describe(cmd *cobra.Command, args []string) error {
	cluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	// call api
	pipeline, err := c.V2Client.GetSdPipeline(c.EnvironmentId(), cluster.ID, args[0])
	if err != nil {
		return err
	}

	element := &Pipeline{
		Id:          *pipeline.Id,
		Name:        *pipeline.Spec.DisplayName,
		Description: *pipeline.Spec.Description,
		KsqlCluster: *&pipeline.Spec.KsqlCluster.Id,
		State:       *pipeline.Status.State,
		CreatedAt:   *pipeline.Metadata.CreatedAt,
		UpdatedAt:   *pipeline.Metadata.UpdatedAt,
	}

	return output.DescribeObject(cmd, element, pipelineDescribeFields, pipelineDescribeHumanLabels, pipelineDescribeStructuredLabels)
}
