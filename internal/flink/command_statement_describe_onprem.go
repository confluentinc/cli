package flink

import (

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newStatementDescribeCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "describe [name]",
		Short:       "Describe a Flink SQL statement.",
		Long:        "Describe a Flink SQL statement in Confluent Platform.",
		Args:        cobra.ExactArgs(1),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogout},
		RunE:        c.statementDescribeOnPrem,
	}

	cmd.Flags().String("environment", "", "Name of the Flink environment.")
	addCmfFlagSet(cmd)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("environment"))

	return cmd
}

func (c *command) statementDescribeOnPrem(cmd *cobra.Command, args []string) error {
	name := args[0]

	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}

	outputStatement, err := client.GetStatement(c.createContext(), environment, name)
	if err != nil {
		return err
	}

	if output.GetFormat(cmd) == output.Human {
		table := output.NewTable(cmd)
		table.Add(&statementOutOnPrem{
			CreationDate: outputStatement.Metadata.GetCreationTimestamp(),
			Name:         outputStatement.Metadata.GetName(),
			Statement:    outputStatement.Spec.GetStatement(),
			ComputePool:  outputStatement.Spec.GetComputePoolName(),
			Status:       outputStatement.Status.GetPhase(),
			StatusDetail: outputStatement.Status.GetDetail(),
			Parallelism:  outputStatement.Spec.GetParallelism(),
			Stopped:      outputStatement.Spec.GetStopped(),
			SqlKind:      outputStatement.Status.Traits.GetSqlKind(),
			AppendOnly:   outputStatement.Status.Traits.GetIsAppendOnly(),
			Bounded:      outputStatement.Status.Traits.GetIsBounded(),
		})
		return table.Print()
	}

	// Create the top-level LocalStatement struct.
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

// Handle the nested Status, which is a pointer.
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
						Type: copyDataType(sdkCol.Type), // Use the helper function here
					})
				}
			}
			localTraits.Schema = localSchema
		}
		localStatus.Traits = localTraits
	}
	localStmt.Status = localStatus
}

// Handle the nested Result, which is a pointer.
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

	return output.SerializedOutput(cmd, localStmt)
}
