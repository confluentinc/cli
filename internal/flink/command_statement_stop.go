package flink

import (
	"github.com/spf13/cobra"

	flinkgatewayv1beta1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1beta1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

func (c *command) newStatementStopCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "stop <name>",
		Short:             "Stop a Flink SQL statement.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validStatementArgs),
		RunE:              c.statementStop,
	}

	pcmd.AddCloudFlag(cmd)
	c.addRegionFlag(cmd)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)

	return cmd
}

func (c *command) statementStop(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	client, err := c.GetFlinkGatewayClient()
	if err != nil {
		return err
	}

	statement, err := client.GetStatement(environmentId, args[0], c.Context.GetCurrentOrganization())
	if err != nil {
		return err
	}
	statement.Spec.Stopped = flinkgatewayv1beta1.PtrBool(true)

	if err := client.UpdateStatement(environmentId, args[0], c.Context.GetCurrentOrganization(), statement); err != nil {
		return err
	}

	output.Printf(c.Config.EnableColor, "Requested to stop %s \"%s\".\n", resource.FlinkStatement, args[0])
	return nil
}
