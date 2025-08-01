package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"

	"github.com/confluentinc/cli/v4/pkg/ccloudv2"
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/kafka"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/serdes"
	"github.com/confluentinc/cli/v4/pkg/types"
	"github.com/confluentinc/cli/v4/pkg/utils"
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
	cmd.Flags().String("api-secret", "", "API secret.")
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
	cmd.Flags().String("availability", "single-zone", fmt.Sprintf("Specify the availability of the cluster as %s.", utils.ArrayToCommaDelimitedString(kafka.Availabilities, "or")))
	RegisterFlagCompletionFunc(cmd, "availability", func(_ *cobra.Command, _ []string) []string { return kafka.Availabilities })
}

func AddByokKeyFlag(cmd *cobra.Command, command *AuthenticatedCLICommand) {
	cmd.Flags().String("byok", "", `Confluent Cloud Key ID of a registered encryption key (use "confluent byok create" to register a key).`)

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

func AddByokStateFlag(cmd *cobra.Command) {
	cmd.Flags().String("state", "", fmt.Sprintf("Specify the state as %s.", utils.ArrayToCommaDelimitedString([]string{"in-use", "available"}, "or")))
	RegisterFlagCompletionFunc(cmd, "state", func(_ *cobra.Command, _ []string) []string { return []string{"in-use", "available"} })
}

func AddCloudFlag(cmd *cobra.Command) {
	cmd.Flags().String("cloud", "", fmt.Sprintf("Specify the cloud provider as %s.", utils.ArrayToCommaDelimitedString(kafka.Clouds, "or")))
	RegisterFlagCompletionFunc(cmd, "cloud", func(_ *cobra.Command, _ []string) []string { return kafka.Clouds })
}

func AddCloudAwsAzureFlag(cmd *cobra.Command) {
	cmd.Flags().String("cloud", "", fmt.Sprintf("Specify the cloud provider as %s.", utils.ArrayToCommaDelimitedString(kafka.Clouds[:2], "or")))
	RegisterFlagCompletionFunc(cmd, "cloud", func(_ *cobra.Command, _ []string) []string { return kafka.Clouds[:2] })
}

func AddCloudAwsFlag(cmd *cobra.Command) {
	cmd.Flags().String("cloud", "", fmt.Sprintf("Specify the cloud provider as %s.", utils.ArrayToCommaDelimitedString(kafka.Clouds[:1], "or")))
	RegisterFlagCompletionFunc(cmd, "cloud", func(_ *cobra.Command, _ []string) []string { return kafka.Clouds[:1] })
}

func AddListCloudFlag(cmd *cobra.Command) {
	cmd.Flags().StringSlice("cloud", nil, "A comma-separated list of cloud providers.")
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

func AddEndpointFlag(cmd *cobra.Command, c *AuthenticatedCLICommand) {
	cmd.Flags().String("kafka-endpoint", "", "Endpoint to be used for this Kafka cluster.")
}

func AutocompleteEndpoints(environmentId string, client *ccloudv2.Client) []string {
	// nice-to-have, tracked by JIRA ticket APIE-439
	return nil
}

func AddConfigFlag(cmd *cobra.Command) {
	cmd.Flags().StringSlice("config", []string{}, `A comma-separated list of "key=value" pairs, or path to a configuration file containing a newline-separated list of "key=value" pairs.`)
}

func AddContextFlag(cmd *cobra.Command, command *CLICommand) {
	cmd.Flags().String("context", "", "CLI context name.")

	RegisterFlagCompletionFunc(cmd, "context", func(cmd *cobra.Command, args []string) []string {
		if err := command.PersistentPreRunE(cmd, args); err != nil {
			return nil
		}

		return AutocompleteContexts(command.Config)
	})
}

func AutocompleteContexts(cfg *config.Config) []string {
	return types.GetSortedKeys(cfg.Contexts)
}

func AddEnvironmentFlag(cmd *cobra.Command, command *AuthenticatedCLICommand) {
	cmd.Flags().String("environment", "", "Environment ID.")
	RegisterFlagCompletionFunc(cmd, "environment", func(cmd *cobra.Command, args []string) []string {
		if err := command.PersistentPreRunE(cmd, args); err != nil {
			return nil
		}

		return AutocompleteEnvironments(command.Client, command.V2Client)
	})
}

func AutocompleteEnvironments(v1Client *ccloudv1.Client, v2Client *ccloudv2.Client) []string {
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

func AddDryRunFlag(cmd *cobra.Command) {
	cmd.Flags().Bool("dry-run", false, "Run the command without committing changes.")
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

func AddFilterFlag(cmd *cobra.Command) {
	cmd.Flags().String("filter", "true", "A supported Common Expression Language (CEL) filter expression.")
}

func AddExternalIdentifierFlag(cmd *cobra.Command) {
	cmd.Flags().String("external-identifier", "", "External Identifier for this pool.")
}

func AutocompleteGroupMappings(client *ccloudv2.Client) []string {
	groupMappings, err := client.ListGroupMappings()
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(groupMappings))
	for i, groupMapping := range groupMappings {
		description := fmt.Sprintf("%s: %s", groupMapping.GetDisplayName(), groupMapping.GetDescription())
		suggestions[i] = fmt.Sprintf("%s\t%s", groupMapping.GetId(), description)
	}
	return suggestions
}

func AutocompleteCertificatePool(client *ccloudv2.Client, provider string) []string {
	certificatePools, err := client.ListCertificatePool(provider)
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(certificatePools))
	for i, certificatePool := range certificatePools {
		description := fmt.Sprintf("%s: %s", certificatePool.GetDisplayName(), certificatePool.GetDescription())
		suggestions[i] = fmt.Sprintf("%s\t%s", certificatePool.GetId(), description)
	}
	return suggestions
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
	case "SASL_SSL":
		return []string{"PLAIN", "OAUTHBEARER"}
	default:
		return nil
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

func AddOutputFlagWithHumanRestricted(cmd *cobra.Command) {
	cmd.Flags().StringP(output.FlagName, "o", output.JSON.String(), fmt.Sprintf("Specify the output format as %s.", utils.ArrayToCommaDelimitedString(output.ValidFlagValuesHumanRestricted, "or")))
	RegisterFlagCompletionFunc(cmd, output.FlagName, func(_ *cobra.Command, _ []string) []string { return output.ValidFlagValuesHumanRestricted })
}

func AddOutputFlagWithDefaultValue(cmd *cobra.Command, defaultValue string) {
	cmd.Flags().StringP(output.FlagName, "o", defaultValue, fmt.Sprintf("Specify the output format as %s.", utils.ArrayToCommaDelimitedString(output.ValidFlagValues, "or")))
	RegisterFlagCompletionFunc(cmd, output.FlagName, func(_ *cobra.Command, _ []string) []string { return output.ValidFlagValues })
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
	protocols := []string{"PLAINTEXT", "SASL_SSL", "SSL"}
	cmd.Flags().String("protocol", "SSL", fmt.Sprintf("Specify the broker communication protocol as %s.", utils.ArrayToCommaDelimitedString(protocols, "or")))
	RegisterFlagCompletionFunc(cmd, "protocol", func(_ *cobra.Command, _ []string) []string { return protocols })
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

func AutocompleteCertificateAuthorities(client *ccloudv2.Client) []string {
	certificateAuthorities, err := client.ListCertificateAuthorities()
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(certificateAuthorities))
	for i, certificateAuthority := range certificateAuthorities {
		suggestions[i] = fmt.Sprintf("%s\t%s", certificateAuthority.GetId(), certificateAuthority.GetDisplayName())
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

func AddResourceGroupFlag(cmd *cobra.Command) {
	var arr []string = []string{"management", "multiple"}
	cmd.Flags().String("resource-group", "multiple", fmt.Sprintf("Name of resource group: %s.", utils.ArrayToCommaDelimitedString(arr, "or")))
	RegisterFlagCompletionFunc(cmd, "resource-group", func(_ *cobra.Command, _ []string) []string {
		return arr
	})
}

func AutocompleteIpFilters(client *ccloudv2.Client) []string {
	ipFilters, err := client.ListIamIpFilters("", "")
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(ipFilters))
	for i, ipFilter := range ipFilters {
		var ipGroupIds []string
		for _, ipGroup := range ipFilter.GetIpGroups() {
			ipGroupIds = append(ipGroupIds, ipGroup.GetId())
		}
		description := fmt.Sprintf("%s: %s, %s", ipFilter.GetFilterName(), ipFilter.GetResourceGroup(), ipGroupIds)
		suggestions[i] = fmt.Sprintf("%s\t%s", ipFilter.GetId(), description)
	}
	return suggestions
}

func AutocompleteIpGroups(client *ccloudv2.Client) []string {
	ipGroups, err := client.ListIamIpGroups()
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(ipGroups))
	for i, ipGroup := range ipGroups {
		description := fmt.Sprintf("%s: %s", ipGroup.GetGroupName(), ipGroup.GetCidrBlocks())
		suggestions[i] = fmt.Sprintf("%s\t%s", ipGroup.GetId(), description)
	}
	return suggestions
}

func AddRegionFlagKafka(cmd *cobra.Command, command *AuthenticatedCLICommand) {
	cmd.Flags().String("region", "", `Cloud region for Kafka (use "confluent kafka region list" to see all).`)
	RegisterFlagCompletionFunc(cmd, "region", func(cmd *cobra.Command, args []string) []string {
		if err := command.PersistentPreRunE(cmd, args); err != nil {
			return nil
		}

		cloud, _ := cmd.Flags().GetString("cloud")

		regions, err := kafka.ListRegions(command.Client, cloud)
		if err != nil {
			return nil
		}

		suggestions := make([]string, len(regions))
		for i, region := range regions {
			suggestions[i] = region.RegionId
		}
		return suggestions
	})
}

func AddRegionFlagFlink(cmd *cobra.Command, command *AuthenticatedCLICommand) {
	cmd.Flags().String("region", "", `Cloud region for Flink (use "confluent flink region list" to see all).`)
	RegisterFlagCompletionFunc(cmd, "region", func(cmd *cobra.Command, args []string) []string {
		if err := command.PersistentPreRunE(cmd, args); err != nil {
			return nil
		}

		cloud, _ := cmd.Flags().GetString("cloud")
		regionName, _ := cmd.Flags().GetString("region")

		regions, err := command.V2Client.ListFlinkRegions(strings.ToUpper(cloud), regionName)
		if err != nil {
			return nil
		}

		suggestions := make([]string, len(regions))
		for i, region := range regions {
			suggestions[i] = region.GetRegionName()
		}
		return suggestions
	})
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
	cmd.Flags().String("type", "basic", fmt.Sprintf("Specify the type of the Kafka cluster as %s.", utils.ArrayToCommaDelimitedString(kafka.Types, "or")))
	RegisterFlagCompletionFunc(cmd, "type", func(_ *cobra.Command, _ []string) []string { return kafka.Types })
}

func AddKeyFormatFlag(cmd *cobra.Command) {
	cmd.Flags().String("key-format", "string", fmt.Sprintf("Format of message key as %s. Note that schema references are not supported for Avro.", utils.ArrayToCommaDelimitedString(serdes.Formats, "or")))
	RegisterFlagCompletionFunc(cmd, "key-format", func(_ *cobra.Command, _ []string) []string { return serdes.Formats })
}

func AddValueFormatFlag(cmd *cobra.Command) {
	cmd.Flags().String("value-format", "string", fmt.Sprintf("Format message value as %s. Note that schema references are not supported for Avro.", utils.ArrayToCommaDelimitedString(serdes.Formats, "or")))
	RegisterFlagCompletionFunc(cmd, "value-format", func(_ *cobra.Command, _ []string) []string { return serdes.Formats })
}

func AddLinkFlag(cmd *cobra.Command, c *AuthenticatedCLICommand) {
	cmd.Flags().String("link", "", "Name of cluster link.")

	RegisterFlagCompletionFunc(cmd, "link", func(cmd *cobra.Command, args []string) []string {
		if err := c.PersistentPreRunE(cmd, args); err != nil {
			return nil
		}

		return AutocompleteLinks(cmd, c)
	})
}

func AddAlgorithmFlag(cmd *cobra.Command) {
	cmd.Flags().String("algorithm", "", fmt.Sprintf("Use algorithm %s for the Data Encryption Key (DEK).", utils.ArrayToCommaDelimitedString(serdes.DekAlgorithms, "or")))
	RegisterFlagCompletionFunc(cmd, "algorithm", func(_ *cobra.Command, _ []string) []string { return serdes.DekAlgorithms })
}

func AddKmsTypeFlag(cmd *cobra.Command) {
	cmd.Flags().String("kms-type", "", fmt.Sprintf("The type of Key Management Service (KMS), typically one of %s.", utils.ArrayToCommaDelimitedString(serdes.KmsTypes, "or")))
	RegisterFlagCompletionFunc(cmd, "kms-type", func(_ *cobra.Command, _ []string) []string { return serdes.KmsTypes })
}

func AutocompleteLinks(cmd *cobra.Command, c *AuthenticatedCLICommand) []string {
	kafkaREST, err := c.GetKafkaREST(cmd)
	if err != nil {
		return nil
	}

	links, err := kafkaREST.CloudClient.ListKafkaLinks()
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(links))
	for i, link := range links {
		suggestions[i] = link.GetLinkName()
	}
	return suggestions
}

func AutocompleteConsumerGroups(cmd *cobra.Command, c *AuthenticatedCLICommand) []string {
	kafkaREST, err := c.GetKafkaREST(cmd)
	if err != nil {
		return nil
	}

	consumerGroups, err := kafkaREST.CloudClient.ListKafkaConsumerGroups()
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(consumerGroups))
	for i, consumerGroup := range consumerGroups {
		suggestions[i] = consumerGroup.GetConsumerGroupId()
	}
	return suggestions
}

func AddNetworkFlag(cmd *cobra.Command, c *AuthenticatedCLICommand) {
	cmd.Flags().String("network", "", "Network ID.")
	RegisterFlagCompletionFunc(cmd, "network", func(cmd *cobra.Command, args []string) []string {
		if err := c.PersistentPreRunE(cmd, args); err != nil {
			return nil
		}

		environmentId, err := c.Context.EnvironmentId()
		if err != nil {
			return nil
		}
		return AutocompleteNetworks(environmentId, c.V2Client)
	})
}

func AutocompleteNetworks(environmentId string, client *ccloudv2.Client) []string {
	networks, err := client.ListNetworks(environmentId, nil, nil, nil, nil, nil, nil)
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(networks))
	for i, network := range networks {
		suggestions[i] = fmt.Sprintf("%s\t%s", network.GetId(), network.Spec.GetDisplayName())
	}
	return suggestions
}

func AddMDSOnPremMTLSFlags(cmd *cobra.Command) {
	cmd.Flags().String("client-cert-path", "", "Path to client cert to be verified by MDS. Include for mTLS authentication.")
	cmd.Flags().String("client-key-path", "", "Path to client private key, include for mTLS authentication.")
	cmd.MarkFlagsRequiredTogether("client-cert-path", "client-key-path")
}
