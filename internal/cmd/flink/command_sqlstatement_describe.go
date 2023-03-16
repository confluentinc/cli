package flink

import (
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/spf13/cobra"
)

func (c *command) newSqlStatementDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <name>",
		Short: "Describe a Flink SQL statement.",
		RunE:  c.sqlStatementDescribe,
	}

	return cmd
}

func (c *command) sqlStatementDescribe(cmd *cobra.Command, args []string) error {
	statement, err := c.V2Client.GetSqlStatement(c.EnvironmentId(), args[0])
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&sqlStatementOut{
		Name:        statement.Spec.GetStatementName(),
		Statement:   statement.Spec.GetStatement(),
		ComputePool: statement.Spec.GetComputePoolId(),
	})
	return table.Print()
}
