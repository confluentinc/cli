package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newSavepointDetachCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "detach <name>",
		Short:       "Detach Flink Savepoint in Confluent Platform.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogout},
		RunE:        c.savepointDetach,
	}

	cmd.Flags().String("environment", "", "Name of the Flink environment.")
	cmd.Flags().String("application", "", "The name of the Flink application to create the savepoint for.")
	addCmfFlagSet(cmd)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("environment"))
	cobra.CheckErr(cmd.MarkFlagRequired("application"))

	return cmd
}

func (c *command) savepointDetach(cmd *cobra.Command, args []string) error {
	name := args[0]

	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}

	application, err := cmd.Flags().GetString("application")
	if err != nil {
		return err
	}

	cmfSavepoint, err := client.DetachSavepointApplication(c.createContext(), name, environment, application)
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&savepointOut{
		Name:        cmfSavepoint.Metadata.GetName(),
		Application: application,
		Path:        cmfSavepoint.Spec.GetPath(),
		Format:      cmfSavepoint.Spec.GetFormatType(),
		Limit:       cmfSavepoint.Spec.GetBackoffLimit(),
		Uid:         cmfSavepoint.Metadata.GetUid(),
		State:       cmfSavepoint.Status.GetState(),
	})
	return table.Print()
}
