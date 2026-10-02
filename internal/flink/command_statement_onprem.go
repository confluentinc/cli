package flink

import (
	"github.com/spf13/cobra"

	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

type statementOutOnPrem struct {
	CreationDate string `human:"Creation Date" serialized:"creation_date"`
	Name         string `human:"Name" serialized:"name"`
	Statement    string `human:"Statement" serialized:"statement"`
	ComputePool  string `human:"Compute Pool" serialized:"compute_pool"`
	Status       string `human:"Status" serialized:"status"`
	StatusDetail string `human:"Status Detail,omitempty" serialized:"status_detail,omitempty"`
	Parallelism  int32  `human:"Parallelism" serialized:"parallelism"`
	Stopped      bool   `human:"Stopped" serialized:"stopped"`
	SqlKind      string `human:"SQL Kind,omitempty" serialized:"sql_kind,omitempty"`
	AppendOnly   bool   `human:"Append Only,omitempty" serialized:"append_only,omitempty"`
	Bounded      bool   `human:"Bounded,omitempty" serialized:"bounded,omitempty"`
}

// The RequireCloudLogout annotation covers every command below it, the `exception` subtree
// included: ErrIfMissingRunRequirement walks up from the leaf through each parent.
func (c *command) newStatementCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "statement",
		Short:       "Manage Flink SQL statements in Confluent Platform.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogout},
	}

	cmd.AddCommand(c.newStatementCreateCommandOnPrem())
	cmd.AddCommand(c.newStatementDeleteCommandOnPrem())
	cmd.AddCommand(c.newStatementDescribeCommandOnPrem())
	cmd.AddCommand(c.newStatementExceptionCommandOnPrem())
	cmd.AddCommand(c.newStatementListCommandOnPrem())
	cmd.AddCommand(c.newStatementRescaleCommandOnPrem())
	cmd.AddCommand(c.newStatementResumeCommandOnPrem())
	cmd.AddCommand(c.newStatementStopCommandOnPrem())
	cmd.AddCommand(c.newStatementWebUiForwardCommand())

	return cmd
}

func convertSdkStatementToLocalStatement(outputStatement cmfsdk.Statement) LocalStatement {
	localStmt := LocalStatement{
		ApiVersion: outputStatement.ApiVersion,
		Kind:       outputStatement.Kind,
		Metadata: LocalStatementMetadata{
			Name:              outputStatement.Metadata.Name,
			CreationTimestamp: outputStatement.Metadata.CreationTimestamp,
			UpdateTimestamp:   outputStatement.Metadata.UpdateTimestamp,
			Uid:               outputStatement.Metadata.Uid,
			Labels:            outputStatement.Metadata.Labels,
			Annotations:       outputStatement.Metadata.Annotations,
		},
		Spec: LocalStatementSpec{
			Statement:          outputStatement.Spec.Statement,
			Properties:         outputStatement.Spec.Properties,
			FlinkConfiguration: outputStatement.Spec.FlinkConfiguration,
			ComputePoolName:    outputStatement.Spec.ComputePoolName,
			Parallelism:        outputStatement.Spec.Parallelism,
			Stopped:            outputStatement.Spec.Stopped,
		},
	}

	if outputStatement.Status != nil {
		localStatus := &LocalStatementStatus{
			Phase:  outputStatement.Status.Phase,
			Detail: outputStatement.Status.Detail,
		}

		if outputStatement.Status.Traits != nil {
			localTraits := &LocalStatementTraits{
				SqlKind:       outputStatement.Status.Traits.SqlKind,
				IsBounded:     outputStatement.Status.Traits.IsBounded,
				IsAppendOnly:  outputStatement.Status.Traits.IsAppendOnly,
				UpsertColumns: outputStatement.Status.Traits.UpsertColumns,
			}

			if outputStatement.Status.Traits.Schema != nil {
				localSchema := &LocalResultSchema{}
				if outputStatement.Status.Traits.Schema.Columns != nil {
					localSchema.Columns = make([]LocalResultSchemaColumn, 0, len(outputStatement.Status.Traits.Schema.Columns))
					for _, sdkCol := range outputStatement.Status.Traits.Schema.Columns {
						localSchema.Columns = append(localSchema.Columns, LocalResultSchemaColumn{
							Name: sdkCol.Name,
							Type: copyDataType(sdkCol.Type),
						})
					}
				}
				localTraits.Schema = localSchema
			}
			localStatus.Traits = localTraits
		}
		localStmt.Status = localStatus
	}

	if outputStatement.Result != nil {
		localStmt.Result = &LocalStatementResult{
			ApiVersion: outputStatement.Result.ApiVersion,
			Kind:       outputStatement.Result.Kind,
			Metadata: LocalStatementResultMetadata{
				CreationTimestamp: outputStatement.Result.Metadata.CreationTimestamp,
				Annotations:       outputStatement.Result.Metadata.Annotations,
			},
			Results: LocalStatementResults{
				Data: outputStatement.Result.Results.Data,
			},
		}
	}

	return localStmt
}
