package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *command) newStatementExceptionListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "list <statement-name>",
		Short:             "List exceptions for a Flink SQL statement.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validStatementArgs),
		RunE:              c.statementExceptionList,
	}

	pcmd.AddCloudFlag(cmd)
	pcmd.AddRegionFlagFlink(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) statementExceptionList(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	client, err := c.GetFlinkGatewayClient(false)
	if err != nil {
		return err
	}

	exceptions, err := client.GetExceptions(environmentId, args[0], c.Context.GetCurrentOrganization())
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, exception := range exceptions {
		list.Add(&exceptionOut{
			Name:       exception.GetName(),
			Timestamp:  exception.GetTimestamp(),
			StackTrace: exception.GetMessage(),
		})
	}
	return list.Print()
}
