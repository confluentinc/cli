package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/deletion"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/resource"
)

func (c *command) newSecretMappingDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <name-1> [name-2] ... [name-n]",
		Short: "Delete one or more Flink secret mappings.",
		Long:  "Delete one or more Flink environment secret mappings in Confluent Platform.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.secretMappingDelete,
	}

	cmd.Flags().String("environment", "", "Name of the Flink environment.")
	cobra.CheckErr(cmd.MarkFlagRequired("environment"))
	addCmfFlagSet(cmd)
	pcmd.AddForceFlag(cmd)

	return cmd
}

func (c *command) secretMappingDelete(cmd *cobra.Command, args []string) error {
	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}

	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	existenceFunc := func(name string) bool {
		_, err := client.DescribeSecretMapping(c.createContext(), environment, name)
		return err == nil
	}

	if err := deletion.ValidateAndConfirm(cmd, args, existenceFunc, resource.FlinkSecretMapping); err != nil {
		suggestions := "List available Flink secret mappings with `confluent flink secret-mapping list --environment <envName>`."
		suggestions += "\nCheck that CMF is running and accessible."
		return errors.NewErrorWithSuggestions(err.Error(), suggestions)
	}

	deleteFunc := func(name string) error {
		return client.DeleteSecretMapping(c.createContext(), environment, name)
	}

	_, err = deletion.Delete(cmd, args, deleteFunc, resource.FlinkSecretMapping)
	return err
}
