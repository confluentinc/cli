package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/deletion"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/resource"
)

func (c *command) newCatalogDatabaseDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a Flink database in Confluent Platform.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.catalogDatabaseDelete,
	}

	cmd.Flags().String("catalog", "", "Name of the catalog.")
	cobra.CheckErr(cmd.MarkFlagRequired("catalog"))
	addCmfFlagSet(cmd)
	pcmd.AddForceFlag(cmd)

	return cmd
}

func (c *command) catalogDatabaseDelete(cmd *cobra.Command, args []string) error {
	catalogName, err := cmd.Flags().GetString("catalog")
	if err != nil {
		return err
	}

	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	existenceFunc := func(name string) bool {
		_, err := client.DescribeDatabase(c.createContext(), catalogName, name)
		return err == nil
	}

	// Block deletion of CFK-owned resources before prompting for confirmation. The
	// whole batch fails if any resource is CFK-owned. Resources that cannot be
	// fetched are skipped here and reported by ValidateAndConfirm below.
	for _, name := range args {
		database, describeErr := client.DescribeDatabase(c.createContext(), catalogName, name)
		if describeErr != nil {
			continue
		}
		if err := errIfCfkManaged(resource.FlinkDatabase, name, database.Metadata.GetAnnotations()); err != nil {
			return err
		}
	}

	if err := deletion.ValidateAndConfirm(cmd, args, existenceFunc, resource.FlinkDatabase); err != nil {
		// We are validating only the existence of the resources (there is no prefix validation).
		// Thus, we can add some extra context for the error.
		suggestions := "List available Flink databases with `confluent flink catalog database list --catalog <name>`."
		suggestions += "\nCheck that CMF is running and accessible."
		return errors.NewErrorWithSuggestions(err.Error(), suggestions)
	}

	deleteFunc := func(name string) error {
		return client.DeleteDatabase(c.createContext(), catalogName, name)
	}

	_, err = deletion.Delete(cmd, args, deleteFunc, resource.FlinkDatabase)
	return err
}
