package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newSavepointDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "describe <name>",
		Short:       "Describe a Flink savepoint in Confluent Platform.",
		Args:        cobra.ExactArgs(1),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogout},
		RunE:        c.savepointDescribe,
	}

	cmd.Flags().String("environment", "", "Name of the Flink environment.")
	cmd.Flags().String("application", "", "Name of the application to which the savepoint is attached to.")
	cmd.Flags().String("statement", "", "Name of the statement to which the savepoint is attached to.")
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
			Name:         outputSavepoint.Metadata.GetName(),
			Statement:    statement,
			Application:  application,
			Path:         outputSavepoint.Spec.GetPath(),
			Format:       outputSavepoint.Spec.GetFormatType(),
			BackoffLimit: outputSavepoint.Spec.GetBackoffLimit(),
			Uid:          outputSavepoint.Metadata.GetUid(),
			State:        outputSavepoint.Status.GetState(),
		})
		return table.Print()
	}

	localStmt := convertSdkSavepointToLocalSavepoint(outputSavepoint)
	return output.SerializedOutput(cmd, localStmt)
}
