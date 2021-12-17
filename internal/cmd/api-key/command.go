package apikey

import (
	"context"
	"strings"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/ccloud-sdk-go-v1"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/keystore"
)

type command struct {
	*pcmd.AuthenticatedStateFlagCommand
	keystore        keystore.KeyStore
	flagResolver    pcmd.FlagResolver
	analyticsClient analytics.Client
}

const resourceFlagName = "resource"

func New(prerunner pcmd.PreRunner, keystore keystore.KeyStore, resolver pcmd.FlagResolver, analyticsClient analytics.Client) *command {
	cmd := &cobra.Command{
		Use:         "api-key",
		Short:       "Manage the API keys.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &command{
		AuthenticatedStateFlagCommand: pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner),
		keystore:                      keystore,
		flagResolver:                  resolver,
		analyticsClient:               analyticsClient,
	}

	c.AddCommand(c.newCreateCommand())
	c.AddCommand(c.newDeleteCommand())
	c.AddCommand(c.newListCommand())
	c.AddCommand(c.newStoreCommand())
	c.AddCommand(c.newUpdateCommand())
	c.AddCommand(c.newUseCommand())

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
