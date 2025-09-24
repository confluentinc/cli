package tableflow

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	tableflowv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/tableflow/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/kafka"
)

func (c *command) newCatalogIntegrationCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a catalog integration.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.createCatalogIntegration,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Create an Aws Glue catalog integration.",
				Code: "confluent tableflow catalog-integration create my-catalog-integration --type aws --provider-integration cspi-stgce89r7",
			},
			examples.Example{
				Text: "Create a Snowflake catalog integration.",
				Code: "confluent tableflow catalog-integration create my-catalog-integration --type snowflake --endpoint https://vuser1_polaris.snowflakecomputing.com/ --warehouse catalog-name --allowed-scope session:role:R1 --client-id $CLIENT_ID --client-secret $CLIENT_SECRET",
			},
			examples.Example{
				Text: "Create a Unity catalog integration.",
				Code: "confluent tableflow catalog-integration create my-catalog-integration --type unity --workspace-endpoint https://dbc-1.cloud.databricks.com --catalog-name tableflow-quickstart-catalog --unity-client-id $CLIENT_ID --unity-client-secret $CLIENT_SECRET",
			},
		),
	}

	addCatalogIntegrationTypeFlag(cmd)
	cmd.Flags().String("provider-integration", "", "Specify the provider integration id.")
	cmd.Flags().String("endpoint", "", "Specify the The catalog integration connection endpoint for Snowflake Open Catalog.")
	cmd.Flags().String("warehouse", "", "Specify the warehouse name of the Snowflake Open Catalog.")
	cmd.Flags().String("allowed-scope", "", "Specify the allowed scope of the Snowflake Open Catalog.")
	cmd.Flags().String("client-id", "", "Specify the client id.")
	cmd.Flags().String("client-secret", "", "Specify the client secret.")
	cmd.Flags().String("workspace-endpoint", "", "Specify the Databricks workspace URL associated with the Unity Catalog.")
	cmd.Flags().String("catalog-name", "", "Specify the name of the catalog.")
	cmd.Flags().String("unity-client-id", "", "Specify the Unity client id.")
	cmd.Flags().String("unity-client-secret", "", "Specify the Unity client secret.")

	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("type"))
	cmd.MarkFlagsRequiredTogether("endpoint", "warehouse", "allowed-scope", "client-id", "client-secret")
	cmd.MarkFlagsRequiredTogether("workspace-endpoint", "catalog-name", "unity-client-id", "unity-client-secret")

	return cmd
}

func (c *command) createCatalogIntegration(cmd *cobra.Command, args []string) error {
	name := args[0]

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	cluster, err := kafka.GetClusterForCommand(c.V2Client, c.Context)
	if err != nil {
		return err
	}

	catalogIntegrationType, err := cmd.Flags().GetString("type")
	if err != nil {
		return err
	}

	createCatalogIntegration := tableflowv1.TableflowV1CatalogIntegration{
		Spec: &tableflowv1.TableflowV1CatalogIntegrationSpec{
			DisplayName:  tableflowv1.PtrString(name),
			Environment:  &tableflowv1.GlobalObjectReference{Id: environmentId},
			KafkaCluster: &tableflowv1.EnvScopedObjectReference{Id: cluster.GetId()},
		},
	}

	if strings.ToLower(catalogIntegrationType) == aws {
		if !cmd.Flags().Changed("provider-integration") {
			return fmt.Errorf("`--provider-integration` flag is required for catalog integration type `aws`.")
		}
		providerIntegration, err := cmd.Flags().GetString("provider-integration")
		if err != nil {
			return err
		}

		createCatalogIntegration.Spec.Config = &tableflowv1.TableflowV1CatalogIntegrationSpecConfigOneOf{
			TableflowV1CatalogIntegrationAwsGlueSpec: &tableflowv1.TableflowV1CatalogIntegrationAwsGlueSpec{
				Kind:                  awsGlueKind,
				ProviderIntegrationId: providerIntegration,
			},
		}
	} else if strings.ToLower(catalogIntegrationType) == snowflake {
		if !cmd.Flags().Changed("endpoint") { // we only need to check for one since this flag set is marked as required together
			return fmt.Errorf("`--endpoint`, `--warehouse`, `--allowed-scope`, `--client-id` and `--client-secret` flags are required for catalog integration type `snowflake`.")
		}
		endpoint, err := cmd.Flags().GetString("endpoint")
		if err != nil {
			return err
		}
		warehouse, err := cmd.Flags().GetString("warehouse")
		if err != nil {
			return err
		}
		allowedScope, err := cmd.Flags().GetString("allowed-scope")
		if err != nil {
			return err
		}
		clientId, err := cmd.Flags().GetString("client-id")
		if err != nil {
			return err
		}
		clientSecret, err := cmd.Flags().GetString("client-secret")
		if err != nil {
			return err
		}

		createCatalogIntegration.Spec.Config = &tableflowv1.TableflowV1CatalogIntegrationSpecConfigOneOf{
			TableflowV1CatalogIntegrationSnowflakeSpec: &tableflowv1.TableflowV1CatalogIntegrationSnowflakeSpec{
				Kind:         snowflakeKind,
				Endpoint:     endpoint,
				Warehouse:    warehouse,
				AllowedScope: allowedScope,
				ClientId:     clientId,
				ClientSecret: clientSecret,
			},
		}
	} else if strings.ToLower(catalogIntegrationType) == unity {
		if !cmd.Flags().Changed("workspace-endpoint") { // we only need to check for one since this flag set is marked as required together
			return fmt.Errorf("`--workspace-endpoint`, `--catalog-name`, `--unity-client-id` and `--unity-client-secret` flags are required for catalog integration type `unity`.")
		}
		workspaceEndpoint, err := cmd.Flags().GetString("workspace-endpoint")
		if err != nil {
			return err
		}
		catalogName, err := cmd.Flags().GetString("catalog-name")
		if err != nil {
			return err
		}
		clientId, err := cmd.Flags().GetString("unity-client-id")
		if err != nil {
			return err
		}
		clientSecret, err := cmd.Flags().GetString("unity-client-secret")
		if err != nil {
			return err
		}

		createCatalogIntegration.Spec.Config = &tableflowv1.TableflowV1CatalogIntegrationSpecConfigOneOf{
			TableflowV1CatalogIntegrationUnitySpec: &tableflowv1.TableflowV1CatalogIntegrationUnitySpec{
				Kind:              unityKind,
				WorkspaceEndpoint: workspaceEndpoint,
				CatalogName:       catalogName,
				ClientId:          clientId,
				ClientSecret:      clientSecret,
			},
		}
	}

	catalogIntegration, err := c.V2Client.CreateCatalogIntegration(createCatalogIntegration)
	if err != nil {
		return err
	}

	return printCatalogIntegrationTable(cmd, catalogIntegration)
}
