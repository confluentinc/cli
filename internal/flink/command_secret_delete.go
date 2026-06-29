package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/deletion"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/resource"
)

func (c *command) newSecretDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <name-1> [name-2] ... [name-n]",
		Short: "Delete one or more Flink secrets in Confluent Platform.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.secretDelete,
	}

	addCmfFlagSet(cmd)
	pcmd.AddForceFlag(cmd)

	return cmd
}

func (c *command) secretDelete(cmd *cobra.Command, args []string) error {
	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	existenceFunc := func(name string) bool {
		_, err := client.DescribeSecret(c.createContext(), name)
		return err == nil
	}

	// Block deletion of CFK-owned resources before prompting for confirmation. The
	// whole batch fails if any resource is CFK-owned. Resources that cannot be
	// fetched are skipped here and reported by ValidateAndConfirm below.
	for _, name := range args {
		secret, describeErr := client.DescribeSecret(c.createContext(), name)
		if describeErr != nil {
			continue
		}
		if err := errIfCfkManaged(resource.FlinkSecret, name, secret.Metadata.GetAnnotations()); err != nil {
			return err
		}
	}

	if err := deletion.ValidateAndConfirm(cmd, args, existenceFunc, resource.FlinkSecret); err != nil {
		suggestions := "List available Flink secrets with `confluent flink secret list`."
		suggestions += "\nCheck that CMF is running and accessible."
		return errors.NewErrorWithSuggestions(err.Error(), suggestions)
	}

	deleteFunc := func(name string) error {
		return client.DeleteSecret(c.createContext(), name)
	}

	_, err = deletion.Delete(cmd, args, deleteFunc, resource.FlinkSecret)
	return err
}
