package pipeline

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *command) newDescribeCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <pipeline-id>",
		Short: "Describe a pipeline, and optionally save it locally.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.describe,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Describe a pipeline, and save the source code to current local directory.",
				Code: `confluent pipeline describe pipe-12345 --output-directory .`,
			},
		),
	}

	cmd.Flags().String("output-directory", "", "Path to save pipeline model.")

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) describe(cmd *cobra.Command, args []string) error {
	outputDirectory, _ := cmd.Flags().GetString("output-directory")

	// get kafka cluster
	cluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	// call api
	pipeline, err := c.V2Client.GetSdPipeline(c.EnvironmentId(), cluster.ID, args[0])
	if err != nil {
		return err
	}

	outputWriter, err := output.NewListOutputWriter(cmd, pipelineListFields, pipelineListHumanLabels, pipelineListStructuredLabels)
	if err != nil {
		return err
	}

	element := &Pipeline{Id: *pipeline.Id, Name: *pipeline.Spec.DisplayName, State: *pipeline.Status.State}
	outputWriter.AddElement(element)

	if outputDirectory != "" {
		// add pipelineId.sql to user chosen directory
		filepath := filepath.Join(outputDirectory, args[0]+".sql")
		// create the file at generated filepath
		out, err := os.Create(filepath)
		if err != nil {
			return err
		}

		defer out.Close()

		// replace pipeline.Name with pipeline.Spec.Sql once official minispec used
		_, err = out.Write([]byte(*pipeline.Spec.DisplayName))
		if err != nil {
			return err
		}
		utils.Printf(cmd, "Saved SQL file for pipeline %s.\n", args[0])
	}

	return outputWriter.Out()
}
