package apikey

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"

	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/keystore"
	presource "github.com/confluentinc/cli/internal/pkg/resource"
)

type command struct {
	*pcmd.AuthenticatedCLICommand
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
		AuthenticatedCLICommand: pcmd.NewAuthenticatedCLICommand(cmd, prerunner),
		keystore:                keystore,
		flagResolver:            resolver,
	}

	cmd.AddCommand(c.newCreateCommand())
	cmd.AddCommand(c.newDeleteCommand())
	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(c.newListCommand())
	cmd.AddCommand(c.newStoreCommand())
	cmd.AddCommand(c.newUpdateCommand())
	cmd.AddCommand(c.newUseCommand())

	return cmd
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

	return pcmd.AutocompleteApiKeys(c.V2Client)
}

func (c *command) getAllUsers() ([]*ccloudv1.User, error) {
	users, err := c.Client.User.GetServiceAccounts(context.Background())
	if err != nil {
		return nil, err
	}

	user, err := c.Client.Auth.User(context.Background())
	if err != nil {
		return nil, err
	}

	if auditLog := user.GetOrganization().GetAuditLog(); auditLog != nil {
		serviceAccount, err := c.Client.User.GetServiceAccount(context.Background(), auditLog.GetServiceAccountId())
		if err != nil {
			// ignore 403s so we can still get other users
			if !strings.Contains(err.Error(), "Forbidden Access") {
				return nil, err
			}
		} else {
			users = append(users, serviceAccount)
		}
	}

	adminUsers, err := c.Client.User.List(context.Background())
	if err != nil {
		return nil, err
	}
	users = append(users, adminUsers...)

	if currentUser := c.Context.GetUser(); currentUser != nil {
		users = append(users, currentUser)
	}

	return users, nil
}

func (c *command) resolveResourceId(cmd *cobra.Command, v2Client *ccloudv2.Client) (string, string, string, error) {
	resource, err := cmd.Flags().GetString("resource")
	if err != nil {
		return "", "", "", err
	}
	if resource == "" {
		return "", "", "", nil
	}

	resourceType := presource.LookupType(resource)

	var clusterId string
	var apiKey string

	switch resourceType {
	case presource.Cloud:
		break
	case presource.KafkaCluster:
		cluster, err := c.Context.FindKafkaCluster(resource)
		if err != nil {
			return "", "", "", errors.CatchResourceNotFoundError(err, resource)
		}
		clusterId = cluster.ID
		apiKey = cluster.APIKey
	case presource.KsqlCluster:
		environmentId, err := c.Context.EnvironmentId()
		if err != nil {
			return "", "", "", err
		}
		cluster, err := v2Client.DescribeKsqlCluster(resource, environmentId)
		if err != nil {
			return "", "", "", errors.CatchResourceNotFoundError(err, resource)
		}
		clusterId = cluster.GetId()
	case presource.SchemaRegistryCluster:
		cluster, err := c.Context.SchemaRegistryCluster(cmd)
		if err != nil {
			return "", "", "", errors.CatchResourceNotFoundError(err, resource)
		}
		clusterId = cluster.Id
		if cluster.SrCredentials != nil {
			apiKey = cluster.SrCredentials.Key
		}
	default:
		return "", "", "", fmt.Errorf(`unsupported resource type for resource "%s"`, resource)
	}

	return resourceType, clusterId, apiKey, nil
}
