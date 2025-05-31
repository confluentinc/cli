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
	statements, err := client.ListStatements(c.createContext(), environment, computePool, status)
	if err != nil {
		return err
	}

	if output.GetFormat(cmd) == output.Human {
		list := output.NewList(cmd)
		for _, statement := range statements {
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

	return output.SerializedOutput(cmd, statements)
}
