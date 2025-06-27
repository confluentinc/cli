package flink

import (
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

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

	if output.GetFormat(cmd) == output.YAML {
		// Convert the outputStatement to our local struct for correct YAML field names
		// We need to manually map the fields to preserve all data including nil fields
		var outputLocalStmt localStatement
		outputLocalStmt.ApiVersion = outputStatement.ApiVersion
		outputLocalStmt.Kind = outputStatement.Kind

		// Map metadata
		outputLocalStmt.Metadata.Name = outputStatement.Metadata.Name
		outputLocalStmt.Metadata.CreationTimestamp = outputStatement.Metadata.CreationTimestamp
		outputLocalStmt.Metadata.UpdateTimestamp = outputStatement.Metadata.UpdateTimestamp
		outputLocalStmt.Metadata.Uid = outputStatement.Metadata.Uid
		outputLocalStmt.Metadata.Labels = outputStatement.Metadata.Labels
		outputLocalStmt.Metadata.Annotations = outputStatement.Metadata.Annotations

		// Map spec
		outputLocalStmt.Spec.Statement = outputStatement.Spec.Statement
		outputLocalStmt.Spec.Properties = outputStatement.Spec.Properties
		outputLocalStmt.Spec.FlinkConfiguration = outputStatement.Spec.FlinkConfiguration
		outputLocalStmt.Spec.ComputePoolName = outputStatement.Spec.ComputePoolName
		outputLocalStmt.Spec.Parallelism = outputStatement.Spec.Parallelism
		outputLocalStmt.Spec.Stopped = outputStatement.Spec.Stopped

		// Map status if present
		if outputStatement.Status != nil {
			outputLocalStmt.Status = &localStatementStatus{
				Phase:  outputStatement.Status.Phase,
				Detail: outputStatement.Status.Detail,
			}

			// Map traits if present
			if outputStatement.Status.Traits != nil {
				outputLocalStmt.Status.Traits = &localStatementTraits{
					SqlKind:       outputStatement.Status.Traits.SqlKind,
					IsBounded:     outputStatement.Status.Traits.IsBounded,
					IsAppendOnly:  outputStatement.Status.Traits.IsAppendOnly,
					UpsertColumns: outputStatement.Status.Traits.UpsertColumns,
				}

				// Map schema if present
				if outputStatement.Status.Traits.Schema != nil {
					outputLocalStmt.Status.Traits.Schema = &localResultSchema{
						Columns: make([]localResultSchemaColumn, len(outputStatement.Status.Traits.Schema.Columns)),
					}
					for i, col := range outputStatement.Status.Traits.Schema.Columns {
						outputLocalStmt.Status.Traits.Schema.Columns[i] = localResultSchemaColumn{
							Name: col.Name,
							Type: localDataType{
								Type:                col.Type.Type,
								Nullable:            col.Type.Nullable,
								Length:              col.Type.Length,
								Precision:           col.Type.Precision,
								Scale:               col.Type.Scale,
								KeyType:             nil, // Would need recursive mapping for complex types
								ValueType:           nil, // Would need recursive mapping for complex types
								ElementType:         nil, // Would need recursive mapping for complex types
								Fields:              nil, // Would need recursive mapping for complex types
								Resolution:          col.Type.Resolution,
								FractionalPrecision: col.Type.FractionalPrecision,
							},
						}
					}
				}
			}
		}

		// Map result if present
		if outputStatement.Result != nil {
			outputLocalStmt.Result = &localStatementResult{
				ApiVersion: outputStatement.Result.ApiVersion,
				Kind:       outputStatement.Result.Kind,
				Metadata: localStatementResultMetadata{
					CreationTimestamp: outputStatement.Result.Metadata.CreationTimestamp,
					Annotations:       outputStatement.Result.Metadata.Annotations,
				},
				Results: localStatementResults{
					Data: outputStatement.Result.Results.Data,
				},
			}
		}

		// Output the local struct for correct YAML field names
		out, err := yaml.Marshal(outputLocalStmt)
		if err != nil {
			return err
		}
		output.Print(false, string(out))
		return nil
	}

	return output.SerializedOutput(cmd, outputStatement)
}
