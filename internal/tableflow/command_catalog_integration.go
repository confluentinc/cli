package tableflow

import (
	"fmt"

	"github.com/spf13/cobra"

	tableflowv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/tableflow/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/kafka"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/utils"
)

const (
	awsGlueKind   = "AwsGlue"
	snowflakeKind = "Snowflake"

	aws       = "aws"
	snowflake = "snowflake"
)

var createCatalogIntegrationTypes = []string{aws, snowflake}

type catalogIntegrationOut struct {
	Id                    string `human:"ID" serialized:"id"`
	Name                  string `human:"Name" serialized:"name"`
	Environment           string `human:"Environment" serialized:"environment"`
	KafkaCluster          string `human:"Kafka Cluster" serialized:"kafka_cluster"`
	Type                  string `human:"Type" serialized:"type"`
	ProviderIntegrationId string `human:"Provider Integration ID,omitempty" serialized:"provider_integration_id,omitempty"`
	Endpoint              string `human:"Endpoint,omitempty" serialized:"endpoint,omitempty"`
	Warehouse             string `human:"Warehouse,omitempty" serialized:"warehouse,omitempty"`
	AllowedScope          string `human:"Allowed Scope,omitempty" serialized:"allowed_scope,omitempty"`
	Suspended             bool   `human:"Suspended" serialized:"suspended"`
	Phase                 string `human:"Phase" serialized:"phase"`
	ErrorMessage          string `human:"Error Message,omitempty" serialized:"error_message,omitempty"`
}

func (c *command) newCatalogIntegrationCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "catalog-integration",
		Short: "Manage Tableflow catalog integrations.",
	}

	cmd.AddCommand(c.newCatalogIntegrationCreateCommand())
	cmd.AddCommand(c.newCatalogIntegrationDeleteCommand())
	cmd.AddCommand(c.newCatalogIntegrationDescribeCommand())
	cmd.AddCommand(c.newCatalogIntegrationListCommand())
	cmd.AddCommand(c.newCatalogIntegrationUpdateCommand())

	return cmd
}

func addCatalogIntegrationTypeFlag(cmd *cobra.Command) {
	cmd.Flags().String("type", "", fmt.Sprintf("Specify the catalog integration type as %s.", utils.ArrayToCommaDelimitedString(createCatalogIntegrationTypes, "or")))
	pcmd.RegisterFlagCompletionFunc(cmd, "type", func(_ *cobra.Command, _ []string) []string { return createCatalogIntegrationTypes })
}

func (c *command) validCatalogIntegrationArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	return c.validCatalogIntegrationArgsMultiple(cmd, args)
}

func (c *command) validCatalogIntegrationArgsMultiple(cmd *cobra.Command, args []string) []string {
	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return c.autocompleteCatalogIntegrations()
}

func (c *command) autocompleteCatalogIntegrations() []string {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return nil
	}

	cluster, err := kafka.GetClusterForCommand(c.V2Client, c.Context)
	if err != nil {
		return nil
	}

	catalogIntegrations, err := c.V2Client.ListCatalogIntegrations(environmentId, cluster.GetId())
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(catalogIntegrations))
	for i, catalogIntegration := range catalogIntegrations {
		suggestions[i] = fmt.Sprintf("%s\t%s", catalogIntegration.GetId(), catalogIntegration.Spec.GetDisplayName())
	}
	return suggestions
}

func getCatalogIntegrationType(catalogIntegration tableflowv1.TableflowV1CatalogIntegration) (string, error) {
	config := catalogIntegration.Spec.GetConfig()

	if config.TableflowV1CatalogIntegrationAwsGlueSpec != nil {
		return aws, nil
	}

	if config.TableflowV1CatalogIntegrationSnowflakeSpec != nil {
		return snowflake, nil
	}

	return "", fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "config")
}

func printCatalogIntegrationTable(cmd *cobra.Command, catalogIntegration tableflowv1.TableflowV1CatalogIntegration) error {
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

	table := output.NewTable(cmd)
	table.Add(out)
	return table.Print()
}
