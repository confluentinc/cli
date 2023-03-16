package flink

import "github.com/spf13/cobra"

type sqlStatementOut struct {
	Name        string `human:"Name" serialized:"name"`
	Statement   string `human:"Statement" serialized:"statement"`
	ComputePool string `human:"Compute Pool" serialized:"compute_pool"`
}

func (c *command) newSqlStatementCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sql-statement",
		Short: "Manage Flink SQL statements.",
	}

	cmd.AddCommand(c.newSqlStatementCreateCommand())
	cmd.AddCommand(c.newSqlStatementDeleteCommand())
	cmd.AddCommand(c.newSqlStatementDescribeCommand())
	cmd.AddCommand(c.newSqlStatementListCommand())

	return cmd
}
