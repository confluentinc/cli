package pipeline

import (
	"io/ioutil"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *command) newSaveCommand(prerunner pcmd.PreRunner, enableSourceCode bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "save <pipeline-id>",
		Short: "Save a Stream Designer pipeline's source code to a local file.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.save,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Save the source code for Stream Designer pipeline "pipe-12345" to the default file at "./pipe-12345.sql".`,
				Code: "confluent pipeline save pipe-12345",
			},
			examples.Example{
				Text: `Save the source code for Stream Designer pipeline "pipe-12345" to "/tmp/pipeline-source-code.sql".`,
				Code: "confluent pipeline save pipe-12345 --source-code-sql /tmp/pipeline-source-code.sql",
			},
		),
		Hidden: !enableSourceCode,
	}

	cmd.Flags().String("source-code-sql", "", "Path to save the pipeline's source code at. (default \"./<pipeline-id>.sql\")")
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) save(cmd *cobra.Command, args []string) error {
	sourceCodeSql, _ := cmd.Flags().GetString("source-code-sql")

	cluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	pipeline, err := c.V2Client.GetSdPipeline(c.EnvironmentId(), cluster.ID, args[0])
	if err != nil {
		return err
	}

	path := args[0] + ".sql"
	if sourceCodeSql != "" {
		if path, err = homedir.Expand(sourceCodeSql); err != nil {
			path = sourceCodeSql
		}
	}

	if err := ioutil.WriteFile(path, []byte(pipeline.Spec.SourceCode.GetSql()), 0644); err != nil {
		return err
	}

	utils.Printf(cmd, "Saved source code for pipeline \"%s\" at \"%s\".\n", args[0], path)
	return nil
}
