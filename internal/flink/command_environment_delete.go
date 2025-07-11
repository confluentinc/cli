package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/deletion"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/resource"
)

func (c *command) newEnvironmentDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <name-1> [name-2] ... [name-n]",
		Short: "Delete one or more Flink environments.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.environmentDelete,
	}

	addCmfFlagSet(cmd)

	pcmd.AddForceFlag(cmd)

	return cmd
}

func (c *command) environmentDelete(cmd *cobra.Command, args []string) error {
	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	existenceFunc := func(name string) bool {
		_, err := client.DescribeEnvironment(c.createContext(), name)
		return err == nil
	}

	if err := deletion.ValidateAndConfirm(cmd, args, existenceFunc, resource.FlinkEnvironment); err != nil {
		// We are validating only the existence of the resources (there is no prefix validation). Thus we can
		// add some extra context for the error.
		suggestions := "List available Flink environments with `confluent flink environment list`."
		suggestions += "\nCheck that CMF is running and accessible."
		return errors.NewErrorWithSuggestions(err.Error(), suggestions)
	}

	deleteFunc := func(name string) error {
		return client.DeleteEnvironment(c.createContext(), name)
	}

	_, err = deletion.Delete(cmd, args, deleteFunc, resource.FlinkEnvironment)
	return err
}
