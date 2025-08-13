package flink

import (
	"slices"
	"strings"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/log"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/utils"
)

func (c *command) newStatementListCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "list",
		Short:       "List Flink SQL statements in Confluent Platform.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogout},
		RunE:        c.statementListOnPrem,
	}

	cmd.Flags().String("environment", "", "Name of the Flink environment.")
	cmd.Flags().String("compute-pool", "", "Optional flag to filter the Flink statements by compute pool ID.")
	cmd.Flags().String("status", "", "Optional flag to filter the Flink statements by statement status.")
	addCmfFlagSet(cmd)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("environment"))

	return cmd
}

func (c *command) statementListOnPrem(cmd *cobra.Command, _ []string) error {
	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}

	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	status, err := cmd.Flags().GetString("status")
	if err != nil {
		return err
	}
	status = strings.ToLower(status)

	if status != "" && !slices.Contains(allowedStatuses, status) {
		log.CliLogger.Warnf(`Invalid status "%s". Valid statuses are %s.`, status, utils.ArrayToCommaDelimitedString(allowedStatuses, "and"))
	}

	computePool, err := cmd.Flags().GetString("compute-pool")
	if err != nil {
		return err
	}

	sdkStatements, err := client.ListStatements(c.createContext(), environment, computePool, status)
	if err != nil {
		return err
	}

	if output.GetFormat(cmd) == output.Human {
		list := output.NewList(cmd)
		for _, statement := range sdkStatements {
			list.Add(&statementOutOnPrem{
				CreationDate: statement.Metadata.GetCreationTimestamp(),
				Name:         statement.Metadata.Name,
				Statement:    statement.Spec.Statement,
				ComputePool:  statement.Spec.ComputePoolName,
				Status:       statement.Status.Phase,
				StatusDetail: statement.Status.GetDetail(),
				Parallelism:  statement.Spec.GetParallelism(),
				Stopped:      statement.Spec.GetStopped(),
				SqlKind:      statement.Status.Traits.GetSqlKind(),
				AppendOnly:   statement.Status.Traits.GetIsAppendOnly(),
				Bounded:      statement.Status.Traits.GetIsBounded(),
			})
		}
		return list.Print()
	}

	localStmts := make([]LocalStatement, 0, len(sdkStatements))

	for _, sdkStmt := range sdkStatements {
		localStmt := LocalStatement{
			ApiVersion: sdkStmt.ApiVersion,
			Kind:       sdkStmt.Kind,
			Metadata: LocalStatementMetadata{
				Name:              sdkStmt.Metadata.Name,
				CreationTimestamp: sdkStmt.Metadata.CreationTimestamp,
				UpdateTimestamp:   sdkStmt.Metadata.UpdateTimestamp,
				Uid:               sdkStmt.Metadata.Uid,
				Labels:            sdkStmt.Metadata.Labels,
				Annotations:       sdkStmt.Metadata.Annotations,
			},
			Spec: LocalStatementSpec{
				Statement:          sdkStmt.Spec.Statement,
				Properties:         sdkStmt.Spec.Properties,
				FlinkConfiguration: sdkStmt.Spec.FlinkConfiguration,
				ComputePoolName:    sdkStmt.Spec.ComputePoolName,
				Parallelism:        sdkStmt.Spec.Parallelism,
				Stopped:            sdkStmt.Spec.Stopped,
			},
		}

		if sdkStmt.Status != nil {
			localStatus := &LocalStatementStatus{
				Phase:  sdkStmt.Status.Phase,
				Detail: sdkStmt.Status.Detail,
			}
			if sdkStmt.Status.Traits != nil {
				localTraits := &LocalStatementTraits{
					SqlKind:       sdkStmt.Status.Traits.SqlKind,
					IsBounded:     sdkStmt.Status.Traits.IsBounded,
					IsAppendOnly:  sdkStmt.Status.Traits.IsAppendOnly,
					UpsertColumns: sdkStmt.Status.Traits.UpsertColumns,
				}
				if sdkStmt.Status.Traits.Schema != nil {
					localSchema := &LocalResultSchema{}
					if sdkStmt.Status.Traits.Schema.Columns != nil {
						localSchema.Columns = make([]LocalResultSchemaColumn, 0, len(sdkStmt.Status.Traits.Schema.Columns))
						for _, sdkCol := range sdkStmt.Status.Traits.Schema.Columns {
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

		if sdkStmt.Result != nil {
			localStmt.Result = &LocalStatementResult{
				ApiVersion: sdkStmt.Result.ApiVersion,
				Kind:       sdkStmt.Result.Kind,
				Metadata: LocalStatementResultMetadata{
					CreationTimestamp: sdkStmt.Result.Metadata.CreationTimestamp,
					Annotations:       sdkStmt.Result.Metadata.Annotations,
				},
				Results: LocalStatementResults{
					Data: sdkStmt.Result.Results.Data,
				},
			}
		}

		localStmts = append(localStmts, localStmt)
	}

	return output.SerializedOutput(cmd, localStmts)
}
