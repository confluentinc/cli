package flink

import (
	"github.com/spf13/cobra"

	flinkgatewayv1beta1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1beta1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

func (c *command) newStatementUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "update <name>",
		Short:             "Move a Flink SQL statement to another compute pool.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validStatementArgs),
		RunE:              c.statementUpdate,
	}

	pcmd.AddCloudFlag(cmd)
	c.addRegionFlag(cmd)
	c.addComputePoolFlag(cmd)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)

	cobra.CheckErr(cmd.MarkFlagRequired("compute-pool"))

	return cmd
}

func (c *command) statementUpdate(cmd *cobra.Command, args []string) error {
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

	computePool, err := cmd.Flags().GetString("compute-pool")
	if err != nil {
		return err
	}
	statement.Spec.ComputePoolId = flinkgatewayv1beta1.PtrString(computePool)

	if err := client.UpdateStatement(environmentId, args[0], c.Context.GetCurrentOrganization(), statement); err != nil {
		return err
	}

	output.Printf(c.Config.EnableColor, "Requested to move %s \"%s\" to %s \"%s\".\n", resource.FlinkStatement, args[0], resource.FlinkComputePool, computePool)
	return nil
}
