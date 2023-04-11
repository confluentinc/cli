package pipeline

import (
	"os"
	"strings"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *command) newSaveCommand(enableSourceCode bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "save <pipeline-id>",
		Short:             "Save a Stream Designer pipeline's source code to a local file.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.save,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Save the source code for Stream Designer pipeline "pipe-12345" to the default file at "./pipe-12345.sql".`,
				Code: "confluent pipeline save pipe-12345",
			},
			examples.Example{
				Text: `Save the source code for Stream Designer pipeline "pipe-12345" to "/tmp/pipeline-source-code.sql".`,
				Code: "confluent pipeline save pipe-12345 --sql-file /tmp/pipeline-source-code.sql",
			},
		),
		Hidden: !enableSourceCode,
	}

	cmd.Flags().String("sql-file", "", "Path to save the pipeline's source code at. (default \"./<pipeline-id>.sql\")")
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagFilename("sql-file", "sql"))

	return cmd
}

func (c *command) save(cmd *cobra.Command, args []string) error {
	cluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	environmentId, err := c.EnvironmentId()
	if err != nil {
		return err
	}

	pipeline, err := c.V2Client.GetSdPipeline(environmentId, cluster.ID, args[0])
	if err != nil {
		return err
	}

	path := args[0] + ".sql"

	sqlFile, _ := cmd.Flags().GetString("sql-file")
	if sqlFile != "" {
		path = getPath(sqlFile)
	}

	if err := os.WriteFile(path, []byte(pipeline.Spec.SourceCode.GetSql()), 0644); err != nil {
		return err
	}

	output.Printf("Saved source code for pipeline \"%s\" at \"%s\".\n", args[0], path)
	return nil
}

func getPath(sqlFile string) string {
	if strings.HasPrefix(sqlFile, "~") {
		if home, err := os.UserHomeDir(); err == nil {
			return strings.Replace(sqlFile, "~", home, 1)
		}
	}
	return sqlFile
}
