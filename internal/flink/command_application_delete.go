package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/deletion"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/resource"
)

func (c *command) newApplicationDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <name-1> [name-2] ... [name-n]",
		Short: "Delete one or more Flink applications.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.applicationDelete,
	}

	cmd.Flags().String("environment", "", "Name of the Flink environment.")
	addCmfFlagSet(cmd)
	pcmd.AddForceFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("environment"))

	return cmd
}

func (c *command) applicationDelete(cmd *cobra.Command, args []string) error {
	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}
	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	existenceFunc := func(name string) bool {
		_, err := client.DescribeApplication(c.createContext(), environment, name)
		return err == nil
	}

	if err := deletion.ValidateAndConfirm(cmd, args, existenceFunc, resource.FlinkApplication); err != nil {
		// We are validating only the existence of the resources (there is no prefix validation).
		// Thus we can add some extra context for the error.
		suggestions := "List available Flink applications with `confluent flink application list`."
		suggestions += "\nCheck that CMF is running and accessible."
		return errors.NewErrorWithSuggestions(err.Error(), suggestions)
	}

	deleteFunc := func(name string) error {
		return client.DeleteApplication(c.createContext(), environment, name)
	}

	_, err = deletion.Delete(cmd, args, deleteFunc, resource.FlinkApplication)
	return err
}
