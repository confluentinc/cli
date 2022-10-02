package pipeline

import (
	"fmt"
	"github.com/spf13/cobra"

	streamdesignerv1 "github.com/confluentinc/ccloud-sdk-go-v2/stream-designer/v1"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *command) newUpdateCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <pipeline-id>",
		Short: "Update an existing pipeline.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.update,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Request to update Stream Designer pipeline "pipe-12345", with new name and new description.`,
				Code: `confluent pipeline update pipe-12345 --name test-pipeline --description "Description of the pipeline"`,
			},
		),
	}

	cmd.Flags().String("name", "", "Name of the pipeline.")
	cmd.Flags().String("description", "", "Description of the pipeline.")

	pcmd.AddOutputFlag(cmd)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	return cmd
}

func (c *command) update(cmd *cobra.Command, args []string) error {
	name, _ := cmd.Flags().GetString("name")
	description, _ := cmd.Flags().GetString("description")

	cluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	if name == "" && description == "" {
		return fmt.Errorf("one of `--name` or `--description` must be provided")
	}

	updatePipeline := streamdesignerv1.SdV1PipelineUpdate{Spec: &streamdesignerv1.SdV1PipelineSpecUpdate{}}
	if name != "" {
		updatePipeline.Spec.SetDisplayName(name)
	}
	if description != "" {
		updatePipeline.Spec.SetDescription(description)
	}

	// call api
	pipeline, err := c.V2Client.UpdateSdPipeline(c.EnvironmentId(), cluster.ID, args[0], updatePipeline)
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
