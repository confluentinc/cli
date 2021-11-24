package apikey

import (
	"context"
	"fmt"

	"github.com/c-bata/go-prompt"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/ccloud-sdk-go-v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/shell/completer"
)

func (c *command) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return autocompleteApiKeys(c.Client, c.EnvironmentId())
}

func autocompleteApiKeys(client *ccloud.Client, environment string) []string {
	apiKey := &schedv1.ApiKey{AccountId: environment}

	apiKeys, err := client.APIKey.List(context.Background(), apiKey)
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(apiKeys))
	for i, apiKey := range apiKeys {
		suggestions[i] = fmt.Sprintf("%s\t%s", apiKey.Key, apiKey.Description)
	}
	return suggestions
}

func (c *command) Cmd() *cobra.Command {
	return c.Command
}

func (c *command) ServerComplete() []prompt.Suggest {
	var suggests []prompt.Suggest
	apiKeys, err := c.fetchAPIKeys()
	if err != nil {
		return suggests
	}
	for _, key := range apiKeys {
		suggests = append(suggests, prompt.Suggest{
			Text:        key.Key,
			Description: key.Description,
		})
	}
	return suggests
}

func (c *command) ServerCompletableChildren() []*cobra.Command {
	return c.completableChildren
}

func (c *command) ServerCompletableFlagChildren() map[string][]*cobra.Command {
	return c.completableFlagChildren
}

func (c *command) ServerFlagComplete() map[string]func() []prompt.Suggest {
	return map[string]func() []prompt.Suggest{
		resourceFlagName:  c.resourceFlagCompleterFunc,
		"service-account": completer.ServiceAccountFlagCompleterFunc(c.Client),
	}
}

func (c *command) resourceFlagCompleterFunc() []prompt.Suggest {
	suggestions := completer.ClusterFlagServerCompleterFunc(c.Client, c.EnvironmentId())()

	ctx := context.Background()
	ctxClient := pcmd.NewContextClient(c.Context)
	cluster, err := ctxClient.FetchSchemaRegistryByAccountId(ctx, c.EnvironmentId())
	if err == nil {
		suggestions = append(suggestions, prompt.Suggest{
			Text:        cluster.Id,
			Description: cluster.Name,
		})
	}
	req := &schedv1.KSQLCluster{AccountId: c.EnvironmentId()}
	clusters, err := c.Client.KSQL.List(context.Background(), req)
	if err == nil {
		for _, cluster := range clusters {
			suggestions = append(suggestions, prompt.Suggest{
				Text:        cluster.Id,
				Description: cluster.Name,
			})
		}
	}
	return suggestions
}

func (c *command) fetchAPIKeys() ([]*schedv1.ApiKey, error) {
	apiKeys, err := c.Client.APIKey.List(context.Background(), &schedv1.ApiKey{AccountId: c.EnvironmentId(), LogicalClusters: nil, UserId: 0})
	if err != nil {
		return nil, errors.HandleCommon(err, c.Command)
	}

	var userApiKeys []*schedv1.ApiKey
	for _, key := range apiKeys {
		if key.UserId != 0 {
			userApiKeys = append(userApiKeys, key)
		}
	}
	return userApiKeys, nil
}
