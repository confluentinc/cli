package flink

import (
	"github.com/spf13/cobra"

	flinkgatewayv1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *statementCommand) newResumeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "resume <name>",
		Short:             "Resume a Flink SQL statement.",
		Long:              "Resume a stopped Flink SQL statement, optionally under a different principal or compute pool. The principal can only be changed while resuming, so `--principal` is accepted here and on `flink statement update --stopped=false`. The resume is accepted asynchronously. Check the result with `flink statement describe <name>`.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.resume,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Request to resume the currently stopped statement "my-statement" using original principal id and under the original compute pool.`,
				Code: "confluent flink statement resume my-statement",
			},
			examples.Example{
				Text: `Request to resume the currently stopped statement "my-statement" using service account "sa-123456".`,
				Code: "confluent flink statement resume my-statement --principal sa-123456",
			},
			examples.Example{
				Text: `Request to resume the currently stopped statement "my-statement" using user account "u-987654".`,
				Code: "confluent flink statement resume my-statement --principal u-987654",
			},
			examples.Example{
				Text: `Request to resume the currently stopped statement "my-statement" and under a different compute pool "lfcp-123456".`,
				Code: "confluent flink statement resume my-statement --compute-pool lfcp-123456",
			},
			examples.Example{
				Text: `Request to resume the currently stopped statement "my-statement" using service account "sa-123456" and under a different compute pool "lfcp-123456".`,
				Code: "confluent flink statement resume my-statement --principal sa-123456 --compute-pool lfcp-123456",
			},
		),
	}

	pcmd.AddFlinkStatementPrincipalFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddComputePoolFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddCloudFlag(cmd)
	pcmd.AddRegionFlagFlink(cmd, c.AuthenticatedCLICommand)
	deprecateStatementCloudAndRegionFlags(cmd)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)

	return cmd
}

func (c *statementCommand) resume(cmd *cobra.Command, args []string) error {
	statementName := args[0]
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	client, err := c.GetFlinkGatewayClient(false)
	if err != nil {
		return err
	}

	// Full-object update: the gateway rejects a partial body, so seed every field from the current
	// statement and let the flags below overlay it.
	current, err := client.GetStatement(environmentId, statementName, c.Context.GetCurrentOrganization())
	if err != nil {
		return err
	}
	updateReq := current
	if updateReq.Spec == nil {
		updateReq.Spec = &flinkgatewayv1.SqlV1StatementSpec{}
	}

	if cmd.Flags().Changed("compute-pool") {
		computePoolId, err := cmd.Flags().GetString("compute-pool")
		if err != nil {
			return err
		}
		updateReq.Spec.ComputePoolId = flinkgatewayv1.PtrString(computePoolId)
	}

	if cmd.Flags().Changed("principal") {
		principal, err := cmd.Flags().GetString("principal")
		if err != nil {
			return err
		}
		updateReq.Spec.Principal = flinkgatewayv1.PtrString(principal)
	}

	updateReq.Spec.Stopped = flinkgatewayv1.PtrBool(false)

	if err := client.UpdateStatement(environmentId, statementName, c.Context.GetCurrentOrganization(), updateReq); err != nil {
		return err
	}
	output.Printf(c.Config.EnableColor, "Requested to resume Flink SQL statement \"%s\".\n", statementName)
	output.Printf(c.Config.EnableColor, "Please use `confluent flink statement describe %s` to check the latest status.\n", statementName)
	return nil
}
