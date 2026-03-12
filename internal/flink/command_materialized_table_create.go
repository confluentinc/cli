package flink

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	flinkgatewayv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/flink-gateway/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/properties"
)

func (c *command) newMaterializedTableCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a Flink materialized table.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.materializedTableCreate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create Flink materialized table "my-connection" in AWS us-west-2.`,
				Code: "flink materialized-table create my-table --cloud aws --region  us-west-2 --database lkc01 --compute-pool pool1 --service-account principal1 --query query1",
			},
		),
	}

	cmd.Flags().String("database", "", "The ID of Kafka cluster hosting the Materialized Table's topic.")
	cmd.Flags().String("compute-pool", "", "The id associated with the compute pool in context.")
	cmd.Flags().String("service-account", "", "The id of a principal this Materialized Table query runs as.")
	cmd.Flags().String("query", "", "The query section of the latest Materialized Table.")

	pcmd.AddCloudFlag(cmd)
	pcmd.AddRegionFlagFlink(cmd, c.AuthenticatedCLICommand)

	c.addOptionalMaterializedTableFlags(cmd)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("database"))
	cobra.CheckErr(cmd.MarkFlagRequired("compute-pool"))
	cobra.CheckErr(cmd.MarkFlagRequired("service-account"))
	cobra.CheckErr(cmd.MarkFlagRequired("query"))

	return cmd
}

func (c *command) materializedTableCreate(cmd *cobra.Command, args []string) error {
	kafkaId, err := cmd.Flags().GetString("database")
	if err != nil {
		return err
	}

	computePool, err := cmd.Flags().GetString("compute-pool")
	if err != nil {
		return err
	}

	serviceAccount, err := cmd.Flags().GetString("service-account")
	if err != nil {
		return err
	}

	query, err := cmd.Flags().GetString("query")
	if err != nil {
		return err
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	if _, err := c.V2Client.GetOrgEnvironment(environmentId); err != nil {
		return errors.NewErrorWithSuggestions(err.Error(), fmt.Sprintf(envNotFoundErrorMsg, environmentId))
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

	var colDetails []flinkgatewayv1.SqlV1MaterializedTableColumnDetails
	if columnComputed != "" {
		colDetails, err = addComputedColumns(columnComputed, colDetails)
	}

	if columnPhysical != "" {
		colDetails, err = addPhysicalColumns(columnPhysical, colDetails)
	}
	if columnMetadata != "" {
		colDetails, err = addMetadataColumns(columnMetadata, colDetails)
	}
	if err != nil {
		return err
	}

	watermarkColumnName, err := cmd.Flags().GetString("watermark-column-name")
	if err != nil {
		return err
	}

	watermarkExpression, err := cmd.Flags().GetString("watermark-expression")
	if err != nil {
		return err
	}

	constraints, err := cmd.Flags().GetString("constraints")
	if err != nil {
		return err
	}

	var constr []flinkgatewayv1.SqlV1MaterializedTableConstraint
	if constraints != "" {
		constr, err = addConstraints(constraints, constr)
	}
	if err != nil {
		return err
	}

	distributedByColumnNames, err := cmd.Flags().GetString("distributed-by-column-names")
	if err != nil {
		return err
	}

	distributedByColumnNamesArray := csvToStringSlicePtr(distributedByColumnNames)

	distributedByBuckets, err := cmd.Flags().GetInt("distributed-by-buckets")
	if err != nil {
		return err
	}

	distributedByBucketsInt32 := int32(distributedByBuckets)

	name := args[0]

	orgId := c.Context.GetCurrentOrganization()
	if err != nil {
		return err
	}

	client, err := c.GetFlinkGatewayClientInternal(false)
	if err != nil {
		return err
	}

	table := flinkgatewayv1.SqlV1MaterializedTable{
		Name:           name,
		EnvironmentId:  environmentId,
		OrganizationId: orgId,

		Spec: flinkgatewayv1.SqlV1MaterializedTableSpec{
			KafkaClusterId: &kafkaId,
			ComputePoolId:  &computePool,
			Principal:      &serviceAccount,
			Query:          &query,
			Columns:        &colDetails,
			Watermark: &flinkgatewayv1.SqlV1MaterializedTableWatermark{
				ColumnName: &watermarkColumnName,
				Expression: &watermarkExpression,
			},
			DistributedBy: &flinkgatewayv1.SqlV1MaterializedTableDistribution{
				ColumnNames: distributedByColumnNamesArray,
				Buckets:     &distributedByBucketsInt32,
			},
			Constraints: &constr,
		},
	}

	materializedTable, err := client.CreateMaterializedTable(table, environmentId, orgId, kafkaId)
	if err != nil {
		//panic("HERE")
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

func addComputedColumns(path string, colDetails []flinkgatewayv1.SqlV1MaterializedTableColumnDetails) ([]flinkgatewayv1.SqlV1MaterializedTableColumnDetails, error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	tablesContent := properties.ParseLines(string(buf))
	for index := range len(tablesContent) {
		columnComputed := tablesContent[index]
		values := strings.Split(columnComputed, ",")
		if len(values) != 6 {
			return nil, fmt.Errorf("the computed column must be in the format [name,type,comment,kind,expression,virtual]")
		}
		columnName := values[0]
		columnType := values[1]
		columnComment := values[2]
		columnKind := values[3]
		columnExpression := values[4]
		virtual := values[5]
		var columnVirtual bool
		if virtual != "" {
			columnVirtual, err = strconv.ParseBool(values[5])
			if err != nil {
				return nil, fmt.Errorf("please enter true or false for virtual field")
			}
		}
		computedColumn := flinkgatewayv1.SqlV1ComputedColumn{
			Name:       columnName,
			Type:       columnType,
			Comment:    &columnComment,
			Kind:       columnKind,
			Expression: columnExpression,
			Virtual:    &columnVirtual,
		}
		column := flinkgatewayv1.SqlV1MaterializedTableColumnDetails{
			SqlV1ComputedColumn: &computedColumn,
		}
		colDetails = append(colDetails, column)
	}
	return colDetails, nil
}

func addMetadataColumns(path string, colDetails []flinkgatewayv1.SqlV1MaterializedTableColumnDetails) ([]flinkgatewayv1.SqlV1MaterializedTableColumnDetails, error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	tablesContent := properties.ParseLines(string(buf))
	for index := range len(tablesContent) {
		columnComputed := tablesContent[index]
		values := strings.Split(columnComputed, ",")
		if len(values) != 6 {
			return nil, fmt.Errorf("the metadata column must be in the format [name,type,comment,kind,key,virtual]")
		}
		columnName := values[0]
		columnType := values[1]
		columnComment := values[2]
		columnKind := values[3]
		columnKey := values[4]
		virtual := values[5]
		var columnVirtual bool
		if virtual != "" {
			columnVirtual, err = strconv.ParseBool(values[5])
			if err != nil {
				return nil, fmt.Errorf("please enter true or false for virtual field")
			}
		}
		metadataColumn := flinkgatewayv1.SqlV1MetadataColumn{
			Name:        columnName,
			Type:        columnType,
			Comment:     &columnComment,
			Kind:        columnKind,
			MetadataKey: columnKey,
			Virtual:     &columnVirtual,
		}
		column := flinkgatewayv1.SqlV1MaterializedTableColumnDetails{
			SqlV1MetadataColumn: &metadataColumn,
		}
		colDetails = append(colDetails, column)
	}
	return colDetails, nil
}

func addPhysicalColumns(path string, colDetails []flinkgatewayv1.SqlV1MaterializedTableColumnDetails) ([]flinkgatewayv1.SqlV1MaterializedTableColumnDetails, error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	tablesContent := properties.ParseLines(string(buf))
	for index := range len(tablesContent) {
		columnComputed := tablesContent[index]
		values := strings.Split(columnComputed, ",")
		if len(values) != 4 {
			return nil, fmt.Errorf("the physical column must be in the format [name,type,comment,kind]")
		}
		columnName := values[0]
		columnType := values[1]
		columnComment := values[2]
		columnKind := values[3]
		physicalColumn := flinkgatewayv1.SqlV1PhysicalColumn{
			Name:    columnName,
			Type:    columnType,
			Comment: &columnComment,
			Kind:    columnKind,
		}
		column := flinkgatewayv1.SqlV1MaterializedTableColumnDetails{
			SqlV1PhysicalColumn: &physicalColumn,
		}
		colDetails = append(colDetails, column)
	}
	return colDetails, nil
}

func addConstraints(path string, constr []flinkgatewayv1.SqlV1MaterializedTableConstraint) ([]flinkgatewayv1.SqlV1MaterializedTableConstraint, error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	tablesContent := properties.ParseLines(string(buf))
	for index := range len(tablesContent) {
		content := tablesContent[index]
		values := strings.Split(content, ",")
		if len(values) != 4 {
			return nil, fmt.Errorf("the constraints must be in the format [name,type,colNmae1|colName2,enforced]")
		}
		constraintsName := values[0]
		constraintsType := values[1]
		constraintsColumnNames := values[2]
		constraintsColumnNameArray := csvToStringSlicePtrConstraint(constraintsColumnNames)
		constraintsEnforced := values[3]
		constraintsEnforcedBool, err := strconv.ParseBool(constraintsEnforced)
		if err != nil {
			return nil, err
		}
		constraint := flinkgatewayv1.SqlV1MaterializedTableConstraint{
			Name:        &constraintsName,
			Kind:        &constraintsType,
			ColumnNames: constraintsColumnNameArray,
			Enforced:    &constraintsEnforcedBool,
		}
		constr = append(constr, constraint)
	}
	return constr, nil
}

func csvToStringSlicePtr(csv string) *[]string {
	if csv == "" {
		return &[]string{}
	}
	values := strings.Split(csv, ",")
	return &values
}

func csvToStringSlicePtrConstraint(csv string) *[]string {
	if csv == "" {
		return &[]string{}
	}
	values := strings.Split(csv, "|")
	return &values
}

func convertToArrayColumns(columns []flinkgatewayv1.SqlV1MaterializedTableColumnDetails) []string {
	var cols []string
	for _, value := range columns {
		if value.SqlV1PhysicalColumn != nil {
			cols = append(cols, fmt.Sprintf("{%s, %s, %s, %s}", value.SqlV1PhysicalColumn.GetName(), value.SqlV1PhysicalColumn.GetType(), value.SqlV1PhysicalColumn.GetComment(), value.SqlV1PhysicalColumn.GetKind()))
		}
		if value.SqlV1ComputedColumn != nil {
			cols = append(cols, fmt.Sprintf("{%s, %s, %s, %s, %s, %t}", value.SqlV1ComputedColumn.GetName(), value.SqlV1ComputedColumn.GetType(), value.SqlV1ComputedColumn.GetComment(), value.SqlV1ComputedColumn.GetKind(), value.SqlV1ComputedColumn.GetExpression(), value.SqlV1ComputedColumn.GetVirtual()))
		}
		if value.SqlV1MetadataColumn != nil {
			cols = append(cols, fmt.Sprintf("{%s, %s, %s, %s, %s, %t}", value.SqlV1MetadataColumn.GetName(), value.SqlV1MetadataColumn.GetType(), value.SqlV1MetadataColumn.GetComment(), value.SqlV1MetadataColumn.GetKind(), value.SqlV1MetadataColumn.GetMetadataKey(), value.SqlV1MetadataColumn.GetVirtual()))
		}
	}
	return cols
}

func convertToArrayConstraints(constraints []flinkgatewayv1.SqlV1MaterializedTableConstraint) []string {
	constr := make([]string, 0, len(constraints))
	for _, value := range constraints {
		constr = append(constr, fmt.Sprintf("{%s, %s, %s, %t}", value.GetName(), value.GetKind(), value.GetColumnNames(), value.GetEnforced()))
	}
	return constr
}
