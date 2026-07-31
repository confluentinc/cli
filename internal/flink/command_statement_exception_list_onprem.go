package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newStatementExceptionListCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "list <statement-name>",
		Short:       "List exceptions for a Flink SQL statement.",
		Long:        "List exceptions for a Flink SQL statement in Confluent Platform.",
		Args:        cobra.ExactArgs(1),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogout},
		RunE:        c.statementExceptionListOnPrem,
	}

	cmd.Flags().String("environment", "", "Name of the Flink environment.")
	addLimitFlag(cmd)
	addCmfFlagSet(cmd)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("environment"))

	return cmd
}

func (c *command) statementExceptionListOnPrem(cmd *cobra.Command, args []string) error {
	name := args[0]

	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}

	limit, err := getLimit(cmd)
	if err != nil {
		return err
	}

	exceptionList, err := client.ListStatementExceptions(c.createContext(), environment, name)
	if err != nil {
		return err
	}

	// ListStatementExceptions is a single, non-paginated call, so cap the result client-side.
	if data := exceptionList.GetData(); limit > 0 && int32(len(data)) > limit {
		exceptionList.SetData(data[:limit])
	}

	if output.GetFormat(cmd) == output.Human {
		list := output.NewList(cmd)
		for _, exception := range exceptionList.GetData() {
			list.Add(&exceptionOutOnPrem{
				Name:      exception.GetName(),
				Timestamp: exception.GetTimestamp(),
				Message:   exception.GetMessage(),
			})
		}
		return list.Print()
	}

	return output.SerializedOutput(cmd, exceptionList)
}
