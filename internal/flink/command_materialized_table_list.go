package flink

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newMaterializedTableListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List materialized tables.",
		Args:  cobra.NoArgs,
		RunE:  c.tableList,
	}

	cmd.Flags().String("database", "", "The ID of Kafka cluster hosting the Materialized Table's topic.")
	pcmd.AddCloudFlag(cmd)
	pcmd.AddRegionFlagFlink(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("database"))

	return cmd
}

func (c *command) tableList(cmd *cobra.Command, _ []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	if _, err := c.V2Client.GetOrgEnvironment(environmentId); err != nil {
		return errors.NewErrorWithSuggestions(err.Error(), fmt.Sprintf(envNotFoundErrorMsg, environmentId))
	}

	client, err := c.GetFlinkGatewayClientInternal(false)
	if err != nil {
		return err
	}

	tables, err := client.ListMaterializedTable(environmentId, c.Context.GetCurrentOrganization())
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, materializedTable := range tables {
		mtableOut := materializedTableOut{
			Name:           materializedTable.GetName(),
			ClusterID:      materializedTable.Spec.GetKafkaClusterId(),
			Environment:    materializedTable.GetEnvironmentId(),
			ComputePool:    materializedTable.Spec.GetComputePoolId(),
			ServiceAccount: materializedTable.Spec.GetPrincipal(),
			Query:          materializedTable.Spec.GetQuery(),
			Columns:        convertToArrayColumns(materializedTable.Spec.GetColumns()),
			Constraints:    convertToArrayConstraints(materializedTable.Spec.GetConstraints()),
		}

		if materializedTable.Spec.Watermark != nil {
			wm := materializedTable.Spec.GetWatermark()
			mtableOut.WaterMarkColumnName = wm.GetColumnName()
			mtableOut.WaterMarkExpression = wm.GetExpression()
		}

		if materializedTable.Spec.DistributedBy != nil {
			db := materializedTable.Spec.GetDistributedBy()
			mtableOut.DistributedByColumnNames = db.GetColumnNames()
			mtableOut.DistributedByBuckets = int(db.GetBuckets())
		}
		list.Add(&mtableOut)
	}
	return list.Print()
}
