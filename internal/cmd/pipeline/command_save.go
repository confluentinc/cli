package pipeline

import (
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/utils"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"io/ioutil"
)

func (c *command) newSaveCommand(prerunner pcmd.PreRunner, enableSourceCode bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "save <pipeline-id>",
		Short: "Save a Stream Designer pipeline's source code to a local file.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.save,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Save the source code for Stream Designer pipeline "pipe-12345" to "/tmp/pipeline-source-code.sql".`,
				Code: "confluent pipeline save pipe-12345 --source-code-file /tmp/pipeline-source-code.sql",
			},
		),
	}

	cmd.Flags().String("source-code-file", "", "Path to save the pipeline's source code at. (default \"./<pipeline-id>.sql\")")
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cmd.Hidden = !enableSourceCode
	return cmd
}

func (c *command) save(cmd *cobra.Command, args []string) error {
	sourceCodeFile, _ := cmd.Flags().GetString("source-code-file")

	cluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	pipeline, err := c.V2Client.GetSdPipeline(c.EnvironmentId(), cluster.ID, args[0])
	if err != nil {
		return err
	}

	path := args[0] + ".sql"
	if sourceCodeFile != "" {
		if path, err = homedir.Expand(sourceCodeFile); err != nil {
			return err
		}
	}

	if err := ioutil.WriteFile(path, []byte(pipeline.Spec.GetSourceCode()), 0644); err != nil {
		return err
	}

	utils.Printf(cmd, "Saved source code for pipeline \"%s\" at \"%s\".\n", args[0], path)
	return nil
}
