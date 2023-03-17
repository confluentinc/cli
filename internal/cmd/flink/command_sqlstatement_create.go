package flink

import (
	"github.com/spf13/cobra"

	flinkgatewayv1alpha1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/flink-gateway/v1alpha1"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *command) newSqlStatementCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <sql>",
		Short: "Create a Flink SQL statement.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.sqlStatementCreate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Create a Flink SQL statement.",
				Code: `confluent flink sql-statement create "SELECT * FROM table;"`,
			},
		),
	}

	return cmd
}

func (c *command) sqlStatementCreate(cmd *cobra.Command, args []string) error {
	statement := flinkgatewayv1alpha1.SqlV1alpha1Statement{
		Spec: &flinkgatewayv1alpha1.SqlV1alpha1StatementSpec{
			Statement: flinkgatewayv1alpha1.PtrString(args[0]),
		},
	}

	statement, err := c.V2Client.CreateSqlStatement(c.EnvironmentId(), statement)
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&sqlStatementOut{
		Name:         statement.Spec.GetStatementName(),
		Statement:    statement.Spec.GetStatement(),
		ComputePool:  statement.Spec.GetComputePoolId(),
		Status:       statement.Status.GetPhase(),
		StatusDetail: statement.Status.GetDetail(),
	})
	return table.Print()
}
