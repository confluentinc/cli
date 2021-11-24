package apikey

import (
	"context"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/ccloud-sdk-go-v1"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/keystore"
)

const longDescription = `Use this command to register an API secret created by another
process and store it locally.

When you create an API key with the CLI, it is automatically stored locally.
However, when you create an API key using the UI, API, or with the CLI on another
machine, the secret is not available for CLI use until you "store" it. This is because
secrets are irretrievable after creation.

You must have an API secret stored locally for certain CLI commands to
work. For example, the Kafka topic consume and produce commands require an API secret.`

const resourceFlagName = "resource"

var subcommandFlags = map[string]*pflag.FlagSet{
	"create": pcmd.EnvironmentContextSet(),
	"store":  pcmd.EnvironmentContextSet(),
}

type command struct {
	*pcmd.AuthenticatedStateFlagCommand
	keystore                keystore.KeyStore
	flagResolver            pcmd.FlagResolver
	completableChildren     []*cobra.Command
	completableFlagChildren map[string][]*cobra.Command
	analyticsClient         analytics.Client
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
		cluster, err := c.Context.FindKafkaCluster(cmd, resourceId)
		if err != nil {
			return "", "", "", errors.CatchResourceNotFoundError(err, resourceId)
		}
		clusterId = cluster.ID
		currentKey = cluster.APIKey
	}
	return resourceType, clusterId, currentKey, nil
}
