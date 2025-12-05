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

	localStmt := convertSdkStatementToLocalStatement(outputStatement)
	return output.SerializedOutput(cmd, localStmt)
}
