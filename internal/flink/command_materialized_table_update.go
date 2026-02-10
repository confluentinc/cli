package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newMaterializedTableUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <name>",
		Short: "Update a Flink materialized table.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.materializedTableUpdate,
	}

	cmd.Flags().String("database", "", "The ID of Kafka cluster hosting the Materialized Table's topic.")
	cmd.Flags().String("compute-pool", "", "The id associated with the compute pool in context.")
	cmd.Flags().String("service-account", "", "The id of a principal this Materialized Table query runs as.")
	cmd.Flags().String("query", "", "The query section of the latest Materialized Table.")
	cmd.Flags().Bool("stopped", false, "Determine whether stopped or not.")

	c.addOptionalMaterializedTableFlags(cmd)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddCloudFlag(cmd)
	pcmd.AddRegionFlagFlink(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("database"))

	return cmd
}

func (c *command) materializedTableUpdate(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	client, err := c.GetFlinkGatewayClientInternal(false)
	if err != nil {
		return err
	}

	kafkaId, err := cmd.Flags().GetString("database")
	if err != nil {
		return err
	}

	table, err := client.DescribeMaterializedTable(environmentId, c.Context.GetCurrentOrganization(), kafkaId, args[0])
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
	if columnComputed != "" {
		colDetails, _ = addComputedColumns(columnComputed, colDetails)
	}

	if columnPhysical != "" {
		colDetails, _ = addPhysicalColumns(columnPhysical, colDetails)
	}

	if columnMetadata != "" {
		colDetails, _ = addMetadataColumns(columnMetadata, colDetails)
	}

	table.Spec.SetColumns(colDetails)
	watermarkColumnName, err := cmd.Flags().GetString("watermark-column-name")
	if err != nil {
		return err
	}
	if watermarkColumnName != "" {
		table.Spec.Watermark.SetColumnName(watermarkColumnName)
	}

	watermarkExpression, err := cmd.Flags().GetString("watermark-expression")
	if err != nil {
		return err
	}
	if watermarkExpression != "" {
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
	distributedByColumnNamesArray := csvToStringSlicePtr(distributedByColumnNames)

	if err != nil {
		return err
	}
	if distributedByColumnNames != "" {
		table.Spec.DistributedBy.SetColumnNames(*distributedByColumnNamesArray)
	}

	distributedByBuckets, err := cmd.Flags().GetInt("distributed-by-buckets")
	distributedByBucketsInt32 := int32(distributedByBuckets)

	if err != nil {
		return err
	}
	if distributedByBuckets > 0 {
		table.Spec.DistributedBy.SetBuckets(distributedByBucketsInt32)
	}

	if err != nil {
		return err
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

	outputTable.Add(&mtableOut)

	return outputTable.Print()
}
