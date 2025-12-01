package flink

import (
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/spf13/cobra"
)

func (c *command) newSavepointListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "list",
		Short:       "List Flink SQL statements in Confluent Platform.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogout},
		RunE:        c.savepointList,
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

func (c *command) savepointList(cmd *cobra.Command, _ []string) error {
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

	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	sdkSavepoints, err := client.ListSavepoint(c.createContext(), environment, statement, application, statement != "")
	if err != nil {
		return err
	}

	if output.GetFormat(cmd) == output.Human {
		list := output.NewList(cmd)
		for _, savepoint := range sdkSavepoints {
			list.Add(&savepointOut{
				Name:        savepoint.Metadata.GetName(),
				Statement:   statement,
				Application: application,
				Path:        savepoint.Spec.GetPath(),
				Format:      savepoint.Spec.GetFormatType(),
				Limit:       savepoint.Spec.GetBackoffLimit(),
			})
		}
		return list.Print()
	}

	savepoints := make([]LocalSavepoint, 0, len(sdkSavepoints))
	for _, sdkSavepoint := range sdkSavepoints {
		savepoint := convertSdkSavepointToLocalSavepoint(sdkSavepoint)
		savepoints = append(savepoints, savepoint)
	}

	return output.SerializedOutput(cmd, savepoints)
}
