package cmd

import (
	"context"
	"fmt"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	dynamicconfig "github.com/confluentinc/cli/internal/pkg/dynamic-config"
	"github.com/confluentinc/cli/internal/pkg/kafka"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func AddApiKeyFlag(cmd *cobra.Command, command *AuthenticatedCLICommand) {
	cmd.Flags().String("api-key", "", "API key.")

	RegisterFlagCompletionFunc(cmd, "api-key", func(cmd *cobra.Command, args []string) []string {
		if err := command.PersistentPreRunE(cmd, args); err != nil {
			return nil
		}

		return AutocompleteApiKeys(command.EnvironmentId(cmd), command.V2Client)
	})
}

func AddApiSecretFlag(cmd *cobra.Command) {
	cmd.Flags().String("api-secret", "", "API key secret.")
}

func AutocompleteApiKeys(environment string, client *ccloudv2.Client) []string {
	apiKeys, err := client.ListApiKeys("", "")
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(apiKeys))
	for i, apiKey := range apiKeys {
		if !apiKey.Spec.HasOwner() {
			continue
		}
		suggestions[i] = fmt.Sprintf("%s\t%s", *apiKey.Id, *apiKey.GetSpec().Description)
	}
	return suggestions
}

func AddAvailabilityFlag(cmd *cobra.Command) {
	cmd.Flags().String("availability", kafka.Availabilities[0], fmt.Sprintf("Specify the availability of the cluster as %s.", utils.ArrayToCommaDelimitedString(kafka.Availabilities)))
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
	cmd.Flags().String("provider", "", fmt.Sprintf("Specify the provider as %s.", utils.ArrayToCommaDelimitedString([]string{"aws", "azure"})))

	RegisterFlagCompletionFunc(cmd, "provider", func(_ *cobra.Command, _ []string) []string {
		return []string{"aws", "azure"}
	})
}

func AddByokStateFlag(cmd *cobra.Command) {
	cmd.Flags().String("state", "", fmt.Sprintf("Specify the state as %s.", utils.ArrayToCommaDelimitedString([]string{"in-use", "available"})))

	RegisterFlagCompletionFunc(cmd, "state", func(_ *cobra.Command, _ []string) []string {
		return []string{"in-use", "available"}
	})
}

func AddCloudFlag(cmd *cobra.Command) {
	cmd.Flags().String("cloud", "", fmt.Sprintf("Specify the cloud provider as %s.", utils.ArrayToCommaDelimitedString(kafka.Clouds)))
	RegisterFlagCompletionFunc(cmd, "cloud", func(_ *cobra.Command, _ []string) []string { return kafka.Clouds })
}

func AddClusterFlag(cmd *cobra.Command, command *AuthenticatedCLICommand) {
	cmd.Flags().String("cluster", "", "Kafka cluster ID.")
	RegisterFlagCompletionFunc(cmd, "cluster", func(cmd *cobra.Command, args []string) []string {
		if err := command.PersistentPreRunE(cmd, args); err != nil {
			return nil
		}

		return AutocompleteClusters(command.EnvironmentId(cmd), command.V2Client)
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

func AutocompleteCmkClusters(environmentId string, client *ccloudv2.Client) []string {
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
	suggestions := make([]string, len(cfg.Contexts))
	i := 0
	for ctx := range cfg.Contexts {
		suggestions[i] = ctx
		i++
	}
	return suggestions
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

	if auditLog := v1.GetAuditLog(ctx.Context); auditLog != nil {
		auditLogAccount, err := v1Client.Account.Get(context.Background(), &ccloudv1.Account{Id: auditLog.GetAccountId()})
		if err != nil {
			return nil
		}
		suggestions = append(suggestions, fmt.Sprintf("%s\t%s", auditLog.GetAccountId(), auditLogAccount.GetName()))
	}

	return suggestions
}

func AddForceFlag(cmd *cobra.Command) {
	cmd.Flags().Bool("force", false, "Skip the deletion confirmation prompt.")
}

func AddKsqlClusterFlag(cmd *cobra.Command, command *AuthenticatedCLICommand) {
	cmd.Flags().String("ksql-cluster", "", "KSQL cluster for the pipeline.")
	RegisterFlagCompletionFunc(cmd, "ksql-cluster", func(cmd *cobra.Command, args []string) []string {
		if err := command.PersistentPreRunE(cmd, args); err != nil {
			return nil
		}

		return autocompleteKSQLClusters(command.EnvironmentId(cmd), command.V2Client)
	})
}

func autocompleteKSQLClusters(environmentId string, client *ccloudv2.Client) []string {
	clusters, err := client.ListKsqlClusters(environmentId)
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(clusters.Data))
	for i, cluster := range clusters.Data {
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

func AddOutputFlag(cmd *cobra.Command) {
	AddOutputFlagWithDefaultValue(cmd, output.Human.String())
}

func AddOutputFlagWithDefaultValue(cmd *cobra.Command, defaultValue string) {
	cmd.Flags().StringP(output.FlagName, "o", defaultValue, fmt.Sprintf("Specify the output format as %s.", utils.ArrayToCommaDelimitedString(output.ValidFlagValues)))

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
		description := fmt.Sprintf("%s: %s", *identityProvider.DisplayName, *identityProvider.Description)
		suggestions[i] = fmt.Sprintf("%s\t%s", *identityProvider.Id, description)
	}
	return suggestions
}

func AutocompleteIdentityPools(client *ccloudv2.Client, providerID string) []string {
	identityPools, err := client.ListIdentityPools(providerID)
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(identityPools))
	for i, identityPool := range identityPools {
		description := fmt.Sprintf("%s: %s", *identityPool.DisplayName, *identityPool.Description)
		suggestions[i] = fmt.Sprintf("%s\t%s", *identityPool.Id, description)
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
	str := utils.ArrayToCommaDelimitedString(arr)

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
		description := fmt.Sprintf("%s: %s", *serviceAccount.DisplayName, *serviceAccount.Description)
		suggestions[i] = fmt.Sprintf("%s\t%s", *serviceAccount.Id, description)
	}
	return suggestions
}

func AddTypeFlag(cmd *cobra.Command) {
	cmd.Flags().String("type", kafka.Types[0], fmt.Sprintf("Specify the type of the Kafka cluster as %s.", utils.ArrayToCommaDelimitedString(kafka.Types)))
	RegisterFlagCompletionFunc(cmd, output.FlagName, func(_ *cobra.Command, _ []string) []string { return kafka.Types })
}

func AddValueFormatFlag(cmd *cobra.Command) {
	arr := []string{"string", "avro", "jsonschema", "protobuf"}
	str := utils.ArrayToCommaDelimitedString(arr)

	cmd.Flags().String("value-format", "string", fmt.Sprintf("Format of message value as %s. Note that schema references are not supported for avro.", str))

	RegisterFlagCompletionFunc(cmd, "value-format", func(_ *cobra.Command, _ []string) []string {
		return arr
	})
}
