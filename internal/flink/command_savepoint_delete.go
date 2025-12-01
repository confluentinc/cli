package flink

import (
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/deletion"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/resource"
	"github.com/spf13/cobra"
)

func (c *command) newSavepointDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete Flink Savepoint in Confluent Platform.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  c.savepointDelete,
	}

	cmd.Flags().String("environment", "", "Name of the Flink environment.")
	cmd.Flags().Bool("force", false, "Force delete the savepoint.")
	cmd.Flags().String("application", "", "The name of the Flink application to create the savepoint for.")
	cmd.Flags().String("statement", "", "The name of the Flink statement to create the savepoint for.")
	addCmfFlagSet(cmd)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("environment"))
	cmd.MarkFlagsOneRequired("application", "statement")
	cmd.MarkFlagsMutuallyExclusive("application", "statement")

	return cmd
}

func (c *command) savepointDelete(cmd *cobra.Command, args []string) error {
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

	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return err
	}

	existenceFunc := func(name string) bool {
		_, err := client.DescribeSavepoint(c.createContext(), environment, name, application, statement)
		return err == nil
	}

	if err := deletion.ValidateAndConfirm(cmd, args, existenceFunc, resource.FlinkSavepoint); err != nil {
		// We are validating only the existence of the resources (there is no prefix validation).
		// Thus, we can add some extra context for the error.
		suggestions := "List available Flink savepoints with `confluent flink savepoint list`."
		suggestions += "\nCheck that CMF is running and accessible."
		return errors.NewErrorWithSuggestions(err.Error(), suggestions)
	}

	deleteFunc := func(name string) error {
		return client.DeleteSavepoint(c.createContext(), environment, name, application, statement, force)
	}

	_, err = deletion.Delete(cmd, args, deleteFunc, resource.FlinkSavepoint)
	return err
}
