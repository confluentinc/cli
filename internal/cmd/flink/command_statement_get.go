package flink

import (
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/spf13/cobra"
)

func (c *command) newStatementGetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "get <name>",
		Short:             "Get details of a Flink SQL statement.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validStatementArgs),
		RunE:              c.statementGet,
	}

	pcmd.AddForceFlag(cmd)
	c.addComputePoolFlag(cmd)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)

	return cmd
}

func (c *command) statementGet(cmd *cobra.Command, args []string) error {
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
	output.Printf("Details for statement %s\n\n", args[0])
	list := output.NewList(cmd)
	list.Add(&statementOut{
		CreationDate: statement.Metadata.GetCreatedAt(),
		Name:         statement.Spec.GetStatementName(),
		Statement:    statement.Spec.GetStatement(),
		ComputePool:  statement.Spec.GetComputePoolId(),
		Status:       statement.Status.GetPhase(),
		StatusDetail: statement.Status.GetDetail(),
	})

	_ = list.Print()

	orgId := c.Context.GetCurrentOrganization()

	list = output.NewList(cmd)
	exceptions, err := client.GetExceptions(environmentId, args[0], orgId)
	if err != nil {
		return err
	}

	if len(exceptions.GetData()) > 0 {
		list = output.NewList(cmd)
		output.Println()
		output.Printf("Exceptions for statement %s\n\n", args[0])
		for _, exception := range exceptions.Data {
			list.Add(&exceptionOut{
				Name:       exception.GetName(),
				Timestamp:  exception.GetTimestamp(),
				Stacktrace: exception.GetStacktrace(),
			})
		}
		_ = list.Print()
	}

	return nil
}
