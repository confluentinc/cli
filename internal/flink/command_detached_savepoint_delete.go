package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/deletion"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/resource"
)

func (c *command) newDetachedSavepointDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <name-1> [name-2] ... [name-n]",
		Short: "Delete Flink detached savepoints in Confluent Platform.",
		Long:  "Deleting a detached savepoint does not delete the actual physical data.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.detachedSavepointDelete,
	}
	addCmfFlagSet(cmd)
	pcmd.AddForceFlag(cmd)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) detachedSavepointDelete(cmd *cobra.Command, args []string) error {
	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}
	existenceFunc := func(name string) bool {
		_, err := client.DescribeDetachedSavepoint(c.createContext(), name)
		return err == nil
	}

	if err := deletion.ValidateAndConfirm(cmd, args, existenceFunc, resource.FlinkDetachedSavepoint); err != nil {
		// We are validating only the existence of the resources (there is no prefix validation).
		// Thus, we can add some extra context for the error.
		suggestions := "List available Flink detached savepoints with `confluent flink detached-savepoint list`."
		suggestions += "\nCheck that CMF is running and accessible."
		return errors.NewErrorWithSuggestions(err.Error(), suggestions)
	}

	deleteFunc := func(name string) error {
		return client.DeleteDetachedSavepoint(c.createContext(), name)
	}

	_, err = deletion.Delete(cmd, args, deleteFunc, resource.FlinkDetachedSavepoint)
	return err
}
