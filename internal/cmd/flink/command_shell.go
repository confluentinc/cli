package flink

import (
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	client "github.com/confluentinc/flink-sql-client"
	"github.com/spf13/cobra"
)

func (c *command) newStartFlinkSqlClientCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "shell",
		Short: "Start Flink interactive SQL client.",
		RunE:  c.startFlinkSqlClient,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Start Flink interactive SQL client.",
				Code: `confluent flink shell"`,
			},
		),
	}

	cmd.Flags().String("compute-pool", "", "Flink compute pool ID.")
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) startFlinkSqlClient(cmd *cobra.Command, args []string) error {
	computePool, err := cmd.Flags().GetString("compute-pool")
	if err != nil {
		return err
	}

	client.StartApp(c.EnvironmentId(), computePool, c.AuthToken(), nil)

	return nil
}
