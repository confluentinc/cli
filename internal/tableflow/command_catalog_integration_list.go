package tableflow

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/kafka"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newCatalogIntegrationListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List catalog integrations.",
		Args:  cobra.NoArgs,
		RunE:  c.listCatalogIntegration,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List catalog integrations.",
				Code: "confluent tableflow catalog-integration list",
			},
		),
	}

	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) listCatalogIntegration(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	cluster, err := kafka.GetClusterForCommand(c.V2Client, c.Context)
	if err != nil {
		return err
	}

	catalogIntegrations, err := c.V2Client.ListCatalogIntegrations(environmentId, cluster.GetId())
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, catalogIntegration := range catalogIntegrations {
		catalogIntegrationType, err := getCatalogIntegrationType(catalogIntegration)
		if err != nil {
			return err
		}

		out := &catalogIntegrationOut{
			Id:           catalogIntegration.GetId(),
			Name:         catalogIntegration.Spec.GetDisplayName(),
			Type:         catalogIntegrationType,
			Environment:  catalogIntegration.GetSpec().Environment.GetId(),
			KafkaCluster: catalogIntegration.GetSpec().KafkaCluster.GetId(),
			Suspended:    catalogIntegration.Spec.GetSuspended(),
			Phase:        catalogIntegration.Status.GetPhase(),
			ErrorMessage: catalogIntegration.Status.GetErrorMessage(),
		}

		if catalogIntegrationType == aws {
			out.ProviderIntegrationId = catalogIntegration.Spec.GetConfig().TableflowV1CatalogIntegrationAwsGlueSpec.GetProviderIntegrationId()
		}
		if catalogIntegrationType == snowflake {
			out.Endpoint = catalogIntegration.Spec.GetConfig().TableflowV1CatalogIntegrationSnowflakeSpec.GetEndpoint()
			out.Warehouse = catalogIntegration.Spec.GetConfig().TableflowV1CatalogIntegrationSnowflakeSpec.GetWarehouse()
			out.AllowedScope = catalogIntegration.Spec.GetConfig().TableflowV1CatalogIntegrationSnowflakeSpec.GetAllowedScope()
		}
		if catalogIntegrationType == unity {
			out.WorkspaceEndpoint = catalogIntegration.Spec.GetConfig().TableflowV1CatalogIntegrationUnitySpec.GetWorkspaceEndpoint()
			out.CatalogName = catalogIntegration.Spec.GetConfig().TableflowV1CatalogIntegrationUnitySpec.GetCatalogName()
			out.ClientId = catalogIntegration.Spec.GetConfig().TableflowV1CatalogIntegrationUnitySpec.GetClientId()
		}

		list.Add(out)
	}
	return list.Print()
}
