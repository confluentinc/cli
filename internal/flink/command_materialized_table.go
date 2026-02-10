package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

type materializedTableOut struct {
	Name                     string   `human:"Name" serialized:"name"`
	ClusterID                string   `human:"Kafka Cluster ID" serialized:"kafka_cluster_id"`
	Environment              string   `human:"Environment" serialized:"environment"`
	ComputePool              string   `human:"Compute Pool" serialized:"compute_pool"`
	ServiceAccount           string   `human:"Service Account" serialized:"service_account"`
	Query                    string   `human:"Query,omitempty" serialized:"query,omitempty"`
	Columns                  []string `human:"Columns,omitempty" serialized:"columns,omitempty"`
	WaterMarkColumnName      string   `human:"Watermark Column Name,omitempty" serialized:"watermark_column_name,omitempty"`
	WaterMarkExpression      string   `human:"Watermark Expression,omitempty" serialized:"watermark_expression,omitempty"`
	Constraints              []string `human:"Constraints,omitempty" serialized:"constraints,omitempty"`
	DistributedByColumnNames []string `human:"Distributed By Column Names,omitempty" serialized:"distributed_by_column_names,omitempty"`
	DistributedByBuckets     int      `human:"Distributed By Buckets,omitempty" serialized:"distributed_by_buckets,omitempty"`
}

func (c *command) newMaterializedTableCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "materialized-table",
		Short:       "Manage Flink materialized tables.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	cmd.AddCommand(c.newMaterializedTableCreateCommand())
	cmd.AddCommand(c.newMaterializedTableDeleteCommand())
	cmd.AddCommand(c.newMaterializedTableDescribeCommand())
	cmd.AddCommand(c.newMaterializedTableListCommand())
	cmd.AddCommand(c.newMaterializedTableUpdateCommand())
	cmd.AddCommand(c.newMaterializedTableStopCommand())
	cmd.AddCommand(c.newMaterializedTableResumeCommand())

	return cmd
}

func (c *command) validMaterializedTableArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	return c.validMaterializedTablesArgsMultiple(cmd, args)
}

func (c *command) validMaterializedTablesArgsMultiple(cmd *cobra.Command, args []string) []string {
	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return nil
	}

	client, err := c.GetFlinkGatewayClientInternal(false)
	if err != nil {
		return nil
	}

	tables, err := client.ListMaterializedTable(environmentId, c.Context.GetCurrentOrganization())
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(tables))
	for i, table := range tables {
		suggestions[i] = table.GetName()
	}
	return suggestions
}

func (c *command) addOptionalMaterializedTableFlags(cmd *cobra.Command) {
	cmd.Flags().String("column-physical", "", "Path to the columns data for type physical.")
	cmd.Flags().String("column-metadata", "", "Path to the columns data for type metadata.")
	cmd.Flags().String("column-computed", "", "Path to the columns data for type computed.")
	cmd.Flags().String("watermark-column-name", "", "The name of the watermark columns.")
	cmd.Flags().String("watermark-expression", "", "The watermark expression.")
	cmd.Flags().String("constraints", "", "Path to the constraints.")
	cmd.Flags().String("distributed-by-column-names", "", "The names of the columns the table is distributed by.")
	cmd.Flags().Int("distributed-by-buckets", 0, "The number of buckets.")
}
