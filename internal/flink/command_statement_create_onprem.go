package flink

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/flink/types"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/retry"
)

func (c *command) newStatementCreateCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "create [name]",
		Short:       "Create a Flink SQL statement.",
		Long:        "Create a Flink SQL statement in Confluent Platform.",
		Args:        cobra.MaximumNArgs(1),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogout},
		RunE:        c.statementCreateOnPrem,
	}

	cmd.Flags().String("sql", "", "The Flink SQL statement.")
	cmd.Flags().String("environment", "", "Name of the Flink environment.")
	cmd.Flags().String("compute-pool", "", "The compute pool name to execute the Flink SQL statement.")
	cmd.Flags().Uint16("parallelism", 1, "The parallelism the statement, default value is 1.")
	cmd.Flags().String("catalog", "", "The name of the default catalog.")
	cmd.Flags().String("database", "", "The name of the default database.")
	cmd.Flags().String("flink-configuration", "", "The file path to hold the Flink configuration for the statement.")
	cmd.Flags().Bool("wait", false, "Boolean flag to block until the statement is running or has failed.")
	addCmfFlagSet(cmd)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("sql"))
	cobra.CheckErr(cmd.MarkFlagRequired("environment"))
	cobra.CheckErr(cmd.MarkFlagRequired("compute-pool"))

	return cmd
}

func (c *command) statementCreateOnPrem(cmd *cobra.Command, args []string) error {
	// Flink statement name can be automatically generated or provided by the user
	name := types.GenerateStatementName()
	if len(args) == 1 {
		name = args[0]
	}

	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	sql, err := cmd.Flags().GetString("sql")
	if err != nil {
		return err
	}

	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}

	computePool, err := cmd.Flags().GetString("compute-pool")
	if err != nil {
		return err
	}

	parallelism, err := cmd.Flags().GetUint16("parallelism")
	if err != nil {
		return err
	}

	catalog, err := cmd.Flags().GetString("catalog")
	if err != nil {
		return err
	}

	database, err := cmd.Flags().GetString("database")
	if err != nil {
		return err
	}

	properties := map[string]string{}
	if database != "" {
		properties["sql.current-database"] = database
	}
	if catalog != "" {
		properties["sql.current-catalog"] = catalog
	}

	configFilePath, err := cmd.Flags().GetString("flink-configuration")
	if err != nil {
		return err
	}

	var flinkConfiguration = map[string]string{}
	if configFilePath != "" {
		var data []byte
		// Read configuration file contents
		data, err = os.ReadFile(configFilePath)
		if err != nil {
			return fmt.Errorf("failed to read Flink configuration file: %v", err)
		}
		ext := filepath.Ext(configFilePath)
		switch ext {
		case ".json":
			err = json.Unmarshal(data, &flinkConfiguration)
		case ".yaml", ".yml":
			err = yaml.Unmarshal(data, &flinkConfiguration)
		default:
			return errors.NewErrorWithSuggestions(fmt.Sprintf("unsupported file format: %s", ext), "Supported file formats are .json, .yaml, and .yml.")
		}
		if err != nil {
			return err
		}
	}

	statement := cmfsdk.Statement{
		ApiVersion: "cmf.confluent.io/v1",
		Kind:       "Statement",
		Metadata: cmfsdk.StatementMetadata{
			Name: name,
		},
		Spec: cmfsdk.StatementSpec{
			Statement:          sql,
			Properties:         &properties,
			FlinkConfiguration: &flinkConfiguration,
			ComputePoolName:    computePool,
			Parallelism:        cmfsdk.PtrInt32(int32(parallelism)),
			Stopped:            cmfsdk.PtrBool(false),
		},
	}

	wait, err := cmd.Flags().GetBool("wait")
	if err != nil {
		return err
	}

	outputStatement, err := client.CreateStatement(c.createContext(), environment, statement)
	if err != nil {
		return err
	}

	// CreateStatement() is async API, add wait logic below when the statement is still PENDING
	if wait {
		err := retry.Retry(time.Second*2, time.Minute, func() error {
			statement, err = client.GetStatement(c.createContext(), environment, name)
			if err != nil {
				return err
			}

			if statement.GetStatus().Phase == "PENDING" {
				return fmt.Errorf(`statement phase is "%s"`, statement.GetStatus().Phase)
			}

			return nil
		})
		if err != nil {
			return err
		}
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
