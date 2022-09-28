package pipeline

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	sdv1 "github.com/confluentinc/ccloud-sdk-go-v2/stream-designer/v1"
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
				Text: `Request to update Stream Designer pipeline "pipe-12345", with new name, description, and source code located in current directory.`,
				Code: `confluent pipeline update pipe-12345 --name "NewPipeline" -- description "NewDescription" -- sql-file`,
			},
		),
	}

	cmd.Flags().String("name", "", "New pipeline name.")
	cmd.Flags().String("description", "", "New pipeline description.")
	cmd.Flags().String("sql-file", "", "Path to the new pipeline model file.")

	pcmd.AddOutputFlag(cmd)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	return cmd
}

func (c *command) update(cmd *cobra.Command, args []string) error {
	name, _ := cmd.Flags().GetString("name")
	description, _ := cmd.Flags().GetString("description")
	sqlFile, _ := cmd.Flags().GetString("sql-file")

	cluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	if name == "" && description == "" && sqlFile == "" {
		return fmt.Errorf("At least one field must be specified with --name, --description, or --sql-file")
	}

	updatePipeline := sdv1.SdV1PipelineUpdate{
		Spec: &sdv1.SdV1PipelineSpecUpdate{},
	}
	if name != "" {
		updatePipeline.Spec.SetDisplayName(name)
	}
	if description != "" {
		updatePipeline.Spec.SetDescription(description)
	}
	if sqlFile != "" {
		// get SQL content from file
		sqlData, err := os.ReadFile(sqlFile)
		if err != nil {
			return err
		}

		updatePipeline.Spec.SetSourceCode(string(sqlData))
	}

	// call api
	pipeline, err := c.V2Client.UpdateSdPipeline(c.EnvironmentId(), cluster.ID, args[0], updatePipeline)
	if err != nil {
		return err
	}

	outputWriter, err := output.NewListOutputWriter(cmd, pipelineListFields, pipelineListHumanLabels, pipelineListStructuredLabels)
	if err != nil {
		return err
	}

	element := &Pipeline{Id: *pipeline.Id, Name: *pipeline.Spec.DisplayName, State: *pipeline.Status.State}
	outputWriter.AddElement(element)

	return outputWriter.Out()
}
