package apikey

import (
	"context"
	"fmt"
	"strings"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/keystore"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

type command struct {
	*pcmd.AuthenticatedStateFlagCommand
	keystore     keystore.KeyStore
	flagResolver pcmd.FlagResolver
}

const resourceFlagName = "resource"

const (
	deleteOperation = "deleting"
	getOperation    = "getting"
	updateOperation = "updating"
)

func New(prerunner pcmd.PreRunner, keystore keystore.KeyStore, resolver pcmd.FlagResolver) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "api-key",
		Short:       "Manage API keys.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &command{
		AuthenticatedStateFlagCommand: pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner),
		keystore:                      keystore,
		flagResolver:                  resolver,
	}

	c.AddCommand(c.newCreateCommand())
	c.AddCommand(c.newDeleteCommand())
	c.AddCommand(c.newDescribeCommand())
	c.AddCommand(c.newListCommand())
	c.AddCommand(c.newStoreCommand())
	c.AddCommand(c.newUpdateCommand())
	c.AddCommand(c.newUseCommand())

	return c.Command
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

	return pcmd.AutocompleteApiKeys(c.EnvironmentId(), c.V2Client)
}

func (c *command) getAllUsers() ([]*ccloudv1.User, error) {
	users, err := c.Client.User.GetServiceAccounts(context.Background())
	if err != nil {
		return nil, err
	}

	if auditLog := v1.GetAuditLog(c.Context.Context); auditLog != nil {
		serviceAccount, err := c.Client.User.GetServiceAccount(context.Background(), auditLog.GetServiceAccountId())
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

func (c *command) resolveResourceId(cmd *cobra.Command, v2Client *ccloudv2.Client) (string, string, string, error) {
	resourceId, err := cmd.Flags().GetString("resource")
	if err != nil {
		return "", "", "", err
	}
	if resourceId == "" {
		return "", "", "", nil
	}

	resourceType := resource.LookupType(resourceId)

	var clusterId string
	var apiKey string

	switch resourceType {
	case resource.Cloud:
		break
	case resource.KafkaCluster:
		cluster, err := c.Context.FindKafkaCluster(resourceId)
		if err != nil {
			return "", "", "", errors.CatchResourceNotFoundError(err, resourceId)
		}
		clusterId = cluster.ID
		apiKey = cluster.APIKey
	case resource.KsqlCluster:
		cluster, err := v2Client.DescribeKsqlCluster(resourceId, c.EnvironmentId())
		if err != nil {
			return "", "", "", errors.CatchResourceNotFoundError(err, resourceId)
		}
		clusterId = cluster.GetId()
	case resource.SchemaRegistryCluster:
		cluster, err := c.Context.SchemaRegistryCluster(cmd)
		if err != nil {
			return "", "", "", errors.CatchResourceNotFoundError(err, resourceId)
		}
		clusterId = cluster.Id
		if cluster.SrCredentials != nil {
			apiKey = cluster.SrCredentials.Key
		}
	default:
		return "", "", "", fmt.Errorf(`unsupported resource type for resource "%s"`, resourceId)
	}

	return resourceType, clusterId, apiKey, nil
}
