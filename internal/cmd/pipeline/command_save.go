package pipeline

import (
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/utils"
	"github.com/spf13/cobra"
	"io/ioutil"
)

func (c *command) newSaveCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "save <pipeline-id>",
		Short: "Save a Stream Designer pipeline's source code to a local file.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.save,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Save Stream Designer pipeline's source code for "pipe-12345".`,
				Code: `confluent pipeline save pipe-12345`,
			},
		),
	}

	cmd.Flags().String("source-code-file", "", "Path to save the pipeline's source code at. (default \"./<pipeline-id>.sql\")")

	pcmd.AddOutputFlag(cmd)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	return cmd
}

func (c *command) save(cmd *cobra.Command, args []string) error {
	outputFile, _ := cmd.Flags().GetString("source-code-file")

	cluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	// call api
	pipeline, err := c.V2Client.GetSdPipeline(c.EnvironmentId(), cluster.ID, args[0])
	if err != nil {
		return err
	}

	file := args[0] + ".sql"
	if outputFile != "" {
		file = outputFile
	}

	err = ioutil.WriteFile(file, []byte(pipeline.Spec.GetSourceCode()), 0644)
	if err != nil {
		return err
	}

	utils.Printf(cmd, "Pipeline's source code is now saved at '%s'.", file)
	return nil
}
