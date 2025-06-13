package tableflow

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/deletion"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/kafka"
	"github.com/confluentinc/cli/v4/pkg/resource"
)

func (c *command) newCatalogIntegrationDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id-1> [id-2] ... [id-n]",
		Short:             "Delete catalog integrations.",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validCatalogIntegrationArgsMultiple),
		RunE:              c.deleteCatalogIntegration,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Delete a catalog integration.",
				Code: "confluent tableflow catalog-integration delete tci-abc123",
			},
		),
	}

	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddForceFlag(cmd)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) deleteCatalogIntegration(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	cluster, err := kafka.GetClusterForCommand(c.V2Client, c.Context)
	if err != nil {
		return err
	}

	existenceFunc := func(id string) bool {
		_, err := c.V2Client.GetCatalogIntegration(environmentId, cluster.GetId(), id)
		return err == nil
	}

	if err := deletion.ValidateAndConfirm(cmd, args, existenceFunc, resource.CatalogIntegration); err != nil {
		return err
	}

	deleteFunc := func(id string) error {
		return c.V2Client.DeleteCatalogIntegration(environmentId, cluster.GetId(), id)
	}

	_, err = deletion.Delete(cmd, args, deleteFunc, resource.CatalogIntegration)
	return err
}
