package flink

import (
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/spf13/cobra"
)

func (c *command) newSqlStatementListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Flink SQL statements.",
		RunE:  c.sqlStatementCreate,
	}

	return cmd
}

func (c *command) sqlStatementList(cmd *cobra.Command, args []string) error {
	statements, err := c.V2Client.ListSqlStatements(c.EnvironmentId())
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, statement := range statements {
		list.Add(&sqlStatementOut{
			Name:        statement.Spec.GetStatementName(),
			Statement:   statement.Spec.GetStatement(),
			ComputePool: statement.Spec.GetComputePoolId(),
		})
	}
	return list.Print()
}
