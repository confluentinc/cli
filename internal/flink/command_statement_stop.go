package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/resource"
)

func (c *statementCommand) newStatementStopCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "stop <name>",
		Short:             "Stop a Flink SQL statement.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.statementStop,
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

	return cmd
}

func (c *statementCommand) statementStop(_ *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	client, err := c.GetFlinkGatewayClient(false)
	if err != nil {
		return err
	}

	if err := client.StopStatement(environmentId, args[0], c.Context.GetCurrentOrganization()); err != nil {
		return err
	}

	output.Printf(c.Config.EnableColor, "Requested to stop %s \"%s\".\n", resource.FlinkStatement, args[0])
	return nil
}
