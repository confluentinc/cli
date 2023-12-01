package pipeline

import (
	"os"
	"strings"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/kafka"
	"github.com/confluentinc/cli/v3/pkg/output"
)

const sqlFileTemplate = "./<pipeline-id>.sql"

func (c *command) newSaveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "save <id>",
		Short:             "Save the source code of a Stream Designer pipeline locally.",
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
	}

	cmd.Flags().String("sql-file", sqlFileTemplate, `Path to save the pipeline's source code at.`)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagFilename("sql-file", "sql"))

	return cmd
}

func (c *command) save(cmd *cobra.Command, args []string) error {
	cluster, err := kafka.GetClusterForCommand(c.V2Client, c.Context)
	if err != nil {
		return err
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	pipeline, err := c.V2Client.GetSdPipeline(environmentId, cluster.ID, args[0])
	if err != nil {
		return err
	}

	path := args[0] + ".sql"

	sqlFile, err := cmd.Flags().GetString("sql-file")
	if err != nil {
		return err
	}
	if sqlFile != "" && sqlFile != sqlFileTemplate {
		path = expandHomeDir(sqlFile)
	}

	if err := os.WriteFile(path, []byte(pipeline.Spec.SourceCode.GetSql()), 0644); err != nil {
		return err
	}

	output.Printf(c.Config.EnableColor, "Saved source code for pipeline \"%s\" at \"%s\".\n", args[0], path)
	return nil
}

func expandHomeDir(path string) string {
	if strings.HasPrefix(path, "~") {
		if home, err := os.UserHomeDir(); err == nil {
			return strings.Replace(path, "~", home, 1)
		}
	}

	return path
}
