package apikey

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"

	"github.com/confluentinc/cli/v3/pkg/ccloudv2"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	dynamicconfig "github.com/confluentinc/cli/v3/pkg/dynamic-config"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/keystore"
	presource "github.com/confluentinc/cli/v3/pkg/resource"
)

type command struct {
	*pcmd.AuthenticatedCLICommand
	keystore     *keystore.ConfigKeyStore
	flagResolver pcmd.FlagResolver
}

const (
	deleteOperation = "deleting"
	getOperation    = "getting"
	updateOperation = "updating"
)

const (
	apiKeyNotValidForClusterSuggestions = "Specify the cluster this API key belongs to using the `--resource` flag. Alternatively, first execute the `confluent kafka cluster use` command to set the context to the proper cluster for this key and retry the `confluent api-key store` command."
	apiKeyUseFailedErrorMsg             = "unable to set active API key"
	apiKeyUseFailedSuggestions          = "If you did not create this API key with the CLI or created it on another computer, you must first store the API key and secret locally with `confluent api-key store %s <secret>`."
	nonKafkaNotImplementedErrorMsg      = "functionality not yet available for non-Kafka cluster resources"
	refuseToOverrideSecretSuggestions   = "If you would like to override the existing secret stored for API key \"%s\", use the `--force` flag."
	unableToStoreApiKeyErrorMsg         = "unable to store API key locally: %w"
)

func New(prerunner pcmd.PreRunner, resolver pcmd.FlagResolver) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "api-key",
		Short:       "Manage API keys.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &command{
		AuthenticatedCLICommand: pcmd.NewAuthenticatedCLICommand(cmd, prerunner),
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

func (c *command) addResourceFlag(cmd *cobra.Command, addCloud bool) {
	description := "The ID of the resource the API key is for."
	if addCloud {
		description += ` Use "cloud" for a Cloud API key.`
	}

	cmd.Flags().String("resource", "", description)

	pcmd.RegisterFlagCompletionFunc(cmd, "resource", func(cmd *cobra.Command, args []string) []string {
		if err := c.PersistentPreRunE(cmd, args); err != nil {
			return nil
		}

		environmentId, err := c.Context.EnvironmentId()
		if err != nil {
			return nil
		}

		kafkaClusters, err := c.V2Client.ListKafkaClusters(environmentId)
		if err != nil {
			return nil
		}

		schemaRegistryClusters, err := c.V2Client.GetSchemaRegistryClustersByEnvironment(environmentId)
		if err != nil {
			return nil
		}

		ksqlClusters, err := c.V2Client.ListKsqlClusters(environmentId)
		if err != nil {
			return nil
		}

		suggestions := make([]string, len(kafkaClusters)+len(schemaRegistryClusters)+len(ksqlClusters))
		i := 0

		for _, cluster := range kafkaClusters {
			suggestions[i] = fmt.Sprintf("%s\t%s", cluster.GetId(), cluster.Spec.GetDisplayName())
			i++
		}

		for _, cluster := range schemaRegistryClusters {
			suggestions[i] = fmt.Sprintf("%s\t%s", cluster.GetId(), cluster.Spec.GetDisplayName())
			i++
		}

		for _, cluster := range ksqlClusters {
			suggestions[i] = fmt.Sprintf("%s\t%s", cluster.GetId(), cluster.Spec.GetDisplayName())
			i++
		}

		if addCloud {
			suggestions = append(suggestions, "cloud")
		}

		return suggestions
	})
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

	return c.validArgsMultiple(cmd, args)
}

func (c *command) validArgsMultiple(cmd *cobra.Command, args []string) []string {
	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return pcmd.AutocompleteApiKeys(c.V2Client)
}

func (c *command) getAllUsers() ([]*ccloudv1.User, error) {
	users, err := c.Client.User.GetServiceAccounts()
	if err != nil {
		return nil, err
	}

	user, err := c.Client.Auth.User()
	if err != nil {
		return nil, err
	}

	if auditLog := user.GetOrganization().GetAuditLog(); auditLog.GetServiceAccountId() != 0 {
		serviceAccount, err := c.Client.User.GetServiceAccount(auditLog.GetServiceAccountId())
		if err != nil {
			// ignore 403s so we can still get other users
			if !strings.Contains(err.Error(), "Forbidden Access") {
				return nil, err
			}
		} else {
			users = append(users, serviceAccount)
		}
	}

	adminUsers, err := c.Client.User.List()
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
		cluster, err := dynamicconfig.FindKafkaCluster(c.V2Client, c.Context, resource)
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
		environmentId, err := c.Context.EnvironmentId()
		if err != nil {
			return "", "", "", err
		}
		cluster, err := v2Client.GetSchemaRegistryClusterById(resource, environmentId)
		if err != nil {
			return "", "", "", errors.CatchResourceNotFoundError(err, resource)
		}
		clusterId = cluster.GetId()
	default:
		return "", "", "", fmt.Errorf(`unsupported resource type for resource "%s"`, resource)
	}

	return resourceType, clusterId, apiKey, nil
}
