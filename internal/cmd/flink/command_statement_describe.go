package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *command) newStatementGetDetailsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <name>",
		Short:             "Get details of a Flink SQL statement.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validStatementArgs),
		RunE:              c.statementDescribe,
	}

	c.addComputePoolFlag(cmd)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)

	return cmd
}

func (c *command) statementDescribe(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	client, err := c.GetFlinkGatewayClient()
	if err != nil {
		return err
	}

	statement, err := client.GetStatement(environmentId, args[0], c.Context.LastOrgId)
	if err != nil {
		return err
	}
	list := output.NewList(cmd)
	list.Add(&statementOut{
		CreationDate: statement.Metadata.GetCreatedAt(),
		Name:         statement.Spec.GetStatementName(),
		Statement:    statement.Spec.GetStatement(),
		ComputePool:  statement.Spec.GetComputePoolId(),
		Status:       statement.Status.GetPhase(),
		StatusDetail: statement.Status.GetDetail(),
	})

	return list.Print()
}
