package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newStatementExceptionListCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "list <statement-name>",
		Short:             "List exceptions for a Flink SQL statement in Confluent Platform.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validStatementArgs),
		Annotations:       map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogout},
		RunE:              c.statementExceptionListOnPrem,
	}

	cmd.Flags().String("environment", "", "Name of the environment to list the Flink SQL statement exceptions.")
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

	exceptionList, err := client.ListStatementExceptions(c.createContext(), environment, name)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)

	for _, exception := range exceptionList.Data {
		list.Add(&exceptionOutOnPrem{
			Name:      exception.Name,
			Timestamp: exception.Timestamp,
			Message:   exception.Message,
		})
	}

	return list.Print()
}
