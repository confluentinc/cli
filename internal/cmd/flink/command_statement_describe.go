package flink

import (
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/spf13/cobra"
)

func (c *command) newStatementDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <name>",
		Short: "Describe a Flink SQL statement.",
		RunE:  c.statementDescribe,
	}

	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) statementDescribe(cmd *cobra.Command, args []string) error {
	environment, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	statement, err := c.V2Client.GetStatement(environment, args[0])
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&statementOut{
		Name:         statement.Spec.GetStatementName(),
		Statement:    statement.Spec.GetStatement(),
		ComputePool:  statement.Spec.GetComputePoolId(),
		Status:       statement.Status.GetPhase(),
		StatusDetail: statement.Status.GetDetail(),
	})
	return table.Print()
}
