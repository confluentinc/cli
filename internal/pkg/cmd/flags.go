package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"

	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	dynamicconfig "github.com/confluentinc/cli/internal/pkg/dynamic-config"
	"github.com/confluentinc/cli/internal/pkg/kafka"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/types"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func AddApiKeyFlag(cmd *cobra.Command, c *AuthenticatedCLICommand) {
	cmd.Flags().String("api-key", "", "API key.")

	RegisterFlagCompletionFunc(cmd, "api-key", func(cmd *cobra.Command, args []string) []string {
		if err := c.PersistentPreRunE(cmd, args); err != nil {
			return nil
		}

		return AutocompleteApiKeys(c.V2Client)
	})
}

func AddApiSecretFlag(cmd *cobra.Command) {
	cmd.Flags().String("api-secret", "", "API key secret.")
}

func AutocompleteApiKeys(client *ccloudv2.Client) []string {
	apiKeys, err := client.ListApiKeys("", "")
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(apiKeys))
	for i, apiKey := range apiKeys {
		if !apiKey.Spec.HasOwner() {
			continue
		}
		suggestions[i] = fmt.Sprintf("%s\t%s", apiKey.GetId(), apiKey.Spec.GetDescription())
	}
	return suggestions
}

func AddAvailabilityFlag(cmd *cobra.Command) {
	cmd.Flags().String("availability", kafka.Availabilities[0], fmt.Sprintf("Specify the availability of the cluster as %s.", utils.ArrayToCommaDelimitedString(kafka.Availabilities, "or")))
	RegisterFlagCompletionFunc(cmd, output.FlagName, func(_ *cobra.Command, _ []string) []string { return kafka.Availabilities })
}

func AddByokKeyFlag(cmd *cobra.Command, command *AuthenticatedCLICommand) {
	cmd.Flags().String("byok", "", `Confluent Cloud Key ID of a registered encryption key (AWS and Azure only, use "confluent byok create" to register a key).`)

	RegisterFlagCompletionFunc(cmd, "byok", func(cmd *cobra.Command, args []string) []string {
		if err := command.PersistentPreRunE(cmd, args); err != nil {
			return nil
		}

		return AutocompleteByokKeyIds(command.V2Client)
	})
}

func AutocompleteByokKeyIds(client *ccloudv2.Client) []string {
	keys, err := client.ListByokKeys("", "")
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(keys))
	for i, key := range keys {
		suggestions[i] = key.GetId()
	}
	return suggestions
}

func AddByokProviderFlag(cmd *cobra.Command) {
	cmd.Flags().String("provider", "", fmt.Sprintf("Specify the provider as %s.", utils.ArrayToCommaDelimitedString([]string{"aws", "azure"}, "or")))

	RegisterFlagCompletionFunc(cmd, "provider", func(_ *cobra.Command, _ []string) []string {
		return []string{"aws", "azure"}
	})
}

func AddByokStateFlag(cmd *cobra.Command) {
	cmd.Flags().String("state", "", fmt.Sprintf("Specify the state as %s.", utils.ArrayToCommaDelimitedString([]string{"in-use", "available"}, "or")))

	RegisterFlagCompletionFunc(cmd, "state", func(_ *cobra.Command, _ []string) []string {
		return []string{"in-use", "available"}
	})
}

func AddCloudFlag(cmd *cobra.Command) {
	cmd.Flags().String("cloud", "", fmt.Sprintf("Specify the cloud provider as %s.", utils.ArrayToCommaDelimitedString(kafka.Clouds, "or")))
	RegisterFlagCompletionFunc(cmd, "cloud", func(_ *cobra.Command, _ []string) []string { return kafka.Clouds })
}

func AddClusterFlag(cmd *cobra.Command, c *AuthenticatedCLICommand) {
	cmd.Flags().String("cluster", "", "Kafka cluster ID.")
	RegisterFlagCompletionFunc(cmd, "cluster", func(cmd *cobra.Command, args []string) []string {
		if err := c.PersistentPreRunE(cmd, args); err != nil {
			return nil
		}

		environmentId, err := c.Context.EnvironmentId()
		if err != nil {
			return nil
		}
		return AutocompleteClusters(environmentId, c.V2Client)
	})
}

func AutocompleteClusters(environmentId string, client *ccloudv2.Client) []string {
	clusters, err := client.ListKafkaClusters(environmentId)
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(clusters))
	for i, cluster := range clusters {
		suggestions[i] = fmt.Sprintf("%s\t%s", cluster.GetId(), cluster.Spec.GetDisplayName())
	}
	return suggestions
}

func AddContextFlag(cmd *cobra.Command, command *CLICommand) {
	cmd.Flags().String("context", "", "CLI context name.")

	RegisterFlagCompletionFunc(cmd, "context", func(cmd *cobra.Command, args []string) []string {
		if err := command.PersistentPreRunE(cmd, args); err != nil {
			return nil
		}

		return AutocompleteContexts(command.Config.Config)
	})
}

func AutocompleteContexts(cfg *v1.Config) []string {
	return types.GetSortedKeys(cfg.Contexts)
}

func AddEnvironmentFlag(cmd *cobra.Command, command *AuthenticatedCLICommand) {
	cmd.Flags().String("environment", "", "Environment ID.")
	RegisterFlagCompletionFunc(cmd, "environment", func(cmd *cobra.Command, args []string) []string {
		if err := command.PersistentPreRunE(cmd, args); err != nil {
			return nil
		}

		return AutocompleteEnvironments(command.Client, command.V2Client, command.Context)
	})
}

func AutocompleteEnvironments(v1Client *ccloudv1.Client, v2Client *ccloudv2.Client, ctx *dynamicconfig.DynamicContext) []string {
	environments, err := v2Client.ListOrgEnvironments()
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(environments))
	for i, environment := range environments {
		suggestions[i] = fmt.Sprintf("%s\t%s", environment.GetId(), environment.GetDisplayName())
	}

	user, err := v1Client.Auth.User()
	if err != nil {
		return nil
	}

	if auditLog := user.GetOrganization().GetAuditLog(); auditLog.GetServiceAccountId() != 0 {
		environment, err := v2Client.GetOrgEnvironment(auditLog.GetAccountId())
		if err != nil {
			return nil
		}
		suggestions = append(suggestions, fmt.Sprintf("%s\t%s", auditLog.GetAccountId(), environment.GetDisplayName()))
	}

	return suggestions
}

func AddForceFlag(cmd *cobra.Command) {
	cmd.Flags().Bool("force", false, "Skip the deletion confirmation prompt.")
}

func AddKsqlClusterFlag(cmd *cobra.Command, c *AuthenticatedCLICommand) {
	cmd.Flags().String("ksql-cluster", "", "KSQL cluster for the pipeline.")
	RegisterFlagCompletionFunc(cmd, "ksql-cluster", func(cmd *cobra.Command, args []string) []string {
		if err := c.PersistentPreRunE(cmd, args); err != nil {
			return nil
		}

		environmentId, err := c.Context.EnvironmentId()
		if err != nil {
			return nil
		}
		return autocompleteKSQLClusters(environmentId, c.V2Client)
	})
}

func autocompleteKSQLClusters(environmentId string, client *ccloudv2.Client) []string {
	clusters, err := client.ListKsqlClusters(environmentId)
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(clusters))
	for i, cluster := range clusters {
		suggestions[i] = fmt.Sprintf("%s\t%s", cluster.GetId(), cluster.Spec.GetDisplayName())
	}
	return suggestions
}

func AddMechanismFlag(cmd *cobra.Command, command *AuthenticatedCLICommand) {
	cmd.Flags().String("sasl-mechanism", "PLAIN", "SASL_SSL mechanism used for authentication.")
	RegisterFlagCompletionFunc(cmd, "sasl-mechanism", func(cmd *cobra.Command, args []string) []string {
		if err := command.PersistentPreRunE(cmd, args); err != nil {
			return nil
		}

		protocol, _ := cmd.Flags().GetString("protocol")
		return autocompleteMechanisms(protocol)
	})
}

func autocompleteMechanisms(protocol string) []string {
	switch protocol {
	default:
		return nil
	case "SSL":
		return nil
	case "SASL_SSL":
		return []string{"PLAIN", "OAUTHBEARER"}
	}
}

func AddProducerConfigFileFlag(cmd *cobra.Command) {
	cmd.Flags().String("config-file", "", "The path to the configuration file for the producer client, in JSON or Avro format.")
}

func AddConsumerConfigFileFlag(cmd *cobra.Command) {
	cmd.Flags().String("config-file", "", "The path to the configuration file for the consumer client, in JSON or Avro format.")
}

func AddOutputFlag(cmd *cobra.Command) {
	AddOutputFlagWithDefaultValue(cmd, output.Human.String())
}

func AddOutputFlagWithDefaultValue(cmd *cobra.Command, defaultValue string) {
	cmd.Flags().StringP(output.FlagName, "o", defaultValue, fmt.Sprintf("Specify the output format as %s.", utils.ArrayToCommaDelimitedString(output.ValidFlagValues, "or")))

	RegisterFlagCompletionFunc(cmd, output.FlagName, func(_ *cobra.Command, _ []string) []string {
		return output.ValidFlagValues
	})
}

func AddPrincipalFlag(cmd *cobra.Command, command *AuthenticatedCLICommand) {
	cmd.Flags().String("principal", "", "Principal ID.")
	RegisterFlagCompletionFunc(cmd, "principal", func(cmd *cobra.Command, args []string) []string {
		if err := command.PersistentPreRunE(cmd, args); err != nil {
			return nil
		}

		return AutocompleteServiceAccounts(command.V2Client)
	})
}

func AddProtocolFlag(cmd *cobra.Command) {
	cmd.Flags().String("protocol", "SSL", "Security protocol used to communicate with brokers.")
	RegisterFlagCompletionFunc(cmd, "protocol", func(_ *cobra.Command, _ []string) []string { return kafka.Protocols })
}

func AddProviderFlag(cmd *cobra.Command, command *AuthenticatedCLICommand) {
	cmd.Flags().String("provider", "", "ID of this pool's identity provider.")

	RegisterFlagCompletionFunc(cmd, "provider", func(cmd *cobra.Command, args []string) []string {
		if err := command.PersistentPreRunE(cmd, args); err != nil {
			return nil
		}

		return AutocompleteIdentityProviders(command.V2Client)
	})
}

func AutocompleteIdentityProviders(client *ccloudv2.Client) []string {
	identityProviders, err := client.ListIdentityProviders()
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(identityProviders))
	for i, identityProvider := range identityProviders {
		description := fmt.Sprintf("%s: %s", identityProvider.GetDisplayName(), identityProvider.GetDescription())
		suggestions[i] = fmt.Sprintf("%s\t%s", identityProvider.GetId(), description)
	}
	return suggestions
}

func AutocompleteIdentityPools(client *ccloudv2.Client, providerId string) []string {
	identityPools, err := client.ListIdentityPools(providerId)
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(identityPools))
	for i, identityPool := range identityPools {
		description := fmt.Sprintf("%s: %s", identityPool.GetDisplayName(), identityPool.GetDescription())
		suggestions[i] = fmt.Sprintf("%s\t%s", identityPool.GetId(), description)
	}
	return suggestions
}

func AddRegionFlag(cmd *cobra.Command, command *AuthenticatedCLICommand) {
	cmd.Flags().String("region", "", `Cloud region ID for cluster (use "confluent kafka region list" to see all).`)
	RegisterFlagCompletionFunc(cmd, "region", func(cmd *cobra.Command, args []string) []string {
		if err := command.PersistentPreRunE(cmd, args); err != nil {
			return nil
		}

		cloud, _ := cmd.Flags().GetString("cloud")
		return autocompleteRegions(command.Client, cloud)
	})
}

func autocompleteRegions(client *ccloudv1.Client, cloud string) []string {
	regions, err := kafka.ListRegions(client, cloud)
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(regions))
	for i, region := range regions {
		suggestions[i] = region.RegionId
	}
	return suggestions
}

func AddSchemaTypeFlag(cmd *cobra.Command) {
	arr := []string{"avro", "json", "protobuf"}
	str := utils.ArrayToCommaDelimitedString(arr, "or")

	cmd.Flags().String("type", "", fmt.Sprintf("Specify the schema type as %s.", str))

	RegisterFlagCompletionFunc(cmd, "type", func(_ *cobra.Command, _ []string) []string {
		return arr
	})
}

func AddServiceAccountFlag(cmd *cobra.Command, command *AuthenticatedCLICommand) {
	cmd.Flags().String("service-account", "", "Service account ID.")

	RegisterFlagCompletionFunc(cmd, "service-account", func(cmd *cobra.Command, args []string) []string {
		if err := command.PersistentPreRunE(cmd, args); err != nil {
			return nil
		}

		return AutocompleteServiceAccounts(command.V2Client)
	})
}

func AutocompleteServiceAccounts(client *ccloudv2.Client) []string {
	serviceAccounts, err := client.ListIamServiceAccounts()
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(serviceAccounts))
	for i, serviceAccount := range serviceAccounts {
		description := fmt.Sprintf("%s: %s", serviceAccount.GetDisplayName(), serviceAccount.GetDescription())
		suggestions[i] = fmt.Sprintf("%s\t%s", serviceAccount.GetId(), description)
	}
	return suggestions
}

func AutocompleteUsers(client *ccloudv2.Client) []string {
	users, err := client.ListIamUsers()
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(users))
	for i, user := range users {
		suggestions[i] = fmt.Sprintf("%s\t%s", user.GetId(), user.GetFullName())
	}
	return suggestions
}

func AddTypeFlag(cmd *cobra.Command) {
	cmd.Flags().String("type", kafka.Types[0], fmt.Sprintf("Specify the type of the Kafka cluster as %s.", utils.ArrayToCommaDelimitedString(kafka.Types, "or")))
	RegisterFlagCompletionFunc(cmd, output.FlagName, func(_ *cobra.Command, _ []string) []string { return kafka.Types })
}

func AddValueFormatFlag(cmd *cobra.Command) {
	arr := []string{"string", "avro", "jsonschema", "protobuf"}
	str := utils.ArrayToCommaDelimitedString(arr, "or")

	cmd.Flags().String("value-format", "string", fmt.Sprintf("Format of message value as %s. Note that schema references are not supported for avro.", str))

	RegisterFlagCompletionFunc(cmd, "value-format", func(_ *cobra.Command, _ []string) []string {
		return arr
	})
}

func AddLinkFlag(cmd *cobra.Command, command *AuthenticatedCLICommand) {
	cmd.Flags().String("link", "", "Name of cluster link.")

	RegisterFlagCompletionFunc(cmd, "link", func(cmd *cobra.Command, args []string) []string {
		if err := command.PersistentPreRunE(cmd, args); err != nil {
			return nil
		}

		return AutocompleteLinks(command)
	})
}

func AutocompleteLinks(command *AuthenticatedCLICommand) []string {
	kafkaREST, err := command.GetKafkaREST()
	if err != nil || kafkaREST == nil {
		return nil
	}

	kafkaClusterConfig, err := command.Context.GetKafkaClusterForCommand()
	if err != nil {
		return nil
	}

	links, _, err := kafkaREST.CloudClient.ListKafkaLinks(kafkaClusterConfig.ID)
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(links.Data))
	for i, link := range links.Data {
		suggestions[i] = link.GetLinkName()
	}
	return suggestions
}
