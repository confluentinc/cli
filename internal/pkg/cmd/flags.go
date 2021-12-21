package cmd

import (
	"context"
	"fmt"
	"strings"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/ccloud-sdk-go-v1"
	"github.com/spf13/cobra"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/kafka"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func AddApiKeyFlag(cmd *cobra.Command, command *AuthenticatedCLICommand) {
	cmd.Flags().String("api-key", "", "API key.")

	RegisterFlagCompletionFunc(cmd, "api-key", func(cmd *cobra.Command, args []string) []string {
		if err := command.PersistentPreRunE(cmd, args); err != nil {
			return nil
		}

		return AutocompleteApiKeys(command.EnvironmentId(), command.Client)
	})
}

func AddApiSecretFlag(cmd *cobra.Command) {
	cmd.Flags().String("api-secret", "", "API key secret.")
}

func AutocompleteApiKeys(environment string, client *ccloud.Client) []string {
	apiKeys, err := client.APIKey.List(context.Background(), &schedv1.ApiKey{AccountId: environment})
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(apiKeys))
	for i, apiKey := range apiKeys {
		if apiKey.UserId == 0 {
			continue
		}
		suggestions[i] = fmt.Sprintf("%s\t%s", apiKey.Key, apiKey.Description)
	}
	return suggestions
}

func AddCloudFlag(cmd *cobra.Command) {
	cmd.Flags().String("cloud", "", fmt.Sprintf("Cloud provider (%s).", strings.Join(kafka.Clouds, ", ")))
	RegisterFlagCompletionFunc(cmd, "cloud", func(_ *cobra.Command, _ []string) []string { return kafka.Clouds })
}

func AddClusterFlag(cmd *cobra.Command, command *AuthenticatedCLICommand) {
	cmd.Flags().String("cluster", "", "Kafka cluster ID.")
	RegisterFlagCompletionFunc(cmd, "cluster", func(cmd *cobra.Command, args []string) []string {
		if err := command.PersistentPreRunE(cmd, args); err != nil {
			return nil
		}

		return AutocompleteClusters(command.EnvironmentId(), command.Client)
	})
}

func AutocompleteClusters(environmentId string, client *ccloud.Client) []string {
	clusters, err := kafka.ListKafkaClusters(client, environmentId)
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(clusters))
	for i, cluster := range clusters {
		suggestions[i] = fmt.Sprintf("%s\t%s", cluster.Id, cluster.Name)
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

		return AutocompleteEnvironments(command.Client)
	})
}

func AutocompleteEnvironments(client *ccloud.Client) []string {
	environments, err := client.Account.List(context.Background(), &orgv1.Account{})
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(environments))
	for i, environment := range environments {
		suggestions[i] = fmt.Sprintf("%s\t%s", environment.Id, environment.Name)
	}
	return suggestions
}

func AddOutputFlag(cmd *cobra.Command) {
	AddOutputFlagWithDefaultValue(cmd, output.Human.String())
}

func AddOutputFlagWithDefaultValue(cmd *cobra.Command, defaultValue string) {
	cmd.Flags().StringP(output.FlagName, "o", defaultValue, `Specify the output format as "human", "json", or "yaml".`)

	RegisterFlagCompletionFunc(cmd, output.FlagName, func(_ *cobra.Command, _ []string) []string {
		return output.ValidFlagValues
	})
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

func autocompleteRegions(client *ccloud.Client, cloud string) []string {
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

func AddServiceAccountFlag(cmd *cobra.Command, command *AuthenticatedCLICommand) {
	cmd.Flags().String("service-account", "", "Service account ID.")

	RegisterFlagCompletionFunc(cmd, "service-account", func(cmd *cobra.Command, args []string) []string {
		if err := command.PersistentPreRunE(cmd, args); err != nil {
			return nil
		}

		return AutocompleteServiceAccounts(command.Client)
	})
}

func AutocompleteServiceAccounts(client *ccloud.Client) []string {
	serviceAccounts, err := client.User.GetServiceAccounts(context.Background())
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(serviceAccounts))
	for i, serviceAccount := range serviceAccounts {
		description := fmt.Sprintf("%s: %s", serviceAccount.ServiceName, serviceAccount.ServiceDescription)
		suggestions[i] = fmt.Sprintf("%s\t%s", serviceAccount.ResourceId, description)
	}
	return suggestions
}
