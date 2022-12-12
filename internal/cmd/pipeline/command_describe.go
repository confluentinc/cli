package pipeline

import (
	"github.com/spf13/cobra"
	"io/ioutil"
	"path/filepath"

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

	cmd.Flags().Bool("save-source-code", false, "Save the pipeline source code in a local file with name as pipeline_id.sql.")
	cmd.Flags().String("output-directory", "./", "Path to save pipeline source code.")

	pcmd.AddOutputFlag(cmd)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	return cmd
}

func (c *command) describe(cmd *cobra.Command, args []string) error {
	saveSource, _ := cmd.Flags().GetBool("save-source-code")
	outputDir, _ := cmd.Flags().GetString("output-directory")

	cluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	// call api
	pipeline, err := c.V2Client.GetSdPipeline(c.EnvironmentId(), cluster.ID, args[0])
	if err != nil {
		return err
	}

	if saveSource {
		file := args[0] + ".sql"

		if outputDir != "" {
			file = filepath.Join(outputDir, file)
		}

		err = ioutil.WriteFile(file, []byte(pipeline.Spec.GetSourceCode()), 0644)
		if err != nil {
			return err
		}
	}

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
