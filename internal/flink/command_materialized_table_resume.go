package flink

import (
	"github.com/spf13/cobra"

	flinkgatewayv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/flink-gateway/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/resource"
)

func (c *command) newMaterializedTableResumeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "resume <name>",
		Short:             "Resume a Flink materialized table.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validMaterializedTableArgs),
		RunE:              c.materializedTableResume,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Request to resume the materialized table "materialized-table-1".`,
				Code: "confluent flink materialized-table resume materialized-table-1 --database lk01",
			},
		),
	}

	cmd.Flags().String("database", "", "The ID of Kafka cluster hosting the Materialized Table's topic.")

	pcmd.AddCloudFlag(cmd)
	pcmd.AddRegionFlagFlink(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)

	cobra.CheckErr(cmd.MarkFlagRequired("database"))

	return cmd
}

func (c *command) materializedTableResume(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	client, err := c.GetFlinkGatewayClientInternal(false)
	if err != nil {
		return err
	}

	kafkaID, err := cmd.Flags().GetString("database")
	if err != nil {
		return err
	}

	table, err := client.DescribeMaterializedTable(environmentId, c.Context.GetCurrentOrganization(), kafkaID, args[0])
	if err != nil {
		return err
	}
	table.Spec.Stopped = flinkgatewayv1.PtrBool(false)

	if _, err := client.UpdateMaterializedTable(table, environmentId, c.Context.GetCurrentOrganization(), kafkaID, args[0]); err != nil {
		return err
	}

	output.Printf(c.Config.EnableColor, "Requested to resume %s \"%s\".\n", resource.MaterializedTable, args[0])
	return nil
}
