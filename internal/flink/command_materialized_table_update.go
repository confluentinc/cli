package flink

import (
	"fmt"

	"github.com/spf13/cobra"

	flinkgatewayv1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newMaterializedTableUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "update <name>",
		Short:             "Update a Flink materialized table.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validMaterializedTableArgs),
		RunE:              c.materializedTableUpdate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Stop the Flink materialized table "my-table".`,
				Code: "confluent flink materialized-table update my-table --database lkc01 --stopped=true",
			},
			examples.Example{
				Text: `Resume the Flink materialized table "my-table".`,
				Code: "confluent flink materialized-table update my-table --database lkc01 --stopped=false",
			},
		),
	}

	cmd.Flags().String("database", "", "The ID of Kafka cluster hosting the Materialized Table's topic.")
	cmd.Flags().String("compute-pool", "", "The ID associated with the compute pool in context.")
	cmd.Flags().String("service-account", "", "The ID of a principal this Materialized Table query runs as.")
	cmd.Flags().String("query", "", "The query section of the latest Materialized Table.")
	cmd.Flags().Bool("stopped", false, "Determine whether stopped or not.")

	c.addOptionalMaterializedTableFlags(cmd)
	pcmd.AddCloudFlag(cmd)
	pcmd.AddRegionFlagFlink(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("database"))

	return cmd
}

func (c *command) materializedTableUpdate(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	if _, err := c.V2Client.GetOrgEnvironment(environmentId); err != nil {
		return errors.NewErrorWithSuggestions(err.Error(), fmt.Sprintf(envNotFoundErrorMsg, environmentId))
	}

	client, err := c.GetFlinkGatewayClient(false)
	if err != nil {
		return err
	}

	kafkaId, err := cmd.Flags().GetString("database")
	if err != nil {
		return err
	}

	table, err := client.GetMaterializedTable(environmentId, c.Context.GetCurrentOrganization(), kafkaId, args[0])
	if err != nil {
		return err
	}

	principal, err := cmd.Flags().GetString("service-account")
	if err != nil {
		return err
	}
	if principal != "" {
		table.Spec.SetPrincipal(principal)
	}

	computePool, err := cmd.Flags().GetString("compute-pool")
	if err != nil {
		return err
	}
	if computePool != "" {
		table.Spec.SetComputePoolId(computePool)
	}

	query, err := cmd.Flags().GetString("query")
	if err != nil {
		return err
	}
	if query != "" {
		table.Spec.SetQuery(query)
	}

	if cmd.Flags().Changed("stopped") {
		stopped, err := cmd.Flags().GetBool("stopped")
		if err != nil {
			return err
		}
		table.Spec.SetStopped(stopped)
	}

	columnComputed, err := cmd.Flags().GetString("column-computed")
	if err != nil {
		return err
	}

	columnPhysical, err := cmd.Flags().GetString("column-physical")
	if err != nil {
		return err
	}

	columnMetadata, err := cmd.Flags().GetString("column-metadata")
	if err != nil {
		return err
	}

	colDetails := table.Spec.GetColumns()
	if columnPhysical != "" {
		colDetails, err = addPhysicalColumns(columnPhysical, colDetails)
		if err != nil {
			return err
		}
	}

	if columnComputed != "" {
		colDetails, err = addComputedColumns(columnComputed, colDetails)
		if err != nil {
			return err
		}
	}

	if columnMetadata != "" {
		colDetails, err = addMetadataColumns(columnMetadata, colDetails)
		if err != nil {
			return err
		}
	}

	table.Spec.SetColumns(colDetails)
	watermarkColumnName, err := cmd.Flags().GetString("watermark-column-name")
	if err != nil {
		return err
	}
	if watermarkColumnName != "" {
		if table.Spec.Watermark == nil {
			table.Spec.Watermark = &flinkgatewayv1.SqlV1Watermark{}
		}
		table.Spec.Watermark.SetColumn(watermarkColumnName)
	}

	watermarkExpression, err := cmd.Flags().GetString("watermark-expression")
	if err != nil {
		return err
	}
	if watermarkExpression != "" {
		if table.Spec.Watermark == nil {
			table.Spec.Watermark = &flinkgatewayv1.SqlV1Watermark{}
		}
		table.Spec.Watermark.SetExpression(watermarkExpression)
	}

	constraints, err := cmd.Flags().GetString("constraints")
	if err != nil {
		return err
	}

	constr := table.Spec.GetConstraints()
	if constraints != "" {
		constr, err = addConstraints(constraints, constr)
		if err != nil {
			return err
		}
	}

	table.Spec.SetConstraints(constr)
	distributedByColumnNames, err := cmd.Flags().GetString("distributed-by-column-names")
	if err != nil {
		return err
	}
	distributedByColumnNamesArray := csvToStringSlicePtr(distributedByColumnNames)

	if distributedByColumnNames != "" {
		if table.Spec.Distribution == nil {
			table.Spec.Distribution = &flinkgatewayv1.SqlV1Distribution{}
		}
		table.Spec.Distribution.SetKeys(*distributedByColumnNamesArray)
	}

	distributedByBuckets, err := cmd.Flags().GetInt("distributed-by-buckets")
	if err != nil {
		return err
	}
	if distributedByBuckets > 0 {
		if table.Spec.Distribution == nil {
			table.Spec.Distribution = &flinkgatewayv1.SqlV1Distribution{}
		}
		table.Spec.Distribution.SetBucketCount(int32(distributedByBuckets))
	}

	materializedTable, err := client.UpdateMaterializedTable(table, environmentId, c.Context.GetCurrentOrganization(), kafkaId, args[0])
	if err != nil {
		return err
	}

	outputTable := output.NewTable(cmd)
	mtableOut := materializedTableOut{
		Name:           materializedTable.GetName(),
		ClusterID:      materializedTable.Spec.GetKafkaClusterId(),
		Environment:    materializedTable.GetEnvironmentId(),
		ComputePool:    materializedTable.Spec.GetComputePoolId(),
		ServiceAccount: materializedTable.Spec.GetPrincipal(),
		Stopped:        materializedTable.Spec.GetStopped(),
		Query:          materializedTable.Spec.GetQuery(),
		Columns:        convertToArrayColumns(materializedTable.Spec.GetColumns()),
		Constraints:    convertToArrayConstraints(materializedTable.Spec.GetConstraints()),
	}

	if materializedTable.Spec.Watermark != nil {
		wm := materializedTable.Spec.GetWatermark()
		mtableOut.WaterMarkColumnName = wm.GetColumn()
		mtableOut.WaterMarkExpression = wm.GetExpression()
	}

	if materializedTable.Spec.Distribution != nil {
		db := materializedTable.Spec.GetDistribution()
		mtableOut.DistributionKeys = db.GetKeys()
		mtableOut.DistributionBucketCount = int(db.GetBucketCount())
	}

	outputTable.Add(&mtableOut)

	return outputTable.Print()
}
