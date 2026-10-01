package flink

import (
	"github.com/spf13/cobra"

	flinkgatewayv1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *statementCommand) newStopCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "stop <name>",
		Short:             "Stop a Flink SQL statement.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.stop,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Request to stop the currently running statement "my-statement".`,
				Code: "confluent flink statement stop my-statement",
			},
		),
	}

	pcmd.AddCloudFlag(cmd)
	pcmd.AddRegionFlagFlink(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *statementCommand) stop(cmd *cobra.Command, args []string) error {
	statementName := args[0]
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	client, err := c.GetFlinkGatewayClient(false)
	if err != nil {
		return err
	}

	// Full-object update: the gateway rejects a spec.stopped-only body, so send the current
	// statement back with stopped set.
	current, err := client.GetStatement(environmentId, statementName, c.Context.GetCurrentOrganization())
	if err != nil {
		return err
	}
	updateReq := current
	if updateReq.Spec == nil {
		updateReq.Spec = &flinkgatewayv1.SqlV1StatementSpec{}
	}
	updateReq.Spec.Stopped = flinkgatewayv1.PtrBool(true)

	if err := client.UpdateStatement(environmentId, statementName, c.Context.GetCurrentOrganization(), updateReq); err != nil {
		return err
	}
	if output.GetFormat(cmd) == output.Human {
		output.Printf(c.Config.EnableColor, "Requested to stop Flink SQL statement \"%s\".\n", statementName)
	}
	statement, err := client.GetStatement(environmentId, statementName, c.Context.GetCurrentOrganization())
	if err != nil {
		return err
	}
	return printStatement(cmd, statement)
}
