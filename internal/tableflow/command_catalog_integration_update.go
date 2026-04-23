package tableflow

import (
	"github.com/spf13/cobra"

	tableflowv1 "github.com/confluentinc/ccloud-sdk-go-v2/tableflow/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/kafka"
)

func (c *command) newCatalogIntegrationUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "update <id>",
		Short:             "Update a catalog integration.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validCatalogIntegrationArgs),
		RunE:              c.updateCatalogIntegration,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Update a catalog integration name.",
				Code: "confluent tableflow catalog-integration update tci-abc123 --name new-name",
			},
			examples.Example{
				Text: "Create a Snowflake catalog integration.",
				Code: "confluent tableflow catalog-integration update tc-abc123 --endpoint https://vuser1_polaris.snowflakecomputing.com/ --warehouse catalog-name --allowed-scope session:role:R1 --client-id $CLIENT_ID --client-secret $CLIENT_SECRET",
			},
		),
	}

	cmd.Flags().String("name", "", "Name of the catalog integration.")
	cmd.Flags().String("endpoint", "", "Specify the The catalog integration connection endpoint for Snowflake Open Catalog.")
	cmd.Flags().String("warehouse", "", "Specify the warehouse name of the Snowflake Open Catalog.")
	cmd.Flags().String("allowed-scope", "", "Specify the allowed scope of the Snowflake Open Catalog.")
	cmd.Flags().String("client-id", "", "Specify the client id.")
	cmd.Flags().String("client-secret", "", "Specify the client secret.")
	cmd.Flags().String("custom-database", "", "Specify the custom database name for AWS Glue.")
	cmd.Flags().String("custom-namespace", "", "Specify the custom namespace for Snowflake Open Catalog.")
	cmd.Flags().String("custom-schema", "", "Specify the custom schema name for Unity Catalog.")

	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cmd.MarkFlagsOneRequired("name", "endpoint", "warehouse", "allowed-scope", "client-id", "client-secret", "custom-database", "custom-namespace", "custom-schema")

	return cmd
}

func (c *command) updateCatalogIntegration(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	cluster, err := kafka.GetClusterForCommand(c.V2Client, c.Context)
	if err != nil {
		return err
	}

	updateCatalogIntegration := tableflowv1.TableflowV1CatalogIntegrationUpdateRequest{
		Spec: &tableflowv1.TableflowV1CatalogIntegrationUpdateSpec{
			Environment:  tableflowv1.GlobalObjectReference{Id: environmentId},
			KafkaCluster: tableflowv1.EnvScopedObjectReference{Id: cluster.GetId()},
		},
	}

	if cmd.Flags().Changed("name") {
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			return err
		}

		updateCatalogIntegration.Spec.SetDisplayName(name)

		// Config is required, so when only updating the name, we retrieve the config type from the backend
		catalogIntegration, err := c.V2Client.GetCatalogIntegration(environmentId, cluster.GetId(), args[0])
		if err != nil {
			return err
		}
		catalogIntegrationType, err := getCatalogIntegrationType(catalogIntegration)
		if err != nil {
			return err
		}
		if catalogIntegrationType == aws {
			updateCatalogIntegration.Spec.SetConfig(tableflowv1.TableflowV1CatalogIntegrationUpdateSpecConfigOneOf{
				TableflowV1CatalogIntegrationAwsGlueUpdateSpec: &tableflowv1.TableflowV1CatalogIntegrationAwsGlueUpdateSpec{
					Kind: awsGlueKind,
				},
			})
		}
		if catalogIntegrationType == snowflake {
			updateCatalogIntegration.Spec.SetConfig(tableflowv1.TableflowV1CatalogIntegrationUpdateSpecConfigOneOf{
				TableflowV1CatalogIntegrationSnowflakeUpdateSpec: &tableflowv1.TableflowV1CatalogIntegrationSnowflakeUpdateSpec{
					Kind: snowflakeKind,
				},
			})
		}
		if catalogIntegrationType == unity {
			updateCatalogIntegration.Spec.SetConfig(tableflowv1.TableflowV1CatalogIntegrationUpdateSpecConfigOneOf{
				TableflowV1CatalogIntegrationUnityUpdateSpec: &tableflowv1.TableflowV1CatalogIntegrationUnityUpdateSpec{
					Kind: unityKind,
				},
			})
		}
	}

	if cmd.Flags().Changed("custom-database") {
		customDatabase, err := cmd.Flags().GetString("custom-database")
		if err != nil {
			return err
		}
		updateCatalogIntegration.Spec.SetConfig(tableflowv1.TableflowV1CatalogIntegrationUpdateSpecConfigOneOf{
			TableflowV1CatalogIntegrationAwsGlueUpdateSpec: &tableflowv1.TableflowV1CatalogIntegrationAwsGlueUpdateSpec{
				Kind: awsGlueKind,
			},
		})
		updateCatalogIntegration.Spec.Config.TableflowV1CatalogIntegrationAwsGlueUpdateSpec.SetCustomDatabase(customDatabase)
	}

	if cmd.Flags().Changed("endpoint") || cmd.Flags().Changed("warehouse") || cmd.Flags().Changed("allowed-scope") || cmd.Flags().Changed("client-id") || cmd.Flags().Changed("client-secret") || cmd.Flags().Changed("custom-namespace") {
		updateCatalogIntegration.Spec.SetConfig(tableflowv1.TableflowV1CatalogIntegrationUpdateSpecConfigOneOf{
			TableflowV1CatalogIntegrationSnowflakeUpdateSpec: &tableflowv1.TableflowV1CatalogIntegrationSnowflakeUpdateSpec{
				Kind: snowflakeKind,
			},
		})
		if cmd.Flags().Changed("endpoint") {
			endpoint, err := cmd.Flags().GetString("endpoint")
			if err != nil {
				return err
			}
			updateCatalogIntegration.Spec.Config.TableflowV1CatalogIntegrationSnowflakeUpdateSpec.SetEndpoint(endpoint)
		}
		if cmd.Flags().Changed("warehouse") {
			warehouse, err := cmd.Flags().GetString("warehouse")
			if err != nil {
				return err
			}
			updateCatalogIntegration.Spec.Config.TableflowV1CatalogIntegrationSnowflakeUpdateSpec.SetWarehouse(warehouse)
		}
		if cmd.Flags().Changed("allowed-scope") {
			allowedScope, err := cmd.Flags().GetString("allowed-scope")
			if err != nil {
				return err
			}
			updateCatalogIntegration.Spec.Config.TableflowV1CatalogIntegrationSnowflakeUpdateSpec.SetAllowedScope(allowedScope)
		}
		if cmd.Flags().Changed("client-id") {
			clientId, err := cmd.Flags().GetString("client-id")
			if err != nil {
				return err
			}
			updateCatalogIntegration.Spec.Config.TableflowV1CatalogIntegrationSnowflakeUpdateSpec.SetClientId(clientId)
		}
		if cmd.Flags().Changed("client-secret") {
			clientSecret, err := cmd.Flags().GetString("client-secret")
			if err != nil {
				return err
			}
			updateCatalogIntegration.Spec.Config.TableflowV1CatalogIntegrationSnowflakeUpdateSpec.SetClientSecret(clientSecret)
		}
		if cmd.Flags().Changed("custom-namespace") {
			customNamespace, err := cmd.Flags().GetString("custom-namespace")
			if err != nil {
				return err
			}
			updateCatalogIntegration.Spec.Config.TableflowV1CatalogIntegrationSnowflakeUpdateSpec.SetCustomNamespace(customNamespace)
		}
	}

	if cmd.Flags().Changed("custom-schema") {
		customSchema, err := cmd.Flags().GetString("custom-schema")
		if err != nil {
			return err
		}
		updateCatalogIntegration.Spec.SetConfig(tableflowv1.TableflowV1CatalogIntegrationUpdateSpecConfigOneOf{
			TableflowV1CatalogIntegrationUnityUpdateSpec: &tableflowv1.TableflowV1CatalogIntegrationUnityUpdateSpec{
				Kind: unityKind,
			},
		})
		updateCatalogIntegration.Spec.Config.TableflowV1CatalogIntegrationUnityUpdateSpec.SetCustomSchema(customSchema)
	}

	catalogIntegration, err := c.V2Client.UpdateCatalogIntegration(args[0], updateCatalogIntegration)
	if err != nil {
		return err
	}

	return printCatalogIntegrationTable(cmd, catalogIntegration)
}
