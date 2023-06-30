package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *command) newStatementExceptionsListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "list <statement name>",
		Short:             "Get exceptions for a Flink SQL statement.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validStatementArgs),
		RunE:              c.exceptionsList,
	}

	c.addComputePoolFlag(cmd)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)

	return cmd
}

func (c *command) exceptionsList(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	client, err := c.GetFlinkGatewayClient()
	if err != nil {
		return err
	}

	orgId := c.Context.GetCurrentOrganization()

	exceptions, err := client.GetExceptions(environmentId, args[0], orgId)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)

	for _, exception := range exceptions.Data {
		list.Add(&exceptionOut{
			Name:       exception.GetName(),
			Timestamp:  exception.GetTimestamp(),
			Stacktrace: exception.GetStacktrace(),
		})
	}

	return list.Print()
}
