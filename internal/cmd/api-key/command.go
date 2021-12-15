package apikey

import (
	"context"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/confluentinc/ccloud-sdk-go-v1"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/keystore"
	"github.com/confluentinc/cli/internal/pkg/shell/completer"
)

type command struct {
	*pcmd.AuthenticatedStateFlagCommand
	keystore                keystore.KeyStore
	flagResolver            pcmd.FlagResolver
	completableChildren     []*cobra.Command
	completableFlagChildren map[string][]*cobra.Command
	analyticsClient         analytics.Client
}

const resourceFlagName = "resource"

var subcommandFlags = map[string]*pflag.FlagSet{
	"create": pcmd.EnvironmentContextSet(),
	"store":  pcmd.EnvironmentContextSet(),
}

func New(prerunner pcmd.PreRunner, keystore keystore.KeyStore, resolver pcmd.FlagResolver, analyticsClient analytics.Client) *command {
	cmd := &cobra.Command{
		Use:         "api-key",
		Short:       "Manage the API keys.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &command{
		AuthenticatedStateFlagCommand: pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner, subcommandFlags),
		keystore:                      keystore,
		flagResolver:                  resolver,
		analyticsClient:               analyticsClient,
	}

	createCmd := c.newCreateCommand()
	deleteCmd := c.newDeleteCommand()
	listCmd := c.newListCommand()
	storeCmd := c.newStoreCommand()
	updateCmd := c.newUpdateCommand()
	useCmd := c.newUseCommand()

	c.AddCommand(createCmd)
	c.AddCommand(deleteCmd)
	c.AddCommand(listCmd)
	c.AddCommand(storeCmd)
	c.AddCommand(updateCmd)
	c.AddCommand(useCmd)

	c.completableChildren = append(c.completableChildren, updateCmd, deleteCmd, storeCmd, useCmd)
	c.completableFlagChildren = map[string][]*cobra.Command{
		resourceFlagName:  {createCmd, listCmd, storeCmd, useCmd},
		"service-account": {createCmd},
	}

	return c
}

func (c *command) setKeyStoreIfNil() {
	if c.keystore == nil {
		c.keystore = &keystore.ConfigKeyStore{Config: c.Config}
	}
}

func (c *command) parseFlagResolverPromptValue(source, prompt string, secure bool) (string, error) {
	val, err := c.flagResolver.ValueFrom(source, prompt, secure)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(val), nil
}

func (c *command) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return pcmd.AutocompleteApiKeys(c.EnvironmentId(), c.Client)
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

func (c *command) getAllUsers() ([]*orgv1.User, error) {
	users, err := c.Client.User.GetServiceAccounts(context.Background())
	if err != nil {
		return nil, err
	}

	if auditLog, ok := pcmd.AreAuditLogsEnabled(c.State); ok {
		serviceAccount, err := c.Client.User.GetServiceAccount(context.Background(), auditLog.ServiceAccountId)
		if err != nil {
			return nil, err
		}
		users = append(users, serviceAccount)
	}

	adminUsers, err := c.Client.User.List(context.Background())
	if err != nil {
		return nil, err
	}
	users = append(users, adminUsers...)

	return users, nil
}

func (c *command) resolveResourceId(cmd *cobra.Command, resolver pcmd.FlagResolver, client *ccloud.Client) (resourceType string, clusterId string, currentKey string, err error) {
	resourceType, resourceId, err := resolver.ResolveResourceId(cmd)
	if err != nil || resourceType == "" {
		return "", "", "", err
	}
	if resourceType == pcmd.SrResourceType {
		cluster, err := c.Context.SchemaRegistryCluster(cmd)
		if err != nil {
			return "", "", "", errors.CatchResourceNotFoundError(err, resourceId)
		}
		clusterId = cluster.Id
		if cluster.SrCredentials != nil {
			currentKey = cluster.SrCredentials.Key
		}
	} else if resourceType == pcmd.KSQLResourceType {
		ctx := context.Background()
		cluster, err := client.KSQL.Describe(
			ctx, &schedv1.KSQLCluster{
				Id:        resourceId,
				AccountId: c.EnvironmentId(),
			})
		if err != nil {
			return "", "", "", errors.CatchResourceNotFoundError(err, resourceId)
		}
		clusterId = cluster.Id
	} else if resourceType == pcmd.CloudResourceType {
		return resourceType, "", "", nil
	} else {
		// Resource is of KafkaResourceType.
		cluster, err := c.Context.FindKafkaCluster(resourceId)
		if err != nil {
			return "", "", "", errors.CatchResourceNotFoundError(err, resourceId)
		}
		clusterId = cluster.ID
		currentKey = cluster.APIKey
	}
	return resourceType, clusterId, currentKey, nil
}
