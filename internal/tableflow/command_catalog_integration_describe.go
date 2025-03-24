package tableflow

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/kafka"
)

func (c *command) newCatalogIntegrationDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <id>",
		Short:             "Describe a catalog integration.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validCatalogIntegrationArgs),
		RunE:              c.describeCatalogIntegration,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Describe a catalog integration.",
				Code: "confluent tableflow catalog-integration describe tci-abc123",
			},
		),
	}

	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) describeCatalogIntegration(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	cluster, err := kafka.GetClusterForCommand(c.V2Client, c.Context)
	if err != nil {
		return err
	}

	catalogIntegration, err := c.V2Client.GetCatalogIntegration(environmentId, cluster.GetId(), args[0])
	if err != nil {
		return err
	}

	return printCatalogIntegrationTable(cmd, catalogIntegration)
}
