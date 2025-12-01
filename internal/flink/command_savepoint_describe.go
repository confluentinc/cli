package flink

import (
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/spf13/cobra"
)

func (c *command) newSavepointDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "describe <name>",
		Short:       "Describe flink savepoint in Confluent Platform.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogout},
		RunE:        c.savepointDescribe,
	}

	cmd.Flags().String("environment", "", "Name of the Flink environment.")
	cmd.Flags().String("application", "", "The name of the Flink application to create the savepoint for.")
	cmd.Flags().String("statement", "", "The name of the Flink statement to create the savepoint for.")
	addCmfFlagSet(cmd)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("environment"))
	cmd.MarkFlagsOneRequired("application", "statement")
	cmd.MarkFlagsMutuallyExclusive("application", "statement")

	return cmd
}

func (c *command) savepointDescribe(cmd *cobra.Command, args []string) error {
	name := args[0]

	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}

	statement, err := cmd.Flags().GetString("statement")
	if err != nil {
		return err
	}

	application, err := cmd.Flags().GetString("application")
	if err != nil {
		return err
	}

	outputSavepoint, err := client.DescribeSavepoint(c.createContext(), environment, name, application, statement)
	if err != nil {
		return err
	}

	if output.GetFormat(cmd) == output.Human {
		table := output.NewTable(cmd)
		table.Add(&savepointOut{
			Name:        outputSavepoint.Metadata.GetName(),
			Statement:   statement,
			Application: application,
			Path:        outputSavepoint.Spec.GetPath(),
			Format:      outputSavepoint.Spec.GetFormatType(),
			Limit:       outputSavepoint.Spec.GetBackoffLimit(),
		})
		return table.Print()
	}

	localStmt := convertSdkSavepointToLocalSavepoint(outputSavepoint)
	return output.SerializedOutput(cmd, localStmt)
}
