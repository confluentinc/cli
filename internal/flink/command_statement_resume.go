package flink

import (
	"fmt"

	"github.com/spf13/cobra"

	flinkgatewayv1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/resource"
)

func (c *command) newStatementResumeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "resume <name>",
		Short:             "Resume a Flink SQL statement.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validStatementArgs),
		RunE:              c.statementResume,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Request to resume statement "my-statement" with original principal id and compute pool.`,
				Code: "confluent flink statement resume my-statement",
			},
			examples.Example{
				Text: `Request to resume statement "my-statement" with service account "sa-123456".`,
				Code: "confluent flink statement resume my-statement --principal sa-123456",
			},
			examples.Example{
				Text: `Request to resume statement "my-statement" with user account "u-987654".`,
				Code: "confluent flink statement resume my-statement --principal u-987654",
			},
			examples.Example{
				Text: `Request to resume statement "my-statement" and move to compute pool "lfcp-123456".`,
				Code: "confluent flink statement resume my-statement --compute-pool lfcp-123456",
			},
			examples.Example{
				Text: `Request to resume statement "my-statement" with service account "sa-123456" and move to compute pool "lfcp-123456".`,
				Code: "confluent flink statement resume my-statement --principal sa-123456 --compute-pool lfcp-123456",
			},
		),
	}

	c.addPrincipalFlag(cmd)
	c.addComputePoolFlag(cmd)
	pcmd.AddCloudFlag(cmd)
	pcmd.AddRegionFlagFlink(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)

	return cmd
}

func (c *command) statementResume(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	client, err := c.GetFlinkGatewayClient(false)
	if err != nil {
		return err
	}

	statement, err := client.GetStatement(environmentId, args[0], c.Context.GetCurrentOrganization())
	if err != nil {
		return err
	}

	// Support resume a Flink statement with a different principal and/or compute-pool
	principal, err := cmd.Flags().GetString("principal")
	if err != nil {
		return err
	}
	if principal != "" {
		statement.Spec.SetPrincipal(principal)
	}

	computePool, err := cmd.Flags().GetString("compute-pool")
	if err != nil {
		return err
	}
	if computePool != "" {
		statement.Spec.SetComputePoolId(computePool)
	}

	statement.Spec.Stopped = flinkgatewayv1.PtrBool(false)

	// the UPDATE statement is an async API
	// An accepted response 202 doesn't necessarily mean the UPDATE will be successful/complete
	if err := client.UpdateStatement(environmentId, args[0], c.Context.GetCurrentOrganization(), statement); err != nil {
		return fmt.Errorf("failed to resume %s \"%s\": %w", resource.FlinkStatement, args[0], err)
	}

	output.Printf(c.Config.EnableColor, "Requested to resume %s \"%s\".\n", resource.FlinkStatement, args[0])
	return nil
}
