package flink

import (
	"time"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newStatementDescribeCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "describe [name]",
		Short:       "Describe a Flink SQL statement in Confluent Platform.",
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

	table := output.NewTable(cmd)
	table.Add(&statementOutOnPrem{
		CreationDate: time.Now().Format(time.RFC3339),
		Name:         outputStatement.Metadata.Name,
		Statement:    outputStatement.Spec.Statement,
		ComputePool:  outputStatement.Spec.ComputePoolName,
		Status:       outputStatement.Status.Phase,
		StatusDetail: outputStatement.Status.GetDetail(),
		Properties:   outputStatement.Spec.GetProperties(),
	})
	table.Filter([]string{"CreationDate", "Name", "Statement", "ComputePool", "Status", "StatusDetail", "Properties"})

	return table.Print()
}
